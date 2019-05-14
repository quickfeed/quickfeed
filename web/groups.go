package web

import (
	"context"
	"log"

	"github.com/autograde/aguis/scm"

	"google.golang.org/grpc/codes"

	pb "github.com/autograde/aguis/ag"
	"github.com/autograde/aguis/database"
	"github.com/jinzhu/gorm"
	"google.golang.org/grpc/status"
)

// group request validation
func validGroup(grp *pb.Group) bool {
	return grp != nil &&
		grp.Name != "" &&
		len(grp.Users) > 0
}

// CreateGroup creates a new group for the given course id
func CreateGroup(request *pb.Group, db database.Database, currentUser *pb.User) (*pb.Group, error) {
	if _, err := db.GetCourse(request.CourseId); err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Errorf(codes.NotFound, "Course not found")
		}
		return nil, err
	}

	// validating received group request
	if !validGroup(request) {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid payload")
	}

	var userIds []uint64
	for _, user := range request.Users {
		userIds = append(userIds, user.Id)
	}
	users, err := db.GetUsers(userIds...)
	if err != nil {
		return nil, err
	}

	if len(users) != len(request.Users) {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid payload")
	}

	signedInUserEnrollment, err := db.GetEnrollmentByCourseAndUser(request.CourseId, currentUser.Id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "Not able to retreive enrollment for signed in user")
	}
	signedInUserInGroup := false

	for _, user := range users {
		enrollment, err := db.GetEnrollmentByCourseAndUser(request.CourseId, user.Id)
		switch {
		case err == gorm.ErrRecordNotFound:
			return nil, status.Errorf(codes.NotFound, "User is not enrolled in this course")
		case err != nil:
			return nil, err
		case enrollment.GroupId > 0:
			return nil, status.Errorf(codes.InvalidArgument, "User is already in another group")
		case enrollment.Status < pb.Enrollment_STUDENT:
			return nil, status.Errorf(codes.InvalidArgument, "User is not yet accepted to this course")
		case enrollment.Status == pb.Enrollment_TEACHER && signedInUserEnrollment.Status != pb.Enrollment_TEACHER:
			return nil, status.Errorf(codes.InvalidArgument, "A teacher has to create this group")
		case currentUser.Id == user.Id && enrollment.Status == pb.Enrollment_STUDENT:
			signedInUserInGroup = true
		}
	}

	// If user is a teacher it should be allowed to proceed and create a group with only the "enrolled" persons.
	if signedInUserEnrollment.Status == pb.Enrollment_TEACHER {
		signedInUserInGroup = true
	}

	if !signedInUserInGroup {
		return nil, status.Errorf(codes.FailedPrecondition, "student must be member of new group")
	}

	group := pb.Group{
		Name:     request.Name,
		CourseId: request.CourseId,
		Users:    users,
	}

	// CreateGroup creates a new group and update group_id in enrollment table
	if err := db.CreateGroup(&group); err != nil {
		if err == database.ErrDuplicateGroup {
			return nil, status.Errorf(codes.InvalidArgument, err.Error())
		}
		return nil, err
	}

	// database method returns error, front end method wants a new group to be returned. Get it from the database as an extra check for success
	newGroup, err := db.GetGroup(request.Id)
	if err != nil {
		return nil, err
	}
	return newGroup, nil
}

// UpdateGroup updates status of a group
func UpdateGroup(ctx context.Context, request *pb.Group, db database.Database, s scm.SCM, currentUser *pb.User) (*pb.StatusCode, error) {

	if !currentUser.IsAdmin {
		return &pb.StatusCode{StatusCode: int32(codes.PermissionDenied)}, status.Errorf(codes.PermissionDenied, "Only teacher can update status of a group")
	}
	course, err := db.GetCourse(request.CourseId)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return &pb.StatusCode{StatusCode: int32(codes.NotFound)}, status.Errorf(codes.NotFound, "course not found")
		}
		return &pb.StatusCode{StatusCode: int32(codes.InvalidArgument)}, err
	}

	oldgrp, err := db.GetGroup(request.Id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return &pb.StatusCode{StatusCode: int32(codes.NotFound)}, status.Errorf(codes.NotFound, "group not found")
		}
		return nil, err
	}
	enrollment, err := db.GetEnrollmentByCourseAndUser(request.CourseId, currentUser.Id)
	if err != nil {
		return &pb.StatusCode{StatusCode: int32(codes.InvalidArgument)}, err
	}
	if enrollment.Status != pb.Enrollment_TEACHER {
		return &pb.StatusCode{StatusCode: int32(codes.PermissionDenied)}, status.Errorf(codes.PermissionDenied, "only teacher can update a group")
	}

	if !validGroup(request) {
		return &pb.StatusCode{StatusCode: int32(codes.InvalidArgument)}, status.Errorf(codes.InvalidArgument, "Invalid payload")
	}

	var userIds []uint64
	for _, user := range request.Users {
		userIds = append(userIds, user.Id)
	}
	users, err := db.GetUsers(userIds...)
	if err != nil {
		return &pb.StatusCode{StatusCode: int32(codes.Aborted)}, err
	}

	group, err := db.GetGroup(request.Id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return &pb.StatusCode{StatusCode: int32(codes.NotFound)}, status.Errorf(codes.NotFound, "Group not found")
		}
		return &pb.StatusCode{StatusCode: int32(codes.Aborted)}, err
	}

	if len(group.Users) != len(userIds) {
		return &pb.StatusCode{StatusCode: int32(codes.InvalidArgument)}, status.Errorf(codes.InvalidArgument, "Invalid payload")
	}

	for _, user := range group.Users {
		enrollment, err := db.GetEnrollmentByCourseAndUser(request.CourseId, user.Id)
		switch {
		case err == gorm.ErrRecordNotFound:
			return &pb.StatusCode{StatusCode: int32(codes.NotFound)}, status.Errorf(codes.NotFound, "User is not enrolled in this course")
		case err != nil:
			return &pb.StatusCode{StatusCode: int32(codes.Aborted)}, err
		case enrollment.GroupId > 0 && enrollment.GroupId != request.Id:
			return &pb.StatusCode{StatusCode: int32(codes.InvalidArgument)}, status.Errorf(codes.InvalidArgument, "User is already in another group")
		case enrollment.Status < pb.Enrollment_STUDENT:
			return &pb.StatusCode{StatusCode: int32(codes.InvalidArgument)}, status.Errorf(codes.InvalidArgument, "User is not yet accepted to this course")
		}
	}

	if err := db.UpdateGroup(&pb.Group{
		Id:       oldgrp.Id,
		Name:     request.Name,
		CourseId: request.CourseId,
		Users:    users,
	}); err != nil {
		if err == database.ErrDuplicateGroup {
			return &pb.StatusCode{StatusCode: int32(codes.InvalidArgument)}, status.Errorf(codes.InvalidArgument, err.Error())
		}
		return &pb.StatusCode{StatusCode: int32(codes.Aborted)}, err
	}

	var userRemoteIdentity []*pb.RemoteIdentity
	// TODO move this into the for loop above, modify db.GetUsers() to also retreive RemoteIdentity
	// so we can remove individual GetUser calls
	for _, user := range users {
		remoteIdentityUser, _ := db.GetUser(user.Id)
		if err != nil {
			return &pb.StatusCode{StatusCode: int32(codes.InvalidArgument)}, err
		}
		if len(remoteIdentityUser.RemoteIdentities) > 0 {
			userRemoteIdentity = append(userRemoteIdentity, remoteIdentityUser.RemoteIdentities[0])
		}
	}

	contextWithTimeout, cancel := context.WithTimeout(ctx, MaxWait)
	defer cancel()
	// TODO move this functionality down into SCM?
	// Note: This Requires alot of calls to git.
	// Figure out all group members git-username
	var gitUserNames []string
	for _, identity := range userRemoteIdentity {
		gitName, err := s.GetUserNameByID(contextWithTimeout, identity.RemoteId)
		if err != nil {
			return &pb.StatusCode{StatusCode: int32(codes.Aborted)}, err
		}
		gitUserNames = append(gitUserNames, gitName)
	}

	// Create and add repo to autograder group
	dir, err := s.GetDirectory(contextWithTimeout, course.DirectoryId)
	if err != nil {
		return &pb.StatusCode{StatusCode: int32(codes.Aborted)}, err
	}
	repo, err := s.CreateRepository(contextWithTimeout, &scm.CreateRepositoryOptions{
		Directory: dir,
		Path:      request.Name,
		Private:   true,
	})
	if err != nil {
		return &pb.StatusCode{StatusCode: int32(codes.Aborted)}, err
	}

	// Add repo to DB
	dbRepo := pb.Repository{
		DirectoryId:  course.DirectoryId,
		RepositoryId: repo.ID,
		RepoType:     pb.Repository_USER,
		UserId:       userIds[0], // Should this be groupID ????
	}
	if err := db.CreateRepository(&dbRepo); err != nil {
		return &pb.StatusCode{StatusCode: int32(codes.Aborted)}, err
	}

	// Create git-team
	team, err := s.CreateTeam(contextWithTimeout, &scm.CreateTeamOptions{
		Directory: &scm.Directory{Path: dir.Path},
		TeamName:  request.Name,
		Users:     gitUserNames,
	})
	if err != nil {
		return &pb.StatusCode{StatusCode: int32(codes.Aborted)}, err
	}
	// Adding Repo to git-team
	if err = s.AddTeamRepo(contextWithTimeout, &scm.AddTeamRepoOptions{
		TeamID: team.ID,
		Owner:  repo.Owner,
		Repo:   repo.Path,
	}); err != nil {
		return &pb.StatusCode{StatusCode: int32(codes.Aborted)}, err
	}

	return &pb.StatusCode{StatusCode: int32(codes.OK)}, nil
}

// TODO: Finish function, will also require context for provider info
func createGroupRepoAndTeam(course *pb.Course, group *pb.Group) (*scm.Repository, error) {
	return nil, status.Errorf(codes.Unimplemented, "Function not yet implemented")
}

// PatchGroup updates status of a group
func PatchGroup(ctx context.Context, request *pb.Group, db database.Database, currentUser *pb.User, s scm.SCM) (*pb.StatusCode, error) {
	oldGrp, err := db.GetGroup(request.Id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return &pb.StatusCode{StatusCode: int32(codes.NotFound)}, status.Errorf(codes.NotFound, "group not found")
		}
		return &pb.StatusCode{StatusCode: int32(codes.Aborted)}, err
	}

	if request.Status > pb.Group_DELETED {
		return &pb.StatusCode{StatusCode: int32(codes.InvalidArgument)}, status.Errorf(codes.InvalidArgument, "Invalid payload")
	}
	if !currentUser.IsAdmin {
		return &pb.StatusCode{StatusCode: int32(codes.PermissionDenied)}, status.Errorf(codes.PermissionDenied, "Only teacher can update status of a group")
	}

	users := oldGrp.Users

	course, err := db.GetCourse(oldGrp.CourseId)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return &pb.StatusCode{StatusCode: int32(codes.NotFound)}, status.Errorf(codes.NotFound, "course not found")
		}
		return &pb.StatusCode{StatusCode: int32(codes.InvalidArgument)}, err
	}

	var userRemoteIdentity []*pb.RemoteIdentity
	for _, user := range users {
		remoteIdentityUser, _ := db.GetUser(user.Id)
		if err != nil {
			return &pb.StatusCode{StatusCode: int32(codes.InvalidArgument)}, err
		}
		// TODO, figure out which remote identity to be used!
		if len(remoteIdentityUser.RemoteIdentities) > 0 {
			userRemoteIdentity = append(userRemoteIdentity, remoteIdentityUser.RemoteIdentities[0])
		}
	}

	contextWithTimeout, cancel := context.WithTimeout(ctx, MaxWait)
	defer cancel()

	// TODO move this functionality down into SCM?
	// Note: This Requires alot of calls to git.
	// Figure out all group members git-username
	var gitUserNames []string
	for _, identity := range userRemoteIdentity {
		gitName, err := s.GetUserNameByID(contextWithTimeout, identity.RemoteId)
		if err != nil {
			return &pb.StatusCode{StatusCode: int32(codes.Aborted)}, err
		}
		gitUserNames = append(gitUserNames, gitName)
	}
	// Create and add repo to autograder group
	dir, err := s.GetDirectory(ctx, course.DirectoryId)
	if err != nil {
		return &pb.StatusCode{StatusCode: int32(codes.Aborted)}, err
	}

	repos, err := s.GetRepositories(contextWithTimeout, dir)
	if err != nil {
		return &pb.StatusCode{StatusCode: int32(codes.Aborted)}, err
	}

	existing := make(map[string]*scm.Repository)
	for _, repo := range repos {
		existing[repo.Path] = repo
	}
	repo, created := existing[oldGrp.Name]
	if !created {
		repo, err = s.CreateRepository(contextWithTimeout, &scm.CreateRepositoryOptions{
			Directory: dir,
			Path:      oldGrp.Name,
			Private:   true,
		})
		if err != nil {
			log.Println("Failed to create repository")
			//TODO(meling) this does not seem to hold group repos for unknown reasons
			repo = existing[oldGrp.Name]
			return &pb.StatusCode{StatusCode: int32(codes.Aborted)}, err
		}
		log.Println("Created new group repository")
		// Add repo to DB
		dbRepo := pb.Repository{
			DirectoryId:  course.DirectoryId,
			RepositoryId: repo.ID,
			HtmlUrl:      repo.WebURL,
			RepoType:     pb.Repository_USER,
			UserId:       0,
			GroupId:      oldGrp.Id,
		}
		if err := db.CreateRepository(&dbRepo); err != nil {
			log.Println("Failed to create repository in database")
			return &pb.StatusCode{StatusCode: int32(codes.Aborted)}, err
		}
		log.Println("Created new group repository in database")

		if err := db.UpdateGroupStatus(&pb.Group{
			Id:     oldGrp.Id,
			Status: request.Status,
		}); err != nil {
			log.Println("Failed to update group status in database")
			return &pb.StatusCode{StatusCode: int32(codes.Aborted)}, err
		}

		// Create git-team
		team, err := s.CreateTeam(contextWithTimeout, &scm.CreateTeamOptions{
			Directory: &scm.Directory{Path: dir.Path},
			TeamName:  oldGrp.Name,
			Users:     gitUserNames,
		})
		if err != nil {
			log.Println("Failed to create git-team for ", oldGrp.Name, " with ", gitUserNames)
			return &pb.StatusCode{StatusCode: int32(codes.Aborted)}, err
		}

		// Adding Repo to git-team
		if err = s.AddTeamRepo(contextWithTimeout, &scm.AddTeamRepoOptions{
			TeamID: team.ID,
			Owner:  repo.Owner,
			Repo:   repo.Path,
		}); err != nil {
			log.Println("Failed to add repo to git-team for ", repo.Owner)
			return &pb.StatusCode{StatusCode: int32(codes.Aborted)}, err
		}
	}
	return &pb.StatusCode{StatusCode: int32(codes.OK)}, nil

}

// GetGroup returns a group
func GetGroup(request *pb.RecordRequest, db database.Database) (*pb.Group, error) {
	group, err := db.GetGroup(request.Id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Errorf(codes.NotFound, "Group not found")
		}
		return nil, err
	}
	return group, nil
}

// GetGroups returns all groups in a given course
func GetGroups(request *pb.RecordRequest, db database.Database) (*pb.Groups, error) {
	groups, err := db.GetGroupsByCourse(request.Id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Errorf(codes.NotFound, "course not found")
		}
		return nil, err
	}
	return &pb.Groups{Groups: groups}, nil
}

// DeleteGroup deletes a pending or rejected group
func DeleteGroup(request *pb.Group, db database.Database) (*pb.StatusCode, error) {
	group, err := db.GetGroup(request.Id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return &pb.StatusCode{StatusCode: int32(codes.NotFound)}, status.Errorf(codes.NotFound, "Group not found")
		}
		return &pb.StatusCode{StatusCode: int32(codes.Aborted)}, err
	}
	if group.Status > pb.Group_REJECTED_GROUP {
		return &pb.StatusCode{StatusCode: int32(codes.Aborted)}, status.Errorf(codes.Aborted, "Accepted group cannot be deleted")
	}
	if err := db.DeleteGroup(request.Id); err != nil {
		return &pb.StatusCode{StatusCode: int32(codes.Aborted)}, err
	}
	return &pb.StatusCode{StatusCode: int32(codes.OK)}, nil
}

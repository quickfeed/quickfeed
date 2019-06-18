package web

import (
	"context"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/autograde/aguis/ag"
	"github.com/autograde/aguis/database"
	"github.com/autograde/aguis/scm"
	"github.com/jinzhu/gorm"
)

// ListCourses returns a JSON object containing all the courses in the database.
func ListCourses(db database.Database) (*pb.Courses, error) {
	courses, err := db.GetCourses()
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "course not found")
	}
	return &pb.Courses{Courses: courses}, nil
}

// ListCoursesWithEnrollment lists all existing courses with the provided users
// enrollment status.
// If status query param is provided, lists only courses of the student filtered by the query param.
func ListCoursesWithEnrollment(request *pb.RecordRequest, db database.Database) (*pb.Courses, error) {
	courses, err := db.GetCoursesByUser(request.ID, request.Statuses...)
	if err != nil {
		err = status.Errorf(codes.NotFound, "no courses found")
	}
	return &pb.Courses{Courses: courses}, err
}

// ListAssignments lists the assignments for the provided course.
func ListAssignments(request *pb.RecordRequest, db database.Database) (*pb.Assignments, error) {
	assignments, err := db.GetAssignmentsByCourse(request.ID)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "no assignments found")
	}
	return &pb.Assignments{Assignments: assignments}, nil
}

// CreateEnrollment enrolls a user in a course.
func CreateEnrollment(request *pb.ActionRequest, db database.Database) error {

	enrollment := pb.Enrollment{
		UserID:   request.UserID,
		CourseID: request.CourseID,
		Status:   pb.Enrollment_PENDING,
	}

	if err := db.CreateEnrollment(&enrollment); err != nil {
		if err == gorm.ErrRecordNotFound {
			return status.Errorf(codes.NotFound, "record not found")

		}
		return status.Errorf(codes.Internal, "could not create enrollment")
	}
	return nil
}

// UpdateEnrollment accepts or rejects a user to enroll in a course.
func UpdateEnrollment(ctx context.Context, request *pb.ActionRequest, db database.Database, s scm.SCM) error {
	if _, err := db.GetEnrollmentByCourseAndUser(request.CourseID, request.UserID); err != nil {
		if err == gorm.ErrRecordNotFound {
			return status.Errorf(codes.NotFound, "not found")
		}
		return status.Errorf(codes.Internal, "could not update enrollment")
	}

	// TODO If the enrollment is accepted, create repositories and permissions for them with webooks.
	var err error
	switch request.Status {
	case pb.Enrollment_STUDENT:
		// Update enrollment for student in DB.
		err = db.EnrollStudent(request.UserID, request.CourseID)
		if err != nil {
			return status.Errorf(codes.Internal, "could not enroll student")
		}
		course, err := db.GetCourse(request.CourseID)
		if err != nil {
			return status.Errorf(codes.NotFound, "course not found")
		}
		student, err := db.GetUser(request.UserID)
		if err != nil {
			return status.Errorf(codes.NotFound, "user not found")
		}

		// check whether user repo already exists (happens when accepting previously rejected student)
		userRepoQuery := &pb.Repository{
			OrganizationID: course.GetOrganizationID(),
			UserID:         request.GetUserID(),
			RepoType:       pb.Repository_USER,
		}
		repos, err := db.GetRepositories(userRepoQuery)
		if err != nil {
			return status.Errorf(codes.NotFound, "repository not found")
		}
		if len(repos) == 0 {
			/*
				// create user repo and team on SCM.
				repo, _, err := createUserRepoAndTeam(ctx, s, course, student)
				if err != nil {
					return status.Errorf(codes.Internal, "could not create user repository and team")
				}*/
			// create only repo to avoid creating personal team
			repo, err := createUserRepo(ctx, s, course.OrganizationID, student)
			if err != nil {
				return status.Errorf(codes.Internal, "could not create user repository")
			}
			// add student repo to database if SCM interaction was successful
			dbRepo := pb.Repository{
				OrganizationID: course.GetOrganizationID(),
				UserID:         request.GetUserID(),
				RepoType:       pb.Repository_USER,
				RepositoryID:   repo.ID,
				HTMLURL:        repo.WebURL,
			}
			if err := db.CreateRepository(&dbRepo); err != nil {
				return status.Errorf(codes.Internal, "could not create user repository")
			}
			// instead of creating personal team we will add all students to students team
			return addToUserTeam(ctx, s, course.GetOrganizationID(), student, pb.Enrollment_STUDENT)

		}

	case pb.Enrollment_TEACHER:
		if err = db.EnrollTeacher(request.UserID, request.CourseID); err != nil {
			return status.Errorf(codes.Internal, "could not enroll teacher")
		}
		course, err := db.GetCourse(request.CourseID)
		if err != nil {
			return status.Errorf(codes.NotFound, "course not found")
		}
		student, err := db.GetUser(request.UserID)
		if err != nil {
			return status.Errorf(codes.NotFound, "user not found")
		}
		orgUpdate := &scm.OrgMembership{Username: student.GetLogin(), OrgID: course.GetOrganizationID(), Role: "admin"}
		if err = s.UpdateOrgMembership(ctx, orgUpdate); err != nil {
			return status.Errorf(codes.Internal, "could not update user membership")
		}
		// add all teachers to teachers team
		return addToUserTeam(ctx, s, course.GetOrganizationID(), student, pb.Enrollment_TEACHER)

	case pb.Enrollment_REJECTED:
		if err = db.RejectEnrollment(request.UserID, request.CourseID); err != nil {
			err = status.Errorf(codes.Internal, "could not reject user")
		}
	case pb.Enrollment_PENDING:
		if err = db.SetPendingEnrollment(request.UserID, request.CourseID); err != nil {
			err = status.Errorf(codes.Internal, "could not set pending")
		}
	}
	// it will be nil or error, as expected by calling function
	return err
}

func createUserRepoAndTeam(c context.Context, s scm.SCM, course *pb.Course, student *pb.User) (*scm.Repository, *scm.Team, error) {

	org, err := s.GetOrganization(c, course.OrganizationID)
	if err != nil {
		return nil, nil, err
	}

	// the student's git user name is the same as the team name
	teamName := student.Login

	opt := &scm.CreateRepositoryOptions{
		Organization: org,
		Path:         pb.StudentRepoName(teamName),
		Private:      true,
	}

	return s.CreateRepoAndTeam(c, opt, teamName, []string{teamName})
}

func createUserRepo(c context.Context, s scm.SCM, orgID uint64, student *pb.User) (*scm.Repository, error) {
	org, err := s.GetOrganization(c, orgID)
	if err != nil {
		return nil, err
	}
	return s.CreateRepository(c, &scm.CreateRepositoryOptions{Path: pb.StudentRepoName(student.GetLogin()), Organization: org, Private: true})
}

func addToUserTeam(c context.Context, s scm.SCM, orgID uint64, user *pb.User, status pb.Enrollment_UserStatus) error {

	// get course organization
	org, err := s.GetOrganization(c, orgID)
	if err != nil {
		return err
	}

	var slug string
	// check whether user is teacher or not
	switch status {
	case pb.Enrollment_STUDENT:
		slug = "students"
	case pb.Enrollment_TEACHER:
		slug = "teachers"
	}

	opt := &scm.AddMemberOptions{
		Organization: org,
		TeamSlug:     slug,
		Username:     user.GetLogin(),
	}

	return s.AddTeamMember(c, opt)
}

// GetCourse find course by id and return JSON object.
func GetCourse(query *pb.RecordRequest, db database.Database) (*pb.Course, error) {

	course, err := db.GetCourse(query.ID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Errorf(codes.NotFound, "course not found")
		}
		return nil, status.Errorf(codes.NotFound, "failed to get course from database")
	}
	return course, nil
}

// RefreshCourse updates the course assignments (and possibly other course information).
func RefreshCourse(ctx context.Context, request *pb.RecordRequest, s scm.SCM, db database.Database, currentUser *pb.User) (*pb.Assignments, error) {
	course, err := db.GetCourse(request.ID)
	if err != nil {
		return nil, err
	}

	assignments, err := FetchAssignments(ctx, s, course)
	if err != nil {
		return nil, err
	}

	if err = db.UpdateAssignments(assignments); err != nil {
		return nil, err
	}
	//TODO(meling) Are the assignments (previously it was yamlparser.NewAssignmentRequest)
	// needed by the frontend? Or can we use models.Assignment instead through db.GetAssignmentsByCourse()?
	// Currently the frontend looks faulty, i.e. doesn't use the returned results from this
	// function; see 'refreshCoursesFor(course_CourseID: number): Promise<any>' in ServerProvider.ts,
	// which does a 'return this.makeUserInfo(result.data);', indicating that the result is
	// converted into a UserInfo type, which probably fails??

	return &pb.Assignments{Assignments: assignments}, nil
}

// GetSubmission returns a single submission for a assignment and a user
func GetSubmission(request *pb.RecordRequest, db database.Database, currentUser *pb.User) (*pb.Submission, error) {
	submission, err := db.GetSubmissionForUser(request.ID, currentUser.ID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Errorf(codes.NotFound, "not found")
		}
		return nil, err
	}

	return submission, nil

}

// ListSubmissions returns all the latests submissions for a user to a course
func ListSubmissions(request *pb.ActionRequest, db database.Database) (*pb.Submissions, error) {

	submissions, err := db.GetSubmissions(request.UserID, request.CourseID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Errorf(codes.NotFound, "not found")
		}
		return nil, err
	}
	return &pb.Submissions{Submissions: submissions}, nil
}

// UpdateCourse updates an existing course
func UpdateCourse(ctx context.Context, request *pb.Course, db database.Database, s scm.SCM) error {
	_, err := db.GetCourse(request.ID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return status.Errorf(codes.NotFound, "Course not found")
		}
		return err
	}

	// Check that the directory exists.
	_, err = s.GetOrganization(ctx, request.OrganizationID)
	if err != nil {
		return status.Errorf(codes.Aborted, "no directory found")
	}

	if err := db.UpdateCourse(request); err != nil {
		return status.Errorf(codes.Aborted, "could not create course")
	}
	return nil
}

// GetEnrollmentsByCourse get all enrollments for a course.
func GetEnrollmentsByCourse(request *pb.EnrollmentRequest, db database.Database) (*pb.Enrollments, error) {

	enrollments, err := db.GetEnrollmentsByCourse(request.CourseID, request.States...)
	if err != nil {
		return nil, err
	}

	// to populate response only with users who are not member of any group, we must filter the result
	if request.FilterOutGroupMembers {
		enrollmentsWithoutGroups := make([]*pb.Enrollment, 0)
		for _, enrollment := range enrollments {
			if enrollment.GroupID == 0 {
				enrollmentsWithoutGroups = append(enrollmentsWithoutGroups, enrollment)
			}
		}
		enrollments = enrollmentsWithoutGroups
	}

	for _, enrollment := range enrollments {
		enrollment.User, err = db.GetUser(enrollment.UserID)
		if err != nil {
			return nil, err
		}
	}
	return &pb.Enrollments{Enrollments: enrollments}, nil
}

// UpdateSubmission updates a submission
func UpdateSubmission(request *pb.RecordRequest, db database.Database) error {
	return db.UpdateSubmissionByID(request.ID, true)
}

// ListGroupSubmissions fetches all submissions from specific group
func ListGroupSubmissions(request *pb.ActionRequest, db database.Database) (*pb.Submissions, error) {
	submissions, err := db.GetGroupSubmissions(request.CourseID, request.UserID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Errorf(codes.NotFound, "not found")
		}
		return nil, status.Errorf(codes.NotFound, "no submissions found")
	}

	return &pb.Submissions{Submissions: submissions}, nil

}

// GetRepositoryURL returns the repository information
func GetRepositoryURL(currentUser *pb.User, request *pb.RepositoryRequest, db database.Database) (*pb.URLResponse, error) {
	course, err := db.GetCourse(request.GetCourseID())
	if err != nil {
		return nil, err
	}
	userRepoQuery := &pb.Repository{
		OrganizationID: course.GetOrganizationID(),
		RepoType:       request.GetType(),
	}
	if request.Type == pb.Repository_USER {
		userRepoQuery.UserID = currentUser.GetID()
	}

	repos, err := db.GetRepositories(userRepoQuery)
	if err != nil {
		return nil, err
	}
	if len(repos) != 1 {
		return nil, fmt.Errorf("found %d repositories for query %+v", len(repos), userRepoQuery)
	}
	return &pb.URLResponse{URL: repos[0].HTMLURL}, nil
}

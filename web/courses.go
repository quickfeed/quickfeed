package web

import (
	"context"
	"fmt"
	"sort"

	pb "github.com/autograde/aguis/ag"
	"github.com/autograde/aguis/scm"
)

// getCourses returns all courses.
func (s *AutograderService) getCourses() (*pb.Courses, error) {
	courses, err := s.db.GetCourses()
	if err != nil {
		return nil, err
	}
	return &pb.Courses{Courses: courses}, nil
}

// getCoursesWithEnrollment returns all courses that match the provided enrollment status.
func (s *AutograderService) getCoursesWithEnrollment(request *pb.CoursesListRequest) (*pb.Courses, error) {
	courses, err := s.db.GetCoursesByUser(request.GetUserID(), request.States...)
	if err != nil {
		return nil, err
	}
	return &pb.Courses{Courses: courses}, nil
}

// createEnrollment creates a pending enrollment for the given user and course.
func (s *AutograderService) createEnrollment(request *pb.Enrollment) error {
	enrollment := pb.Enrollment{
		UserID:   request.GetUserID(),
		CourseID: request.GetCourseID(),
		Status:   pb.Enrollment_PENDING,
	}
	return s.db.CreateEnrollment(&enrollment)
}

// updateEnrollment changes the status of the given course enrollment.
func (s *AutograderService) updateEnrollment(ctx context.Context, sc scm.SCM, curUser string, request *pb.Enrollment) error {
	enrollment, err := s.db.GetEnrollmentByCourseAndUser(request.CourseID, request.UserID)
	if err != nil {
		return err
	}
	// log changes to teacher status
	if enrollment.Status == pb.Enrollment_TEACHER || request.Status == pb.Enrollment_TEACHER {
		s.logger.Debugf("User %s attempting to change enrollment status of user %d from %s to %s", curUser, enrollment.UserID, enrollment.Status, request.Status)
	}

	switch request.Status {
	case pb.Enrollment_NONE:
		return s.rejectEnrollment(ctx, sc, enrollment)

	case pb.Enrollment_STUDENT:
		return s.enrollStudent(ctx, sc, enrollment)

	case pb.Enrollment_TEACHER:
		return s.enrollTeacher(ctx, sc, enrollment)
	}
	return fmt.Errorf("unknown enrollment")
}

func (s *AutograderService) updateEnrollments(ctx context.Context, sc scm.SCM, cid uint64) error {
	enrolls, err := s.db.GetEnrollmentsByCourse(cid, pb.Enrollment_PENDING)
	if err != nil {
		return err
	}
	for _, enrol := range enrolls {
		enrol.Status = pb.Enrollment_STUDENT
		if err = s.updateEnrollment(ctx, sc, "", enrol); err != nil {
			return err
		}
	}
	return nil
}

// updateReposAndTeams changes access to the course repositories and team memberships of the given user
// depending on the given enrollment status.
func updateReposAndTeams(ctx context.Context, sc scm.SCM, course *pb.Course, login string, state pb.Enrollment_UserStatus) (*scm.Repository, error) {
	if course.GetOrganizationPath() == "" {
		org, err := sc.GetOrganization(ctx, &scm.GetOrgOptions{ID: course.OrganizationID})
		if err != nil {
			return nil, err
		}
		course.OrganizationPath = org.GetPath()
	}
	var err error

	switch state {
	case pb.Enrollment_STUDENT:
		// get all repositories for organization
		repos, err := sc.GetRepositories(ctx, &pb.Organization{ID: course.GetOrganizationID(), Path: course.GetOrganizationPath()})
		if err != nil {
			return nil, err
		}
		// grant read access to assignments and course-info repositories
		for _, r := range repos {
			if r.Path == pb.AssignmentRepo || r.Path == pb.InfoRepo {
				if err = sc.UpdateRepoAccess(ctx, &scm.Repository{Owner: r.Owner, Path: r.Path}, login, scm.RepoPull); err != nil {
					return nil, fmt.Errorf("updateReposAndTeams: failed to update repo access to repo %s for user %s: %w ", r.Path, login, err)
				}
			}
		}

		// add student to the organization's "students" team
		if err = addUserToStudentsTeam(ctx, sc, course.GetOrganizationPath(), login); err != nil {
			return nil, err
		}
		org := &pb.Organization{ID: course.GetOrganizationID(), Path: course.GetOrganizationPath()}
		return createStudentRepo(ctx, sc, org, pb.StudentRepoName(login), login)

	case pb.Enrollment_TEACHER:
		// if teacher, promote to owner, remove from students team, add to teachers team
		orgUpdate := &scm.OrgMembershipOptions{
			Organization: course.GetOrganizationPath(),
			Username:     login,
			Role:         scm.OrgOwner,
		}
		// when promoting to teacher, promote to organization owner as well
		if err := sc.UpdateOrgMembership(ctx, orgUpdate); err != nil {
			return nil, fmt.Errorf("UpdateReposAndTeams: failed to update org membership for %s: %w", login, err)
		}
		err = promoteUserToTeachersTeam(ctx, sc, course.OrganizationPath, login)
	}
	return nil, err
}

// GetCourse returns a course object for the given course id.
func (s *AutograderService) getCourse(courseID uint64) (*pb.Course, error) {
	return s.db.GetCourse(courseID, false)
}

// getSubmissions returns all the latests submissions for a user of the given course.
func (s *AutograderService) getSubmissions(request *pb.SubmissionRequest) (*pb.Submissions, error) {
	// only one of user ID and group ID will be set; enforced by IsValid on pb.SubmissionRequest
	query := &pb.Submission{
		UserID:  request.GetUserID(),
		GroupID: request.GetGroupID(),
	}
	submissions, err := s.db.GetSubmissions(request.GetCourseID(), query)
	if err != nil {
		return nil, err
	}
	return &pb.Submissions{Submissions: submissions}, nil
}

// getAllLabs returns all individual lab submissions by students enrolled in the specified course.
func (s *AutograderService) getAllLabs(request *pb.LabRequest) ([]*pb.LabResultLink, error) {
	allLabs, err := s.db.GetCourseSubmissions(request.GetCourseID(), request.GetGroupLabs())
	if err != nil {
		return nil, err
	}

	//TODO(meling): Not sure this cache is effective, since the map is created on every call! Consider options!

	// make a local map to store database values to avoid querying the database multiple times
	// format: [studentID][assignmentID]{latest submission}
	labCache := make(map[uint64]map[uint64]pb.Submission)

	// populate cache map with student labs, filtering the latest submissions for every assignment
	for _, lab := range allLabs {
		labID := lab.GetUserID()
		if request.GroupLabs {
			labID = lab.GetGroupID()
		}
		_, ok := labCache[labID]
		if !ok {
			labCache[labID] = make(map[uint64]pb.Submission)
		}
		labCache[labID][lab.GetAssignmentID()] = lab
	}

	// fetch course record with all assignments and active enrollments
	course, err := s.db.GetCourse(request.GetCourseID(), true)
	if err != nil {
		return nil, err
	}

	return makeLabResults(course, labCache, request.GetGroupLabs()), nil
}

// updateSubmission approves the given submission or undoes a previous approval.
func (s *AutograderService) updateSubmission(submissionID uint64, approve bool) error {
	return s.db.UpdateSubmission(submissionID, approve)
}

// updateCourse updates an existing course.
func (s *AutograderService) updateCourse(ctx context.Context, sc scm.SCM, request *pb.Course) error {
	// ensure the course exists
	_, err := s.db.GetCourse(request.ID, false)
	if err != nil {
		return err
	}
	// ensure the organization exists
	org, err := sc.GetOrganization(ctx, &scm.GetOrgOptions{ID: request.OrganizationID})
	if err != nil {
		return err
	}
	request.OrganizationPath = org.GetPath()
	return s.db.UpdateCourse(request)
}

// getEnrollmentsByCourse get all enrollments for a course that match the given enrollment request.
func (s *AutograderService) getEnrollmentsByCourse(request *pb.EnrollmentRequest) (*pb.Enrollments, error) {
	enrollments, err := s.db.GetEnrollmentsByCourse(request.CourseID, request.States...)
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
	return &pb.Enrollments{Enrollments: enrollments}, nil
}

// getRepositoryURL returns URL of a course repository of the given type.
func (s *AutograderService) getRepositoryURL(currentUser *pb.User, courseID uint64, repoType pb.Repository_Type) (string, error) {
	course, err := s.db.GetCourse(courseID, false)
	if err != nil {
		return "", err
	}
	userRepoQuery := &pb.Repository{
		OrganizationID: course.GetOrganizationID(),
		RepoType:       repoType,
	}

	switch repoType {
	case pb.Repository_USER:
		userRepoQuery.UserID = currentUser.GetID()
	case pb.Repository_GROUP:
		enrol, err := s.db.GetEnrollmentByCourseAndUser(courseID, currentUser.GetID())
		if err != nil {
			return "", err
		}
		if enrol.GetGroupID() > 0 {
			userRepoQuery.GroupID = enrol.GroupID
		}
	}

	repos, err := s.db.GetRepositories(userRepoQuery)
	if err != nil {
		return "", err
	}
	if len(repos) != 1 {
		return "", fmt.Errorf("found %d repositories for query %+v", len(repos), userRepoQuery)
	}
	return repos[0].HTMLURL, nil
}

// isEmptyRepo returns nil if all repositories for the given course and student or group are empty,
// returns an error otherwise.
func (s *AutograderService) isEmptyRepo(ctx context.Context, sc scm.SCM, request *pb.RepositoryRequest) error {
	course, err := s.db.GetCourse(request.GetCourseID(), false)
	if err != nil {
		return err
	}
	repos, err := s.db.GetRepositories(&pb.Repository{OrganizationID: course.GetOrganizationID(), UserID: request.GetUserID(), GroupID: request.GetGroupID()})
	if err != nil {
		return err
	}
	if len(repos) < 1 {
		return fmt.Errorf("no repositories found")
	}
	return isEmpty(ctx, sc, repos)
}

// rejectEnrollment rejects a student enrollment, if a student repo exists for the given course, removes it from the SCM and database.
func (s *AutograderService) rejectEnrollment(ctx context.Context, sc scm.SCM, enrolled *pb.Enrollment) error {
	// course and user are both preloaded, no need to query the database
	course, user := enrolled.GetCourse(), enrolled.GetUser()
	repos, err := s.db.GetRepositories(&pb.Repository{
		UserID:         user.GetID(),
		OrganizationID: course.GetOrganizationID(),
		RepoType:       pb.Repository_USER,
	})
	if err != nil {
		return err
	}
	for _, repo := range repos {
		// we do not care about errors here, even if the github repo does not exists,
		// log the error and go on with deleting database entries
		if err := removeUserFromCourse(ctx, sc, user.GetLogin(), repo); err != nil {
			s.logger.Debug("updateEnrollment: rejectUserFromCourse failed (expected behavior): ", err)
		}

		if err := s.db.DeleteRepositoryByRemoteID(repo.GetRepositoryID()); err != nil {
			return err
		}
	}
	return s.db.RejectEnrollment(user.ID, course.ID)
}

// enrollStudent enrolls the given user as a student into the given course.
func (s *AutograderService) enrollStudent(ctx context.Context, sc scm.SCM, enrolled *pb.Enrollment) error {
	// course and user are both preloaded, no need to query the database
	course, user := enrolled.GetCourse(), enrolled.GetUser()

	// check whether user repo already exists,
	// which could happen if accepting a previously rejected student
	userRepoQuery := &pb.Repository{
		OrganizationID: course.GetOrganizationID(),
		UserID:         user.GetID(),
		RepoType:       pb.Repository_USER,
	}
	repos, err := s.db.GetRepositories(userRepoQuery)
	if err != nil {
		return err
	}
	s.logger.Debug("Enrolling student: ", user.GetLogin(), " have database repos: ", len(repos))
	if len(repos) > 0 {
		// repo already exist, update enrollment in database
		return s.db.EnrollStudent(user.ID, course.ID)
	}

	// create user repo, user team, and add user to students team
	repo, err := updateReposAndTeams(ctx, sc, course, user.GetLogin(), pb.Enrollment_STUDENT)
	if err != nil {
		s.logger.Errorf("failed to update repos or team membersip for student %s: %s", user.Login, err.Error())
		return err
	}
	s.logger.Debug("Enrolling student: ", user.GetLogin(), " repo and team update done")

	// add student repo to database if SCM interaction above was successful
	userRepo := pb.Repository{
		OrganizationID: course.GetOrganizationID(),
		RepositoryID:   repo.ID,
		UserID:         user.ID,
		HTMLURL:        repo.WebURL,
		RepoType:       pb.Repository_USER,
	}

	// only create database record if there are no user repos
	// TODO(vera): this can be set as a unique constraint in go tag in proto
	// but will it be compatible with the database created without this constraint?
	if dbRepo, _ := s.db.GetRepositories(&userRepo); len(dbRepo) < 1 {
		if err := s.db.CreateRepository(&userRepo); err != nil {
			return err
		}
	}
	return s.db.EnrollStudent(user.ID, course.ID)
}

// enrollTeacher promotes the given user to teacher of the given course
func (s *AutograderService) enrollTeacher(ctx context.Context, sc scm.SCM, enrolled *pb.Enrollment) error {
	// course and user are both preloaded, no need to query the database
	course, user := enrolled.GetCourse(), enrolled.GetUser()

	// make owner, remove from students, add to teachers
	if _, err := updateReposAndTeams(ctx, sc, course, user.GetLogin(), pb.Enrollment_TEACHER); err != nil {
		s.logger.Errorf("failed to update team membership for teacher %s: %s", user.Login, err.Error())
		return err
	}
	return s.db.EnrollTeacher(user.ID, course.ID)
}

func makeLabResults(course *pb.Course, labCache map[uint64]map[uint64]pb.Submission, groupLab bool) []*pb.LabResultLink {
	allCourseLabs := make([]*pb.LabResultLink, 0)

	if groupLab {
		for _, grp := range course.Groups {
			groupSubmissions := make([]*pb.Submission, 0)
			for _, v := range labCache[grp.GetID()] {
				tmp := v
				groupSubmissions = append(groupSubmissions, &tmp)
			}

			// sort latest submissions by lab ID
			sort.Slice(groupSubmissions, func(i, j int) bool {
				return groupSubmissions[i].GetAssignmentID() < groupSubmissions[j].GetAssignmentID()
			})

			labResult := &pb.LabResultLink{
				AuthorName: grp.GetName(),
				Enrollment: &pb.Enrollment{
					CourseID: course.ID,
					GroupID:  grp.GetID(),
					Group:    grp,
				},
				Submissions: groupSubmissions,
			}
			allCourseLabs = append(allCourseLabs, labResult)
		}
	} else {
		for _, usr := range course.Enrollments {
			// collect all submission values for the user
			userSubmissions := make([]*pb.Submission, 0)
			for _, v := range labCache[usr.GetUserID()] {
				tmp := v
				userSubmissions = append(userSubmissions, &tmp)
			}

			// sort latest submissions by lab ID
			sort.Slice(userSubmissions, func(i, j int) bool {
				return userSubmissions[i].GetAssignmentID() < userSubmissions[j].GetAssignmentID()
			})

			labResult := &pb.LabResultLink{
				AuthorName:  usr.GetUser().GetName(),
				Enrollment:  usr,
				Submissions: userSubmissions,
			}
			allCourseLabs = append(allCourseLabs, labResult)
		}
	}

	return allCourseLabs
}

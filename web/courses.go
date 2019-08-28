package web

import (
	"context"
	"fmt"

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
func (s *AutograderService) getCoursesWithEnrollment(request *pb.RecordRequest) (*pb.Courses, error) {
	courses, err := s.db.GetCoursesByUser(request.ID, request.Statuses...)
	if err != nil {
		return nil, err
	}
	return &pb.Courses{Courses: courses}, nil
}

// createEnrollment enrolls a user in a course.
func (s *AutograderService) createEnrollment(request *pb.Enrollment) error {
	enrollment := pb.Enrollment{
		UserID:   request.GetUserID(),
		CourseID: request.GetCourseID(),
		Status:   pb.Enrollment_PENDING,
	}
	return s.db.CreateEnrollment(&enrollment)
}

// updateEnrollment accepts or rejects a user to enroll in a course.
func (s *AutograderService) updateEnrollment(ctx context.Context, sc scm.SCM, request *pb.Enrollment) error {
	enrollment, err := s.db.GetEnrollmentByCourseAndUser(request.CourseID, request.UserID)
	if err != nil {
		return err
	}

	switch request.Status {
	case pb.Enrollment_REJECTED:
		return s.db.RejectEnrollment(request.UserID, request.CourseID)

	case pb.Enrollment_PENDING:
		return s.db.SetPendingEnrollment(request.UserID, request.CourseID)

	case pb.Enrollment_STUDENT:
		course, student := enrollment.GetCourse(), enrollment.GetUser()

		// check whether user repo already exists,
		// which could happen if accepting a previously rejected student
		userRepoQuery := &pb.Repository{
			OrganizationID: course.GetOrganizationID(),
			UserID:         request.GetUserID(),
			RepoType:       pb.Repository_USER,
		}
		repos, err := s.db.GetRepositories(userRepoQuery)
		if err != nil {
			return err
		}
		if len(repos) > 0 {
			// repo already exist, update enrollment in database
			return s.db.EnrollStudent(request.UserID, request.CourseID)
		}

		// create user repo, user team, and add user to students team
		// TODO(vera): creation of a single user team can be replaced by adding student as collaborator with push permission to user repo
		repo, err := updateReposAndTeams(ctx, sc, course, student.GetLogin(), request.GetStatus())
		if err != nil {
			s.logger.Errorf("failed to update repos or team membersip for student %s: %s", student.Login, err.Error())
			return err
		}

		// add student repo to database if SCM interaction above was successful
		userRepo := pb.Repository{
			OrganizationID: course.GetOrganizationID(),
			RepositoryID:   repo.ID,
			UserID:         request.GetUserID(),
			HTMLURL:        repo.WebURL,
			RepoType:       pb.Repository_USER,
		}

		// only create database record if there are no user repos
		// TODO(vera): this can be set as a unique constraint in go tag in proto
		// but will it be compatible with the database created without this constraint?
		if dbRepo, _ := s.db.GetRepositories(&userRepo); dbRepo == nil {
			if err := s.db.CreateRepository(&userRepo); err != nil {
				return err
			}
		}

		return s.db.EnrollStudent(request.UserID, request.CourseID)

	case pb.Enrollment_TEACHER:
		course, teacher := enrollment.GetCourse(), enrollment.GetUser()

		// make owner, remove from students, add to teachers
		if _, err := updateReposAndTeams(ctx, sc, course, teacher.GetLogin(), request.GetStatus()); err != nil {
			s.logger.Errorf("failed to update team membersip for teacher %s: %s", teacher.Login, err.Error())
			return err
		}
		return s.db.EnrollTeacher(teacher.ID, course.ID)
	}

	return fmt.Errorf("unknown enrollment")
}

func (s *AutograderService) updateEnrollments(ctx context.Context, sc scm.SCM, cid uint64) error {
	// get all pending enrolmets for the course ID
	enrolls, err := s.db.GetEnrollmentsByCourse(cid, pb.Enrollment_PENDING)
	if err != nil {
		return err
	}
	// TODO(vera): feels like a really bad idea to pass the same context multiple times
	// we probably need a new scm method to batch scm requests
	// it is certainly possible to add all approved students to org and student team in one request
	// but can be a bit more tricky with repos
	for _, enrol := range enrolls {
		if err = s.updateEnrollment(ctx, sc, enrol); err != nil {
			return err
		}
	}
	return nil
}

func updateReposAndTeams(ctx context.Context, sc scm.SCM, course *pb.Course, login string, state pb.Enrollment_UserStatus) (*scm.Repository, error) {
	org, err := sc.GetOrganization(ctx, course.OrganizationID)
	if err != nil {
		return nil, err
	}

	switch state {
	case pb.Enrollment_STUDENT:
		// get repos for organization
		repos, err := sc.GetRepositories(ctx, &pb.Organization{ID: org.GetID(), Path: org.GetPath()})
		if err != nil {
			return nil, err
		}
		// grant read access to Assignments and Course-info repositories
		for _, r := range repos {
			if r.Path == "assignments" || r.Path == "course-info" {
				if err = sc.UpdateRepoAccess(ctx, &scm.Repository{Owner: r.Owner, Path: r.Path}, login, scm.RepoPull); err != nil {
					return nil, fmt.Errorf("updateReposAndTeams: failed to update repo access to repo %s for user %s: %w ", r.Path, login, err)
				}
			}
		}

		// add student to the organization's "students" team
		if err = addUserToStudentsTeam(ctx, sc, org, login); err != nil {
			return nil, err
		}

		return createStudentRepo(ctx, sc, org, pb.StudentRepoName(login), login)

	case pb.Enrollment_TEACHER:
		// if teacher, promote to owner, remove from students team, add to teachers team
		orgUpdate := &scm.OrgMembershipOptions{
			Organization: org,
			Username:     login,
			Role:         scm.OrgOwner,
		}
		// when promoting to teacher, promote to organization owner as well
		if err = sc.UpdateOrgMembership(ctx, orgUpdate); err != nil {
			return nil, fmt.Errorf("UpdateReposAndTeams: failed to update org membership for %s: %w", login, err)
		}
		err = promoteUserToTeachersTeam(ctx, sc, org, login)
	}
	return nil, err
}

// GetCourse returns a course object for the given course id.
func (s *AutograderService) getCourse(courseID uint64) (*pb.Course, error) {
	return s.db.GetCourse(courseID)
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

// approveSubmission approves the given submission.
func (s *AutograderService) approveSubmission(submissionID uint64) error {
	return s.db.UpdateSubmission(submissionID, true)
}

// updateCourse updates an existing course.
func (s *AutograderService) updateCourse(ctx context.Context, sc scm.SCM, request *pb.Course) error {
	// ensure the course exists
	_, err := s.db.GetCourse(request.ID)
	if err != nil {
		return err
	}
	// ensure the organization exists
	_, err = sc.GetOrganization(ctx, request.OrganizationID)
	if err != nil {
		return err
	}
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

// getRepositoryURL returns the repository information
func (s *AutograderService) getRepositoryURL(currentUser *pb.User, courseID uint64, repoType pb.Repository_Type) (string, error) {
	course, err := s.db.GetCourse(courseID)
	if err != nil {
		return "", err
	}
	userRepoQuery := &pb.Repository{
		OrganizationID: course.GetOrganizationID(),
		RepoType:       repoType,
	}
	if repoType == pb.Repository_USER {
		userRepoQuery.UserID = currentUser.GetID()
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

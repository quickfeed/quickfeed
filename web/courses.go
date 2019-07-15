package web

import (
	"context"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/autograde/aguis/ag"
	"github.com/autograde/aguis/database"
	"github.com/autograde/aguis/scm"
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
// TODO(meling) simplify the flow of this func; too long; split into sub-functions
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
	}

	// TODO If the enrollment is accepted, create repositories and permissions for them with webhooks.
	switch request.Status {
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
			// if a user repo already exists it most likely means that the student
			// was previously accepted to course, then rejected, and is now being
			// accepted again. Also, it means that user team has already been created
			// and invitation to organization has been issued.
			// we only need to update enrollment in the database
			return s.db.EnrollStudent(request.UserID, request.CourseID)
		}

		// create user repo and team on SCM.
		// personal team is being created because user repo access is based on team role now
		// TODO(vera): we could avoid creating personal teams for every student if there is a way
		// to automate granting the user write access to user repo on repo creation
		// TODO(meling): clean up these three methods (integrate into one helper method)
		repo, _, err := createUserRepoAndTeam(ctx, sc, course, student)
		if err != nil {
			return err
		}
		// send invitation to course organization to student (will return nil if successful or already a member)
		if err = addUserToOrg(ctx, sc, course.GetOrganizationID(), student); err != nil {
			return err
		}
		// then add to the 'students' team
		if err = addToUserTeam(ctx, sc, course.GetOrganizationID(), student, pb.Enrollment_STUDENT); err != nil {
			return err
		}

		// add student repo to database if SCM interaction above was successful
		dbRepo := pb.Repository{
			OrganizationID: course.GetOrganizationID(),
			UserID:         request.GetUserID(),
			RepoType:       pb.Repository_USER,
			RepositoryID:   repo.ID,
			HTMLURL:        repo.WebURL,
		}
		if err := s.db.CreateRepository(&dbRepo); err != nil {
			return err
		}
		return s.db.EnrollStudent(request.UserID, request.CourseID)

	case pb.Enrollment_TEACHER:
		course, teacher := enrollment.GetCourse(), enrollment.GetUser()
		// we want to update org membership first, as it is more prone to errors user could react to
		orgUpdate := &scm.OrgMembership{
			Username: teacher.GetLogin(),
			OrgID:    course.GetOrganizationID(),
			Role:     "admin", //TODO(meling) is admin a role?
		}
		if err = sc.UpdateOrgMembership(ctx, orgUpdate); err != nil {
			return err
		}
		// add all teachers to 'teachers' team (will remove them from 'students' team)
		if err = addToUserTeam(ctx, sc, course.GetOrganizationID(), teacher, pb.Enrollment_TEACHER); err != nil {
			return err
		}
		return s.db.EnrollTeacher(teacher.ID, course.ID)
	}

	return fmt.Errorf("unknown enrollment")
}

func createUserRepoAndTeam(ctx context.Context, sc scm.SCM, course *pb.Course, student *pb.User) (*scm.Repository, *scm.Team, error) {
	org, err := sc.GetOrganization(ctx, course.OrganizationID)
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

	return sc.CreateRepoAndTeam(ctx, opt, teamName, []string{teamName})
}

// TODO(meling) why is not used?
func createUserRepo(c context.Context, sc scm.SCM, orgID uint64, student *pb.User) (*scm.Repository, error) {
	org, err := sc.GetOrganization(c, orgID)
	if err != nil {
		return nil, err
	}

	opt := &scm.CreateRepositoryOptions{
		Organization: org,
		Path:         pb.StudentRepoName(student.GetLogin()),
		Private:      true,
		Owner:        student.GetLogin(),
	}

	return sc.CreateRepository(c, opt)
}

func addToUserTeam(c context.Context, s scm.SCM, orgID uint64, user *pb.User, status pb.Enrollment_UserStatus) error {
	// get course organization
	org, err := s.GetOrganization(c, orgID)
	if err != nil {
		return err
	}

	opt := &scm.TeamMembershipOptions{
		Organization: org,
		TeamSlug:     "students",
		Username:     user.GetLogin(),
	}

	// check whether user is teacher or not
	if status == pb.Enrollment_TEACHER {
		// remove user from students
		if err = s.RemoveTeamMember(c, opt); err != nil {
			return err
		}
		// add user to teachers
		opt.TeamSlug = "teachers"
		opt.Role = "maintainer"
	}
	return s.AddTeamMember(c, opt)
}

func addUserToOrg(ctx context.Context, s scm.SCM, orgID uint64, user *pb.User) error {
	// get course organization
	org, err := s.GetOrganization(ctx, orgID)
	if err != nil {
		return err
	}
	return s.CreateOrgMembership(ctx, &scm.OrgMembershipOptions{Organization: org, Username: user.GetLogin()})
}

// GetCourse returns a course object for the given course id.
func (s *AutograderService) getCourse(courseID uint64) (*pb.Course, error) {
	return s.db.GetCourse(courseID)
}

// getSubmission returns a single submission for a assignment and a user
func (s *AutograderService) getSubmission(currentUser *pb.User, request *pb.RecordRequest) (*pb.Submission, error) {
	// ensure that the submission belongs to the current user
	query := &pb.Submission{AssignmentID: request.ID, UserID: currentUser.ID}
	return s.db.GetSubmission(query)
}

// getSubmissions returns all the latests submissions for a user to a course
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

// getEnrollmentsByCourse get all enrollments for a course.
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
func (s *AutograderService) getRepositoryURL(currentUser *pb.User, request *pb.RepositoryRequest) (*pb.URLResponse, error) {
	course, err := s.db.GetCourse(request.GetCourseID())
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

	repos, err := s.db.GetRepositories(userRepoQuery)
	if err != nil {
		return nil, err
	}
	if len(repos) != 1 {
		return nil, fmt.Errorf("found %d repositories for query %+v", len(repos), userRepoQuery)
	}
	return &pb.URLResponse{URL: repos[0].HTMLURL}, nil
}

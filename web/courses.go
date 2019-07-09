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
func CreateEnrollment(request *pb.Enrollment, db database.Database) error {
	enrollment := pb.Enrollment{
		UserID:   request.GetUserID(),
		CourseID: request.GetCourseID(),
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
// TODO(meling) simplify the flow of this func; too long; split into sub-functions
func UpdateEnrollment(ctx context.Context, request *pb.Enrollment, db database.Database, s scm.SCM) error {
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
			// create user repo and team on SCM.
			// personal team is being created because user repo access is based on team role now
			// TODO(vera): we could avoid creating personal teams for every student if there is a way
			// to automate granting the user write access to user repo on repo creation
			repo, _, err := createUserRepoAndTeam(ctx, s, course, student)
			if err != nil {
				return status.Errorf(codes.Internal, "could not create user repository and team")
			}

			err = db.EnrollStudent(request.UserID, request.CourseID)
			if err != nil {
				return status.Errorf(codes.Internal, "could not enroll student")
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
			// along with personal team we will add all students to students team
			if err = addToUserTeam(ctx, s, course.GetOrganizationID(), student, pb.Enrollment_STUDENT); err != nil {
				return err
			}
			// then send invitation to course organization to student (will return nil if successful or already a member)
			return addUserToOrg(ctx, s, course.GetOrganizationID(), student)
		}
		// if repo already exists (student was previously accepted to course, then rejected, and now is being accepted again),
		// it means that user team has already been created and invitation to organization has been issued
		// we only need to update enrollment in the database
		return db.EnrollStudent(request.UserID, request.CourseID)

	case pb.Enrollment_TEACHER:
		course, err := db.GetCourse(request.CourseID)
		if err != nil {
			return status.Errorf(codes.NotFound, "course not found")
		}
		student, err := db.GetUser(request.UserID)
		if err != nil {
			return status.Errorf(codes.NotFound, "user not found")
		}
		// we want to update org membership first, as it is more prone to errors user could react to
		orgUpdate := &scm.OrgMembership{
			Username: student.GetLogin(),
			OrgID:    course.GetOrganizationID(),
			Role:     "admin",
		}
		if err = s.UpdateOrgMembership(ctx, orgUpdate); err != nil {
			return status.Errorf(codes.Internal, fmt.Sprintln("could not update github membership, reason: ", err.Error()))
		}
		if err = db.EnrollTeacher(student.ID, course.ID); err != nil {
			return status.Errorf(codes.Internal, "could not enroll teacher")
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

// TODO(meling) why is not used?
func createUserRepo(c context.Context, s scm.SCM, orgID uint64, student *pb.User) (*scm.Repository, error) {
	org, err := s.GetOrganization(c, orgID)
	if err != nil {
		return nil, err
	}

	opt := &scm.CreateRepositoryOptions{
		Organization: org,
		Path:         pb.StudentRepoName(student.GetLogin()),
		Private:      true,
		Owner:        student.GetLogin(),
	}

	return s.CreateRepository(c, opt)
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
		opt := &scm.TeamMembershipOptions{
			Organization: org,
			TeamSlug:     "students",
			Username:     user.GetLogin(),
			Role:         "member",
		}
		return s.AddTeamMember(c, opt)

	case pb.Enrollment_TEACHER:
		// remove user from students
		opt := &scm.TeamMembershipOptions{
			Organization: org,
			TeamSlug:     "students",
			Username:     user.GetLogin(),
		}
		if err = s.RemoveTeamMember(c, opt); err != nil {
			return err
		}
		// add user to teachers
		opt.TeamSlug = "teachers"
		// TODO(meling) should this have Role: "admin" ??
		return s.AddTeamMember(c, opt)
	}

	opt := &scm.TeamMembershipOptions{
		Organization: org,
		TeamSlug:     slug, //TODO(meling) this is uninitialized, why?
		Username:     user.GetLogin(),
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

// RefreshCourse updates the course assignments (and possibly other course information).
func RefreshCourse(ctx context.Context, request *pb.RecordRequest, s scm.SCM, db database.Database, currentUser *pb.User) (*pb.Assignments, error) {
	course, err := db.GetCourse(request.ID)
	if err != nil {
		return nil, err
	}

	assignments, err := fetchAssignments(ctx, s, course)
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
	query := &pb.Submission{AssignmentID: request.ID, UserID: currentUser.ID}
	submission, err := db.GetSubmission(query)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Errorf(codes.NotFound, "not found")
		}
		return nil, err
	}
	return submission, nil
}

// getSubmissions returns all the latests submissions for a user to a course
func (s *AutograderService) getSubmissions(request *pb.SubmissionRequest) (*pb.Submissions, error) {
	// only one of user ID and group ID will be set; enforced by the IsValid
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

// UpdateSubmission updates a submission
func UpdateSubmission(request *pb.RecordRequest, db database.Database) error {
	return db.UpdateSubmission(request.ID, true)
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

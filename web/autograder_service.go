package web

import (
	"context"
	"errors"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/autograde/quickfeed/ag"
	pb "github.com/autograde/quickfeed/ag/types"
	"github.com/autograde/quickfeed/ci"
	"github.com/autograde/quickfeed/database"
	"github.com/autograde/quickfeed/scm"
	"github.com/autograde/quickfeed/web/auth"
	"github.com/autograde/quickfeed/web/auth/tokens"
	"github.com/autograde/quickfeed/web/config"
)

// AutograderService holds references to the database and
// other shared data structures.
type AutograderService struct {
	logger       *zap.SugaredLogger
	db           database.Database
	scmMaker     *scm.SCMMaker
	Config       *config.Config // TODO(vera): make unexported again after refactoring the startup method
	tokenManager *tokens.TokenManager
	runner       ci.Runner
	ag.UnimplementedAutograderServiceServer
}

// NewAutograderService returns an AutograderService object.
func NewAutograderService(logger *zap.Logger, db database.Database, app *scm.SCMMaker, config *config.Config, tokens *tokens.TokenManager, runner ci.Runner) *AutograderService {
	return &AutograderService{
		logger:       logger.Sugar(),
		db:           db,
		scmMaker:     app,
		Config:       config,
		tokenManager: tokens,
		runner:       runner,
	}
}

// GetUser will return current user with active course enrollments
// to use in separating teacher and admin roles
func (s *AutograderService) GetUser(ctx context.Context, _ *pb.Void) (*pb.User, error) {
	usr, err := s.getCurrentUser(ctx)
	if err != nil {
		s.logger.Errorf("GetUser failed: authentication error: %v", err)
		return nil, ErrInvalidUserInfo
	}
	userInfo, err := s.db.GetUserWithEnrollments(usr.GetID())
	if err != nil {
		s.logger.Errorf("GetUser failed to get user with enrollments: %v ", err)
	}
	return userInfo, nil
}

// GetUsers returns a list of all users.
// Frontend note: This method is called from AdminPage.
func (s *AutograderService) GetUsers(ctx context.Context, _ *pb.Void) (*pb.Users, error) {
	users, err := s.getUsers()
	if err != nil {
		s.logger.Errorf("GetUsers failed: %v", err)
		return nil, status.Error(codes.NotFound, "failed to get users")
	}
	return users, nil
}

// GetUserByCourse returns the user matching the given course name and GitHub login
// specified in CourseUserRequest.
func (s *AutograderService) GetUserByCourse(ctx context.Context, in *pb.CourseUserRequest) (*pb.User, error) {
	usr, err := s.getCurrentUser(ctx)
	if err != nil {
		s.logger.Errorf("GetUserByCourse failed: authentication error: %v", err)
		return nil, ErrInvalidUserInfo
	}
	userInfo, err := s.getUserByCourse(in, usr)
	if err != nil {
		s.logger.Errorf("GetUserByCourse failed: %+v", err)
		return nil, status.Error(codes.FailedPrecondition, "failed to get student information")
	}
	return userInfo, nil
}

// UpdateUser updates the current users's information and returns the updated user.
// This function can also promote a user to admin or demote a user.
func (s *AutograderService) UpdateUser(ctx context.Context, in *pb.User) (*pb.Void, error) {
	usr, err := s.getCurrentUser(ctx)
	if err != nil {
		s.logger.Errorf("UpdateUser failed: authentication error: %v", err)
		return nil, ErrInvalidUserInfo
	}

	if _, err = s.updateUser(usr, in); err != nil {
		s.logger.Errorf("UpdateUser failed to update user %d: %v", in.GetID(), err)
		err = status.Error(codes.InvalidArgument, "failed to update user")
	}
	return &pb.Void{}, err
}

// CreateCourse creates a new course.
func (s *AutograderService) CreateCourse(ctx context.Context, in *pb.Course) (*pb.Course, error) {
	usr, err := s.getCurrentUser(ctx)
	if err != nil {
		s.logger.Errorf("CreateCourse failed: scm authentication error: %v", err)
		return nil, err
	}
	sc, err := s.scmMaker.GetOrCreateSCMEntry(s.logger, in)
	if err != nil {
		s.logger.Errorf("CreateCourse failed: scm client error: %v", err)
		return nil, status.Errorf(codes.InvalidArgument, "failed to create course")
	}
	in.CourseCreatorID = usr.GetID()
	course, err := s.createCourse(ctx, sc, in)
	if err != nil {
		s.logger.Errorf("CreateCourse failed: %v", err)
		// errors informing about requested organization state will have code 9: FailedPrecondition
		// error message will be displayed to the user
		if contextCanceled(ctx) {
			return nil, status.Error(codes.FailedPrecondition, ErrContextCanceled)
		}
		if err == ErrAlreadyExists || err == ErrFreePlan {
			return nil, status.Error(codes.FailedPrecondition, err.Error())
		}
		if ok, parsedErr := parseSCMError(err); ok {
			return nil, parsedErr
		}
		return nil, status.Error(codes.InvalidArgument, "failed to create course")
	}
	s.scmMaker.AddSCM(sc, course.GetID())
	return course, nil
}

// UpdateCourse changes the course information details.
func (s *AutograderService) UpdateCourse(ctx context.Context, in *pb.Course) (*pb.Void, error) {
	scm, err := s.getSCM(in.GetID())
	if err != nil {
		s.logger.Errorf("UpdateCourse failed: csm client for course %s not found: %v", in.GetCode(), err)
		return nil, ErrInvalidUserInfo
	}
	if err := s.updateCourse(ctx, scm, in); err != nil {
		s.logger.Errorf("UpdateCourse failed: %v", err)
		if contextCanceled(ctx) {
			return nil, status.Error(codes.FailedPrecondition, ErrContextCanceled)
		}
		if ok, parsedErr := parseSCMError(err); ok {
			return nil, parsedErr
		}
		return nil, status.Error(codes.InvalidArgument, "failed to update course")
	}
	return &pb.Void{}, nil
}

// GetCourse returns course information for the given course.
func (s *AutograderService) GetCourse(ctx context.Context, in *pb.CourseRequest) (*pb.Course, error) {
	course, err := s.getCourse(in.GetCourseID())
	if err != nil {
		s.logger.Errorf("GetCourse failed: %v", err)
		return nil, status.Error(codes.NotFound, "course not found")
	}
	return course, nil
}

// GetCourses returns a list of all courses.
func (s *AutograderService) GetCourses(_ context.Context, _ *pb.Void) (*pb.Courses, error) {
	courses, err := s.getCourses()
	if err != nil {
		s.logger.Errorf("GetCourses failed: %v", err)
		return nil, status.Error(codes.NotFound, "no courses found")
	}
	return courses, nil
}

// UpdateCourseVisibility allows to edit what courses are visible in the sidebar.
func (s *AutograderService) UpdateCourseVisibility(ctx context.Context, in *pb.Enrollment) (*pb.Void, error) {
	err := s.changeCourseVisibility(in)
	if err != nil {
		s.logger.Errorf("ChangeCourseVisibility failed: %v", err)
		err = status.Error(codes.InvalidArgument, "failed to update course visibility")
	}
	return &pb.Void{}, err
}

// CreateEnrollment enrolls a new student for the course specified in the request.
func (s *AutograderService) CreateEnrollment(_ context.Context, in *pb.Enrollment) (*pb.Void, error) {
	err := s.createEnrollment(in)
	if err != nil {
		s.logger.Errorf("CreateEnrollment failed: %v", err)
		err = status.Error(codes.InvalidArgument, "failed to create enrollment")
	}
	return &pb.Void{}, err
}

// UpdateEnrollments changes status of all pending enrollments for the specified course to approved.
// If the request contains a single enrollment, it will be updated to the specified status.
func (s *AutograderService) UpdateEnrollments(ctx context.Context, in *pb.Enrollments) (*pb.Void, error) {
	usr, err := s.getCurrentUser(ctx)
	if err != nil {
		s.logger.Errorf("UpdateEnrollment failed: scm authentication error: %v", err)
		return nil, err
	}
	scm, err := s.getSCM(in.GetCourseID())
	if err != nil {
		s.logger.Errorf("UpdateEnrollments failed: %v", err)
		return nil, ErrInvalidUserInfo
	}
	for _, enrollment := range in.GetEnrollments() {
		if s.isCourseCreator(enrollment.CourseID, enrollment.UserID) {
			s.logger.Errorf("UpdateEnrollments failed: user %s attempted to demote course creator", usr.GetName())
			return nil, status.Error(codes.PermissionDenied, "course creator cannot be demoted")
		}
		if err = s.updateEnrollment(ctx, scm, usr.GetLogin(), enrollment); err != nil {
			s.logger.Errorf("UpdateEnrollments failed: %v", err)
			if contextCanceled(ctx) {
				return nil, status.Error(codes.FailedPrecondition, ErrContextCanceled)
			}
			if ok, parsedErr := parseSCMError(err); ok {
				return nil, parsedErr
			}
			return nil, status.Error(codes.InvalidArgument, "failed to update enrollment")
		}
	}
	return &pb.Void{}, err
}

// GetCoursesByUser returns all courses the given user is enrolled into with the given status.
func (s *AutograderService) GetCoursesByUser(_ context.Context, in *pb.EnrollmentStatusRequest) (*pb.Courses, error) {
	courses, err := s.getCoursesByUser(in)
	if err != nil {
		s.logger.Errorf("GetCoursesWithEnrollment failed: %v", err)
		return nil, status.Error(codes.NotFound, "no courses with enrollment found")
	}
	return courses, nil
}

// GetEnrollmentsByUser returns all enrollments for the given user and enrollment status with preloaded courses and groups.
func (s *AutograderService) GetEnrollmentsByUser(ctx context.Context, in *pb.EnrollmentStatusRequest) (*pb.Enrollments, error) {
	enrols, err := s.getEnrollmentsByUser(in)
	if err != nil {
		s.logger.Errorf("Get enrollments for user %d failed: %v", in.GetUserID(), err)
	}
	return enrols, nil
}

// GetEnrollmentsByCourse returns all enrollments for the course specified in the request.
func (s *AutograderService) GetEnrollmentsByCourse(ctx context.Context, in *pb.EnrollmentRequest) (*pb.Enrollments, error) {
	enrolls, err := s.getEnrollmentsByCourse(in)
	if err != nil {
		s.logger.Errorf("GetEnrollmentsByCourse failed: %v", err)
		return nil, status.Error(codes.InvalidArgument, "failed to get enrollments for given course")
	}
	return enrolls, nil
}

// GetGroup returns information about a group.
func (s *AutograderService) GetGroup(ctx context.Context, in *pb.GetGroupRequest) (*pb.Group, error) {
	group, err := s.getGroup(in)
	if err != nil {
		s.logger.Errorf("GetGroup failed: %v", err)
		return nil, status.Error(codes.NotFound, "failed to get group")
	}
	return group, nil
}

// GetGroupsByCourse returns a list of groups created for the course id in the record request.
func (s *AutograderService) GetGroupsByCourse(ctx context.Context, in *pb.CourseRequest) (*pb.Groups, error) {
	groups, err := s.getGroups(in)
	if err != nil {
		s.logger.Errorf("GetGroups failed: %v", err)
		return nil, status.Error(codes.NotFound, "failed to get groups")
	}
	return groups, nil
}

// GetGroupByUserAndCourse returns the group of the given student for a given course.
func (s *AutograderService) GetGroupByUserAndCourse(ctx context.Context, in *pb.GroupRequest) (*pb.Group, error) {
	group, err := s.getGroupByUserAndCourse(in)
	if err != nil {
		if err != errUserNotInGroup {
			s.logger.Errorf("GetGroupByUserAndCourse failed: %v", err)
		}
		return nil, status.Error(codes.NotFound, "failed to get group for given user and course")
	}
	return group, nil
}

// CreateGroup creates a new group in the database.
func (s *AutograderService) CreateGroup(ctx context.Context, in *pb.Group) (*pb.Group, error) {
	group, err := s.createGroup(in)
	if err != nil {
		s.logger.Errorf("CreateGroup failed: %v", err)
		if _, ok := status.FromError(err); ok {
			// err was already a status error; return it to client.
			return nil, err
		}
		// err was not a status error; return a generic error to client.
		return nil, status.Error(codes.InvalidArgument, "failed to create group")
	}
	return group, nil
}

// UpdateGroup updates group information.
func (s *AutograderService) UpdateGroup(ctx context.Context, in *pb.Group) (*pb.Void, error) {
	scm, err := s.getSCM(in.GetCourseID())
	if err != nil {
		s.logger.Errorf("UpdateGroup failed: %v", err)
		return nil, ErrInvalidUserInfo
	}
	err = s.updateGroup(ctx, scm, in)
	if err != nil {
		s.logger.Errorf("UpdateGroup failed: %v", err)
		if contextCanceled(ctx) {
			return nil, status.Error(codes.FailedPrecondition, ErrContextCanceled)
		}
		if ok, parsedErr := parseSCMError(err); ok {
			return nil, parsedErr
		}
		if _, ok := status.FromError(err); ok {
			// err was already a status error; return it to client.
			return nil, err
		}
		// err was not a status error; return a generic error to client.
		return nil, status.Error(codes.InvalidArgument, "failed to update group")
	}
	return &pb.Void{}, nil
}

// DeleteGroup removes group record from the database.
func (s *AutograderService) DeleteGroup(ctx context.Context, in *pb.GroupRequest) (*pb.Void, error) {
	scm, err := s.getSCM(in.GetCourseID())
	if err != nil {
		s.logger.Errorf("DeleteGroup failed: %v", err)
		return nil, ErrInvalidUserInfo
	}
	if err := s.deleteGroup(ctx, scm, in); err != nil {
		s.logger.Errorf("DeleteGroup failed: %v", err)
		if contextCanceled(ctx) {
			return nil, status.Error(codes.FailedPrecondition, ErrContextCanceled)
		}
		if ok, parsedErr := parseSCMError(errors.Unwrap(err)); ok {
			return nil, parsedErr
		}
		return nil, status.Error(codes.InvalidArgument, "failed to delete group")
	}
	return &pb.Void{}, nil
}

// GetSubmissions returns the submissions matching the query encoded in the action request.
func (s *AutograderService) GetSubmissions(ctx context.Context, in *pb.SubmissionRequest) (*pb.Submissions, error) {
	s.logger.Debugf("GetSubmissions: %v", in)
	submissions, err := s.getSubmissions(in)
	if err != nil {
		s.logger.Errorf("GetSubmissions failed: %v", err)
		return nil, status.Error(codes.NotFound, "no submissions found")
	}
	return submissions, nil
}

// GetSubmissionsByCourse returns all the latest submissions
// for every individual or group course assignment for all course students/groups.
func (s *AutograderService) GetSubmissionsByCourse(ctx context.Context, in *pb.SubmissionsForCourseRequest) (*pb.CourseSubmissions, error) {
	// TODO(meling) This is a hack to give access to cmd/approvelist via the root usr admin
	// Normally, the root admin should not have access to submissions for all courses.
	//	if usr.IsAdmin {
	//		goto BYPASS
	//	}

	// BYPASS:
	s.logger.Debugf("GetSubmissionsByCourse: %v", in)
	courseLinks, err := s.getAllCourseSubmissions(in)
	if err != nil {
		s.logger.Errorf("GetSubmissionsByCourse failed: %v", err)
		return nil, status.Error(codes.NotFound, "no submissions found")
	}
	return courseLinks, nil
}

// UpdateSubmission is called to approve the given submission or to undo approval.
func (s *AutograderService) UpdateSubmission(ctx context.Context, in *pb.UpdateSubmissionRequest) (*pb.Void, error) {
	// TODO(vera): probably can check in interceptor
	if !s.isValidSubmission(in.SubmissionID) {
		s.logger.Errorf("UpdateSubmission failed: submission author has no access to the course")
		return nil, status.Error(codes.PermissionDenied, "submission author has no course access")
	}
	err := s.updateSubmission(in.GetCourseID(), in.GetSubmissionID(), in.GetStatus(), in.GetReleased(), in.GetScore())
	if err != nil {
		s.logger.Errorf("UpdateSubmission failed: %v", err)
		err = status.Error(codes.InvalidArgument, "failed to approve submission")
	}
	return &pb.Void{}, err
}

// RebuildSubmissions re-runs the tests for the given assignment.
// A single submission is executed again if the request specifies a submission ID
// or all submissions if the request specifies a course ID.
func (s *AutograderService) RebuildSubmissions(ctx context.Context, in *pb.RebuildRequest) (*pb.Void, error) {
	// RebuildType can be either SubmissionID or CourseID, but not both.
	switch in.GetRebuildType().(type) {
	case *pb.RebuildRequest_SubmissionID:
		if !s.isValidSubmission(in.GetSubmissionID()) {
			s.logger.Errorf("RebuildSubmission failed: submitter has no access to the course")
			return nil, status.Error(codes.PermissionDenied, "submitter has no course access")
		}
		if _, err := s.rebuildSubmission(in); err != nil {
			s.logger.Errorf("RebuildSubmission failed: %v", err)
			return nil, status.Error(codes.InvalidArgument, "failed to rebuild submission")
		}
	case *pb.RebuildRequest_CourseID:
		if err := s.rebuildSubmissions(in); err != nil {
			s.logger.Errorf("RebuildSubmissions failed: %v", err)
			return nil, status.Error(codes.InvalidArgument, "failed to rebuild submissions")
		}
	}
	return &pb.Void{}, nil
}

// CreateBenchmark adds a new grading benchmark for an assignment
func (s *AutograderService) CreateBenchmark(_ context.Context, in *pb.BenchmarkRequest) (*pb.GradingBenchmark, error) {
	bm, err := s.createBenchmark(in.GetBenchmark())
	if err != nil {
		s.logger.Errorf("CreateBenchmark failed for %+v: %v", in, err)
		return nil, status.Error(codes.InvalidArgument, "failed to add benchmark")
	}
	return bm, nil
}

// UpdateBenchmark edits a grading benchmark for an assignment
func (s *AutograderService) UpdateBenchmark(_ context.Context, in *pb.BenchmarkRequest) (*pb.Void, error) {
	err := s.updateBenchmark(in.GetBenchmark())
	if err != nil {
		s.logger.Errorf("UpdateBenchmark failed for %+v: %v", in, err)
		err = status.Error(codes.InvalidArgument, "failed to update benchmark")
	}
	return &pb.Void{}, err
}

// DeleteBenchmark removes a grading benchmark
func (s *AutograderService) DeleteBenchmark(_ context.Context, in *pb.BenchmarkRequest) (*pb.Void, error) {
	err := s.deleteBenchmark(in.GetBenchmark())
	if err != nil {
		s.logger.Errorf("DeleteBenchmark failed for %+v: %v", in, err)
		err = status.Error(codes.InvalidArgument, "failed to delete benchmark")
	}
	return &pb.Void{}, err
}

// CreateCriterion adds a new grading criterion for an assignment
func (s *AutograderService) CreateCriterion(_ context.Context, in *pb.CriteriaRequest) (*pb.GradingCriterion, error) {
	c, err := s.createCriterion(in.GetCriterion())
	if err != nil {
		s.logger.Errorf("CreateCriterion failed for %+v: %v", in, err)
		return nil, status.Error(codes.InvalidArgument, "failed to add criterion")
	}
	return c, nil
}

// UpdateCriterion edits a grading criterion for an assignment
func (s *AutograderService) UpdateCriterion(_ context.Context, in *pb.CriteriaRequest) (*pb.Void, error) {
	err := s.updateCriterion(in.GetCriterion())
	if err != nil {
		s.logger.Errorf("UpdateCriterion failed for %+v: %v", in, err)
		err = status.Error(codes.InvalidArgument, "failed to update criterion")
	}
	return &pb.Void{}, err
}

// DeleteCriterion removes a grading criterion for an assignment
func (s *AutograderService) DeleteCriterion(_ context.Context, in *pb.CriteriaRequest) (*pb.Void, error) {
	err := s.deleteCriterion(in.GetCriterion())
	if err != nil {
		s.logger.Errorf("DeleteCriterion failed for %+v: %v", in, err)
		err = status.Error(codes.InvalidArgument, "failed to delete criterion")
	}
	return &pb.Void{}, err
}

// CreateReview adds a new submission review
func (s *AutograderService) CreateReview(ctx context.Context, in *pb.ReviewRequest) (*pb.Review, error) {
	review, err := s.createReview(in.Review)
	if err != nil {
		s.logger.Errorf("CreateReview failed for review %+v: %v", in, err)
		return nil, status.Error(codes.InvalidArgument, "failed to create review")
	}
	return review, nil
}

// UpdateReview updates a submission review
func (s *AutograderService) UpdateReview(ctx context.Context, in *pb.ReviewRequest) (*pb.Review, error) {
	review, err := s.updateReview(in.Review)
	if err != nil {
		s.logger.Errorf("UpdateReview failed for review %+v: %v", in, err)
		err = status.Error(codes.InvalidArgument, "failed to update review")
	}
	return review, err
}

// UpdateSubmissions approves and/or releases all manual reviews for student submission for the given assignment
// with the given score.
func (s *AutograderService) UpdateSubmissions(ctx context.Context, in *pb.UpdateSubmissionsRequest) (*pb.Void, error) {
	err := s.updateSubmissions(in)
	if err != nil {
		s.logger.Errorf("UpdateSubmissions failed for request %+v", in)
		err = status.Error(codes.InvalidArgument, "failed to update submissions")
	}
	return &pb.Void{}, err
}

// GetReviewers returns names of all active reviewers for a student submission
func (s *AutograderService) GetReviewers(ctx context.Context, in *pb.SubmissionReviewersRequest) (*pb.Reviewers, error) {
	reviewers, err := s.getReviewers(in.SubmissionID)
	if err != nil {
		s.logger.Errorf("GetReviewers failed: error fetching from database: %v", err)
		return nil, status.Error(codes.InvalidArgument, "failed to get reviewers")
	}
	return &pb.Reviewers{Reviewers: reviewers}, err
}

// GetAssignments returns a list of all assignments for the given course.
func (s *AutograderService) GetAssignments(_ context.Context, in *pb.CourseRequest) (*pb.Assignments, error) {
	courseID := in.GetCourseID()
	assignments, err := s.getAssignments(courseID)
	if err != nil {
		s.logger.Errorf("GetAssignments failed: %v", err)
		return nil, status.Error(codes.NotFound, "no assignments found for course")
	}
	return assignments, nil
}

// UpdateAssignments updates the assignments record in the database
// by fetching assignment information from the course's test repository.
func (s *AutograderService) UpdateAssignments(ctx context.Context, in *pb.CourseRequest) (*pb.Void, error) {
	courseID := in.GetCourseID()
	err := s.updateAssignments(courseID)
	if err != nil {
		s.logger.Errorf("UpdateAssignments failed: %v", err)
		return nil, status.Error(codes.NotFound, "course not found")
	}
	return &pb.Void{}, nil
}

// TODO(vera): looks like this method is never called?
// GetProviders returns a list of SCM providers supported by the backend.
// Access policy: Any User.
func (s *AutograderService) GetProviders(_ context.Context, _ *pb.Void) (*pb.Providers, error) {
	providers := auth.GetProviders()
	if len(providers.GetProviders()) < 1 {
		s.logger.Error("GetProviders failed: found no enabled SCM providers")
		return nil, status.Error(codes.NotFound, "found no enabled SCM providers")
	}
	return providers, nil
}

// GetOrganization fetches a github organization by name.
func (s *AutograderService) GetOrganization(ctx context.Context, in *pb.OrgRequest) (*pb.Organization, error) {
	usr, err := s.getCurrentUser(ctx)
	if err != nil {
		s.logger.Errorf("GetOrganization failed: scm authentication error: %v", err)
		return nil, err
	}
	courseCreatorToken, err := usr.GetAccessToken("github")
	if err != nil {
		s.logger.Errorf("GetOrganization failed: error getting access token for user %s and organization %s: %v", usr.Login, in.OrgName, err)
		return nil, status.Error(codes.NotFound, "failed to get organization: missing access token")
	}
	// GetOrganization is called before a new course is created. There is no scm client yet,
	// so we need to create a new one from the installation in order to access a private organization.
	sc, err := s.scmMaker.NewSCM(ctx, s.logger, in.OrgName, courseCreatorToken)
	if err != nil {
		s.logger.Errorf("GetOrganization failed: error creating scm client for user %s and organization %s: %v", usr.Login, in.OrgName, err)
		return nil, status.Error(codes.NotFound, "failed to get organization: missing access token")
	}
	org, err := s.getOrganization(ctx, sc, in.GetOrgName(), usr.GetLogin())
	if err != nil {
		s.logger.Errorf("GetOrganization failed: %v", err)
		if contextCanceled(ctx) {
			return nil, status.Error(codes.FailedPrecondition, ErrContextCanceled)
		}
		if err == scm.ErrNotMember {
			return nil, status.Error(codes.NotFound, "organization membership not confirmed, please enable third-party access")
		}
		if err == ErrFreePlan || err == ErrAlreadyExists || err == scm.ErrNotOwner {
			return nil, status.Error(codes.FailedPrecondition, err.Error())
		}
		if ok, parsedErr := parseSCMError(err); ok {
			return nil, parsedErr
		}
		return nil, status.Error(codes.NotFound, "organization not found. Please make sure that 3rd-party access is enabled for your organization")
	}
	return org, nil
}

// GetRepositories returns URL strings for repositories of given type for the given course.
func (s *AutograderService) GetRepositories(ctx context.Context, in *pb.URLRequest) (*pb.Repositories, error) {
	usr, err := s.getCurrentUser(ctx)
	if err != nil {
		s.logger.Errorf("GetRepositories failed: authentication error: %v", err)
		return nil, ErrInvalidUserInfo
	}
	course, err := s.getCourse(in.GetCourseID())
	if err != nil {
		s.logger.Errorf("GetRepositories failed: course %d not found: %v", in.GetCourseID(), err)
		return nil, status.Error(codes.NotFound, "course not found")
	}
	enrol, _ := s.db.GetEnrollmentByCourseAndUser(course.GetID(), usr.GetID())

	urls := make(map[string]string)
	for _, repoType := range in.GetRepoTypes() {
		var id uint64
		switch repoType {
		case pb.Repository_USER:
			id = usr.GetID()
		case pb.Repository_GROUP:
			id = enrol.GetGroupID() // will be 0 if not enrolled in a group
		}
		repo, _ := s.getRepo(course, id, repoType)
		// for repo == nil: will result in an empty URL string, which will be ignored by the frontend
		urls[repoType.String()] = repo.GetHTMLURL()
	}
	return &pb.Repositories{URLs: urls}, nil
}

// IsEmptyRepo ensures that group repository is empty and can be deleted
func (s *AutograderService) IsEmptyRepo(ctx context.Context, in *pb.RepositoryRequest) (*pb.Void, error) {
	scm, err := s.getSCM(in.GetCourseID())
	if err != nil {
		s.logger.Errorf("IsEmptyRepo failed: %v", err)
		return nil, status.Error(codes.FailedPrecondition, "failed to get scm client")
	}
	if err := s.isEmptyRepo(ctx, scm, in); err != nil {
		s.logger.Errorf("IsEmptyRepo failed: %v", err)
		if contextCanceled(ctx) {
			return nil, status.Error(codes.FailedPrecondition, ErrContextCanceled)
		}
		if ok, parsedErr := parseSCMError(err); ok {
			return nil, parsedErr
		}
		return nil, status.Error(codes.FailedPrecondition, "group repository does not exist or not empty")
	}
	return &pb.Void{}, nil
}

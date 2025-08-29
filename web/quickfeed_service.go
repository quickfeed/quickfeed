package web

import (
	"context"
	"errors"

	"go.uber.org/zap"

	"connectrpc.com/connect"
	"github.com/quickfeed/quickfeed/assignments"
	"github.com/quickfeed/quickfeed/ci"
	"github.com/quickfeed/quickfeed/database"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/qf/qfconnect"
	"github.com/quickfeed/quickfeed/scm"
	"github.com/quickfeed/quickfeed/web/stream"
)

var scmConnectErr = connect.NewError(connect.CodeNotFound, errors.New("unable to connect to the GitHub organization for the course"))

// QuickFeedService holds references to the database and
// other shared data structures.
type QuickFeedService struct {
	logger *zap.SugaredLogger
	db     database.Database
	scmMgr *scm.Manager
	bh     BaseHookOptions
	runner ci.Runner
	qfconnect.UnimplementedQuickFeedServiceHandler
	streams *stream.StreamServices
}

// NewQuickFeedService returns a QuickFeedService object.
func NewQuickFeedService(logger *zap.Logger, db database.Database, mgr *scm.Manager, bh BaseHookOptions, runner ci.Runner) *QuickFeedService {
	return &QuickFeedService{
		logger:  logger.Sugar(),
		db:      db,
		scmMgr:  mgr,
		bh:      bh,
		runner:  runner,
		streams: stream.NewStreamServices(),
	}
}

// GetUser will return current user with active course enrollments
// to use in separating teacher and admin roles
func (s *QuickFeedService) GetUser(ctx context.Context, _ *connect.Request[qf.Void]) (*connect.Response[qf.User], error) {
	userInfo, err := s.db.GetUserWithEnrollments(userID(ctx))
	if err != nil {
		s.logger.Errorf("GetUser(%d) failed: %v", userID(ctx), err)
		return nil, connect.NewError(connect.CodeNotFound, errors.New("unknown user"))
	}
	return connect.NewResponse(userInfo), nil
}

// GetUsers returns a list of all users.
// Frontend note: This method is called from AdminPage.
func (s *QuickFeedService) GetUsers(_ context.Context, _ *connect.Request[qf.Void]) (*connect.Response[qf.Users], error) {
	users, err := s.db.GetUsers()
	if err != nil {
		s.logger.Errorf("GetUsers failed: %v", err)
		return nil, connect.NewError(connect.CodeNotFound, errors.New("failed to get users"))
	}
	return connect.NewResponse(&qf.Users{
		Users: users,
	}), nil
}

// UpdateUser updates the current users's information and returns the updated user.
// This function can also promote a user to admin or demote a user.
func (s *QuickFeedService) UpdateUser(ctx context.Context, in *connect.Request[qf.User]) (*connect.Response[qf.Void], error) {
	usr, err := s.db.GetUser(userID(ctx))
	if err != nil {
		s.logger.Errorf("UpdateUser(userID=%d) failed: %v", userID(ctx), err)
		return nil, connect.NewError(connect.CodeNotFound, errors.New("unknown user"))
	}
	if err = s.editUserProfile(usr, in.Msg); err != nil {
		s.logger.Errorf("UpdateUser failed to update user %d: %v", in.Msg.GetID(), err)
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("failed to update user"))
	}
	return &connect.Response[qf.Void]{}, nil
}

// UpdateCourse changes the course information details.
func (s *QuickFeedService) UpdateCourse(ctx context.Context, in *connect.Request[qf.Course]) (*connect.Response[qf.Void], error) {
	scmClient, err := s.getSCM(ctx, in.Msg.GetScmOrganizationName())
	if err != nil {
		s.logger.Errorf("UpdateCourse failed: could not create scm client for organization %s: %v", in.Msg.GetScmOrganizationName(), err)
		return nil, scmConnectErr
	}
	// ensure the course exists
	_, err = s.db.GetCourse(in.Msg.GetID())
	if err != nil {
		s.logger.Errorf("UpdateCourse failed: course %d not found: %v", in.Msg.GetID(), err)
		return nil, connect.NewError(connect.CodeNotFound, errors.New("failed to get course"))
	}
	// ensure the organization exists
	org, err := scmClient.GetOrganization(ctx, &scm.OrganizationOptions{ID: in.Msg.GetScmOrganizationID()})
	if err != nil {
		s.logger.Errorf("UpdateCourse failed: to get organization %s: %v", in.Msg.GetScmOrganizationName(), in.Msg.GetID(), err)
		if ctxErr := ctxErr(ctx); ctxErr != nil {
			s.logger.Error(ctxErr)
			return nil, ctxErr
		}
		if scmErr := userSCMError(err); scmErr != nil {
			return nil, scmErr
		}
		return nil, connect.NewError(connect.CodeNotFound, errors.New("failed to get organization"))
	}
	in.Msg.ScmOrganizationName = org.GetScmOrganizationName()

	if err = s.db.UpdateCourse(in.Msg); err != nil {
		s.logger.Errorf("UpdateCourse failed: %v", err)
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("failed to update course"))
	}
	return &connect.Response[qf.Void]{}, nil
}

// GetCourse returns course information for the given course.
func (s *QuickFeedService) GetCourse(ctx context.Context, in *connect.Request[qf.CourseRequest]) (*connect.Response[qf.Course], error) {
	status := courseStatus(ctx, in.Msg.GetCourseID())
	course, err := s.db.GetCourseByStatus(in.Msg.GetCourseID(), status)
	if err != nil {
		s.logger.Errorf("GetCourse failed: %v", err)
		return nil, connect.NewError(connect.CodeNotFound, errors.New("course not found"))
	}
	if isTeacher(ctx, in.Msg.GetCourseID()) {
		course.Enrollments, err = s.getEnrollmentsWithActivity(in.Msg.GetCourseID())
		if err != nil {
			s.logger.Errorf("GetCourse failed: %v", err)
			return nil, connect.NewError(connect.CodeNotFound, errors.New("failed to get course enrollments"))
		}
	}

	return connect.NewResponse(course), nil
}

// GetCourses returns a list of all courses.
func (s *QuickFeedService) GetCourses(_ context.Context, _ *connect.Request[qf.Void]) (*connect.Response[qf.Courses], error) {
	courses, err := s.db.GetCourses()
	if err != nil {
		s.logger.Errorf("GetCourses failed: %v", err)
		return nil, connect.NewError(connect.CodeNotFound, errors.New("no courses found"))
	}
	return connect.NewResponse(&qf.Courses{
		Courses: courses,
	}), nil
}

// UpdateCourseVisibility allows to edit what courses are visible in the sidebar.
func (s *QuickFeedService) UpdateCourseVisibility(ctx context.Context, in *connect.Request[qf.Enrollment]) (*connect.Response[qf.Void], error) {
	enrollment, err := s.db.GetEnrollmentByCourseAndUser(in.Msg.GetCourseID(), userID(ctx))
	if err != nil {
		s.logger.Errorf("UpdateCourseVisibility failed: %v", err)
		return nil, connect.NewError(connect.CodeNotFound, errors.New("failed to get enrollment"))
	}

	enrollment.State = in.Msg.GetState()
	if err := s.db.UpdateEnrollment(enrollment); err != nil {
		s.logger.Errorf("ChangeCourseVisibility failed: %v", err)
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("failed to update course visibility"))
	}
	return &connect.Response[qf.Void]{}, nil
}

// CreateEnrollment enrolls a new student for the course specified in the request.
func (s *QuickFeedService) CreateEnrollment(_ context.Context, in *connect.Request[qf.Enrollment]) (*connect.Response[qf.Void], error) {
	if err := s.db.CreateEnrollment(in.Msg); err != nil {
		s.logger.Errorf("CreateEnrollment failed: %v", err)
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("failed to create enrollment"))
	}
	return &connect.Response[qf.Void]{}, nil
}

// UpdateEnrollments changes status of all pending enrollments for the specified course to approved.
// If the request contains a single enrollment, it will be updated to the specified status.
func (s *QuickFeedService) UpdateEnrollments(ctx context.Context, in *connect.Request[qf.Enrollments]) (*connect.Response[qf.Void], error) {
	usr, err := s.db.GetUser(userID(ctx))
	if err != nil {
		s.logger.Errorf("UpdateEnrollments(userID=%d) failed: %v", userID(ctx), err)
		return nil, connect.NewError(connect.CodeNotFound, errors.New("unknown user"))
	}
	scmClient, err := s.getSCMForCourse(ctx, in.Msg.GetCourseID())
	if err != nil {
		s.logger.Errorf("UpdateEnrollments failed: could not create scm client for course %d: %v", in.Msg.GetCourseID(), err)
		return nil, scmConnectErr
	}
	for _, enrollment := range in.Msg.GetEnrollments() {
		if s.isCourseCreator(enrollment.GetCourseID(), enrollment.GetUserID()) {
			s.logger.Errorf("UpdateEnrollments failed: user %s attempted to demote course creator", usr.GetName())
			return nil, connect.NewError(connect.CodePermissionDenied, errors.New("course creator cannot be demoted"))
		}
		if err = s.updateEnrollment(ctx, scmClient, usr.GetLogin(), enrollment); err != nil {
			s.logger.Errorf("UpdateEnrollments failed: %v", err)
			if ctxErr := ctxErr(ctx); ctxErr != nil {
				s.logger.Error(ctxErr)
				return nil, ctxErr
			}
			if scmErr := userSCMError(err); scmErr != nil {
				return nil, scmErr
			}
			return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("failed to update enrollments"))
		}
	}
	return &connect.Response[qf.Void]{}, nil
}

// GetEnrollments returns all enrollments for the given course ID or user ID and enrollment status.
func (s *QuickFeedService) GetEnrollments(ctx context.Context, in *connect.Request[qf.EnrollmentRequest]) (*connect.Response[qf.Enrollments], error) {
	var enrollments []*qf.Enrollment
	var err error
	statuses := in.Msg.GetStatuses()
	switch in.Msg.GetFetchMode().(type) {
	case *qf.EnrollmentRequest_UserID:
		userID := in.Msg.GetUserID()
		enrollments, err = s.db.GetEnrollmentsByUser(userID, statuses...)
		if err != nil {
			s.logger.Errorf("GetEnrollments failed: user %d: %v", userID, err)
			return nil, connect.NewError(connect.CodeNotFound, errors.New("no enrollments found for user"))
		}
	case *qf.EnrollmentRequest_CourseID:
		courseID := in.Msg.GetCourseID()
		if isTeacher(ctx, courseID) {
			enrollments, err = s.getEnrollmentsWithActivity(courseID)
		} else {
			enrollments, err = s.db.GetEnrollmentsByCourse(courseID, statuses...)
		}
		if err != nil {
			s.logger.Errorf("GetEnrollments failed: course %d: %v", courseID, err)
			return nil, connect.NewError(connect.CodeNotFound, errors.New("failed to get enrollments for course"))
		}
	default:
		s.logger.Errorf("GetEnrollments failed: unknown message type: %v", in.Msg.GetFetchMode())
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("failed to get enrollments"))
	}
	return connect.NewResponse(&qf.Enrollments{
		Enrollments: enrollments,
	}), nil
}

// GetGroup returns information about the given group ID, or the given user's course group if group ID is 0.
func (s *QuickFeedService) GetGroup(_ context.Context, in *connect.Request[qf.GroupRequest]) (*connect.Response[qf.Group], error) {
	var (
		group   *qf.Group
		err     error
		groupID = in.Msg.GetGroupID()
	)
	if groupID > 0 {
		group, err = s.db.GetGroup(groupID)
	} else {
		group, err = s.getGroupByUserAndCourse(in.Msg)
	}
	if err != nil {
		s.logger.Errorf("GetGroup failed: group %d: %v", in.Msg, err)
		return nil, connect.NewError(connect.CodeNotFound, errors.New("failed to get group"))
	}
	return connect.NewResponse(group), nil
}

// GetGroupsByCourse returns groups created for the given course.
func (s *QuickFeedService) GetGroupsByCourse(_ context.Context, in *connect.Request[qf.CourseRequest]) (*connect.Response[qf.Groups], error) {
	groups, err := s.db.GetGroupsByCourse(in.Msg.GetCourseID())
	if err != nil {
		s.logger.Errorf("GetGroups failed: course %d: %v", in.Msg.GetCourseID(), err)
		return nil, connect.NewError(connect.CodeNotFound, errors.New("failed to get groups"))
	}
	return connect.NewResponse(&qf.Groups{Groups: groups}), nil
}

// CreateGroup creates a new group for the given course and users.
// This function is typically called by a student when creating
// a group, which will later be (optionally) edited and approved
// by a teacher of the course using the updateGroup function below.
// Access policy: Any User enrolled in course and specified as member of the group or a course teacher.
func (s *QuickFeedService) CreateGroup(_ context.Context, in *connect.Request[qf.Group]) (*connect.Response[qf.Group], error) {
	group := in.Msg
	if err := s.checkGroupName(group.GetCourseID(), group.GetName()); err != nil {
		s.logger.Errorf("CreateGroup: failed to validate group %s: %v", group.GetName(), err)
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	// get users of group, check consistency of group request
	if _, err := s.getGroupUsers(group); err != nil {
		s.logger.Errorf("CreateGroup: failed to retrieve users for group %s: %v", group.GetName(), err)
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("failed to create group"))
	}
	// create new group and update groupID in enrollment table
	if err := s.db.CreateGroup(group); err != nil {
		s.logger.Errorf("CreateGroup failed: %v", err)
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("failed to create group"))
	}
	group, err := s.db.GetGroup(group.GetID())
	if err != nil {
		s.logger.Errorf("CreateGroup failed to get group %d: %v", group.GetID(), err)
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("failed to create group"))
	}
	return connect.NewResponse(group), nil
}

// UpdateGroup updates group information, and returns the updated group.
func (s *QuickFeedService) UpdateGroup(ctx context.Context, in *connect.Request[qf.Group]) (*connect.Response[qf.Group], error) {
	scmClient, err := s.getSCMForCourse(ctx, in.Msg.GetCourseID())
	if err != nil {
		s.logger.Errorf("UpdateGroup failed: could not create scm client for group %s and course %d: %v", in.Msg.GetName(), in.Msg.GetCourseID(), err)
		return nil, scmConnectErr
	}
	err = s.internalUpdateGroup(ctx, scmClient, in.Msg)
	if err != nil {
		s.logger.Errorf("UpdateGroup failed: %v", err)
		if ctxErr := ctxErr(ctx); ctxErr != nil {
			s.logger.Error(ctxErr)
			return nil, ctxErr
		}
		if scmErr := userSCMError(err); scmErr != nil {
			return nil, scmErr
		}
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("failed to update group"))
	}
	group, err := s.db.GetGroup(in.Msg.GetID())
	if err != nil {
		s.logger.Errorf("UpdateGroup failed to get group: %d: %v", in.Msg.GetID(), err)
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("failed to get group"))
	}
	return connect.NewResponse(group), nil
}

// DeleteGroup removes group record from the database.
func (s *QuickFeedService) DeleteGroup(ctx context.Context, in *connect.Request[qf.GroupRequest]) (*connect.Response[qf.Void], error) {
	scmClient, err := s.getSCMForCourse(ctx, in.Msg.GetCourseID())
	if err != nil {
		s.logger.Errorf("DeleteGroup failed: could not create scm client for group %d and course %d: %v", in.Msg.GetGroupID(), in.Msg.GetCourseID(), err)
		return nil, scmConnectErr
	}
	if err = s.internalDeleteGroup(ctx, scmClient, in.Msg); err != nil {
		s.logger.Errorf("DeleteGroup failed: %v", err)
		if ctxErr := ctxErr(ctx); ctxErr != nil {
			s.logger.Error(ctxErr)
			return nil, ctxErr
		}
		if scmErr := userSCMError(err); scmErr != nil {
			return nil, scmErr
		}
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("failed to delete group"))
	}
	return &connect.Response[qf.Void]{}, nil
}

// GetSubmission returns a fully populated submission matching the given submission ID if it exists for the given course ID.
// Used in the frontend to fetch a full submission for a given submission ID and course ID.
func (s *QuickFeedService) GetSubmission(_ context.Context, in *connect.Request[qf.SubmissionRequest]) (*connect.Response[qf.Submission], error) {
	submission, err := s.db.GetLastSubmission(in.Msg.GetCourseID(), &qf.Submission{ID: in.Msg.GetSubmissionID()})
	if err != nil {
		s.logger.Errorf("GetSubmission failed: %v", err)
		return nil, connect.NewError(connect.CodeNotFound, errors.New("failed to get submission"))
	}
	return connect.NewResponse(submission), nil
}

// GetSubmissions returns the submissions matching the query encoded in the action request.
func (s *QuickFeedService) GetSubmissions(ctx context.Context, in *connect.Request[qf.SubmissionRequest]) (*connect.Response[qf.Submissions], error) {
	s.logger.Debugf("GetSubmissions: %v", in.Msg)
	query := &qf.Submission{
		UserID:  in.Msg.GetUserID(),
		GroupID: in.Msg.GetGroupID(),
	}
	subs, err := s.db.GetLastSubmissions(in.Msg.GetCourseID(), query)
	if err != nil {
		s.logger.Errorf("GetSubmissions failed: %v", err)
		return nil, connect.NewError(connect.CodeNotFound, errors.New("no submissions found"))
	}
	submissions := &qf.Submissions{Submissions: subs}
	id := userID(ctx)
	// If the user is not a teacher, remove score and reviews from submissions that are not released.
	if !s.isTeacher(id, in.Msg.GetCourseID()) {
		submissions.Clean(id)
	}
	return connect.NewResponse(submissions), nil
}

// GetSubmissionsByCourse returns a map of submissions for the given course ID.
// The map is keyed by either the group ID or enrollment ID depending on request type.
// SubmissionRequest_GROUP returns a map keyed by group ID.
// SubmissionRequest_ALL and SubmissionRequest_USER return a map keyed by enrollment ID.
// The map values are lists of all submissions for the given group or enrollment.
func (s *QuickFeedService) GetSubmissionsByCourse(_ context.Context, in *connect.Request[qf.SubmissionRequest]) (*connect.Response[qf.CourseSubmissions], error) {
	s.logger.Debugf("GetSubmissionsByCourse: %v", in)
	courseLinks, err := s.db.GetCourseSubmissions(in.Msg)
	if err != nil {
		s.logger.Errorf("GetSubmissionsByCourse failed: %v", err)
		return nil, connect.NewError(connect.CodeNotFound, errors.New("no submissions found"))
	}
	return connect.NewResponse(courseLinks), nil
}

// UpdateSubmission is called to approve the given submission or to undo approval.
func (s *QuickFeedService) UpdateSubmission(_ context.Context, in *connect.Request[qf.UpdateSubmissionRequest]) (*connect.Response[qf.Void], error) {
	submission, err := s.db.GetSubmission(&qf.Submission{ID: in.Msg.GetSubmissionID()})
	if err != nil {
		s.logger.Errorf("UpdateSubmission failed to get submission: %v", err)
		return nil, connect.NewError(connect.CodeNotFound, errors.New("failed to update submission"))
	}
	submission.SetGradesAndRelease(in.Msg)
	err = s.db.UpdateSubmission(submission)
	if err != nil {
		s.logger.Errorf("UpdateSubmission failed: %v", err)
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("failed to approve submission"))
	}
	return &connect.Response[qf.Void]{}, nil
}

// RebuildSubmissions re-runs the tests for the given assignment and course.
// A single submission is executed again if the request specifies a submission ID
// or all submissions if no submission ID is specified.
func (s *QuickFeedService) RebuildSubmissions(_ context.Context, in *connect.Request[qf.RebuildRequest]) (*connect.Response[qf.Void], error) {
	if in.Msg.GetSubmissionID() > 0 {
		// Submission ID > 0 ==> rebuild single submission for given CourseID and AssignmentID
		if err := s.internalRebuildSubmission(in.Msg); err != nil {
			s.logger.Errorf("RebuildSubmission failed: %v", err)
			return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("failed to rebuild submission"))
		}
	} else {
		// Submission ID == 0 ==> rebuild all for given CourseID and AssignmentID
		if err := s.internalRebuildAllSubmissions(in.Msg); err != nil {
			s.logger.Errorf("RebuildSubmissions failed: %v", err)
			return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("failed to rebuild submissions"))
		}
	}
	return &connect.Response[qf.Void]{}, nil
}

// CreateBenchmark adds a new grading benchmark for an assignment.
func (s *QuickFeedService) CreateBenchmark(_ context.Context, in *connect.Request[qf.GradingBenchmark]) (*connect.Response[qf.GradingBenchmark], error) {
	benchmark := in.Msg
	if err := s.db.CreateBenchmark(benchmark); err != nil {
		s.logger.Errorf("CreateBenchmark failed for %+v: %v", in, err)
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("failed to create benchmark"))
	}
	return connect.NewResponse(benchmark), nil
}

// UpdateBenchmark edits a grading benchmark for an assignment.
func (s *QuickFeedService) UpdateBenchmark(_ context.Context, in *connect.Request[qf.GradingBenchmark]) (*connect.Response[qf.Void], error) {
	if err := s.db.UpdateBenchmark(in.Msg); err != nil {
		s.logger.Errorf("UpdateBenchmark failed for %+v: %v", in, err)
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("failed to update benchmark"))
	}
	return &connect.Response[qf.Void]{}, nil
}

// DeleteBenchmark removes a grading benchmark.
func (s *QuickFeedService) DeleteBenchmark(_ context.Context, in *connect.Request[qf.GradingBenchmark]) (*connect.Response[qf.Void], error) {
	if err := s.db.DeleteBenchmark(in.Msg); err != nil {
		s.logger.Errorf("DeleteBenchmark failed for %+v: %v", in, err)
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("failed to delete benchmark"))
	}
	return &connect.Response[qf.Void]{}, nil
}

// CreateCriterion adds a new grading criterion for an assignment.
func (s *QuickFeedService) CreateCriterion(_ context.Context, in *connect.Request[qf.GradingCriterion]) (*connect.Response[qf.GradingCriterion], error) {
	criterion := in.Msg
	if err := s.db.CreateCriterion(criterion); err != nil {
		s.logger.Errorf("CreateCriterion failed for %+v: %v", in, err)
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("failed to add criterion"))
	}
	return connect.NewResponse(criterion), nil
}

// UpdateCriterion edits a grading criterion for an assignment.
func (s *QuickFeedService) UpdateCriterion(_ context.Context, in *connect.Request[qf.GradingCriterion]) (*connect.Response[qf.Void], error) {
	if err := s.db.UpdateCriterion(in.Msg); err != nil {
		s.logger.Errorf("UpdateCriterion failed for %+v: %v", in, err)
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("failed to update criterion"))
	}
	return &connect.Response[qf.Void]{}, nil
}

// DeleteCriterion removes a grading criterion for an assignment.
func (s *QuickFeedService) DeleteCriterion(_ context.Context, in *connect.Request[qf.GradingCriterion]) (*connect.Response[qf.Void], error) {
	if err := s.db.DeleteCriterion(in.Msg); err != nil {
		s.logger.Errorf("DeleteCriterion failed for %+v: %v", in, err)
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("failed to delete criterion"))
	}
	return &connect.Response[qf.Void]{}, nil
}

// CreateReview adds a new submission review.
func (s *QuickFeedService) CreateReview(_ context.Context, in *connect.Request[qf.ReviewRequest]) (*connect.Response[qf.Review], error) {
	review := in.Msg.GetReview()
	if err := s.db.CreateReview(review); err != nil {
		s.logger.Errorf("CreateReview failed for review %+v: %v", in, err)
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("failed to create review"))
	}
	return connect.NewResponse(review), nil
}

// UpdateReview updates a submission review.
func (s *QuickFeedService) UpdateReview(_ context.Context, in *connect.Request[qf.ReviewRequest]) (*connect.Response[qf.Review], error) {
	review := in.Msg.GetReview()
	if err := s.db.UpdateReview(review); err != nil {
		s.logger.Errorf("UpdateReview failed for review %+v: %v", in, err)
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("failed to update review"))
	}
	return connect.NewResponse(review), nil
}

// CreateAssignmentFeedback creates a new assignment feedback.
func (s *QuickFeedService) CreateAssignmentFeedback(_ context.Context, in *connect.Request[qf.AssignmentFeedback]) (*connect.Response[qf.AssignmentFeedback], error) {
	feedback := in.Msg
	if err := s.db.CreateAssignmentFeedback(feedback); err != nil {
		s.logger.Errorf("CreateAssignmentFeedback failed for feedback %+v: %v", in, err)
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("failed to create assignment feedback"))
	}
	return connect.NewResponse(feedback), nil
}

// GetAssignmentFeedback returns assignment feedback for the given request.
func (s *QuickFeedService) GetAssignmentFeedback(_ context.Context, in *connect.Request[qf.AssignmentFeedbackRequest]) (*connect.Response[qf.AssignmentFeedback], error) {
	feedback, err := s.db.GetAssignmentFeedback(in.Msg)
	if err != nil {
		s.logger.Errorf("GetAssignmentFeedback failed for request %+v: %v", in, err)
		return nil, connect.NewError(connect.CodeNotFound, errors.New("assignment feedback not found"))
	}
	return connect.NewResponse(feedback), nil
}

// UpdateSubmissions approves and/or releases all manual reviews for student submission for the given assignment
// with the given score.
func (s *QuickFeedService) UpdateSubmissions(_ context.Context, in *connect.Request[qf.UpdateSubmissionsRequest]) (*connect.Response[qf.Void], error) {
	query := &qf.Submission{
		AssignmentID: in.Msg.GetAssignmentID(),
		Score:        in.Msg.GetScoreLimit(),
		Released:     in.Msg.GetRelease(),
	}
	err := s.db.UpdateSubmissions(query, true)
	if err != nil {
		s.logger.Errorf("UpdateSubmissions failed for request %+v: %v", in, err)
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("failed to update submissions"))
	}
	return &connect.Response[qf.Void]{}, nil
}

// GetAssignments returns a list of all assignments for the given course.
func (s *QuickFeedService) GetAssignments(_ context.Context, in *connect.Request[qf.CourseRequest]) (*connect.Response[qf.Assignments], error) {
	assignments, err := s.db.GetAssignmentsByCourse(in.Msg.GetCourseID())
	if err != nil {
		s.logger.Errorf("GetAssignments failed: %v", err)
		return nil, connect.NewError(connect.CodeNotFound, errors.New("no assignments found for course"))
	}
	return connect.NewResponse(&qf.Assignments{Assignments: assignments}), nil
}

// UpdateAssignments updates the course's assignments record in the database
// by fetching assignment information from the course's test repository.
func (s *QuickFeedService) UpdateAssignments(ctx context.Context, in *connect.Request[qf.CourseRequest]) (*connect.Response[qf.Void], error) {
	course, err := s.db.GetCourse(in.Msg.GetCourseID())
	if err != nil {
		s.logger.Errorf("UpdateAssignments failed: course %d: %v", in.Msg.GetCourseID(), err)
		return nil, connect.NewError(connect.CodeNotFound, errors.New("course not found"))
	}
	scmClient, err := s.getSCM(ctx, course.GetScmOrganizationName())
	if err != nil {
		s.logger.Errorf("UpdateAssignments failed: could not create scm client for organization %s: %v", course.GetScmOrganizationName(), err)
		return nil, scmConnectErr
	}
	assignments.UpdateFromTestsRepo(s.logger, s.runner, s.db, scmClient, course)

	clonedAssignmentsRepo, err := scmClient.Clone(ctx, &scm.CloneOptions{
		Organization: course.GetScmOrganizationName(),
		Repository:   qf.AssignmentsRepo,
		DestDir:      course.CloneDir(),
	})
	if err != nil {
		s.logger.Errorf("UpdateAssignments failed: to clone '%s' repository: %v", qf.AssignmentsRepo, err)
		return nil, connect.NewError(connect.CodeNotFound, errors.New("failed to clone assignments repository"))
	}
	s.logger.Debugf("Successfully cloned assignments repository to: %s", clonedAssignmentsRepo)

	return &connect.Response[qf.Void]{}, nil
}

// GetRepositories returns URL strings for repositories of given type for the given course.
func (s *QuickFeedService) GetRepositories(ctx context.Context, in *connect.Request[qf.CourseRequest]) (*connect.Response[qf.Repositories], error) {
	course, err := s.db.GetCourse(in.Msg.GetCourseID())
	if err != nil {
		s.logger.Errorf("GetRepositories failed: course %d not found: %v", in.Msg.GetCourseID(), err)
		return nil, connect.NewError(connect.CodeNotFound, errors.New("course not found"))
	}
	usrID := userID(ctx)
	enrol, err := s.db.GetEnrollmentByCourseAndUser(course.GetID(), usrID)
	if err != nil {
		s.logger.Error("GetRepositories failed: enrollment for user %d and course %d not found: v", usrID, course.GetID(), err)
		return nil, connect.NewError(connect.CodeNotFound, errors.New("enrollment not found"))
	}

	urls := make(map[uint32]string)
	for _, repoType := range repoTypes(enrol) {
		var id uint64
		switch repoType {
		case qf.Repository_USER:
			id = usrID
		case qf.Repository_GROUP:
			id = enrol.GetGroupID() // will be 0 if not enrolled in a group
		}
		repo, _ := s.getRepo(course, id, repoType)
		// for repo == nil: will result in an empty URL string, which will be ignored by the frontend
		urls[uint32(repoType)] = repo.GetHTMLURL()
	}
	return connect.NewResponse(&qf.Repositories{URLs: urls}), nil
}

// IsEmptyRepo ensures that group repository is empty and can be deleted.
func (s *QuickFeedService) IsEmptyRepo(ctx context.Context, in *connect.Request[qf.RepositoryRequest]) (*connect.Response[qf.Void], error) {
	course, err := s.db.GetCourse(in.Msg.GetCourseID())
	if err != nil {
		s.logger.Errorf("IsEmptyRepo failed: course %d not found: %v", in.Msg.GetCourseID(), err)
		return nil, connect.NewError(connect.CodeNotFound, errors.New("course not found"))
	}
	repos, err := s.db.GetRepositories(&qf.Repository{
		ScmOrganizationID: course.GetScmOrganizationID(),
		UserID:            in.Msg.GetUserID(),
		GroupID:           in.Msg.GetGroupID(),
	})
	if err != nil {
		s.logger.Errorf("IsEmptyRepo failed: could not get repositories for course %d, user %d, group %d: %v", in.Msg.GetCourseID(), in.Msg.GetUserID(), in.Msg.GetGroupID(), err)
		return nil, connect.NewError(connect.CodeNotFound, errors.New("repositories not found"))
	}
	if len(repos) < 1 {
		s.logger.Debugf("IsEmptyRepo: no repositories found for course %d, user %d, group %d", in.Msg.GetCourseID(), in.Msg.GetUserID(), in.Msg.GetGroupID())
		// No repository found, nothing to delete
		return &connect.Response[qf.Void]{}, nil
	}
	scmClient, err := s.getSCM(ctx, course.GetScmOrganizationName())
	if err != nil {
		s.logger.Errorf("IsEmptyRepo failed: could not create scm client for course %d: %v", in.Msg.GetCourseID(), err)
		return nil, scmConnectErr
	}

	if err := isEmpty(ctx, scmClient, repos); err != nil {
		s.logger.Errorf("IsEmptyRepo failed: %v", err)
		if ctxErr := ctxErr(ctx); ctxErr != nil {
			s.logger.Error(ctxErr)
			return nil, ctxErr
		}
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("group repository is not empty"))
	}
	return &connect.Response[qf.Void]{}, nil
}

// SubmissionStream adds the the created stream to the stream service.
// The stream may be used to send the submission results to the frontend.
// The stream is closed when the client disconnects.
func (s *QuickFeedService) SubmissionStream(ctx context.Context, _ *connect.Request[qf.Void], st *connect.ServerStream[qf.Submission]) error {
	stream := stream.NewStream(ctx, st)
	s.streams.Submission.Add(stream, userID(ctx))
	return stream.Run()
}

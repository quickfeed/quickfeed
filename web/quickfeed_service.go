package web

import (
	"context"
	"errors"
	"fmt"

	"go.uber.org/zap"

	"github.com/bufbuild/connect-go"
	"github.com/quickfeed/quickfeed/assignments"
	"github.com/quickfeed/quickfeed/ci"
	"github.com/quickfeed/quickfeed/database"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/qf/qfconnect"
	"github.com/quickfeed/quickfeed/scm"
	"github.com/quickfeed/quickfeed/web/stream"
)

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

// GetUserByCourse returns the user for the given SCM login name if enrolled in the given course.
func (s *QuickFeedService) GetUserByCourse(_ context.Context, in *connect.Request[qf.CourseUserRequest]) (*connect.Response[qf.User], error) {
	query := &qf.Course{Code: in.Msg.CourseCode, Year: in.Msg.CourseYear}
	user, err := s.db.GetUserByCourse(query, in.Msg.UserLogin)
	if err != nil {
		s.logger.Errorf("GetUserByCourse failed: %v", err)
		return nil, connect.NewError(connect.CodeNotFound, errors.New("failed to get student information"))
	}
	return connect.NewResponse(user), nil
}

// UpdateUser updates the current users's information and returns the updated user.
// This function can also promote a user to admin or demote a user.
func (s *QuickFeedService) UpdateUser(ctx context.Context, in *connect.Request[qf.User]) (*connect.Response[qf.Void], error) {
	usr, err := s.db.GetUser(userID(ctx))
	if err != nil {
		s.logger.Errorf("UpdateUser(userID=%d) failed: %v", userID(ctx), err)
		return nil, connect.NewError(connect.CodeNotFound, errors.New("unknown user"))
	}
	if _, err = s.updateUser(usr, in.Msg); err != nil {
		s.logger.Errorf("UpdateUser failed to update user %d: %v", in.Msg.GetID(), err)
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("failed to update user"))
	}
	return &connect.Response[qf.Void]{}, nil
}

// CreateCourse creates a new course.
func (s *QuickFeedService) CreateCourse(ctx context.Context, in *connect.Request[qf.Course]) (*connect.Response[qf.Course], error) {
	scmClient, err := s.getSCM(ctx, in.Msg.OrganizationName)
	if err != nil {
		s.logger.Errorf("CreateCourse failed: could not create scm client for organization %s: %v", in.Msg.OrganizationName, err)
		return nil, connect.NewError(connect.CodeNotFound, err)
	}
	// make sure that the current user is set as course creator
	in.Msg.CourseCreatorID = userID(ctx)
	course, err := s.createCourse(ctx, scmClient, in.Msg)
	if err != nil {
		s.logger.Errorf("CreateCourse failed: %v", err)
		if ctxErr := ctxErr(ctx); ctxErr != nil {
			s.logger.Error(ctxErr)
			return nil, ctxErr
		}
		if err == ErrAlreadyExists {
			return nil, connect.NewError(connect.CodeAlreadyExists, err)
		}
		if err == ErrFreePlan {
			return nil, connect.NewError(connect.CodeFailedPrecondition, err)
		}
		if ok, parsedErr := parseSCMError(err); ok {
			return nil, parsedErr
		}
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("failed to create course"))
	}
	return connect.NewResponse(course), nil
}

// UpdateCourse changes the course information details.
func (s *QuickFeedService) UpdateCourse(ctx context.Context, in *connect.Request[qf.Course]) (*connect.Response[qf.Void], error) {
	scmClient, err := s.getSCM(ctx, in.Msg.OrganizationName)
	if err != nil {
		s.logger.Errorf("UpdateCourse failed: could not create scm client for organization %s: %v", in.Msg.OrganizationName, err)
		return nil, connect.NewError(connect.CodeNotFound, err)
	}
	if err = s.updateCourse(ctx, scmClient, in.Msg); err != nil {
		s.logger.Errorf("UpdateCourse failed: %v", err)
		if ctxErr := ctxErr(ctx); ctxErr != nil {
			s.logger.Error(ctxErr)
			return nil, ctxErr
		}
		if ok, parsedErr := parseSCMError(err); ok {
			return nil, parsedErr
		}
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("failed to update course"))
	}
	return &connect.Response[qf.Void]{}, nil
}

// GetCourse returns course information for the given course.
func (s *QuickFeedService) GetCourse(_ context.Context, in *connect.Request[qf.CourseRequest]) (*connect.Response[qf.Course], error) {
	course, err := s.db.GetCourse(in.Msg.GetCourseID(), false)
	if err != nil {
		s.logger.Errorf("GetCourse failed: %v", err)
		return nil, connect.NewError(connect.CodeNotFound, errors.New("course not found"))
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
func (s *QuickFeedService) UpdateCourseVisibility(_ context.Context, in *connect.Request[qf.Enrollment]) (*connect.Response[qf.Void], error) {
	if err := s.db.UpdateEnrollment(in.Msg); err != nil {
		s.logger.Errorf("ChangeCourseVisibility failed: %v", err)
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("failed to update course visibility"))
	}
	return &connect.Response[qf.Void]{}, nil
}

// CreateEnrollment enrolls a new student for the course specified in the request.
func (s *QuickFeedService) CreateEnrollment(_ context.Context, in *connect.Request[qf.Enrollment]) (*connect.Response[qf.Void], error) {
	enrollment := &qf.Enrollment{
		UserID:   in.Msg.GetUserID(),
		CourseID: in.Msg.GetCourseID(),
		Status:   qf.Enrollment_PENDING,
	}
	if err := s.db.CreateEnrollment(enrollment); err != nil {
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
		s.logger.Errorf("UpdateEnrollments failed: could not create scm client: %v", err)
		return nil, connect.NewError(connect.CodeNotFound, err)
	}
	for _, enrollment := range in.Msg.GetEnrollments() {
		if s.isCourseCreator(enrollment.CourseID, enrollment.UserID) {
			s.logger.Errorf("UpdateEnrollments failed: user %s attempted to demote course creator", usr.GetName())
			return nil, connect.NewError(connect.CodePermissionDenied, errors.New("course creator cannot be demoted"))
		}
		if err = s.updateEnrollment(ctx, scmClient, usr.GetLogin(), enrollment); err != nil {
			s.logger.Errorf("UpdateEnrollments failed: %v", err)
			if ctxErr := ctxErr(ctx); ctxErr != nil {
				s.logger.Error(ctxErr)
				return nil, ctxErr
			}
			if ok, parsedErr := parseSCMError(err); ok {
				return nil, parsedErr
			}
			return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("failed to update enrollments"))
		}
	}
	return &connect.Response[qf.Void]{}, nil
}

// GetCoursesByUser returns all courses for the given user that match the provided enrollment status.
func (s *QuickFeedService) GetCoursesByUser(_ context.Context, in *connect.Request[qf.EnrollmentStatusRequest]) (*connect.Response[qf.Courses], error) {
	courses, err := s.db.GetCoursesByUser(in.Msg.GetUserID(), in.Msg.GetStatuses()...)
	if err != nil {
		s.logger.Errorf("GetCoursesByUser failed: user %d: %v", in.Msg.GetUserID(), err)
		return nil, connect.NewError(connect.CodeNotFound, errors.New("no courses with enrollment found"))
	}
	return connect.NewResponse(&qf.Courses{
		Courses: courses,
	}), nil
}

// GetEnrollmentsByUser returns all enrollments for the given user and enrollment status with preloaded courses and groups.
func (s *QuickFeedService) GetEnrollmentsByUser(_ context.Context, in *connect.Request[qf.EnrollmentStatusRequest]) (*connect.Response[qf.Enrollments], error) {
	enrollments, err := s.db.GetEnrollmentsByUser(in.Msg.GetUserID(), in.Msg.GetStatuses()...)
	if err != nil {
		s.logger.Errorf("GetEnrollmentsByUser failed: user %d: %v", in.Msg.GetUserID(), err)
		return nil, connect.NewError(connect.CodeNotFound, errors.New("no enrollments found for user"))
	}
	return connect.NewResponse(&qf.Enrollments{
		Enrollments: enrollments,
	}), nil
}

// GetEnrollmentsByCourse returns all enrollments for the course specified in the request.
func (s *QuickFeedService) GetEnrollmentsByCourse(_ context.Context, in *connect.Request[qf.EnrollmentRequest]) (*connect.Response[qf.Enrollments], error) {
	enrolls, err := s.getEnrollmentsByCourse(in.Msg)
	if err != nil {
		s.logger.Errorf("GetEnrollmentsByCourse failed: course %d: %v", in.Msg.GetCourseID(), err)
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("failed to get enrollments for given course"))
	}
	return connect.NewResponse(enrolls), nil
}

// GetGroup returns information about the given group.
func (s *QuickFeedService) GetGroup(_ context.Context, in *connect.Request[qf.GetGroupRequest]) (*connect.Response[qf.Group], error) {
	group, err := s.db.GetGroup(in.Msg.GetGroupID())
	if err != nil {
		s.logger.Errorf("GetGroup failed: group %d: %v", in.Msg.GetGroupID(), err)
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
	return connect.NewResponse(&qf.Groups{
		Groups: groups,
	}), nil
}

// GetGroupByUserAndCourse returns the group of the given student for a given course.
func (s *QuickFeedService) GetGroupByUserAndCourse(_ context.Context, in *connect.Request[qf.GroupRequest]) (*connect.Response[qf.Group], error) {
	group, err := s.getGroupByUserAndCourse(in.Msg)
	if err != nil {
		if err != errUserNotInGroup {
			s.logger.Errorf("GetGroupByUserAndCourse failed: %v", err)
		}
		return nil, connect.NewError(connect.CodeNotFound, errors.New("failed to get group for given user and course"))
	}
	return connect.NewResponse(group), nil
}

// CreateGroup creates a new group in the database.
// Access policy: Any User enrolled in course and specified as member of the group or a course teacher.
func (s *QuickFeedService) CreateGroup(_ context.Context, in *connect.Request[qf.Group]) (*connect.Response[qf.Group], error) {
	group, err := s.createGroup(in.Msg)
	if err != nil {
		s.logger.Errorf("CreateGroup failed: %v", err)
		if connect.CodeOf(err) != connect.CodeUnknown {
			// err was already a status error; return it to client.
			return nil, err
		}
		// err was not a status error; return a generic error to client.
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("failed to create group"))
	}
	return connect.NewResponse(group), nil
}

// UpdateGroup updates group information, and returns the updated group.
func (s *QuickFeedService) UpdateGroup(ctx context.Context, in *connect.Request[qf.Group]) (*connect.Response[qf.Group], error) {
	scmClient, err := s.getSCMForCourse(ctx, in.Msg.GetCourseID())
	if err != nil {
		s.logger.Errorf("UpdateGroup failed: could not create scm client for group %s and course %d: %v", in.Msg.GetName(), in.Msg.GetCourseID(), err)
		return nil, connect.NewError(connect.CodeNotFound, err)
	}
	err = s.updateGroup(ctx, scmClient, in.Msg)
	if err != nil {
		s.logger.Errorf("UpdateGroup failed: %v", err)
		if ctxErr := ctxErr(ctx); ctxErr != nil {
			s.logger.Error(ctxErr)
			return nil, ctxErr
		}
		if ok, parsedErr := parseSCMError(err); ok {
			return nil, parsedErr
		}
		if connect.CodeOf(err) != connect.CodeUnknown {
			// err was already a status error; return it to client.
			return nil, err
		}
		// err was not a status error; return a generic error to client.
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("failed to update group"))
	}
	group, err := s.db.GetGroup(in.Msg.ID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("failed to get group"))
	}
	return connect.NewResponse(group), nil
}

// DeleteGroup removes group record from the database.
func (s *QuickFeedService) DeleteGroup(ctx context.Context, in *connect.Request[qf.GroupRequest]) (*connect.Response[qf.Void], error) {
	scmClient, err := s.getSCMForCourse(ctx, in.Msg.GetCourseID())
	if err != nil {
		s.logger.Errorf("DeleteGroup failed: could not create scm client for group %d and course %d: %v", in.Msg.GetGroupID(), in.Msg.GetCourseID(), err)
		return nil, connect.NewError(connect.CodeNotFound, err)
	}
	if err = s.deleteGroup(ctx, scmClient, in.Msg); err != nil {
		s.logger.Errorf("DeleteGroup failed: %v", err)
		if ctxErr := ctxErr(ctx); ctxErr != nil {
			s.logger.Error(ctxErr)
			return nil, ctxErr
		}
		if ok, parsedErr := parseSCMError(errors.Unwrap(err)); ok {
			return nil, parsedErr
		}
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("failed to delete group"))
	}
	return &connect.Response[qf.Void]{}, nil
}

// GetSubmission returns a fully populated submission matching the given submission ID if it exists for the given course ID.
// Used in the frontend to fetch a full submission for a given submission ID and course ID.
func (s *QuickFeedService) GetSubmission(_ context.Context, in *connect.Request[qf.SubmissionReviewersRequest]) (*connect.Response[qf.Submission], error) {
	submission, err := s.db.GetLastSubmission(in.Msg.CourseID, &qf.Submission{ID: in.Msg.GetSubmissionID()})
	if err != nil {
		s.logger.Errorf("GetSubmission failed: %v", err)
		return nil, connect.NewError(connect.CodeNotFound, errors.New("failed to get submission"))
	}
	return connect.NewResponse(submission), nil
}

// GetSubmissions returns the submissions matching the query encoded in the action request.
func (s *QuickFeedService) GetSubmissions(ctx context.Context, in *connect.Request[qf.SubmissionRequest]) (*connect.Response[qf.Submissions], error) {
	s.logger.Debugf("GetSubmissions: %v", in)
	submissions, err := s.getSubmissions(in.Msg)
	if err != nil {
		s.logger.Errorf("GetSubmissions failed: %v", err)
		return nil, connect.NewError(connect.CodeNotFound, errors.New("no submissions found"))
	}
	// If the user is not a teacher, remove score and reviews from submissions that are not released.
	if !s.isTeacher(userID(ctx), in.Msg.CourseID) {
		submissions.Clean()
	}
	return connect.NewResponse(submissions), nil
}

// GetSubmissionsByCourse returns all the latest submissions
// for every individual or group course assignment for all course students/groups.
func (s *QuickFeedService) GetSubmissionsByCourse(_ context.Context, in *connect.Request[qf.SubmissionsForCourseRequest]) (*connect.Response[qf.CourseSubmissions], error) {
	s.logger.Debugf("GetSubmissionsByCourse: %v", in)

	courseLinks, err := s.getAllCourseSubmissions(in.Msg)
	if err != nil {
		s.logger.Errorf("GetSubmissionsByCourse failed: %v", err)
		return nil, connect.NewError(connect.CodeNotFound, errors.New("no submissions found"))
	}
	return connect.NewResponse(courseLinks), nil
}

// UpdateSubmission is called to approve the given submission or to undo approval.
func (s *QuickFeedService) UpdateSubmission(_ context.Context, in *connect.Request[qf.UpdateSubmissionRequest]) (*connect.Response[qf.Void], error) {
	if !s.isValidSubmission(in.Msg.SubmissionID) {
		s.logger.Errorf("UpdateSubmission failed: submission author has no access to the course")
		return nil, connect.NewError(connect.CodePermissionDenied, errors.New("submission author has no course access"))
	}
	err := s.updateSubmission(in.Msg.GetCourseID(), in.Msg.GetSubmissionID(), in.Msg.GetStatus(), in.Msg.GetReleased(), in.Msg.GetScore())
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
		if !s.isValidSubmission(in.Msg.GetSubmissionID()) {
			s.logger.Errorf("RebuildSubmission failed: submitter has no access to the course")
			return nil, connect.NewError(connect.CodePermissionDenied, errors.New("submitter has no course access"))
		}
		if _, err := s.rebuildSubmission(in.Msg); err != nil {
			s.logger.Errorf("RebuildSubmission failed: %v", err)
			return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("failed to rebuild submission: %w", err))
		}
	} else {
		// Submission ID == 0 ==> rebuild all for given CourseID and AssignmentID
		if err := s.rebuildSubmissions(in.Msg); err != nil {
			s.logger.Errorf("RebuildSubmissions failed: %v", err)
			return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("failed to rebuild submissions: %w", err))
		}
	}
	return &connect.Response[qf.Void]{}, nil
}

// CreateBenchmark adds a new grading benchmark for an assignment.
func (s *QuickFeedService) CreateBenchmark(_ context.Context, in *connect.Request[qf.GradingBenchmark]) (*connect.Response[qf.GradingBenchmark], error) {
	bm, err := s.createBenchmark(in.Msg)
	if err != nil {
		s.logger.Errorf("CreateBenchmark failed for %+v: %v", in, err)
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("failed to add benchmark"))
	}
	return connect.NewResponse(bm), nil
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
	if err := s.db.CreateCriterion(in.Msg); err != nil {
		s.logger.Errorf("CreateCriterion failed for %+v: %v", in, err)
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("failed to add criterion"))
	}
	return connect.NewResponse(in.Msg), nil
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
	review, err := s.createReview(in.Msg.Review)
	if err != nil {
		s.logger.Errorf("CreateReview failed for review %+v: %v", in, err)
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("failed to create review"))
	}
	return connect.NewResponse(review), nil
}

// UpdateReview updates a submission review.
func (s *QuickFeedService) UpdateReview(_ context.Context, in *connect.Request[qf.ReviewRequest]) (*connect.Response[qf.Review], error) {
	review, err := s.updateReview(in.Msg.Review)
	if err != nil {
		s.logger.Errorf("UpdateReview failed for review %+v: %v", in, err)
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("failed to update review"))
	}
	return connect.NewResponse(review), nil
}

// UpdateSubmissions approves and/or releases all manual reviews for student submission for the given assignment
// with the given score.
func (s *QuickFeedService) UpdateSubmissions(_ context.Context, in *connect.Request[qf.UpdateSubmissionsRequest]) (*connect.Response[qf.Void], error) {
	err := s.updateSubmissions(in.Msg)
	if err != nil {
		s.logger.Errorf("UpdateSubmissions failed for request %+v", in)
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("failed to update submissions"))
	}
	return &connect.Response[qf.Void]{}, nil
}

// GetReviewers returns names of all active reviewers for a student submission.
func (s *QuickFeedService) GetReviewers(_ context.Context, in *connect.Request[qf.SubmissionReviewersRequest]) (*connect.Response[qf.Reviewers], error) {
	reviewers, err := s.getReviewers(in.Msg.SubmissionID)
	if err != nil {
		s.logger.Errorf("GetReviewers failed: error fetching from database: %v", err)
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("failed to get reviewers"))
	}
	return connect.NewResponse(&qf.Reviewers{Reviewers: reviewers}), nil
}

// GetAssignments returns a list of all assignments for the given course.
func (s *QuickFeedService) GetAssignments(_ context.Context, in *connect.Request[qf.CourseRequest]) (*connect.Response[qf.Assignments], error) {
	assignments, err := s.getAssignments(in.Msg.GetCourseID())
	if err != nil {
		s.logger.Errorf("GetAssignments failed: %v", err)
		return nil, connect.NewError(connect.CodeNotFound, errors.New("no assignments found for course"))
	}
	return connect.NewResponse(assignments), nil
}

// UpdateAssignments updates the course's assignments record in the database
// by fetching assignment information from the course's test repository.
func (s *QuickFeedService) UpdateAssignments(ctx context.Context, in *connect.Request[qf.CourseRequest]) (*connect.Response[qf.Void], error) {
	course, err := s.db.GetCourse(in.Msg.GetCourseID(), false)
	if err != nil {
		s.logger.Errorf("UpdateAssignments failed: course %d: %v", in.Msg.GetCourseID(), err)
		return nil, connect.NewError(connect.CodeNotFound, errors.New("course not found"))
	}
	scmClient, err := s.getSCM(ctx, course.GetOrganizationName())
	if err != nil {
		s.logger.Errorf("UpdateAssignments failed: could not create scm client for organization %s: %v", course.GetOrganizationName(), err)
		return nil, connect.NewError(connect.CodeNotFound, err)
	}
	assignments.UpdateFromTestsRepo(s.logger, s.db, scmClient, course)

	clonedAssignmentsRepo, err := scmClient.Clone(ctx, &scm.CloneOptions{
		Organization: course.GetOrganizationName(),
		Repository:   qf.AssignmentRepo,
		DestDir:      course.CloneDir(),
	})
	if err != nil {
		s.logger.Errorf("Failed to clone '%s' repository: %v", qf.AssignmentRepo, err)
		return nil, err
	}
	s.logger.Debugf("Successfully cloned assignments repository to: %s", clonedAssignmentsRepo)

	return &connect.Response[qf.Void]{}, nil
}

// GetOrganization fetches a github organization by name.
func (s *QuickFeedService) GetOrganization(ctx context.Context, in *connect.Request[qf.OrgRequest]) (*connect.Response[qf.Organization], error) {
	usr, err := s.db.GetUser(userID(ctx))
	if err != nil {
		s.logger.Errorf("GetOrganization(userID=%d) failed: %v", userID(ctx), err)
		return nil, connect.NewError(connect.CodeNotFound, errors.New("unknown user"))
	}
	scmClient, err := s.getSCM(ctx, in.Msg.GetOrgName())
	if err != nil {
		s.logger.Errorf("GetOrganization failed: could not create scm client for organization %s: %v", in.Msg.GetOrgName(), err)
		return nil, connect.NewError(connect.CodeNotFound, err)
	}
	org, err := s.getOrganization(ctx, scmClient, in.Msg.GetOrgName(), usr.GetLogin())
	if err != nil {
		s.logger.Errorf("GetOrganization failed: %v", err)
		if ctxErr := ctxErr(ctx); ctxErr != nil {
			s.logger.Error(ctxErr)
			return nil, ctxErr
		}
		if err == scm.ErrNotMember {
			return nil, connect.NewError(connect.CodeNotFound, errors.New("organization membership not confirmed, please enable third-party access"))
		}
		if err == ErrFreePlan || err == ErrAlreadyExists || err == scm.ErrNotOwner {
			return nil, connect.NewError(connect.CodeFailedPrecondition, err)
		}
		if ok, parsedErr := parseSCMError(err); ok {
			return nil, parsedErr
		}
		return nil, connect.NewError(connect.CodeNotFound, errors.New("organization not found. Please make sure that 3rd-party access is enabled for your organization"))
	}
	return connect.NewResponse(org), nil
}

// GetRepositories returns URL strings for repositories of given type for the given course.
func (s *QuickFeedService) GetRepositories(ctx context.Context, in *connect.Request[qf.URLRequest]) (*connect.Response[qf.Repositories], error) {
	course, err := s.db.GetCourse(in.Msg.GetCourseID(), false)
	if err != nil {
		s.logger.Errorf("GetRepositories failed: course %d not found: %v", in.Msg.GetCourseID(), err)
		return nil, connect.NewError(connect.CodeNotFound, errors.New("course not found"))
	}
	usrID := userID(ctx)
	enrol, _ := s.db.GetEnrollmentByCourseAndUser(course.GetID(), usrID)

	urls := make(map[string]string)
	for _, repoType := range in.Msg.GetRepoTypes() {
		var id uint64
		switch repoType {
		case qf.Repository_USER:
			id = usrID
		case qf.Repository_GROUP:
			id = enrol.GetGroupID() // will be 0 if not enrolled in a group
		}
		repo, _ := s.getRepo(course, id, repoType)
		// for repo == nil: will result in an empty URL string, which will be ignored by the frontend
		urls[repoType.String()] = repo.GetHTMLURL()
	}
	return connect.NewResponse(&qf.Repositories{URLs: urls}), nil
}

// IsEmptyRepo ensures that group repository is empty and can be deleted.
func (s *QuickFeedService) IsEmptyRepo(ctx context.Context, in *connect.Request[qf.RepositoryRequest]) (*connect.Response[qf.Void], error) {
	scmClient, err := s.getSCMForCourse(ctx, in.Msg.GetCourseID())
	if err != nil {
		s.logger.Errorf("IsEmptyRepo failed: could not create scm client for course %d: %v", in.Msg.GetCourseID(), err)
		return nil, connect.NewError(connect.CodeNotFound, err)
	}

	if err := s.isEmptyRepo(ctx, scmClient, in.Msg); err != nil {
		s.logger.Errorf("IsEmptyRepo failed: %v", err)
		if ctxErr := ctxErr(ctx); ctxErr != nil {
			s.logger.Error(ctxErr)
			return nil, ctxErr
		}
		if ok, parsedErr := parseSCMError(err); ok {
			return nil, parsedErr
		}
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("group repository does not exist or not empty"))
	}
	return &connect.Response[qf.Void]{}, nil
}

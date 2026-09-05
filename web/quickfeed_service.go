package web

import (
	"context"
	"errors"
	"log/slog"

	"connectrpc.com/connect"
	"github.com/quickfeed/quickfeed/assignments"
	"github.com/quickfeed/quickfeed/ci"
	"github.com/quickfeed/quickfeed/database"
	"github.com/quickfeed/quickfeed/internal/qlog"
	"github.com/quickfeed/quickfeed/internal/qlog/courselog"
	"github.com/quickfeed/quickfeed/internal/qlog/label"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/qf/qfconnect"
	"github.com/quickfeed/quickfeed/scm"
	"github.com/quickfeed/quickfeed/web/auth"
	"github.com/quickfeed/quickfeed/web/stream"
)

var scmConnectErr = connect.NewError(connect.CodeNotFound, errors.New("unable to connect to the GitHub organization for the course"))

// QuickFeedService holds references to the database and
// other shared data structures.
type QuickFeedService struct {
	logger     *slog.Logger
	db         database.Database
	scmMgr     *scm.Manager
	runner     ci.Runner
	tm         *auth.TokenManager
	courseLogs *courselog.Store
	qfconnect.UnimplementedQuickFeedServiceHandler
	streams *stream.StreamServices
}

// NewQuickFeedService returns a QuickFeedService object.
func NewQuickFeedService(logger *slog.Logger, db database.Database, mgr *scm.Manager, runner ci.Runner, tm *auth.TokenManager, courseLogs *courselog.Store) *QuickFeedService {
	return &QuickFeedService{
		logger:     logger,
		db:         db,
		scmMgr:     mgr,
		runner:     runner,
		tm:         tm,
		courseLogs: courseLogs,
		streams:    stream.NewStreamServices(),
	}
}

// GetUser will return current user with active course enrollments
// to use in separating teacher and admin roles
func (s *QuickFeedService) GetUser(ctx context.Context, _ *qf.Void) (*qf.User, error) {
	userInfo, err := s.db.GetUserWithEnrollments(userID(ctx))
	if err != nil {
		qlog.FromContext(ctx).Error("failed to get user with enrollments", label.Error, err)
		return nil, connect.NewError(connect.CodeNotFound, errors.New("unknown user"))
	}
	return userInfo, nil
}

// GetUsers returns a list of all users.
// Frontend note: This method is called from AdminPage.
func (s *QuickFeedService) GetUsers(ctx context.Context, _ *qf.Void) (*qf.Users, error) {
	users, err := s.db.GetUsers()
	if err != nil {
		qlog.FromContext(ctx).Error("failed to get users", label.Error, err)
		return nil, connect.NewError(connect.CodeNotFound, errors.New("failed to get users"))
	}
	return &qf.Users{
		Users: users,
	}, nil
}

// UpdateUser updates the current users's information and returns the updated user.
// This function can also promote a user to admin or demote a user.
func (s *QuickFeedService) UpdateUser(ctx context.Context, in *qf.User) (*qf.Void, error) {
	usr, err := s.db.GetUser(userID(ctx))
	if err != nil {
		qlog.FromContext(ctx).Error("failed to get current user", label.Error, err)
		return nil, connect.NewError(connect.CodeNotFound, errors.New("unknown user"))
	}
	if err = s.editUserProfile(ctx, usr, in); err != nil {
		qlog.FromContext(ctx).Error("failed to update user profile", label.TargetUserID, in.GetID(), label.Error, err)
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("failed to update user"))
	}
	return &qf.Void{}, nil
}

// UpdateCourse changes the course information details.
func (s *QuickFeedService) UpdateCourse(ctx context.Context, in *qf.Course) (*qf.Void, error) {
	ctx, logger := qlog.WithLogger(ctx, label.Organization, in.GetScmOrganizationName())
	scmClient, err := s.getSCM(ctx, in.GetScmOrganizationName())
	if err != nil {
		logger.Error("failed to create SCM client", label.Error, err)
		return nil, scmConnectErr
	}
	// ensure the course exists
	_, err = s.db.GetCourse(in.GetID())
	if err != nil {
		logger.Error("failed to get course", label.Error, err)
		return nil, connect.NewError(connect.CodeNotFound, errors.New("failed to get course"))
	}
	// ensure the organization exists
	org, err := scmClient.GetOrganization(ctx, &scm.OrganizationOptions{ID: in.GetScmOrganizationID()})
	if err != nil {
		logger.Error("failed to get SCM organization", label.Error, err)
		if ctxErr := logCtxErr(ctx); ctxErr != nil {
			return nil, ctxErr
		}
		if scmErr := userSCMError(err); scmErr != nil {
			return nil, scmErr
		}
		return nil, connect.NewError(connect.CodeNotFound, errors.New("failed to get organization"))
	}
	in.ScmOrganizationName = org.GetScmOrganizationName()

	if err = s.db.UpdateCourse(in); err != nil {
		logger.Error("failed to update course", label.Error, err)
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("failed to update course"))
	}
	return &qf.Void{}, nil
}

// GetCourse returns course information for the given course.
func (s *QuickFeedService) GetCourse(ctx context.Context, in *qf.CourseRequest) (*qf.Course, error) {
	status := courseStatus(ctx, in.GetCourseID())
	course, err := s.db.GetCourseByStatus(in.GetCourseID(), status)
	if err != nil {
		qlog.FromContext(ctx).Error("failed to get course by enrollment status", label.Error, err)
		return nil, connect.NewError(connect.CodeNotFound, errors.New("course not found"))
	}
	if isTeacher(ctx, in.GetCourseID()) {
		course.Enrollments, err = s.getEnrollmentsWithActivity(in.GetCourseID())
		if err != nil {
			qlog.FromContext(ctx).Error("failed to get course enrollments with activity", label.Error, err)
			return nil, connect.NewError(connect.CodeNotFound, errors.New("failed to get course enrollments"))
		}
	}

	return course, nil
}

// GetCourses returns a list of all courses.
func (s *QuickFeedService) GetCourses(ctx context.Context, _ *qf.Void) (*qf.Courses, error) {
	courses, err := s.db.GetCourses()
	if err != nil {
		qlog.FromContext(ctx).Error("failed to get courses", label.Error, err)
		return nil, connect.NewError(connect.CodeNotFound, errors.New("no courses found"))
	}
	return &qf.Courses{
		Courses: courses,
	}, nil
}

// UpdateCourseVisibility allows to edit what courses are visible in the sidebar.
func (s *QuickFeedService) UpdateCourseVisibility(ctx context.Context, in *qf.Enrollment) (*qf.Void, error) {
	enrollment, err := s.db.GetEnrollmentByCourseAndUser(in.GetCourseID(), userID(ctx))
	if err != nil {
		qlog.FromContext(ctx).Error("failed to get enrollment", label.Error, err)
		return nil, connect.NewError(connect.CodeNotFound, errors.New("failed to get enrollment"))
	}

	enrollment.State = in.GetState()
	if err := s.db.UpdateEnrollment(enrollment); err != nil {
		qlog.FromContext(ctx).Error("failed to update enrollment visibility", label.Error, err)
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("failed to update course visibility"))
	}
	return &qf.Void{}, nil
}

// CreateEnrollment enrolls a new student for the course specified in the request.
func (s *QuickFeedService) CreateEnrollment(ctx context.Context, in *qf.Enrollment) (*qf.Void, error) {
	if err := s.db.CreateEnrollment(in); err != nil {
		qlog.FromContext(ctx).Error("failed to create enrollment", label.Error, err)
		if errors.Is(err, database.ErrIncompleteProfile) {
			return nil, connect.NewError(connect.CodeFailedPrecondition, err)
		}
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("failed to create enrollment"))
	}
	return &qf.Void{}, nil
}

// UpdateEnrollments changes status of all pending enrollments for the specified course to approved.
// If the request contains a single enrollment, it will be updated to the specified status.
func (s *QuickFeedService) UpdateEnrollments(ctx context.Context, in *qf.Enrollments) (*qf.Void, error) {
	usr, err := s.db.GetUser(userID(ctx))
	if err != nil {
		qlog.FromContext(ctx).Error("failed to get current user", label.Error, err)
		return nil, connect.NewError(connect.CodeNotFound, errors.New("unknown user"))
	}
	scmClient, err := s.getSCMForCourse(ctx, in.GetCourseID())
	if err != nil {
		qlog.FromContext(ctx).Error("failed to create SCM client", label.Error, err)
		return nil, scmConnectErr
	}
	for _, enrollment := range in.GetEnrollments() {
		if s.isCourseCreator(enrollment.GetCourseID(), enrollment.GetUserID()) {
			qlog.FromContext(ctx).Error("course creator demotion rejected", label.User, usr.GetLogin())
			return nil, connect.NewError(connect.CodePermissionDenied, errors.New("course creator cannot be demoted"))
		}
		if err = s.updateEnrollment(ctx, scmClient, usr.GetLogin(), enrollment); err != nil {
			qlog.FromContext(ctx).Error("failed to update enrollment", label.Error, err)
			if ctxErr := logCtxErr(ctx); ctxErr != nil {
				return nil, ctxErr
			}
			if scmErr := userSCMError(err); scmErr != nil {
				return nil, scmErr
			}
			return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("failed to update enrollments"))
		}
	}
	return &qf.Void{}, nil
}

// GetEnrollments returns all enrollments for the given course ID or user ID and enrollment status.
func (s *QuickFeedService) GetEnrollments(ctx context.Context, in *qf.EnrollmentRequest) (*qf.Enrollments, error) {
	var enrollments []*qf.Enrollment
	var err error
	statuses := in.GetStatuses()
	switch in.GetFetchMode().(type) {
	case *qf.EnrollmentRequest_UserID:
		userID := in.GetUserID()
		enrollments, err = s.db.GetEnrollmentsByUser(userID, statuses...)
		if err != nil {
			qlog.FromContext(ctx).Error("failed to get enrollments for user", label.TargetUserID, userID, label.Error, err)
			return nil, connect.NewError(connect.CodeNotFound, errors.New("no enrollments found for user"))
		}
	case *qf.EnrollmentRequest_CourseID:
		courseID := in.GetCourseID()
		if isTeacher(ctx, courseID) {
			enrollments, err = s.getEnrollmentsWithActivity(courseID)
		} else {
			enrollments, err = s.db.GetEnrollmentsByCourse(courseID, statuses...)
		}
		if err != nil {
			qlog.FromContext(ctx).Error("failed to get enrollments for course", label.Error, err)
			return nil, connect.NewError(connect.CodeNotFound, errors.New("failed to get enrollments for course"))
		}
	default:
		qlog.FromContext(ctx).Error("unknown enrollment fetch mode", "fetch_mode", in.GetFetchMode())
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("failed to get enrollments"))
	}
	return &qf.Enrollments{
		Enrollments: enrollments,
	}, nil
}

// GetGroup returns information about the given group ID, or the given user's course group if group ID is 0.
func (s *QuickFeedService) GetGroup(ctx context.Context, in *qf.GroupRequest) (*qf.Group, error) {
	var (
		group   *qf.Group
		err     error
		groupID = in.GetGroupID()
	)
	if groupID > 0 {
		group, err = s.db.GetGroup(groupID)
	} else {
		group, err = s.getGroupByUserAndCourse(in)
	}
	if err != nil {
		qlog.FromContext(ctx).Error("failed to get group", label.GroupID, groupID, label.Error, err)
		return nil, connect.NewError(connect.CodeNotFound, errors.New("failed to get group"))
	}
	return group, nil
}

// GetGroupsByCourse returns groups created for the given course.
func (s *QuickFeedService) GetGroupsByCourse(ctx context.Context, in *qf.CourseRequest) (*qf.Groups, error) {
	groups, err := s.db.GetGroupsByCourse(in.GetCourseID())
	if err != nil {
		qlog.FromContext(ctx).Error("failed to get groups for course", label.Error, err)
		return nil, connect.NewError(connect.CodeNotFound, errors.New("failed to get groups"))
	}
	return &qf.Groups{Groups: groups}, nil
}

// CreateGroup creates a new group for the given course and users.
// This function is typically called by a student when creating
// a group, which will later be (optionally) edited and approved
// by a teacher of the course using the updateGroup function below.
// Access policy: Any User enrolled in course and specified as member of the group or a course teacher.
func (s *QuickFeedService) CreateGroup(ctx context.Context, group *qf.Group) (*qf.Group, error) {
	logger := qlog.FromContext(ctx).With(label.Group, group.GetName())
	if err := s.checkGroupName(group.GetCourseID(), group.GetName()); err != nil {
		logger.Error("failed to validate group name", label.Error, err)
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	// get users of group, check consistency of group request
	if _, err := s.getGroupUsers(group); err != nil {
		logger.Error("failed to get group members", label.Error, err)
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("failed to create group"))
	}
	// create new group and update groupID in enrollment table
	if err := s.db.CreateGroup(group); err != nil {
		logger.Error("failed to create group", label.Error, err)
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("failed to create group"))
	}
	// CreateGroup assigns the group ID; keep it since GetGroup returns nil on failure.
	groupID := group.GetID()
	group, err := s.db.GetGroup(groupID)
	if err != nil {
		logger.Error("failed to reload created group", label.GroupID, groupID, label.Error, err)
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("failed to create group"))
	}
	return group, nil
}

// UpdateGroup updates group information, and returns the updated group.
func (s *QuickFeedService) UpdateGroup(ctx context.Context, in *qf.Group) (*qf.Group, error) {
	ctx, logger := qlog.WithLogger(ctx, label.Group, in.GetName(), label.GroupID, in.GetID())
	scmClient, err := s.getSCMForCourse(ctx, in.GetCourseID())
	if err != nil {
		logger.Error("failed to create SCM client", label.Error, err)
		return nil, scmConnectErr
	}
	err = s.internalUpdateGroup(ctx, scmClient, in)
	if err != nil {
		logger.Error("failed to update group", label.Error, err)
		if ctxErr := logCtxErr(ctx); ctxErr != nil {
			return nil, ctxErr
		}
		if scmErr := userSCMError(err); scmErr != nil {
			return nil, scmErr
		}
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("failed to update group"))
	}
	group, err := s.db.GetGroup(in.GetID())
	if err != nil {
		logger.Error("failed to reload updated group", label.Error, err)
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("failed to get group"))
	}
	return group, nil
}

// DeleteGroup removes group record from the database.
func (s *QuickFeedService) DeleteGroup(ctx context.Context, in *qf.GroupRequest) (*qf.Void, error) {
	ctx, logger := qlog.WithLogger(ctx, label.GroupID, in.GetGroupID())
	scmClient, err := s.getSCMForCourse(ctx, in.GetCourseID())
	if err != nil {
		logger.Error("failed to create SCM client", label.Error, err)
		return nil, scmConnectErr
	}
	if err = s.internalDeleteGroup(ctx, scmClient, in); err != nil {
		logger.Error("failed to delete group", label.Error, err)
		if ctxErr := logCtxErr(ctx); ctxErr != nil {
			return nil, ctxErr
		}
		if scmErr := userSCMError(err); scmErr != nil {
			return nil, scmErr
		}
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("failed to delete group"))
	}
	return &qf.Void{}, nil
}

// GetSubmission returns a fully populated submission matching the given submission ID if it exists for the given course ID.
// Used in the frontend to fetch a full submission for a given submission ID and course ID.
func (s *QuickFeedService) GetSubmission(ctx context.Context, in *qf.SubmissionRequest) (*qf.Submission, error) {
	submission, err := s.db.GetLastSubmission(in.GetCourseID(), &qf.Submission{ID: in.GetSubmissionID()})
	if err != nil {
		qlog.FromContext(ctx).Error("failed to get submission", label.SubmissionID, in.GetSubmissionID(), label.Error, err)
		return nil, connect.NewError(connect.CodeNotFound, errors.New("failed to get submission"))
	}
	return submission, nil
}

// GetSubmissions returns the submissions matching the query encoded in the action request.
func (s *QuickFeedService) GetSubmissions(ctx context.Context, in *qf.SubmissionRequest) (*qf.Submissions, error) {
	qlog.FromContext(ctx).Debug("fetching submissions", label.TargetUserID, in.GetUserID(), label.GroupID, in.GetGroupID())
	query := &qf.Submission{
		UserID:  in.GetUserID(),
		GroupID: in.GetGroupID(),
	}
	subs, err := s.db.GetLastSubmissions(in.GetCourseID(), query)
	if err != nil {
		qlog.FromContext(ctx).Error("failed to get last submissions", label.Error, err)
		return nil, connect.NewError(connect.CodeNotFound, errors.New("no submissions found"))
	}
	submissions := &qf.Submissions{Submissions: subs}
	id := userID(ctx)
	// If the user is not a teacher, remove score and reviews from submissions that are not released.
	if !s.isTeacher(ctx, id, in.GetCourseID()) {
		submissions.Clean(id)
	}
	return submissions, nil
}

// GetSubmissionsByCourse returns a map of submissions for the given course ID.
// The map is keyed by either the group ID or enrollment ID depending on request type.
// SubmissionRequest_GROUP returns a map keyed by group ID.
// SubmissionRequest_ALL and SubmissionRequest_USER return a map keyed by enrollment ID.
// The map values are lists of all submissions for the given group or enrollment.
func (s *QuickFeedService) GetSubmissionsByCourse(ctx context.Context, in *qf.SubmissionRequest) (*qf.CourseSubmissions, error) {
	qlog.FromContext(ctx).Debug("fetching course submissions")
	courseLinks, err := s.db.GetCourseSubmissions(in)
	if err != nil {
		qlog.FromContext(ctx).Error("failed to get course submissions", label.Error, err)
		return nil, connect.NewError(connect.CodeNotFound, errors.New("no submissions found"))
	}
	return courseLinks, nil
}

// UpdateSubmission is called to approve the given submission or to undo approval.
func (s *QuickFeedService) UpdateSubmission(ctx context.Context, in *qf.Grade) (*qf.Void, error) {
	logger := qlog.FromContext(ctx).With(label.SubmissionID, in.GetSubmissionID())
	submission, err := s.db.GetSubmission(&qf.Submission{ID: in.GetSubmissionID()})
	if err != nil {
		logger.Error("failed to get submission", label.Error, err)
		return nil, connect.NewError(connect.CodeNotFound, errors.New("failed to update submission"))
	}
	submission.SetGrade(in)
	err = s.db.UpdateSubmission(submission)
	if err != nil {
		logger.Error("failed to update submission grade", label.Error, err)
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("failed to approve submission"))
	}
	return &qf.Void{}, nil
}

// RebuildSubmissions re-runs the tests for the given assignment and course.
// A single submission is executed again if the request specifies a submission ID
// or all submissions if no submission ID is specified.
func (s *QuickFeedService) RebuildSubmissions(ctx context.Context, in *qf.RebuildRequest) (*qf.Void, error) {
	if in.GetSubmissionID() > 0 {
		// Submission ID > 0 ==> rebuild single submission for given CourseID and AssignmentID
		if err := s.internalRebuildSubmission(ctx, in); err != nil {
			qlog.FromContext(ctx).Error("failed to rebuild submission", label.Error, err)
			return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("failed to rebuild submission"))
		}
	} else {
		// Submission ID == 0 ==> rebuild all for given CourseID and AssignmentID
		if err := s.internalRebuildAllSubmissions(ctx, in); err != nil {
			qlog.FromContext(ctx).Error("failed to rebuild all submissions for assignment", label.Error, err)
			return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("failed to rebuild submissions"))
		}
	}
	return &qf.Void{}, nil
}

// CreateReview adds a new submission review.
func (s *QuickFeedService) CreateReview(ctx context.Context, in *qf.ReviewRequest) (*qf.Review, error) {
	review := in.GetReview()
	if err := s.db.CreateReview(review); err != nil {
		qlog.FromContext(ctx).Error("failed to create review", label.SubmissionID, review.GetSubmissionID(), label.Error, err)
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("failed to create review"))
	}
	return review, nil
}

// UpdateReview updates a submission review.
func (s *QuickFeedService) UpdateReview(ctx context.Context, in *qf.ReviewRequest) (*qf.Review, error) {
	review := in.GetReview()
	if err := s.db.UpdateReview(review); err != nil {
		qlog.FromContext(ctx).Error("failed to update review", label.SubmissionID, review.GetSubmissionID(), label.Error, err)
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("failed to update review"))
	}
	return review, nil
}

// CreateAssignmentFeedback creates a new assignment feedback.
func (s *QuickFeedService) CreateAssignmentFeedback(ctx context.Context, feedback *qf.AssignmentFeedback) (*qf.Void, error) {
	if err := s.db.CreateAssignmentFeedback(feedback, userID(ctx)); err != nil {
		qlog.FromContext(ctx).Error("failed to create assignment feedback", label.AssignmentID, feedback.GetAssignmentID(), label.Error, err)
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("failed to create assignment feedback"))
	}
	return &qf.Void{}, nil
}

// GetAssignmentFeedback returns assignment feedback for the given request.
func (s *QuickFeedService) GetAssignmentFeedback(ctx context.Context, in *qf.CourseRequest) (*qf.AssignmentFeedbacks, error) {
	feedback, err := s.db.GetAssignmentFeedback(in)
	if err != nil {
		qlog.FromContext(ctx).Error("failed to get assignment feedback", label.Error, err)
		return nil, connect.NewError(connect.CodeNotFound, errors.New("assignment feedback not found"))
	}
	if len(feedback.GetFeedbacks()) == 0 {
		return nil, connect.NewError(connect.CodeNotFound, errors.New("assignment feedback not found"))
	}
	return feedback, nil
}

// GetAssignments returns a list of all assignments for the given course.
func (s *QuickFeedService) GetAssignments(ctx context.Context, in *qf.CourseRequest) (*qf.Assignments, error) {
	assignments, err := s.db.GetAssignmentsByCourse(in.GetCourseID())
	if err != nil {
		qlog.FromContext(ctx).Error("failed to get assignments for course", label.Error, err)
		return nil, connect.NewError(connect.CodeNotFound, errors.New("no assignments found for course"))
	}
	return &qf.Assignments{Assignments: assignments}, nil
}

// UpdateAssignments updates the course's assignments record in the database
// by fetching assignment information from the course's test repository.
// The response reports the number of content issues found; details are written
// to the course log so that the teaching staff can fix them.
func (s *QuickFeedService) UpdateAssignments(ctx context.Context, in *qf.CourseRequest) (*qf.RepositoryIssues, error) {
	course, err := s.db.GetCourse(in.GetCourseID())
	if err != nil {
		qlog.FromContext(ctx).Error("failed to get course", label.Error, err)
		return nil, connect.NewError(connect.CodeNotFound, errors.New("course not found"))
	}
	// Scope the remainder of the method to the course; the update below uses a
	// tests-repository scope on top of this.
	// The course ID comes from the request logger; see enrichRequestLogger.
	ctx, logger := qlog.WithCourseLog(ctx, course)
	scmClient, err := s.getSCM(ctx, course.GetScmOrganizationName())
	if err != nil {
		logger.Error("failed to create SCM client", label.Error, err)
		return nil, scmConnectErr
	}
	testsCtx := qlog.With(ctx, label.Repository, qf.TestsRepo, label.RepositoryType, qf.Repository_TESTS.String())
	issueCount, err := assignments.UpdateFromTestsRepo(testsCtx, s.runner, s.db, scmClient, course)
	if err != nil {
		qlog.FromContext(testsCtx).Error("failed to update assignments from tests repository", label.Error, err)
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to update assignments from tests repository"))
	}
	return &qf.RepositoryIssues{Count: uint32(issueCount)}, nil
}

// GetRepositories returns URL strings for repositories of given type for the given course.
func (s *QuickFeedService) GetRepositories(ctx context.Context, in *qf.CourseRequest) (*qf.Repositories, error) {
	course, err := s.db.GetCourse(in.GetCourseID())
	if err != nil {
		qlog.FromContext(ctx).Error("failed to get course", label.Error, err)
		return nil, connect.NewError(connect.CodeNotFound, errors.New("course not found"))
	}
	usrID := userID(ctx)
	enrol, err := s.db.GetEnrollmentByCourseAndUser(course.GetID(), usrID)
	if err != nil {
		qlog.FromContext(ctx).Error("failed to get enrollment", label.Error, err)
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
	return &qf.Repositories{URLs: urls}, nil
}

// IsEmptyRepo ensures that group repository is empty and can be deleted.
func (s *QuickFeedService) IsEmptyRepo(ctx context.Context, in *qf.RepositoryRequest) (*qf.Void, error) {
	ctx, logger := qlog.WithLogger(ctx, label.TargetUserID, in.GetUserID(), label.GroupID, in.GetGroupID())
	course, err := s.db.GetCourse(in.GetCourseID())
	if err != nil {
		logger.Error("failed to get course", label.Error, err)
		return nil, connect.NewError(connect.CodeNotFound, errors.New("course not found"))
	}
	repos, err := s.db.GetRepositories(&qf.Repository{
		ScmOrganizationID: course.GetScmOrganizationID(),
		UserID:            in.GetUserID(),
		GroupID:           in.GetGroupID(),
	})
	if err != nil {
		logger.Error("failed to get repositories", label.Error, err)
		return nil, connect.NewError(connect.CodeNotFound, errors.New("repositories not found"))
	}
	if len(repos) < 1 {
		logger.Debug("no repositories found; nothing to delete")
		// No repository found, nothing to delete
		return &qf.Void{}, nil
	}
	scmClient, err := s.getSCM(ctx, course.GetScmOrganizationName())
	if err != nil {
		logger.Error("failed to create SCM client", label.Error, err)
		return nil, scmConnectErr
	}

	if err := CommitsAhead(ctx, scmClient, repos); err != nil {
		logger.Error("failed to verify that repositories are empty", label.Error, err)
		if ctxErr := logCtxErr(ctx); ctxErr != nil {
			return nil, ctxErr
		}
		return nil, connect.NewError(connect.CodeFailedPrecondition, err)
	}
	return &qf.Void{}, nil
}

// SubmissionStream adds the the created stream to the stream service.
// The stream may be used to send the submission results to the frontend.
// The stream is closed when the client disconnects.
func (s *QuickFeedService) SubmissionStream(ctx context.Context, _ *qf.Void, st *connect.ServerStream[qf.Submission]) error {
	stream := stream.NewStream(ctx, st)
	s.streams.Submission.Add(stream, userID(ctx))
	return stream.Run()
}

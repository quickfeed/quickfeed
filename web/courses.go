package web

import (
	"context"
	"errors"
	"fmt"

	"github.com/quickfeed/quickfeed/internal/qlog"
	"github.com/quickfeed/quickfeed/internal/qlog/label"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/scm"
	"gorm.io/gorm"
)

// updateEnrollment changes the status of the given course enrollment.
func (s *QuickFeedService) updateEnrollment(ctx context.Context, sc scm.SCM, curUser string, request *qf.Enrollment) error {
	enrollment, err := s.db.GetEnrollmentByCourseAndUser(request.GetCourseID(), request.GetUserID())
	if err != nil {
		return err
	}
	// Scope the enrollment change once; the helpers called below log the same
	// course and user, and therefore do not repeat these attributes.
	ctx, logger := qlog.WithLogger(ctx,
		label.CourseCode, enrollment.GetCourse().GetCode(),
		label.TargetUser, enrollment.GetUser().GetLogin(),
	)
	// log changes to teacher status
	if enrollment.IsTeacher() || request.IsTeacher() {
		logger.Debug("changing enrollment status", label.User, curUser, "old_status", enrollment.GetStatus(), "new_status", request.GetStatus())
	}

	// check and update user SCM info before updating enrollment status
	if err := s.updateUserFromSCM(ctx, sc, enrollment.GetUser()); err != nil {
		// A user that has deleted their SCM account can no longer be looked up;
		// their enrollment must still be removable.
		if !(errors.Is(err, scm.ErrNotFound) && request.IsNone()) {
			return fmt.Errorf("updating SCM info for user %d: %w", enrollment.GetUserID(), err)
		}
		logger.Debug("SCM user not found; removing enrollment anyway", label.Error, err)
	}
	switch {
	case (enrollment.IsPending() || enrollment.IsStudent()) && request.IsNone(): // pending or student -> none
		return s.rejectEnrollment(ctx, sc, enrollment)
	case enrollment.IsPending() && request.IsStudent(): // pending -> student
		return s.enrollStudent(ctx, sc, enrollment)
	case enrollment.IsStudent() && request.IsTeacher(): // student -> teacher
		return s.enrollTeacher(ctx, sc, enrollment)
	case enrollment.IsTeacher() && request.IsStudent(): // teacher -> student
		return s.revokeTeacherStatus(ctx, sc, enrollment)
	}
	return fmt.Errorf("unknown enrollment status change from %s to %s", enrollment.GetStatus(), request.GetStatus())
}

// rejectEnrollment rejects a student enrollment, if a student repo exists for the given course, removes it from the SCM and database.
// The GitHub repository and organization membership are removed before the database
// records, so that an interrupted reject leaves the database still referencing the
// repository, allowing the operation to be retried rather than orphaning the repo.
func (s *QuickFeedService) rejectEnrollment(ctx context.Context, sc scm.SCM, enrolled *qf.Enrollment) error {
	// course and user are both preloaded, no need to query the database
	course, user := enrolled.GetCourse(), enrolled.GetUser()
	repo, err := s.getRepo(course, user.GetID(), qf.Repository_USER)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("getting %s repository for %q: %w", course.GetCode(), user.GetLogin(), err)
	}
	if repo == nil {
		qlog.FromContext(ctx).Debug("course repository not found", label.Error, err)
		// no repository to delete; only remove the enrollment from the database
		return s.db.RejectEnrollment(user.GetID(), course.GetID())
	}

	// when deleting a user, remove github repository and organization membership as well
	opt := &scm.RejectEnrollmentOptions{
		User:           user.GetLogin(),
		OrganizationID: repo.GetScmOrganizationID(),
		RepositoryID:   repo.GetScmRepositoryID(),
	}
	if err := sc.RejectEnrollment(ctx, opt); err != nil {
		if !errors.Is(err, scm.ErrNotFound) {
			return fmt.Errorf("removing %s from %q: %w", user.GetLogin(), course.GetCode(), err)
		}
		// repository or membership already removed on GitHub, e.g., by a previously interrupted reject
		qlog.FromContext(ctx).Debug("user already removed from SCM organization", label.Error, err)
	}
	if err = s.db.DeleteRepository(repo.GetScmRepositoryID()); err != nil {
		qlog.FromContext(ctx).Debug("failed to delete course repository from database", label.Error, err)
		// continue with other delete operations
	}
	return s.db.RejectEnrollment(user.GetID(), course.GetID())
}

// enrollStudent enrolls the given user as a student into the given course.
func (s *QuickFeedService) enrollStudent(ctx context.Context, sc scm.SCM, query *qf.Enrollment) error {
	// course and user are both preloaded, no need to query the database
	course, user := query.GetCourse(), query.GetUser()

	// check whether user repo already exists,
	// which could happen if accepting a previously rejected student
	repo, err := s.getRepo(course, user.GetID(), qf.Repository_USER)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("getting %s repository for %q: %w", course.GetCode(), user.GetLogin(), err)
	}
	// Use enrollment with full updated info to ensure that gorm Select.Updates works correctly.
	query.Status = qf.Enrollment_STUDENT
	logger := qlog.FromContext(ctx)
	logger.Debug("enrolling student", "has_repository", repo != nil)
	if repo != nil {
		// repo already exist, update enrollment in database
		return s.db.UpdateEnrollment(query)
	}

	// Exchange refresh token for access token and update the user's refresh token
	accessToken, err := s.scmMgr.ExchangeAndUpdateToken(user)
	if err != nil {
		return err
	}
	// Ensure that the user's refresh token is updated after enrollment.
	if err := s.db.UpdateUser(user); err != nil {
		// Continue with enrollment; token can be manually refreshed later
		logger.Error("failed to update refresh token", label.Error, err)
	}

	// create user scmRepo and add user to course organization as a member
	// Pass the access token so that UpdateEnrollment can accept the org invitation,
	// which grants immediate access to repos the user is added to as a collaborator.
	opt := &scm.UpdateEnrollmentOptions{
		Organization: course.GetScmOrganizationName(),
		User:         user.GetLogin(),
		Status:       qf.Enrollment_STUDENT,
		AccessToken:  accessToken,
	}
	scmRepo, err := sc.UpdateEnrollment(ctx, opt)
	if err != nil {
		return fmt.Errorf("updating %s repository membership for %q: %w", course.GetCode(), user.GetLogin(), err)
	}
	logger.Debug("student enrollment repository updated")

	// add student repo to database if SCM interaction above was successful
	userRepo := qf.Repository{
		ScmRepositoryID:   scmRepo.ID,
		ScmOrganizationID: course.GetScmOrganizationID(),
		UserID:            user.GetID(),
		HTMLURL:           scmRepo.HTMLURL,
		RepoType:          qf.Repository_USER,
	}
	if err := s.db.CreateRepository(&userRepo); err != nil {
		return fmt.Errorf("creating %s repository for %q: %w", course.GetCode(), user.GetLogin(), err)
	}

	return s.db.UpdateEnrollment(query)
}

// enrollTeacher promotes the given user to teacher of the given course
func (s *QuickFeedService) enrollTeacher(ctx context.Context, sc scm.SCM, query *qf.Enrollment) error {
	// course and user are both preloaded, no need to query the database
	course, user := query.GetCourse(), query.GetUser()
	query.Status = qf.Enrollment_TEACHER
	// make owner, remove from students, add to teachers
	if _, err := sc.UpdateEnrollment(ctx, &scm.UpdateEnrollmentOptions{
		Organization: course.GetScmOrganizationName(),
		User:         user.GetLogin(),
		Status:       qf.Enrollment_TEACHER,
	}); err != nil {
		return fmt.Errorf("updating %s repository membership for teacher %q: %w", course.GetCode(), user.GetLogin(), err)
	}
	return s.db.UpdateEnrollment(query)
}

func (s *QuickFeedService) revokeTeacherStatus(ctx context.Context, sc scm.SCM, query *qf.Enrollment) error {
	// course and user are both preloaded, no need to query the database
	course, user := query.GetCourse(), query.GetUser()
	err := sc.DemoteTeacherToStudent(ctx, &scm.UpdateEnrollmentOptions{
		Organization: course.GetScmOrganizationName(),
		User:         user.GetLogin(),
		Status:       qf.Enrollment_STUDENT,
	})
	if err != nil {
		// log error, but continue to update enrollment; we can manually revoke teacher access later
		qlog.FromContext(ctx).Error("failed to revoke teacher status", label.Error, err)
	}
	query.Status = qf.Enrollment_STUDENT
	return s.db.UpdateEnrollment(query)
}

// returns all enrollments for the course ID with last activity date and number of approved assignments
func (s *QuickFeedService) getEnrollmentsWithActivity(courseID uint64) ([]*qf.Enrollment, error) {
	submissions, err := s.db.GetCourseSubmissions(
		&qf.SubmissionRequest{
			CourseID: courseID,
			FetchMode: &qf.SubmissionRequest_Type{
				Type: qf.SubmissionRequest_ALL,
			},
		})
	if err != nil {
		return nil, err
	}
	// fetch course record with all assignments and active enrollments
	course, err := s.db.GetCourseByStatus(courseID, qf.Enrollment_TEACHER)
	if err != nil {
		return nil, err
	}
	for _, enrollment := range course.GetEnrollments() {
		enrollment.UpdateTotalApproved(submissions.For(enrollment.GetID()))
	}
	return course.GetEnrollments(), nil
}

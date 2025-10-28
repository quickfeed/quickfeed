package web

import (
	"context"
	"errors"
	"fmt"

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
	// log changes to teacher status
	if enrollment.IsTeacher() || request.IsTeacher() {
		s.logger.Debugf("User %s attempting to change enrollment status of user %d from %s to %s", curUser, enrollment.GetUserID(), enrollment.GetStatus(), request.GetStatus())
	}

	// check that the user login is valid
	if err := s.updateGitHubInfo(ctx, sc, enrollment.GetUser()); err != nil {
		return fmt.Errorf("failed to update SCM info for user %d: %w", enrollment.GetUserID(), err)
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
func (s *QuickFeedService) rejectEnrollment(ctx context.Context, sc scm.SCM, enrolled *qf.Enrollment) error {
	// course and user are both preloaded, no need to query the database
	course, user := enrolled.GetCourse(), enrolled.GetUser()
	if err := s.db.RejectEnrollment(user.GetID(), course.GetID()); err != nil {
		s.logger.Debugf("Failed to delete %s enrollment for %q from database: %v", course.GetCode(), user.GetLogin(), err)
		// continue with other delete operations
	}
	repo, err := s.getRepo(course, user.GetID(), qf.Repository_USER)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("failed to get %s repository for %q: %w", course.GetCode(), user.GetLogin(), err)
	}
	if repo == nil {
		s.logger.Debugf("No %s repository found for %q: %v", course.GetCode(), user.GetLogin(), err)
		// cannot continue without repository information
		return nil
	}
	if err = s.db.DeleteRepository(repo.GetScmRepositoryID()); err != nil {
		s.logger.Debugf("Failed to delete %s repository for %q from database: %v", course.GetCode(), user.GetLogin(), err)
		// continue with other delete operations
	}
	// when deleting a user, remove github repository and organization membership as well
	opt := &scm.RejectEnrollmentOptions{
		User:           user.GetLogin(),
		OrganizationID: repo.GetScmOrganizationID(),
		RepositoryID:   repo.GetScmRepositoryID(),
	}
	if err := sc.RejectEnrollment(ctx, opt); err != nil {
		s.logger.Debugf("rejectEnrollment: failed to remove %s from %q (expected behavior): %v", user.GetLogin(), course.GetCode(), err)
	}
	return nil
}

// enrollStudent enrolls the given user as a student into the given course.
func (s *QuickFeedService) enrollStudent(ctx context.Context, sc scm.SCM, query *qf.Enrollment) error {
	// course and user are both preloaded, no need to query the database
	course, user := query.GetCourse(), query.GetUser()

	// check whether user repo already exists,
	// which could happen if accepting a previously rejected student
	repo, err := s.getRepo(course, user.GetID(), qf.Repository_USER)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("failed to get %s repository for %q: %w", course.GetCode(), user.GetLogin(), err)
	}
	// Use enrollment with full updated info to ensure that gorm Select.Updates works correctly.
	query.Status = qf.Enrollment_STUDENT
	s.logger.Debugf("Enrolling student %q in %s; has database repo: %t", user.GetLogin(), course.GetCode(), repo != nil)
	if repo != nil {
		// repo already exist, update enrollment in database
		return s.db.UpdateEnrollment(query)
	}
	// create user scmRepo and add user to course organization as a member
	scmRepo, err := sc.UpdateEnrollment(ctx, &scm.UpdateEnrollmentOptions{
		Organization: course.GetScmOrganizationName(),
		User:         user.GetLogin(),
		Status:       qf.Enrollment_STUDENT,
	})
	if err != nil {
		return fmt.Errorf("failed to update %s repository membership for %q: %w", course.GetCode(), user.GetLogin(), err)
	}
	s.logger.Debugf("Enrolling student %q in %s; repo update done", user.GetLogin(), course.GetCode())

	// add student repo to database if SCM interaction above was successful
	userRepo := qf.Repository{
		ScmOrganizationID: course.GetScmOrganizationID(),
		ScmRepositoryID:   scmRepo.ID,
		UserID:            user.GetID(),
		HTMLURL:           scmRepo.HTMLURL,
		RepoType:          qf.Repository_USER,
	}
	if err := s.db.CreateRepository(&userRepo); err != nil {
		return fmt.Errorf("failed to create %s repository for %q: %w", course.GetCode(), user.GetLogin(), err)
	}

	if err := s.acceptRepositoryInvites(ctx, sc, user, course.GetScmOrganizationName()); err != nil {
		// log error, but continue with enrollment; we can manually accept invitations later
		s.logger.Errorf("Failed to accept %s repository invites for %q: %v", course.GetCode(), user.GetLogin(), err)
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
		return fmt.Errorf("failed to update %s repository membership for teacher %q: %w", course.GetCode(), user.GetLogin(), err)
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
		s.logger.Errorf("Failed to revoke %s teacher status for %q: %v", course.GetCode(), user.GetLogin(), err)
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

// acceptRepositoryInvites tries to accept repository invitations for the given course on behalf of the given user.
func (s *QuickFeedService) acceptRepositoryInvites(ctx context.Context, scmApp scm.SCM, user *qf.User, organizationName string) error {
	user, err := s.db.GetUser(user.GetID())
	if err != nil {
		return fmt.Errorf("failed to get user %d: %w", user.GetID(), err)
	}
	newRefreshToken, err := scmApp.AcceptInvitations(ctx, &scm.InvitationOptions{
		Login:        user.GetLogin(),
		Owner:        organizationName,
		RefreshToken: user.GetRefreshToken(),
	})
	if err != nil {
		return fmt.Errorf("failed to accept invites for %s: %w", user.GetLogin(), err)
	}
	// Save the user's new refresh token in the database.
	user.RefreshToken = newRefreshToken
	return s.db.UpdateUser(user)
}

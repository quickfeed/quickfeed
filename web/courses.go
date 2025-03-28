package web

import (
	"context"
	"errors"
	"fmt"
	"sort"

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

// getSubmissions returns all the latests submissions for a user of the given course.
func (s *QuickFeedService) getSubmissions(request *qf.SubmissionRequest) (*qf.Submissions, error) {
	// only one of user ID and group ID will be set; enforced by IsValid on qf.SubmissionRequest
	query := &qf.Submission{
		UserID:  request.GetUserID(),
		GroupID: request.GetGroupID(),
	}
	submissions, err := s.db.GetLastSubmissions(request.GetCourseID(), query)
	if err != nil {
		return nil, err
	}
	return &qf.Submissions{Submissions: submissions}, nil
}

// getAllCourseSubmissions returns all individual lab submissions by students enrolled in the specified course.
func (s *QuickFeedService) getAllCourseSubmissions(request *qf.SubmissionRequest) (*qf.CourseSubmissions, error) {
	submissions, err := s.db.GetCourseSubmissions(request.GetCourseID(), request.GetType())
	if err != nil {
		return nil, err
	}
	// fetch course record with all assignments and active enrollments
	course, err := s.db.GetCourseByStatus(request.GetCourseID(), qf.Enrollment_TEACHER)
	if err != nil {
		return nil, err
	}

	var submissionsMap map[uint64]*qf.Submissions
	switch request.GetType() {
	case qf.SubmissionRequest_GROUP:
		submissionsMap = makeGroupResults(course, submissions)
	case qf.SubmissionRequest_USER:
		submissionsMap = makeUserResults(course, submissions)
	case qf.SubmissionRequest_ALL:
		submissionsMap = makeAllResults(course, submissions)
	}
	return &qf.CourseSubmissions{Submissions: submissionsMap}, nil
}

// makeGroupResults returns a map of group ID to Submissions
// for all course groups and all group assignments.
func makeGroupResults(course *qf.Course, submissions []*qf.Submission) map[uint64]*qf.Submissions {
	submissionsMap := make(map[uint64]*qf.Submissions)
	skipGroup := map[uint64]bool{0: true} // skip group ID 0 (no group)
	om := newOrderMap(course.GetAssignments())
	for _, enrollment := range course.GetEnrollments() {
		if skipGroup[enrollment.GetGroupID()] {
			continue // include group enrollment only once
		}
		skipGroup[enrollment.GetGroupID()] = true
		// Note: we (intentionally) use enrollment.GroupID as the key to the map here.
		// This is primarily a convenience for the frontend, which can then use
		// the group ID as the key to the map.
		submissionsMap[enrollment.GetGroupID()] = &qf.Submissions{
			Submissions: choose(submissions, om, func(submission *qf.Submission) bool {
				// include group submissions for this enrollment
				return submission.ByGroup(enrollment.GetGroupID())
			}),
		}
	}
	return submissionsMap
}

// makeUserResults returns a map of enrollment ID to Submissions
// for all course enrollments (students) and all individual assignments.
func makeUserResults(course *qf.Course, submission []*qf.Submission) map[uint64]*qf.Submissions {
	submissionsMap := make(map[uint64]*qf.Submissions)
	om := newOrderMap(course.GetAssignments())
	for _, enrollment := range course.GetEnrollments() {
		submissionsMap[enrollment.GetID()] = &qf.Submissions{
			Submissions: choose(submission, om, func(submission *qf.Submission) bool {
				// include individual submissions for this enrollment
				return submission.ByUser(enrollment.GetUserID())
			}),
		}
	}
	return submissionsMap
}

// makeAllResults returns a map of enrollment ID to Submissions
// for all course enrollments (students and groups) and all individual and group assignments.
func makeAllResults(course *qf.Course, submissions []*qf.Submission) map[uint64]*qf.Submissions {
	submissionsMap := make(map[uint64]*qf.Submissions)
	om := newOrderMap(course.GetAssignments())
	for _, enrollment := range course.GetEnrollments() {
		submissionsMap[enrollment.GetID()] = &qf.Submissions{
			Submissions: choose(submissions, om, func(submission *qf.Submission) bool {
				// include individual and group submissions for this enrollment
				return submission.ByUser(enrollment.GetUserID()) || submission.ByGroup(enrollment.GetGroupID())
			}),
		}
	}
	return submissionsMap
}

func choose(submissions []*qf.Submission, order *orderMap, include func(*qf.Submission) bool) []*qf.Submission {
	var subs []*qf.Submission
	for _, submission := range submissions {
		if include(submission) {
			subs = append(subs, submission)
		}
	}
	// sort submissions by assignment order
	sort.Slice(subs, func(i, j int) bool {
		return order.Less(subs[i].GetAssignmentID(), subs[j].GetAssignmentID())
	})
	return subs
}

// updateSubmission updates submission status or sets a submission score based on a manual review.
func (s *QuickFeedService) updateSubmission(submissionID uint64, grades []*qf.Grade, released bool, score uint32) error {
	submission, err := s.db.GetSubmission(&qf.Submission{ID: submissionID})
	if err != nil {
		return err
	}

	for _, grade := range grades {
		submission.SetGrade(grade.GetUserID(), grade.GetStatus())
	}
	submission.Released = released
	if score > 0 {
		submission.Score = score
	}
	return s.db.UpdateSubmission(submission)
}

// updateSubmissions updates status and release state of multiple submissions for the
// given course and assignment ID for all submissions with score equal or above the provided score
func (s *QuickFeedService) updateSubmissions(request *qf.UpdateSubmissionRequest) error {
	query := &qf.Submission{
		AssignmentID: request.GetAssignmentID(),
		Score:        request.GetScore(),
		Released:     request.GetRelease(),
	}

	return s.db.UpdateSubmissions(query, request.GetStatus())
}

// updateCourse updates an existing course.
func (s *QuickFeedService) updateCourse(ctx context.Context, sc scm.SCM, request *qf.Course) error {
	// ensure the course exists
	_, err := s.db.GetCourse(request.GetID())
	if err != nil {
		return err
	}
	// ensure the organization exists
	org, err := sc.GetOrganization(ctx, &scm.OrganizationOptions{ID: request.GetScmOrganizationID()})
	if err != nil {
		return err
	}
	request.ScmOrganizationName = org.GetScmOrganizationName()
	return s.db.UpdateCourse(request)
}

// returns all enrollments for the course ID with last activity date and number of approved assignments
func (s *QuickFeedService) getEnrollmentsWithActivity(courseID uint64) ([]*qf.Enrollment, error) {
	submissions, err := s.getAllCourseSubmissions(
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

type orderMap map[uint64]uint32

// newOrderMap creates a new orderMap from a list of assignments.
// The ID of each assignment is mapped to its order.
// Useful for sorting submissions by assignment order
// as the order is not stored in the submission themselves.
func newOrderMap(assignments []*qf.Assignment) *orderMap {
	om := make(orderMap)
	for _, assignment := range assignments {
		om[assignment.GetID()] = assignment.GetOrder()
	}
	return &om
}

func (om orderMap) Less(i, j uint64) bool {
	return om[i] < om[j]
}

package web

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	pb "github.com/autograde/quickfeed/ag"
	"github.com/autograde/quickfeed/scm"
	"gorm.io/gorm"
)

// getCourses returns all courses.
func (s *AutograderService) getCourses() (*pb.Courses, error) {
	courses, err := s.db.GetCourses()
	if err != nil {
		return nil, err
	}
	return &pb.Courses{Courses: courses}, nil
}

// getCoursesByUser returns all courses that match the provided enrollment status.
func (s *AutograderService) getCoursesByUser(request *pb.EnrollmentStatusRequest) (*pb.Courses, error) {
	courses, err := s.db.GetCoursesByUser(request.GetUserID(), request.Statuses...)
	if err != nil {
		return nil, err
	}
	return &pb.Courses{Courses: courses}, nil
}

// getEnrollmentsByUser returns all enrollments for the given user with preloaded
// courses and groups
func (s *AutograderService) getEnrollmentsByUser(request *pb.EnrollmentStatusRequest) (*pb.Enrollments, error) {
	enrollments, err := s.db.GetEnrollmentsByUser(request.UserID, request.Statuses...)
	if err != nil {
		return nil, err
	}
	for _, enrollment := range enrollments {
		enrollment.SetSlipDays(enrollment.Course)
	}
	return &pb.Enrollments{Enrollments: enrollments}, nil
}

// getEnrollmentsByCourse returns all enrollments for a course that match the given enrollment request.
func (s *AutograderService) getEnrollmentsByCourse(request *pb.EnrollmentRequest) (*pb.Enrollments, error) {
	enrollments, err := s.db.GetEnrollmentsByCourse(request.CourseID, request.Statuses...)
	if err != nil {
		return nil, err
	}
	if request.WithActivity {
		enrollments, err = s.getEnrollmentsWithActivity(request.CourseID)
		if err != nil {
			return nil, err
		}
	}

	// to populate response only with users who are not member of any group, we must filter the result
	if request.IgnoreGroupMembers {
		enrollmentsWithoutGroups := make([]*pb.Enrollment, 0)
		for _, enrollment := range enrollments {
			if enrollment.GroupID == 0 {
				enrollmentsWithoutGroups = append(enrollmentsWithoutGroups, enrollment)
			}
		}
		enrollments = enrollmentsWithoutGroups
	}

	for _, enrollment := range enrollments {
		enrollment.SetSlipDays(enrollment.Course)
	}
	return &pb.Enrollments{Enrollments: enrollments}, nil
}

// createEnrollment creates a pending enrollment for the given user and course.
func (s *AutograderService) createEnrollment(request *pb.Enrollment) error {
	enrollment := pb.Enrollment{
		UserID:   request.GetUserID(),
		CourseID: request.GetCourseID(),
		Status:   pb.Enrollment_PENDING,
	}
	return s.db.CreateEnrollment(&enrollment)
}

// updateEnrollment changes the status of the given course enrollment.
func (s *AutograderService) updateEnrollment(ctx context.Context, sc scm.SCM, curUser string, request *pb.Enrollment) error {
	enrollment, err := s.db.GetEnrollmentByCourseAndUser(request.CourseID, request.UserID)
	if err != nil {
		return err
	}
	// log changes to teacher status
	if enrollment.IsTeacher() || request.IsTeacher() {
		s.logger.Debugf("User %s attempting to change enrollment status of user %d from %s to %s", curUser, enrollment.UserID, enrollment.Status, request.Status)
	}

	switch {
	case enrollment.IsPending() && request.IsNone(): // pending -> none
		return s.rejectEnrollment(ctx, sc, enrollment)
	case enrollment.IsPending() && request.IsStudent(): // pending -> student
		return s.enrollStudent(ctx, sc, enrollment)
	case enrollment.IsStudent() && request.IsTeacher(): // student -> teacher
		return s.enrollTeacher(ctx, sc, enrollment)
	case enrollment.IsTeacher() && request.IsStudent(): // teacher -> student
		return s.revokeTeacherStatus(ctx, sc, enrollment)
	}
	return fmt.Errorf("unknown enrollment")
}

// rejectEnrollment rejects a student enrollment, if a student repo exists for the given course, removes it from the SCM and database.
func (s *AutograderService) rejectEnrollment(ctx context.Context, sc scm.SCM, enrolled *pb.Enrollment) error {
	// course and user are both preloaded, no need to query the database
	course, user := enrolled.GetCourse(), enrolled.GetUser()
	if err := s.db.RejectEnrollment(user.ID, course.ID); err != nil {
		s.logger.Debugf("Failed to delete %s enrollment for %q from database: %v", course.Code, user.Login, err)
		// continue with other delete operations
	}
	repo, err := s.getRepo(course, user.GetID(), pb.Repository_USER)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("failed to get %s repository for %q: %w", course.Code, user.Login, err)
	}
	if repo == nil {
		s.logger.Debugf("No %s repository found for %q: %v", course.Code, user.Login, err)
		// cannot continue without repository information
		return nil
	}
	if err = s.db.DeleteRepository(repo.GetRepositoryID()); err != nil {
		s.logger.Debugf("Failed to delete %s repository for %q from database: %v", course.Code, user.Login, err)
		// continue with other delete operations
	}
	// when deleting a user, remove github repository and organization membership as well
	if err := removeUserFromCourse(ctx, sc, user.GetLogin(), repo); err != nil {
		s.logger.Debugf("rejectEnrollment: failed to remove %q from %s (expected behavior): %v", course.Code, user.Login, err)
	}
	return nil
}

// enrollStudent enrolls the given user as a student into the given course.
func (s *AutograderService) enrollStudent(ctx context.Context, sc scm.SCM, enrolled *pb.Enrollment) error {
	// course and user are both preloaded, no need to query the database
	course, user := enrolled.GetCourse(), enrolled.GetUser()

	// check whether user repo already exists,
	// which could happen if accepting a previously rejected student
	repo, err := s.getRepo(course, user.GetID(), pb.Repository_USER)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("failed to get %s repository for %q: %w", course.Code, user.Login, err)
	}

	s.logger.Debugf("Enrolling student %q in %s; has database repo: %t", user.Login, course.Code, repo != nil)
	if repo != nil {
		// repo already exist, update enrollment in database
		return s.db.UpdateEnrollment(&pb.Enrollment{
			UserID:   user.ID,
			CourseID: course.ID,
			Status:   pb.Enrollment_STUDENT,
		})
	}
	// create user scmRepo, user team, and add user to students team
	scmRepo, err := updateReposAndTeams(ctx, sc, course, user.GetLogin(), pb.Enrollment_STUDENT)
	if err != nil {
		return fmt.Errorf("failed to update %s repository or team membership for %q: %w", course.Code, user.Login, err)
	}
	s.logger.Debugf("Enrolling student %q in %s; repo and team update done", user.Login, course.Code)

	// add student repo to database if SCM interaction above was successful
	userRepo := pb.Repository{
		OrganizationID: course.GetOrganizationID(),
		RepositoryID:   scmRepo.ID,
		UserID:         user.ID,
		HTMLURL:        scmRepo.WebURL,
		RepoType:       pb.Repository_USER,
	}
	if err := s.db.CreateRepository(&userRepo); err != nil {
		return fmt.Errorf("failed to create %s repository for %q: %w", course.Code, user.Login, err)
	}
	if err := s.acceptRepositoryInvites(ctx, user, course); err != nil {
		s.logger.Errorf("Failed to accept %s repository invites for %q: %v", course.Code, user.Login, err)
	}

	return s.db.UpdateEnrollment(&pb.Enrollment{
		UserID:   user.ID,
		CourseID: course.ID,
		Status:   pb.Enrollment_STUDENT,
	})
}

// enrollTeacher promotes the given user to teacher of the given course
func (s *AutograderService) enrollTeacher(ctx context.Context, sc scm.SCM, enrolled *pb.Enrollment) error {
	// course and user are both preloaded, no need to query the database
	course, user := enrolled.GetCourse(), enrolled.GetUser()

	// make owner, remove from students, add to teachers
	if _, err := updateReposAndTeams(ctx, sc, course, user.GetLogin(), pb.Enrollment_TEACHER); err != nil {
		return fmt.Errorf("failed to update %s repository or team membership for teacher %q: %w", course.Code, user.Login, err)
	}
	return s.db.UpdateEnrollment(&pb.Enrollment{
		UserID:   user.ID,
		CourseID: course.ID,
		Status:   pb.Enrollment_TEACHER,
	})
}

func (s *AutograderService) revokeTeacherStatus(ctx context.Context, sc scm.SCM, enrolled *pb.Enrollment) error {
	// course and user are both preloaded, no need to query the database
	course, user := enrolled.GetCourse(), enrolled.GetUser()
	err := revokeTeacherStatus(ctx, sc, course.GetOrganizationPath(), user.GetLogin())
	if err != nil {
		s.logger.Errorf("Failed to revoke %s teacher status for %q: %v", course.Code, user.Login, err)
	}
	return s.db.UpdateEnrollment(&pb.Enrollment{
		UserID:   user.ID,
		CourseID: course.ID,
		Status:   pb.Enrollment_STUDENT,
	})
}

// getCourse returns a course object for the given course id.
func (s *AutograderService) getCourse(courseID uint64) (*pb.Course, error) {
	return s.db.GetCourse(courseID, false)
}

// getSubmissions returns all the latests submissions for a user of the given course.
func (s *AutograderService) getSubmissions(request *pb.SubmissionRequest) (*pb.Submissions, error) {
	// only one of user ID and group ID will be set; enforced by IsValid on pb.SubmissionRequest
	query := &pb.Submission{
		UserID:  request.GetUserID(),
		GroupID: request.GetGroupID(),
	}
	submissions, err := s.db.GetLastSubmissions(request.GetCourseID(), query)
	if err != nil {
		return nil, err
	}
	return &pb.Submissions{Submissions: submissions}, nil
}

// getAllCourseSubmissions returns all individual lab submissions by students enrolled in the specified course.
func (s *AutograderService) getAllCourseSubmissions(request *pb.SubmissionsForCourseRequest) (*pb.CourseSubmissions, error) {
	assignments, err := s.db.GetAssignmentsWithSubmissions(request.GetCourseID(), request.Type, request.GetWithBuildInfo())
	if err != nil {
		return nil, err
	}
	// fetch course record with all assignments and active enrollments
	course, err := s.db.GetCourse(request.GetCourseID(), true)
	if err != nil {
		return nil, err
	}

	course.SetSlipDays()

	var enrolLinks []*pb.EnrollmentLink
	switch request.Type {
	case pb.SubmissionsForCourseRequest_GROUP:
		enrolLinks = makeGroupResults(course, assignments)
	case pb.SubmissionsForCourseRequest_INDIVIDUAL:
		enrolLinks = makeIndividualResults(course, assignments)
	default: // case pb.SubmissionsForCourseRequest_ALL:
		enrolLinks = makeAllResults(course, assignments)
	}
	return &pb.CourseSubmissions{Course: course, Links: enrolLinks}, nil
}

// makeGroupResults generates enrollment to assignment to submissions links
// for all course groups and all group assignments
func makeGroupResults(course *pb.Course, assignments []*pb.Assignment) []*pb.EnrollmentLink {
	enrolLinks := make([]*pb.EnrollmentLink, 0)
	seenGroup := make(map[uint64]bool)
	for _, enrollment := range course.Enrollments {
		if seenGroup[enrollment.GroupID] {
			continue // include group enrollment only once
		}
		seenGroup[enrollment.GroupID] = true
		enrolLinks = append(enrolLinks, &pb.EnrollmentLink{
			Enrollment: enrollment,
			Submissions: makeSubmissionLinks(assignments, func(submission *pb.Submission) bool {
				// include group submissions for this enrollment
				return submission.ByGroup(enrollment.GroupID)
			}),
		})
	}
	return enrolLinks
}

// makeIndividualResults returns enrollment links with submissions
// for individual assignments for all students in the course.
func makeIndividualResults(course *pb.Course, assignments []*pb.Assignment) []*pb.EnrollmentLink {
	enrolLinks := make([]*pb.EnrollmentLink, 0)
	for _, enrollment := range course.Enrollments {
		enrolLinks = append(enrolLinks, &pb.EnrollmentLink{
			Enrollment: enrollment,
			Submissions: makeSubmissionLinks(assignments, func(submission *pb.Submission) bool {
				// include individual submissions for this enrollment
				return submission.ByUser(enrollment.UserID)
			}),
		})
	}
	return enrolLinks
}

// makeAllResults returns enrollment links with submissions
// for both individual and group assignments for all students/groups in the course.
func makeAllResults(course *pb.Course, assignments []*pb.Assignment) []*pb.EnrollmentLink {
	enrolLinks := make([]*pb.EnrollmentLink, len(course.Enrollments))
	for i, enrollment := range course.Enrollments {
		enrolLinks[i] = &pb.EnrollmentLink{
			Enrollment: enrollment,
			Submissions: makeSubmissionLinks(assignments, func(submission *pb.Submission) bool {
				// include individual and group submissions for this enrollment
				return submission.ByUser(enrollment.UserID) || submission.ByGroup(enrollment.GroupID)
			}),
		}
	}
	return enrolLinks
}

func makeSubmissionLinks(assignments []*pb.Assignment, include func(*pb.Submission) bool) []*pb.SubmissionLink {
	subLinks := make([]*pb.SubmissionLink, len(assignments))
	for i, assignment := range assignments {
		subLinks[i] = &pb.SubmissionLink{
			Assignment: assignment.CloneWithoutSubmissions(),
		}
		for _, submission := range assignment.Submissions {
			if include(submission) {
				subLinks[i].Submission = submission
			}
		}
	}
	// sort submission links by assignment order
	sort.Slice(subLinks, func(i, j int) bool {
		return subLinks[i].Assignment.Order < subLinks[j].Assignment.Order
	})
	return subLinks
}

// updateSubmission updates submission status or sets a submission score based on a manual review.
func (s *AutograderService) updateSubmission(courseID, submissionID uint64, status pb.Submission_Status, released bool, score uint32) error {
	submission, err := s.db.GetSubmission(&pb.Submission{ID: submissionID})
	if err != nil {
		return err
	}

	// if approving previously unapproved submission
	if status == pb.Submission_APPROVED && submission.Status != pb.Submission_APPROVED {
		submission.ApprovedDate = time.Now().Format(pb.TimeLayout)
		if err := s.setLastApprovedAssignment(submission, courseID); err != nil {
			return err
		}
	}
	submission.Status = status
	submission.Released = released
	if score > 0 {
		submission.Score = score
	}
	return s.db.UpdateSubmission(submission)
}

// updateSubmissions updates status and release state of multiple submissions for the
// given course and assignment ID for all submissions with score equal or above the provided score
func (s *AutograderService) updateSubmissions(request *pb.UpdateSubmissionsRequest) error {
	if _, _, err := s.getAssignmentWithCourse(&pb.Assignment{
		CourseID: request.CourseID,
		ID:       request.AssignmentID,
	}, false); err != nil {
		return err
	}

	query := &pb.Submission{
		AssignmentID: request.AssignmentID,
		Score:        request.ScoreLimit,
		Released:     request.Release,
	}
	if request.Approve {
		query.Status = pb.Submission_APPROVED
	}

	return s.db.UpdateSubmissions(request.CourseID, query)
}

func (s *AutograderService) getReviewers(submissionID uint64) ([]*pb.User, error) {
	submission, err := s.db.GetSubmission(&pb.Submission{ID: submissionID})
	if err != nil {
		return nil, err
	}
	names := make([]*pb.User, 0)
	// TODO: make sure to preload reviews here
	for _, review := range submission.Reviews {
		// ignore possible error, will just add an empty string
		u, _ := s.db.GetUser(review.ReviewerID)
		names = append(names, u)
	}
	return names, nil
}

// updateCourse updates an existing course.
func (s *AutograderService) updateCourse(ctx context.Context, sc scm.SCM, request *pb.Course) error {
	// ensure the course exists
	_, err := s.db.GetCourse(request.ID, false)
	if err != nil {
		return err
	}
	// ensure the organization exists
	org, err := sc.GetOrganization(ctx, &scm.GetOrgOptions{ID: request.OrganizationID})
	if err != nil {
		return err
	}
	request.OrganizationPath = org.GetPath()
	return s.db.UpdateCourse(request)
}

func (s *AutograderService) changeCourseVisibility(enrollment *pb.Enrollment) error {
	return s.db.UpdateEnrollment(enrollment)
}

// returns all enrollments for the course ID with last activity date and number of approved assignments
func (s *AutograderService) getEnrollmentsWithActivity(courseID uint64) ([]*pb.Enrollment, error) {
	allEnrollmentsWithSubmissions, err := s.getAllCourseSubmissions(
		&pb.SubmissionsForCourseRequest{
			CourseID: courseID,
			Type:     pb.SubmissionsForCourseRequest_ALL,
		})
	if err != nil {
		return nil, err
	}
	var enrollmentsWithActivity []*pb.Enrollment
	for _, enrolLink := range allEnrollmentsWithSubmissions.Links {
		enrol := enrolLink.Enrollment
		var totalApproved uint64
		var submissionDate time.Time
		for _, submissionLink := range enrolLink.Submissions {
			submission := submissionLink.Submission
			if submission != nil {
				if submission.Status == pb.Submission_APPROVED {
					totalApproved++
				}
				if enrol.LastActivityDate == "" {
					submissionDate, err = submission.NewestBuildDate(submissionDate)
					if err != nil {
						return nil, err
					}
				}
			}
		}
		enrol.TotalApproved = totalApproved
		if enrol.LastActivityDate == "" && !submissionDate.IsZero() {
			enrol.LastActivityDate = submissionDate.Format("02 Jan")
		}
		enrollmentsWithActivity = append(enrollmentsWithActivity, enrol)
	}
	pending, err := s.db.GetEnrollmentsByCourse(courseID, pb.Enrollment_PENDING)
	if err != nil {
		return nil, err
	}
	// append pending users
	enrollmentsWithActivity = append(enrollmentsWithActivity, pending...)
	return enrollmentsWithActivity, nil
}

func (s *AutograderService) setLastApprovedAssignment(submission *pb.Submission, courseID uint64) error {
	query := &pb.Enrollment{
		CourseID: courseID,
	}
	if submission.GroupID > 0 {
		group, err := s.db.GetGroup(submission.GroupID)
		if err != nil {
			return err
		}
		groupMembers, err := s.getGroupUsers(group)
		if err != nil {
			return err
		}
		for _, member := range groupMembers {
			query.UserID = member.ID
			if err := s.db.UpdateEnrollment(query); err != nil {
				return err
			}
		}
		return nil
	}
	query.UserID = submission.UserID
	return s.db.UpdateEnrollment(query)
}

// acceptRepositoryInvites tries to accept repository invitations for the given course on behalf of the given user.
func (s *AutograderService) acceptRepositoryInvites(ctx context.Context, user *pb.User, course *pb.Course) error {
	user, err := s.db.GetUser(user.ID)
	if err != nil {
		return fmt.Errorf("failed to get user %d: %w", user.ID, err)
	}
	userSCM, err := s.getSCM(ctx, user, "github")
	if err != nil {
		return fmt.Errorf("failed to get SCM for user %d: %w", user.ID, err)
	}
	opts := &scm.RepositoryInvitationOptions{
		Login: user.Login,
		Owner: course.GetOrganizationPath(),
	}
	if err := userSCM.AcceptRepositoryInvites(ctx, opts); err != nil {
		return fmt.Errorf("failed to get repository invites for %s: %w", user.Login, err)
	}
	return nil
}

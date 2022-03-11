package web

import (
	"context"
	"fmt"
	"sort"
	"time"

	pb "github.com/autograde/quickfeed/ag"
	"github.com/autograde/quickfeed/scm"
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
	if enrollment.Status == pb.Enrollment_TEACHER || request.Status == pb.Enrollment_TEACHER {
		s.logger.Debugf("User %s attempting to change enrollment status of user %d from %s to %s", curUser, enrollment.UserID, enrollment.Status, request.Status)
	}

	switch request.Status {
	case pb.Enrollment_NONE:
		return s.rejectEnrollment(ctx, sc, enrollment)

	case pb.Enrollment_STUDENT:
		return s.enrollStudent(ctx, sc, enrollment)

	case pb.Enrollment_TEACHER:
		return s.enrollTeacher(ctx, sc, enrollment)
	}
	return fmt.Errorf("unknown enrollment")
}

// updateEnrollments enrolls all students with pending enrollments into course
func (s *AutograderService) updateEnrollments(ctx context.Context, sc scm.SCM, cid uint64) error {
	enrolls, err := s.db.GetEnrollmentsByCourse(cid, pb.Enrollment_PENDING)
	if err != nil {
		return err
	}
	for _, enrol := range enrolls {
		enrol.Status = pb.Enrollment_STUDENT
		if err = s.updateEnrollment(ctx, sc, "", enrol); err != nil {
			return err
		}
	}
	return nil
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
	enrolLinks := make([]*pb.EnrollmentLink, len(course.Groups))
	for i, group := range course.Groups {
		enrolLinks[i] = &pb.EnrollmentLink{}
		for _, enrollment := range course.Enrollments {
			if enrollment.GroupID > 0 && enrollment.GroupID == group.ID {
				enrolLinks[i].Enrollment = enrollment
			}
		}
		if enrolLinks[i].Enrollment == nil {
			// should not happen that a group is not enrolled
			continue
		}
		enrolLinks[i].Submissions = makeSubmissionLinks(assignments, func(submission *pb.Submission) bool {
			// include group submissions for this enrollment
			return submission.ByGroup(group.GetID())
		})
	}
	return enrolLinks
}

// makeIndividualResults returns enrollment links with submissions
// for individual assignments for all students in the course.
func makeIndividualResults(course *pb.Course, assignments []*pb.Assignment) []*pb.EnrollmentLink {
	enrolLinks := make([]*pb.EnrollmentLink, len(course.Enrollments))
	for i, enrollment := range course.Enrollments {
		enrolLinks[i] = &pb.EnrollmentLink{
			Enrollment: enrollment,
			Submissions: makeSubmissionLinks(assignments, func(submission *pb.Submission) bool {
				// include individual submissions for this enrollment
				return submission.ByUser(enrollment.UserID)
			}),
		}
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

// getRepositoryURL returns URL of a course repository of the given type.
func (s *AutograderService) getRepositoryURL(currentUser *pb.User, courseID uint64, repoType pb.Repository_Type) (string, error) {
	course, err := s.db.GetCourse(courseID, false)
	if err != nil {
		return "", err
	}
	userRepoQuery := &pb.Repository{
		OrganizationID: course.GetOrganizationID(),
		RepoType:       repoType,
	}

	switch repoType {
	case pb.Repository_USER:
		userRepoQuery.UserID = currentUser.GetID()
	case pb.Repository_GROUP:
		enrol, err := s.db.GetEnrollmentByCourseAndUser(courseID, currentUser.GetID())
		if err != nil {
			return "", err
		}
		if enrol.GetGroupID() > 0 {
			userRepoQuery.GroupID = enrol.GroupID
		}
	}

	repos, err := s.db.GetRepositories(userRepoQuery)
	if err != nil {
		return "", err
	}
	if len(repos) != 1 {
		return "", fmt.Errorf("found %d repositories for query %+v", len(repos), userRepoQuery)
	}
	return repos[0].HTMLURL, nil
}

// isEmptyRepo returns nil if all repositories for the given course and student or group are empty,
// returns an error otherwise.
func (s *AutograderService) isEmptyRepo(ctx context.Context, sc scm.SCM, request *pb.RepositoryRequest) error {
	course, err := s.db.GetCourse(request.GetCourseID(), false)
	if err != nil {
		return err
	}
	repos, err := s.db.GetRepositories(&pb.Repository{OrganizationID: course.GetOrganizationID(), UserID: request.GetUserID(), GroupID: request.GetGroupID()})
	if err != nil {
		return err
	}
	if len(repos) < 1 {
		return fmt.Errorf("no repositories found")
	}
	return isEmpty(ctx, sc, repos)
}

// rejectEnrollment rejects a student enrollment, if a student repo exists for the given course, removes it from the SCM and database.
func (s *AutograderService) rejectEnrollment(ctx context.Context, sc scm.SCM, enrolled *pb.Enrollment) error {
	// course and user are both preloaded, no need to query the database
	course, user := enrolled.GetCourse(), enrolled.GetUser()
	repos, err := s.db.GetRepositories(&pb.Repository{
		UserID:         user.GetID(),
		OrganizationID: course.GetOrganizationID(),
		RepoType:       pb.Repository_USER,
	})
	if err != nil {
		return err
	}
	for _, repo := range repos {
		// we do not care about errors here, even if the github repo does not exists,
		// log the error and go on with deleting database entries
		if err := removeUserFromCourse(ctx, sc, user.GetLogin(), repo); err != nil {
			s.logger.Debug("updateEnrollment: rejectUserFromCourse failed (expected behavior): ", err)
		}

		if err := s.db.DeleteRepositoryByRemoteID(repo.GetRepositoryID()); err != nil {
			return err
		}
	}
	return s.db.RejectEnrollment(user.ID, course.ID)
}

// enrollStudent enrolls the given user as a student into the given course.
func (s *AutograderService) enrollStudent(ctx context.Context, sc scm.SCM, enrolled *pb.Enrollment) error {
	// course and user are both preloaded, no need to query the database
	course, user := enrolled.GetCourse(), enrolled.GetUser()

	// check whether user repo already exists,
	// which could happen if accepting a previously rejected student
	userRepoQuery := &pb.Repository{
		OrganizationID: course.GetOrganizationID(),
		UserID:         user.GetID(),
		RepoType:       pb.Repository_USER,
	}
	userEnrolQuery := &pb.Enrollment{
		UserID:   user.ID,
		CourseID: course.ID,
		Status:   pb.Enrollment_STUDENT,
	}
	repos, err := s.db.GetRepositories(userRepoQuery)
	if err != nil {
		return err
	}

	if enrolled.Status == pb.Enrollment_TEACHER {
		err = revokeTeacherStatus(ctx, sc, course.GetOrganizationPath(), user.GetLogin())
		if err != nil {
			s.logger.Errorf("Failed to revoke teacher status for user %s and course %s: %v", user.Login, course.Name, err)
		}
	} else {

		s.logger.Debug("Enrolling student: ", user.GetLogin(), " have database repos: ", len(repos))
		if len(repos) > 0 {
			// repo already exist, update enrollment in database
			return s.db.UpdateEnrollment(userEnrolQuery)
		}
		// create user repo, user team, and add user to students team
		repo, err := updateReposAndTeams(ctx, sc, course, user.GetLogin(), pb.Enrollment_STUDENT)
		if err != nil {
			s.logger.Errorf("Failed to update repos or team membership for student %s: %v", user.Login, err)
			return err
		}
		s.logger.Debug("Enrolling student: ", user.GetLogin(), " repo and team update done")

		// add student repo to database if SCM interaction above was successful
		userRepo := pb.Repository{
			OrganizationID: course.GetOrganizationID(),
			RepositoryID:   repo.ID,
			UserID:         user.ID,
			HTMLURL:        repo.WebURL,
			RepoType:       pb.Repository_USER,
		}

		if err := s.db.CreateRepository(&userRepo); err != nil {
			return err
		}

		if err := s.acceptRepositoryInvites(ctx, user, course); err != nil {
			s.logger.Errorf("Failed to accept repository invites for student %s: %v", user.Login, err)
		}
	}

	return s.db.UpdateEnrollment(userEnrolQuery)
}

// enrollTeacher promotes the given user to teacher of the given course
func (s *AutograderService) enrollTeacher(ctx context.Context, sc scm.SCM, enrolled *pb.Enrollment) error {
	// course and user are both preloaded, no need to query the database
	course, user := enrolled.GetCourse(), enrolled.GetUser()

	// make owner, remove from students, add to teachers
	if _, err := updateReposAndTeams(ctx, sc, course, user.GetLogin(), pb.Enrollment_TEACHER); err != nil {
		s.logger.Errorf("Failed to update team membership for teacher %s: %v", user.Login, err)
		return err
	}
	return s.db.UpdateEnrollment(&pb.Enrollment{
		UserID:   user.ID,
		CourseID: course.ID,
		Status:   pb.Enrollment_TEACHER,
	})
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

package web

import (
	"sort"

	pb "github.com/autograde/quickfeed/ag"
)

// getSubmissions returns all the latests submissions for a user of the given course.
func (s *AutograderService) getSubmissions(request *pb.SubmissionRequest) (*pb.Submissions, error) {
	// only one of user ID and group ID will be set; enforced by IsValid on pb.SubmissionRequest
	query := &pb.Submission{
		UserID:  request.GetUserID(),
		GroupID: request.GetGroupID(),
	}
	submissions, err := s.db.GetSubmissions(request.GetCourseID(), query)
	if err != nil {
		return nil, err
	}
	for _, sbm := range submissions {
		sbm.MakeSubmissionReviews()
	}
	return &pb.Submissions{Submissions: submissions}, nil
}

// getAllCourseSubmissions returns all individual lab submissions by students enrolled in the specified course.
func (s *AutograderService) getAllCourseSubmissions(request *pb.SubmissionsForCourseRequest) (*pb.CourseSubmissions, error) {
	assignments, err := s.db.GetCourseAssignmentsWithSubmissions(request.GetCourseID(), request.Type)
	if err != nil {
		return nil, err
	}
	// fetch course record with all assignments and active enrollments
	course, err := s.db.GetCourse(request.GetCourseID(), true)
	if err != nil {
		return nil, err
	}
	course.SetSlipDays()

	for _, a := range assignments {
		for _, sbm := range a.Submissions {
			sbm.MakeSubmissionReviews()
		}
	}

	enrolLinks := make([]*pb.EnrollmentLink, 0)

	switch request.Type {
	case pb.SubmissionsForCourseRequest_GROUP:
		enrolLinks = append(enrolLinks, s.makeGroupResults(course, assignments)...)
	case pb.SubmissionsForCourseRequest_INDIVIDUAL:
		enrolLinks = append(enrolLinks, makeResults(course, assignments)...)
	default:
		enrolLinks = append(makeResults(course, assignments), s.makeGroupResults(course, assignments)...)
	}
	return &pb.CourseSubmissions{Course: course, Links: enrolLinks}, nil
}

// makeResults generates enrollment to assignment to submissions links
// for all course students and all individual assignments
func makeResults(course *pb.Course, assignments []*pb.Assignment) []*pb.EnrollmentLink {
	enrolLinks := make([]*pb.EnrollmentLink, 0)

	for _, enrol := range course.Enrollments {
		newLink := &pb.EnrollmentLink{Enrollment: enrol}
		allSubmissions := make([]*pb.SubmissionLink, 0)
		for _, a := range assignments {
			copyWithoutSubmissions := a.CloneWithoutSubmissions()
			subLink := &pb.SubmissionLink{
				Assignment: copyWithoutSubmissions,
			}

			for _, sb := range a.Submissions {
				if sb.UserID > 0 && sb.UserID == enrol.UserID {
					subLink.Submission = sb
				}
			}
			allSubmissions = append(allSubmissions, subLink)
		}

		newLink.Submissions = allSubmissions
		enrolLinks = append(enrolLinks, newLink)
	}
	return enrolLinks
}

// makeGroupResults generates enrollment to assignment to submissions links
// for all course groups and all group assignments
func (s *AutograderService) makeGroupResults(course *pb.Course, assignments []*pb.Assignment) []*pb.EnrollmentLink {
	enrolLinks := make([]*pb.EnrollmentLink, 0)
	for _, grp := range course.Groups {

		newLink := &pb.EnrollmentLink{}
		for _, enrol := range course.Enrollments {
			if enrol.GroupID > 0 && enrol.GroupID == grp.ID {
				newLink.Enrollment = enrol
			}
		}
		if newLink.Enrollment == nil {
			s.logger.Debugf("Got empty enrollment for group %+v", grp)
		}

		allSubmissions := make([]*pb.SubmissionLink, 0)
		for _, a := range assignments {
			copyWithoutSubmissions := a.CloneWithoutSubmissions()
			subLink := &pb.SubmissionLink{
				Assignment: copyWithoutSubmissions,
			}
			for _, sb := range a.Submissions {
				if sb.GroupID > 0 && sb.GroupID == grp.ID {
					subLink.Submission = sb
				}
			}
			allSubmissions = append(allSubmissions, subLink)
		}
		sortSubmissionsByAssignmentOrder(allSubmissions)
		newLink.Submissions = allSubmissions
		enrolLinks = append(enrolLinks, newLink)
	}
	return enrolLinks
}

// updateSubmission approves the given submission or undoes a previous approval.
func (s *AutograderService) updateSubmission(request *pb.UpdateSubmissionRequest) error {
	submission, err := s.db.GetSubmission(&pb.Submission{ID: request.SubmissionID})
	if err != nil {
		return err
	}
	submission.Status = request.Status
	submission.Released = request.Released
	if request.Score > 0 {
		submission.Score = request.Score
	}
	return s.db.UpdateSubmission(submission)
}

// updateSubmissions updates status and release state of multiple submissions for the
// given course and assignment ID for all submissions with score equal or above the provided score
func (s *AutograderService) updateSubmissions(request *pb.UpdateSubmissionsRequest) error {
	if _, err := s.db.GetCourse(request.CourseID, false); err != nil {
		return err
	}
	if _, err := s.db.GetAssignment(&pb.Assignment{
		CourseID: request.CourseID,
		ID:       request.AssignmentID,
	}); err != nil {
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

// updateComment creates or updates a comment.
func (s *AutograderService) updateComment(request *pb.Comment) (*pb.Comment, error) {
	return s.db.UpdateComment(request)
}

// deleteComment deletes a comment with the given ID.
func (s *AutograderService) deleteComment(commentID uint64) error {
	return s.db.DeleteComment(commentID)
}

// getReviewersBySubmission returns all users who left a review to the given submission ID.
func (s *AutograderService) getReviewersBySubmission(submissionID uint64) ([]*pb.User, error) {
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

func sortSubmissionsByAssignmentOrder(unsorted []*pb.SubmissionLink) []*pb.SubmissionLink {
	sort.Slice(unsorted, func(i, j int) bool {
		return unsorted[i].Assignment.Order < unsorted[j].Assignment.Order
	})
	return unsorted
}

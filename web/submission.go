package web

import (
	"github.com/quickfeed/quickfeed/qf"
	"google.golang.org/protobuf/proto"
)

// streamUserScopedSubmission sends a copy of the submission to each user,
// keeping only that user's grade (if present).
func (s *QuickFeedService) streamUserScopedSubmission(sub *qf.Submission) {
	if len(sub.GetUserIDs()) == 0 {
		// No users: nothing to send.
		return
	}

	if !(sub.GetReleased() || len(sub.GetReviews()) == 0) {
		// Submission is not released or has reviews, do not send to users.
		return
	}

	grades := sub.GetGrades()
	if len(grades) == 0 {
		// No grades: just send the (cloned) submission to each user without grades.
		s.streams.Submission.SendTo(sub, sub.GetUserIDs()...)
		return
	}

	// Index grades by user ID for quick lookup.
	gradeByUser := make(map[uint64]*qf.Grade, len(grades))
	for _, g := range grades {
		gradeByUser[g.GetUserID()] = g
	}

	for _, uid := range sub.GetUserIDs() {
		us := proto.Clone(sub).(*qf.Submission)
		if g, ok := gradeByUser[uid]; ok {
			us.Grades = []*qf.Grade{g}
		} else {
			us.Grades = nil
		}
		s.streams.Submission.SendTo(us, uid)
	}
}

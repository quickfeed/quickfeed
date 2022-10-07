package interceptor

import (
	"github.com/quickfeed/quickfeed/database"
	"github.com/quickfeed/quickfeed/qf"
)

// isValidSubmission returns true if submitting student has active course enrollment or
// if submitting group belongs to the given course.
func isValidSubmission(db database.Database, req requestID) bool {
	courseID := req.IDFor("course")
	submissionID := req.IDFor("submission")
	sbm, err := db.GetSubmission(&qf.Submission{ID: submissionID})
	if err != nil {
		return false
	}

	if sbm.GroupID > 0 {
		grp, err := db.GetGroup(sbm.GroupID)
		if err != nil || grp.GetCourseID() != courseID {
			return false
		}
		return true
	}

	enrol, err := db.GetEnrollmentByCourseAndUser(courseID, sbm.UserID)
	if err != nil || enrol.IsNone() || enrol.IsPending() {
		return false
	}
	return true
}

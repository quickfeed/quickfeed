package interceptor

import (
	"github.com/quickfeed/quickfeed/database"
	"github.com/quickfeed/quickfeed/qf"
)

// isValidSubmission returns true if the student or group submitting the original push event
// has an active course enrollment in the given course.
func isValidSubmission(db database.Database, req requestID) bool {
	courseID := req.IDFor("course")
	submissionID := req.IDFor("submission")
	sbm, err := db.GetSubmission(&qf.Submission{ID: submissionID})
	if err != nil {
		return false
	}

	if sbm.GetGroupID() > 0 {
		grp, err := db.GetGroup(sbm.GetGroupID())
		if err != nil || grp.GetCourseID() != courseID {
			return false
		}
		return true
	}

	enrol, err := db.GetEnrollmentByCourseAndUser(courseID, sbm.GetUserID())
	if err != nil || enrol.IsNone() || enrol.IsPending() {
		return false
	}
	return true
}

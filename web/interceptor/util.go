package interceptor

import (
	"github.com/quickfeed/quickfeed/database"
	"github.com/quickfeed/quickfeed/qf"
)

type (
	userIDProvider       interface{ GetUserID() uint64 }
	courseIDProvider     interface{ GetCourseID() uint64 }
	submissionIDProvider interface{ GetSubmissionID() uint64 }
)

func getUserID(req any) uint64 {
	if uid, ok := req.(userIDProvider); ok {
		return uid.GetUserID()
	}
	return 0
}

func getCourseID(req any) uint64 {
	if cid, ok := req.(courseIDProvider); ok {
		return cid.GetCourseID()
	}
	return 0
}

func getSubmissionID(req any) uint64 {
	if sid, ok := req.(submissionIDProvider); ok {
		return sid.GetSubmissionID()
	}
	return 0
}

// isValidSubmission returns true if the student or group submitting the original push event
// has an active course enrollment in the given course.
func isValidSubmission(db database.Database, req any) bool {
	courseID := getCourseID(req)
	submissionID := getSubmissionID(req)
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

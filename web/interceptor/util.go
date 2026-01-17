package interceptor

import (
	"github.com/quickfeed/quickfeed/database"
	"github.com/quickfeed/quickfeed/qf"
)

type (
	courseIDProvider     interface{ GetCourseID() uint64 }
	userIDProvider       interface{ GetUserID() uint64 }
	groupIDProvider      interface{ GetGroupID() uint64 }
	submissionIDProvider interface{ GetSubmissionID() uint64 }
)

func getCourseID(req any) uint64 {
	if cid, ok := req.(courseIDProvider); ok {
		return cid.GetCourseID()
	}
	return 0
}

func getUserID(req any) uint64 {
	if uid, ok := req.(userIDProvider); ok {
		return uid.GetUserID()
	}
	return 0
}

func hasGroupID(req any) bool {
	if gid, ok := req.(groupIDProvider); ok {
		return gid.GetGroupID() > 0
	}
	return false
}

func getSubmissionID(req any) uint64 {
	if sid, ok := req.(submissionIDProvider); ok {
		return sid.GetSubmissionID()
	}
	return 0
}

// isValidSubmission returns true if the submission belongs to the given course
// and the student or group has an active course enrollment in that course.
func isValidSubmission(db database.Database, req any) bool {
	courseID := getCourseID(req)
	submissionID := getSubmissionID(req)
	sbm, err := db.GetSubmission(&qf.Submission{ID: submissionID})
	if err != nil {
		return false
	}

	// Check that the submission's assignment belongs to the course
	assignment, err := db.GetAssignment(&qf.Assignment{ID: sbm.GetAssignmentID()})
	if err != nil || assignment.GetCourseID() != courseID {
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

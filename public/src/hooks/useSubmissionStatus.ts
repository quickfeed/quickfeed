import { Submission, Submission_Status } from "../../proto/qf/types_pb"
import { getStatusByUser } from "../Helpers"
import { useAppState } from "../overmind"

// statusPriority defines the order for aggregating a group submission's status.
// The first status found among the grades is used, so APPROVED is only returned
// when every member has been approved (no REJECTED, REVISION, or NONE present).
const statusPriority = [
    Submission_Status.REJECTED,
    Submission_Status.REVISION,
    Submission_Status.NONE,
    Submission_Status.APPROVED,
]

const worstGroupStatus = (submission: Submission): Submission_Status => {
    for (const status of statusPriority) {
        if (submission.Grades.some(g => g.Status === status)) {
            return status
        }
    }
    return Submission_Status.NONE
}

/**
 * Resolves the display status for a submission based on the current viewer context:
 * - Teacher with a specific enrollment selected: that member's individual grade.
 * - Teacher with a group row selected (no enrollment): worst status across all members.
 * - Student: the student's own grade.
 */
export const useSubmissionStatus = (submission: Submission): Submission_Status => {
    const state = useAppState()
    const enrollment = state.selectedEnrollment
    if (enrollment) {
        return getStatusByUser(submission, enrollment.userID)
    }
    if (state.isTeacher) {
        return worstGroupStatus(submission)
    }
    return getStatusByUser(submission, state.self.ID)
}

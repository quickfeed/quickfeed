import { Context } from "."
import { Assignment, Submission } from "../../proto/qf/types_pb"


export const storeSubmissions = ({ state }: Context, { courseID, submissions, assignments }: { courseID: string, submissions: Submission[], assignments: Assignment[] }) => {
    if (!state.submissions[courseID]) {
        state.submissions[courseID] = []
    }
    assignments?.forEach(assignment => {
        const submission = submissions.find(submission => submission.AssignmentID === assignment.ID)
        if (submission) {
            state.submissions[courseID][assignment.order - 1] = submission ? submission : new Submission()
        }
    })
}

import React from "react"
import { useAppState } from "../overmind"
import { Submission, Submission_Status } from "../../proto/qf/types_pb"
import { getStatusByUser } from "../Helpers"

type ProgressIndicatorProps = {
    courseID: string,
    submission: Submission
}

const ProgressIndicator = ({ courseID, submission }: ProgressIndicatorProps) => {
    const state = useAppState()

    const assignment = state.assignments[courseID]?.find(assignment => assignment.ID === submission.AssignmentID)
    const score = submission.score ?? 0
    const scorelimit = assignment?.scoreLimit ?? 0
    const status = getStatusByUser(submission, state.self.ID)

    let borderColor = ""
    switch (status) {
        case Submission_Status.NONE:
            borderColor = score >= scorelimit ? "border-b-success" : "border-b-primary"
            break
        case Submission_Status.APPROVED:
            borderColor = "border-b-success"
            break
        case Submission_Status.REJECTED:
            borderColor = "border-b-error"
            break
        case Submission_Status.REVISION:
            borderColor = "border-b-warning"
            break
    }

    return (
        <div className={`absolute bottom-0 left-0 border-b-4 ${borderColor} opacity-30`}
            style={{
                right: `${100 - score}%`
            }} />
    )
}

export default ProgressIndicator

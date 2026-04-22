import React from "react"
import { Assignment, Submission } from "../../proto/qf/types_pb"
import { useSubmissionStatus } from "../hooks/useSubmissionStatus"
import ProgressBar from "./ProgressBar"
import SubmissionInfo from "./submissions/SubmissionInfo"
import SubmissionScores from "./submissions/SubmissionScores"

type LabProps = {
    submission: Submission
    assignment: Assignment
}

const LabResultTable = ({ submission, assignment }: LabProps) => {
    const status = useSubmissionStatus(submission)
    if (submission && assignment) {
        return (
            <div className="pb-2">
                <div className="pb-2">
                    <ProgressBar score={submission.score} scoreLimit={assignment.scoreLimit} status={status} />
                </div>
                <SubmissionInfo submission={submission} assignment={assignment} />
                <SubmissionScores submission={submission} />
            </div>
        )
    }
    return <div className="container"> No Submission </div>
}

export default LabResultTable

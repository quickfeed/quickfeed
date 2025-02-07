import React from "react"
import { Assignment, Submission } from "../../proto/qf/types_pb"
import ProgressBar, { Progress } from "./ProgressBar"
import SubmissionInfo from "./submissions/SubmissionInfo"
import SubmissionScores from "./submissions/SubmissionScores"

type LabProps = {
    submission: Submission
    assignment: Assignment
}

const LabResultTable = ({ submission, assignment }: LabProps): JSX.Element => {
    if (submission && assignment) {
        return (
            <div className="pb-2">
                <div className="pb-2">
                    <ProgressBar key={"progress-bar"} courseID={assignment.CourseID.toString()} submission={submission} type={Progress.LAB} />
                </div>
                <SubmissionInfo submission={submission} assignment={assignment} />    
                <SubmissionScores submission={submission} />
            </div>
        )
    }
    return <div className="container"> No Submission </div>
}

export default LabResultTable

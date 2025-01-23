import React from 'react'
import { useHistory } from 'react-router'
import { Assignment, Submission } from "../../../proto/qf/types_pb"
import { assignmentStatusText, getStatusByUser, isGroupSubmission, isValidSubmissionForAssignment } from "../../Helpers"
import ProgressBar, { Progress } from "../ProgressBar"
import SubmissionTypeIcon from './SubmissionTypeIcon'

interface SubmissionRowProps {
    submission: Submission
    assignment: Assignment
    courseID: string
    selfID: bigint
}

const SubmissionRow: React.FC<SubmissionRowProps> = ({ submission, assignment, courseID, selfID }) => {
    const history = useHistory()

    const redirectTo = (submission: Submission) => {
        if (submission.groupID !== 0n) {
            history.push(`/course/${courseID}/group-lab/${submission.AssignmentID.toString()}`)
        } else {
            history.push(`/course/${courseID}/lab/${submission.AssignmentID.toString()}`)
        }
    }

    if (!isValidSubmissionForAssignment(submission, assignment)) {
        return null
    }

    return (
        <div
            key={submission.ID.toString()}
            className="row clickable mb-1 py-2 align-items-center text-left"
            onClick={() => redirectTo(submission)}
            role="button"
        >
            <div className="col-8">
                <ProgressBar courseID={courseID} submission={submission} type={Progress.LAB} />
            </div>
                <SubmissionTypeIcon solo={!isGroupSubmission(submission)} />
            <div className="col-3">
                {assignmentStatusText(assignment, submission, getStatusByUser(submission, selfID))}
            </div>
        </div>
    )
}

export default SubmissionRow

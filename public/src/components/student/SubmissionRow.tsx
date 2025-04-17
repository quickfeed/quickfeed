import React from 'react'
import { Assignment, Submission } from "../../../proto/qf/types_pb"
import { assignmentStatusText, getStatusByUser, isGroupSubmission } from "../../Helpers"
import ProgressBar, { Progress } from "../ProgressBar"
import SubmissionTypeIcon from './SubmissionTypeIcon'

interface SubmissionRowProps {
    submission: Submission
    assignment: Assignment
    courseID: string
    selfID: bigint
    redirectTo: (submission: Submission) => void
}

const SubmissionRow: React.FC<SubmissionRowProps> = ({ submission, assignment, courseID, selfID, redirectTo }) => {
    // Should hover individual submissions for group labs, to indicate to the user where they end up after clicking
    // The default behavior is to redirect to the group lab submission page
    const hoverSubmission = assignment.isGroupLab && !isGroupSubmission(submission) ? "hover-effect" : ""
    return (
        <div
            key={submission.ID.toString()}
            className={`row clickable mb-1 py-2 align-items-center text-left ${hoverSubmission}`}
            onClick={(e) => { e.stopPropagation(); redirectTo(submission) }}
            role="button"
            aria-hidden="true"
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

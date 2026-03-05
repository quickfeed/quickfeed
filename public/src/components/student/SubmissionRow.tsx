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
            className={`flex items-center gap-4 py-3 px-2 mb-2 rounded-lg cursor-pointer hover:bg-base-200 transition-colors ${hoverSubmission}`}
            onClick={(e) => { e.stopPropagation(); redirectTo(submission) }}
            role="button"
            aria-hidden="true"
        >
            <div className="flex-1 min-w-0">
                <ProgressBar courseID={courseID} submission={submission} type={Progress.LAB} />
            </div>
            <div className="flex-shrink-0 w-10 flex items-center justify-center">
                <SubmissionTypeIcon solo={!isGroupSubmission(submission)} />
            </div>
            <div className="flex-shrink-0 w-40 text-sm font-medium text-right">
                {assignmentStatusText(assignment, submission, getStatusByUser(submission, selfID))}
            </div>
        </div>
    )
}

export default SubmissionRow

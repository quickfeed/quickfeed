import React from 'react'
import { Assignment, Submission, Submission_Status } from "../../../proto/qf/types_pb"
import { assignmentStatusText, getFormattedTime, getStatusByUser, isGroupSubmission } from "../../Helpers"
import ProgressBar from "../ProgressBar"
import SubmissionTypeIcon from './SubmissionTypeIcon'

interface SubmissionRowProps {
    submission: Submission
    assignment: Assignment
    courseID: string
    selfID: bigint
    redirectTo: (submission: Submission) => void
}

const statusConfig: Record<Submission_Status, { color: string; icon: string | null }> = {
    [Submission_Status.NONE]: { color: "text-base-content/70", icon: null },
    [Submission_Status.APPROVED]: { color: "text-success", icon: "fa-check-circle" },
    [Submission_Status.REJECTED]: { color: "text-error", icon: "fa-times-circle" },
    [Submission_Status.REVISION]: { color: "text-warning", icon: "fa-exclamation-circle" },
}

const SubmissionRow: React.FC<SubmissionRowProps> = ({ submission, assignment, courseID, selfID, redirectTo }) => {
    const submissionDate = submission.BuildInfo?.SubmissionDate
        ? getFormattedTime(submission.BuildInfo.SubmissionDate)
        : null

    const status = getStatusByUser(submission, selfID)
    const { color, icon } = statusConfig[status]

    return (
        <div
            className={`flex items-center gap-4 py-2 px-2 mb-1 rounded-lg cursor-pointer`}
            onClick={(e) => { e.stopPropagation(); redirectTo(submission) }}
            role="button"
            aria-hidden="true"
        >
            <div className="flex-1 min-w-0">
                <ProgressBar courseID={courseID} submission={submission} />
            </div>
            <div className="flex-shrink-0 w-8 flex items-center justify-center">
                <SubmissionTypeIcon solo={!isGroupSubmission(submission)} />
            </div>
            <div className="flex-shrink-0 w-44 text-right">
                <div className={`text-sm font-semibold flex items-center justify-end gap-1.5 ${color}`}>
                    {icon && <i className={`fa ${icon}`} />}
                    {assignmentStatusText(assignment, submission, status)}
                </div>
                <div className="text-xs text-base-content/50 leading-tight">
                    {submissionDate && <div>{submissionDate}</div>}
                </div>
            </div>
        </div>
    )
}

export default SubmissionRow

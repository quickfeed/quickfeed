import React from "react"
import { Review, Submission, Submission_Status } from "../../../proto/qf/types_pb"
import { NoSubmission } from "../../consts"
import { getFormattedTime, getStatusByUser, SubmissionStatus } from "../../Helpers"
import { useAppState } from "../../overmind"
import ManageSubmissionStatus from "../ManageSubmissionStatus"

interface ReviewInfoProps {
    courseID: string
    assignmentName: string
    reviewers: number
    submission: Submission
    review: Review
}

const ReviewInfo = ({ courseID, assignmentName, reviewers, submission, review }: ReviewInfoProps) => {
    const state = useAppState()

    const user = state.selectedEnrollment?.user
    let status = Submission_Status.NONE
    if (user) {
        status = getStatusByUser(submission, user.ID)
    }

    const submissionStatus = submission ? SubmissionStatus[status] : NoSubmission
    const InfoRow = ({ label, value, badge }: { label: string, value: React.ReactNode, badge?: React.ReactNode }) => (
        <div className="flex items-center justify-between py-3 px-4 hover:bg-base-200 transition-colors">
            <span className="text-sm font-semibold text-base-content/70 min-w-[140px]">{label}:</span>
            <div className="flex items-center gap-2 flex-1 justify-end">
                <span className="font-medium">{value}</span>
                {badge}
            </div>
        </div>
    )

    return (
        <div className="card bg-base-100 shadow-xl">
            <div className="card-body p-0">
                <div className="flex items-center justify-between bg-primary text-primary-content px-6 py-4 rounded-t-2xl">
                    <div className="flex items-center gap-2">
                        <i className="fa fa-clipboard-check text-xl"></i>
                        <h3 className="text-lg font-bold">{assignmentName}</h3>
                    </div>
                </div>

                <div className="divide-y divide-base-300">
                    {user && <InfoRow label="User" value={user.Name} />}
                    <InfoRow label="Reviewer" value={state.review.reviewer?.Name} />
                    <InfoRow label="Submission Status" value={submissionStatus} />
                    <InfoRow label="Score" value={review.score} />
                    <InfoRow label="Updated" value={getFormattedTime(review.edited)} />
                    <InfoRow label="Graded" value={`${state.review.graded}/${state.review.criteriaTotal}`} />
                </div>

                <div className="px-4 pb-4 pt-2">
                    {state.review.graded === state.review.criteriaTotal && (
                        <ManageSubmissionStatus courseID={courseID} reviewers={reviewers} />
                    )}
                </div>
            </div>
        </div>
    )
}

export default ReviewInfo

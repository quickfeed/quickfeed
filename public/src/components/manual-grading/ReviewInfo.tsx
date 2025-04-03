import React from "react"
import { Review, Submission, Submission_Status } from "../../../proto/qf/types_pb"
import { NoSubmission } from "../../consts"
import { getFormattedTime, getStatusByUser, SubmissionStatus } from "../../Helpers"
import { useAppState } from "../../overmind"

interface ReviewInfoProps {
    courseID: string
    assignmentName: string
    reviewers: number
    submission: Submission
    review: Review | null
}

const ReviewInfo = ({ submission, review }: ReviewInfoProps) => {
    const state = useAppState()

    const user = state.selectedEnrollment?.user
    let status = Submission_Status.NONE
    let userLi = null
    if (user) {
        status = getStatusByUser(submission, user.ID)
        // List item for the user that submitted the selected submission
        userLi = (
            <li className="list-group-item">
                <span className="w-25 mr-5 float-left">User: </span>
                {user.Name}
            </li>
        )
    }

    return (
        <ul className="list-group">
            {userLi}
            <li className="list-group-item">
                <span className="w-25 mr-5 float-left">Reviewer: </span>
                {state.review.reviewer?.Name ?? "???"}
            </li>
            <li className="list-group-item">
                <span className="w-25 mr-5 float-left">Submission Status: </span>
                {submission ? SubmissionStatus[status] : NoSubmission}
            </li>
            {review ? (
                <>
                    <li className="list-group-item">
                        <span className="w-25 mr-5 float-left">Review Status: </span>
                        <span>{review?.ready ? "Ready" : "In progress"}</span>
                    </li>
                    <li className="list-group-item">
                        <span className="w-25 mr-5 float-left">Score: </span>
                        {review?.score}
                    </li>
                    <li className="list-group-item">
                        <span className="w-25 mr-5 float-left">Updated: </span>
                        {getFormattedTime(review?.edited)}
                    </li>
                    <li className="list-group-item">
                        <span className="w-25 mr-5 float-left">Graded: </span>
                        {state.review.graded}/{state.review.criteriaTotal}
                    </li>
                </>
            ) : null}
        </ul>
    )
}

export default ReviewInfo

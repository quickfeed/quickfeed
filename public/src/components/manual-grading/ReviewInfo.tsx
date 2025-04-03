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
    review: Review
}

const ReviewInfo = ({ submission, review }: ReviewInfoProps) => {
    const state = useAppState()
    const ready = review.ready

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
                {state.review.reviewer?.Name}
            </li>
            <li className="list-group-item">
                <span className="w-25 mr-5 float-left">Submission Status: </span>
                {submission ? SubmissionStatus[status] : NoSubmission}
            </li>
            <li className="list-group-item">
                <span className="w-25 mr-5 float-left">Review Status: </span>
                <span>{ready ? "Ready" : "In progress"}</span>
            </li>
            <li className="list-group-item">
                <span className="w-25 mr-5 float-left">Score: </span>
                {review.score}
            </li>
            <li className="list-group-item">
                <span className="w-25 mr-5 float-left">Updated: </span>
                {getFormattedTime(review.edited)}
            </li>
            <li className="list-group-item">
                <span className="w-25 mr-5 float-left">Graded: </span>
                {state.review.graded}/{state.review.criteriaTotal}
            </li>
        </ul>
    )
}

export default ReviewInfo

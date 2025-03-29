import React from "react"
import { Review, Submission, Submission_Status } from "../../../proto/qf/types_pb"
import { NoSubmission } from "../../consts"
import { Color, getFormattedTime, getStatusByUser, SubmissionStatus } from "../../Helpers"
import { useActions, useAppState } from "../../overmind"
import { ButtonType } from "../admin/Button"
import DynamicButton from "../DynamicButton"
import ManageSubmissionStatus from "../ManageSubmissionStatus"
import MarkReadyButton from "./MarkReadyButton"

interface ReviewInfoProps {
    courseID: string
    assignmentName: string
    reviewers: number
    submission: Submission
    review: Review
}

const ReviewInfo = ({ courseID, assignmentName, reviewers, submission, review }: ReviewInfoProps) => {
    const state = useAppState()
    const actions = useActions()
    const ready = review.ready

    const markReadyButton = <MarkReadyButton review={review} />

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

    const setReadyOrGradeButton = ready ? <ManageSubmissionStatus courseID={courseID} reviewers={reviewers} /> : markReadyButton
    const releaseButton = (
        <DynamicButton
            text={submission.released ? "Released" : "Release"}
            color={submission.released ? Color.WHITE : Color.YELLOW}
            type={ButtonType.BUTTON}
            className={`float-right ${!state.isCourseCreator && "disabled"} `}
            onClick={() => actions.review.release({ submission, owner: state.submissionOwner })}
        />
    )
    return (
        <ul className="list-group">
            <li className="list-group-item active">
                <span className="align-middle">
                    <span style={{ display: "inline-block" }} className="w-25 mr-5 p-3">{assignmentName}</span>
                    {releaseButton}
                </span>
            </li>
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
                {ready && markReadyButton}
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
            <li className="list-group-item">
                {setReadyOrGradeButton}
            </li>
        </ul>
    )
}

export default ReviewInfo

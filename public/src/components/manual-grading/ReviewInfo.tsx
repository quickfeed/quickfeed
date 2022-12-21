import React from "react"
import { Review } from "../../../proto/qf/types_pb"
import { NoSubmission } from "../../consts"
import { Color, getFormattedTime, SubmissionStatus } from "../../Helpers"
import { useActions, useAppState } from "../../overmind"
import Button, { ButtonType } from "../admin/Button"
import ManageSubmissionStatus from "../ManageSubmissionStatus"
import MarkReadyButton from "./MarkReadyButton"


const ReviewInfo = ({ review }: { review?: Review }): JSX.Element | null => {
    const state = useAppState()
    const actions = useActions()

    if (!review) {
        return null
    }

    const assignment = state.selectedAssignment
    const submission = state.currentSubmission
    const ready = review.ready

    const markReadyButton = <MarkReadyButton review={review} />

    const setReadyOrGradeButton = ready ? <ManageSubmissionStatus /> : markReadyButton
    const releaseButton = (
        <Button onclick={() => { state.isCourseCreator && actions.review.release(!submission?.released) }}
            classname={`float-right ${!state.isCourseCreator && "disabled"} `}
            text={submission?.released ? "Released" : "Release"}
            color={submission?.released ? Color.WHITE : Color.YELLOW}
            type={ButtonType.BUTTON} />
    )
    return (
        <ul className="list-group">
            <li className="list-group-item active">
                <span className="align-middle">
                    <span style={{ display: "inline-block" }} className="w-25 mr-5 p-3">{assignment?.name}</span>
                    {releaseButton}
                </span>
            </li>
            <li className="list-group-item">
                <span className="w-25 mr-5 float-left">Reviewer: </span>
                {state.review.reviewer?.Name}
            </li>
            <li className="list-group-item">
                <span className="w-25 mr-5 float-left">Submission Status: </span>
                {submission ? SubmissionStatus[submission.status] : { NoSubmission }}
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

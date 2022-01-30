import React from "react"
import { Review } from "../../../proto/ag/ag_pb"
import { NoSubmission } from "../../consts"
import { Color, getCourseID, isCourseCreator, SubmissionStatus } from "../../Helpers"
import { useActions, useAppState } from "../../overmind"
import Button, { ButtonType } from "../admin/Button"
import ManageSubmissionStatus from "../ManageSubmissionStatus"


const ReviewInfo = ({ review }: { review?: Review }): JSX.Element => {
    const state = useAppState()
    const actions = useActions()

    if (!review) {
        return <></>
    }

    const assignment = state.activeSubmissionLink?.getAssignment()
    const submission = state.activeSubmissionLink?.getSubmission()

    const isTeacher = isCourseCreator(state.self, state.courses[getCourseID()])
    const ready = review.getReady()
    const allCriteriaGraded = state.review.graded === state.review.criteriaTotal

    const markReadyButton = (
        <Button onclick={() => { allCriteriaGraded || ready ? actions.review.updateReady(!ready) : null }}
            classname={ready ? "float-right" : allCriteriaGraded ? "" : "disabled"}
            text={ready ? "Mark In progress" : "Mark Ready"}
            color={ready ? Color.YELLOW : Color.GREEN}
            type={ready ? ButtonType.BADGE : ButtonType.BUTTON}
        />
    )

    const setReadyOrGradeButton = ready ? <ManageSubmissionStatus /> : markReadyButton
    const releaseButton = (
        <Button onclick={() => { isTeacher && actions.review.release(!submission?.getReleased()) }}
            classname={`float-right ${!isTeacher && "disabled"} `}
            text={submission?.getReleased() ? "Released" : "Release"}
            color={submission?.getReleased() ? Color.WHITE : Color.YELLOW}
            type={ButtonType.BUTTON} />
    )
    return (
        <ul className="list-group">
            <li className={`list-group-item active`}>
                <span className="align-middle">
                    <span style={{ display: "inline-block" }} className="w-25 mr-5 p-3">{assignment?.getName()}</span>
                    {releaseButton}
                </span>
            </li>
            <li className="list-group-item">
                <span className="w-25 mr-5 float-left">Reviewer: </span>
                {state.review.reviewer?.getName()}
            </li>
            <li className="list-group-item">
                <span className="w-25 mr-5 float-left">Submission Status: </span>
                {submission ? SubmissionStatus[submission.getStatus()] : { NoSubmission }}
            </li>
            <li className="list-group-item">
                <span className="w-25 mr-5 float-left">Review Status: </span>
                <span>{review.getReady() ? "Ready" : "In progress"}</span>
                {ready && markReadyButton}
            </li>
            <li className="list-group-item">
                <span className="w-25 mr-5 float-left">Score: </span>
                {review.getScore()}
            </li>
            <li className="list-group-item">
                <span className="w-25 mr-5 float-left">Updated: </span>
                {review.getEdited()}
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

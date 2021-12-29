import React from "react"
import { SubmissionStatus } from "../Helpers"
import { useAppState } from "../overmind"

const ReviewInfo = (): JSX.Element => {
    const {review: {currentReview, reviewer}, activeSubmissionLink} = useAppState()

    const assignment = activeSubmissionLink?.getAssignment()
    const submission = activeSubmissionLink?.getSubmission()

    if (currentReview) {
        return (
            <ul className="list-group">
                <li className="list-group-item active">
                    <span className="w-25 mr-5 p-3">{assignment?.getName()}</span>
                </li>
                <li className="list-group-item">
                    <span className="w-25 mr-5 float-left">Reviewer: </span>
                    {reviewer?.getName()}
                </li>
                <li className="list-group-item">
                    <span className="w-25 mr-5 float-left">Submission Status: </span>
                    {submission ? SubmissionStatus[submission.getStatus()] : "None"}
                </li>
                <li className="list-group-item">
                    <span className="w-25 mr-5 float-left">Review Status: </span>
                    {currentReview.getReady() ? "Ready" : "In progress"}
                </li>
                <li className="list-group-item">
                    <span className="w-25 mr-5 float-left">Score: </span>
                    {currentReview.getScore()}
                </li>
                <li className="list-group-item">
                    <span className="w-25 mr-5 float-left">Last Edited: </span>
                    {currentReview.getEdited()}
                </li>
            </ul>
        )
    }
    return <></>
}

export default ReviewInfo
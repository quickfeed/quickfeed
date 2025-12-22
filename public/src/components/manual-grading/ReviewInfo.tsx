import React, { useCallback, useState } from "react"
import { Review, Submission, Submission_Status } from "../../../proto/qf/types_pb"
import { NoSubmission } from "../../consts"
import { Color, getFormattedTime, getStatusByUser, SubmissionStatus } from "../../Helpers"
import { useActions, useAppState } from "../../overmind"
import { ButtonType } from "../admin/Button"
import DynamicButton from "../DynamicButton"
import ManageSubmissionStatus from "../ManageSubmissionStatus"

interface ReviewInfoProps {
    courseID: string
    assignmentName: string
    reviewers: number
    submission: Submission
    review: Review
}

const ReviewHelpText = () => {
    const [showHelp, setShowHelp] = useState(false)

    return (
        <span className="ml-2" style={{ position: "relative" }}>
            <i
                className="fa fa-question-circle text-info"
                style={{ cursor: "pointer" }}
                onClick={() => setShowHelp(!showHelp)}
                title="Click for help"
            />
            {showHelp && (
                <div
                    className="alert alert-info p-2"
                    style={{
                        position: "absolute",
                        zIndex: 1000,
                        width: "350px",
                        fontSize: "0.85rem",
                        left: "20px",
                        top: "-10px"
                    }}
                >
                    <strong>Manual Review Steps:</strong>
                    <ol className="mb-0 pl-3 mt-1">
                        <li>Select a submission to review</li>
                        <li>Grade all criteria (Pass/Fail)</li>
                        <li>Once all criteria are graded, Approve/Revise/Reject buttons appear</li>
                        <li>Set the submission status</li>
                        <li>Release the review to make it visible to the student (if required)</li>
                    </ol>
                    <button
                        className="btn btn-sm btn-outline-secondary mt-2"
                        onClick={() => setShowHelp(false)}
                    >
                        Close
                    </button>
                </div>
            )}
        </span>
    )
}

const ReviewInfo = ({ courseID, assignmentName, reviewers, submission, review }: ReviewInfoProps) => {
    const state = useAppState()
    const actions = useActions()
    const handleRelease = useCallback(() => actions.review.release({ submission, owner: state.submissionOwner }), [actions, submission, state.submissionOwner])
    const allCriteriaGraded = state.review.graded === state.review.criteriaTotal && state.review.criteriaTotal > 0

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

    const buttonText = submission.released ? "Released" : "Release"
    const buttonColor = submission.released ? Color.WHITE : Color.YELLOW
    const releaseButton = (
        <DynamicButton
            text={buttonText}
            color={buttonColor}
            type={ButtonType.BUTTON}
            className={`float-right ${!state.isCourseCreator && "disabled"} `}
            onClick={handleRelease}
        />
    )
    const submissionStatus = submission ? SubmissionStatus[status] : NoSubmission
    const reviewStatus = allCriteriaGraded ? "Graded" : "In progress"
    return (
        <ul className="list-group">
            <li className="list-group-item active">
                <span className="align-middle">
                    <span style={{ display: "inline-block" }} className="w-25 mr-5 p-3">
                        {assignmentName}
                        <ReviewHelpText />
                    </span>
                    {releaseButton}
                </span>
            </li>
            {userLi}
            <li className="list-group-item">
                <span className="w-25 mr-5 float-left">Reviewer: </span>
                {state.review.reviewer?.Name ?? "Not assigned"}
            </li>
            <li className="list-group-item">
                <span className="w-25 mr-5 float-left">Submission Status: </span>
                {submissionStatus}
            </li>
            <li className="list-group-item">
                <span className="w-25 mr-5 float-left">Review Status: </span>
                <span>{reviewStatus}</span>
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
                {allCriteriaGraded && <ManageSubmissionStatus courseID={courseID} reviewers={reviewers} />}
            </li>
        </ul>
    )
}

export default ReviewInfo

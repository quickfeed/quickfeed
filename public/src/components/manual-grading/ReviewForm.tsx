import React from "react"
import { Review, GradingCriterion_Grade } from "../../../proto/qf/types_pb"
import { isManuallyGraded, Color } from "../../Helpers"
import { useActions, useAppState } from "../../overmind"
import Button, { ButtonType } from "../admin/Button"
import ReviewInfo from "./ReviewInfo"
import ReviewResult from "../ReviewResult"
import { CenteredMessage, KnownMessage } from "../CenteredMessage"


/** Check if all criteria in the review are graded */
const isFullyGraded = (review: Review): boolean => {
    for (const bm of review.gradingBenchmarks) {
        for (const c of bm.criteria) {
            if (c.grade === GradingCriterion_Grade.NONE) {
                return false
            }
        }
    }
    return true
}

const ReviewForm = () => {
    const state = useAppState()
    const actions = useActions()

    const selectedSubmission = state.selectedSubmission
    if (!selectedSubmission) {
        return <CenteredMessage message={KnownMessage.TeacherNoSubmission} />
    }

    const selectedAssignment = state.selectedAssignment
    if (!selectedAssignment) {
        return <CenteredMessage message={KnownMessage.TeacherNoAssignment} />
    }

    const reviews = state.review.reviews.get(selectedSubmission.ID) ?? []
    const selectReviewButton: React.JSX.Element[] = []

    reviews.forEach((review, index) => {
        const fullyGraded = isFullyGraded(review)
        const buttonText = fullyGraded ? "Graded" : "In Progress"
        const buttonColor = fullyGraded ? Color.GREEN : Color.YELLOW
        const className = state.review.selectedReview === index ? "active border border-dark" : ""
        selectReviewButton.push(
            <Button key={review.ID.toString()}
                text={buttonText}
                color={buttonColor}
                type={ButtonType.BUTTON}
                className={`mr-1 ${className}`}
                onClick={() => { actions.review.setSelectedReview(index) }}
            />
        )
    })

    if (!isManuallyGraded(selectedAssignment.reviewers)) {
        return <div>This assignment is not for manual grading.</div>
    } else {
        return (
            <div className="col lab-sticky reviewLabResult">
                <div className="mb-1">{selectReviewButton}</div>
                {state.review.currentReview ? (
                    <>
                        <ReviewInfo
                            courseID={selectedAssignment.CourseID.toString()}
                            assignmentName={selectedAssignment.name}
                            reviewers={selectedAssignment.reviewers}
                            submission={selectedSubmission}
                            review={state.review.currentReview}
                        />
                        <ReviewResult review={state.review.currentReview} />
                    </>
                ) : null}
            </div>
        )
    }
}

export default ReviewForm

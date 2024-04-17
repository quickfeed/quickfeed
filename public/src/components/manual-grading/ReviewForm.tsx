import React from "react"
import { Review } from "../../../proto/qf/types_pb"
import { isManuallyGraded, Color } from "../../Helpers"
import { useActions, useAppState } from "../../overmind"
import Button, { ButtonType } from "../admin/Button"
import ReviewInfo from "./ReviewInfo"
import ReviewResult from "../ReviewResult"


const ReviewForm = (): JSX.Element => {
    const state = useAppState()
    const actions = useActions()

    if (!state.selectedSubmission) {
        return <div>No submission selected</div>
    }

    const assignment = state.selectedAssignment
    if (!assignment) {
        return <div>No Submission</div>
    }

    const isAuthor = (review: Review) => {
        return review?.ReviewerID === state.self.ID
    }

    const reviewers = assignment.reviewers ?? 0
    const reviews = state.review.reviews.get(state.selectedSubmission.ID) ?? []
    const selectReviewButton: JSX.Element[] = []

    reviews.forEach((review, index) => {
        selectReviewButton.push(
            <Button key={review.ID.toString()}
                text={review.ready ? "Ready" : "In Progress"}
                color={review.ready ? Color.GREEN : Color.YELLOW}
                type={ButtonType.BUTTON}
                className={`mr-1 ${state.review.selectedReview === index ? "active border border-dark" : ""}`}
                onClick={() => { actions.review.setSelectedReview(index) }}
            />
        )
    })

    if ((reviews.length === 0 || reviews.some(review => !isAuthor(review))) && (reviewers - reviews.length) > 0) {
        // Display a button to create a new review if:
        // there are no reviews or the current user is not the author of the review, and there are still available review slots
        selectReviewButton.push(
            <Button key="add"
                text="Add Review"
                color={Color.BLUE}
                type={ButtonType.BUTTON}
                className="mr-1"
                onClick={async () => { await actions.review.createReview() }}
            />
        )
    }

    if (!isManuallyGraded(assignment)) {
        return <div>This assignment is not for manual grading.</div>
    } else {
        return (
            <div className="col lab-sticky reviewLabResult">
                <div className="mb-1">{selectReviewButton}</div>
                {state.review.currentReview ? (
                    <>
                        <ReviewInfo review={state.review.currentReview} />
                        <ReviewResult review={state.review.currentReview} />
                    </>
                ) : null}
            </div>
        )
    }
}

export default ReviewForm

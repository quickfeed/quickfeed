import React from "react"
import { Review } from "../../../proto/qf/types_pb"
import { isManuallyGraded, Color } from "../../Helpers"
import { useActions, useAppState } from "../../overmind"
import Button from "../admin/Button"
import ReviewInfo from "./ReviewInfo"
import ReviewResult from "../ReviewResult"
import { CenteredMessage, KnownMessage } from "../CenteredMessage"


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

    const isAuthor = (review: Review) => {
        return review?.ReviewerID === state.self.ID
    }

    const reviews = state.review.reviews.get(selectedSubmission.ID) ?? []
    const selectReviewButton: React.JSX.Element[] = []

    reviews.forEach((review, index) => {
        const isSelected = state.review.selectedReview === index
        const className = isSelected ? "active border border-dark" : ""
        selectReviewButton.push(
            <Button key={review.ID.toString()}
                text={`Review ${index + 1}`}
                color={isSelected ? Color.BLUE : Color.GRAY}
                className={`mr-1 ${className}`}
                onClick={() => { actions.review.setSelectedReview(index) }}
            />
        )
    })

    if ((reviews.length === 0 || reviews.some(review => !isAuthor(review))) && (selectedAssignment.reviewers - reviews.length) > 0) {
        // Display a button to create a new review if:
        // there are no reviews or the current user is not the author of the review, and there are still available review slots
        selectReviewButton.push(
            <Button key="add"
                text="Add Review"
                color={Color.BLUE}
                className="mr-1"
                onClick={() => { actions.review.createReview() }}
            />
        )
    }

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

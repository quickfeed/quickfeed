import React from "react"
import { Assignment, Review, Submission } from "../../../proto/qf/types_pb"
import { Color } from "../../Helpers"
import MarkReadyButton from "./MarkReadyButton"
import DynamicButton from "../DynamicButton"
import Button, { ButtonType } from "../admin/Button"
import { useActions, useAppState } from "../../overmind"
import ManageSubmissionStatus from "../ManageSubmissionStatus"

interface ReviewBannerProps {
    assignment: Assignment,
    submission: Submission,
    review: Review | null,
}

const ReviewBanner = ({ assignment, submission, review }: ReviewBannerProps) => {
    const actions = useActions()
    const state = useAppState()

    const isAuthor = (review: Review) => {
        return review?.ReviewerID === state.self.ID
    }

    const reviewers = assignment.reviewers ?? 0
    const reviews = state.review.reviews.get(submission.ID) ?? []
    const selectReviewButtons: React.JSX.Element[] = []

    reviews.forEach((review, index) => {
        const border = state.review.selectedReview === index ? "active border border-dark" : ""
        const reviewStatus = review.ready ? "Ready" : "In Progress"
        const selected = state.review.selectedReview === index ? "disabled" : ""
        selectReviewButtons.push(
            <Button key={review.ID.toString()}
                text={`#${index + 1} ${reviewStatus}`}
                color={review.ready ? Color.GREEN : Color.YELLOW}
                type={ButtonType.BUTTON}
                className={`mr-1 ${border} ${selected}`}
                onClick={function () { actions.review.setSelectedReview(index) }}
            />
        )
    })

    // Display a button to create a new review if:
    // there are no reviews or the current user is not the author of the review, and there are still available review slots
    const newReview = (reviews.length === 0 || reviews.some(review => !isAuthor(review))) && (reviewers - reviews.length) > 0
    const addReviewButton = newReview ?
        <Button key="add"
            text="Add Review"
            color={Color.BLUE}
            type={ButtonType.BUTTON}
            onClick={function () { actions.review.createReview() }}
        /> : null

    return (
        <div className="lab-sticky bg-dark text-white d-flex flex-column">
            <div className="d-flex p-2">
                {selectReviewButtons}
                {addReviewButton}
            </div>
            <div className="d-flex">
                <div className="d-flex p-2 w-40 mr-5">
                    <h4 className="p-2">{assignment.name}</h4>
                    <div className="mt-1">
                        {review ? <MarkReadyButton review={review} /> : null}
                    </div>
                </div>
                <div className="ml-auto mt-2 mr-2">
                    {review?.ready ?
                        <DynamicButton
                            text={submission?.released ? "Revert Release" : "Release Lab"}
                            color={submission?.released ? Color.WHITE : Color.YELLOW}
                            type={ButtonType.BUTTON}
                            className="ml-2"
                            onClick={() => actions.review.release({ submission, owner: state.submissionOwner })}
                        />
                        : null}
                </div>
            </div>
            <div className="container mb-3">
                {review?.ready ? <ManageSubmissionStatus /> : null}
            </div>
        </div >
    )
}

export default ReviewBanner

import React from "react"
import { Assignment, Review } from "../../../proto/ag/ag_pb"
import { getCourseID, isManuallyGraded, Color } from "../../Helpers"
import { useActions, useAppState } from "../../overmind"
import Button, { ButtonType } from "../admin/Button"
import ReviewInfo from "./ReviewInfo"
import ReviewResult from "../ReviewResult"


const ReviewForm = (): JSX.Element => {
    const state = useAppState()
    const actions = useActions()
    const courseID = getCourseID()

    if (!state.activeSubmissionLink) {
        return <div>None</div>
    }

    if (!state.activeSubmissionLink.hasSubmission() || !state.activeSubmissionLink.hasAssignment()) {
        return <div>No Submission</div>
    }

    const isAuthor = (review: Review) => {
        console.log("Checking author")
        return review?.getReviewerid() === state.self.getId()
    }

    const reviewers = state.activeSubmissionLink?.getAssignment()?.getReviewers() ?? 0
    const isCourseCreator = state.courses[courseID].getCoursecreatorid() === state.self.getId()
    const reviews = state.review.reviews[courseID][state.activeSubmission] ?? []
    const selectReviewButton: JSX.Element[] = []

    reviews.forEach((review, index) => {
        if (isCourseCreator || isAuthor(review)) {
            // Teaching assistants can only select their own reviews, and course creators can select any review
            selectReviewButton.push(
                <Button key={review.getId()} onclick={() => { actions.review.setSelectedReview(index) }}
                    classname={`mr-1 ${state.review.selectedReview === index ? "active border border-dark" : ""}`}
                    text={review.getReady() ? "Ready" : "In Progress"}
                    color={review.getReady() ? Color.GREEN : Color.YELLOW}
                    type={ButtonType.BUTTON} />
            )
        }
    })

    if ((reviews.length === 0 || reviews.some(review => !isAuthor(review))) && (reviewers - reviews.length) > 0) {
        // Display a button to create a new reviews if:
        // there are no reviews or the current user is not the author of the review, and there are still available review slots
        selectReviewButton.push(
            <Button key="add" onclick={async () => { await actions.review.createReview() }}
                classname="mr-1" text="Add Review" color={Color.BLUE} type={ButtonType.BUTTON} />
        )
    }

    if (!isManuallyGraded(state.activeSubmissionLink.getAssignment() as Assignment)) {
        return <div>This assignment is not for manual grading.</div>
    } else {
        return (
            <div className="col reviewLab reviewLabResult">
                <div className="mb-1">
                    {selectReviewButton}
                </div>
                {state.review.currentReview ?
                    <>
                        <ReviewInfo review={state.review.currentReview} />
                        <ReviewResult review={state.review.currentReview} />
                    </> : null
                }
            </div>
        )
    }
}

export default ReviewForm

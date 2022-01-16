import React, { useEffect } from "react"
import { Assignment, Review, Submission } from "../../../proto/ag/ag_pb"
import { getCourseID, isManuallyGraded, Color } from "../../Helpers"
import { useActions, useAppState } from "../../overmind"
import Button, { ButtonType } from "../admin/Button"
import ReviewInfo from "./ReviewInfo"
import ReviewResult from "../ReviewResult"


// TODO: Figure out who to hide reviews from

const ReviewForm = (): JSX.Element => {
    const state = useAppState()
    const actions = useActions()
    const courseID = getCourseID()

    const selectReview = () => {
        if (state.activeSubmission) {
            const reviews = state.review.reviews[courseID][state.activeSubmission]
            if (reviews.length > 0) {
                actions.review.setSelectedReview(0)
            }
        }
    }

    useEffect(() => {
        if (state.activeSubmissionLink) {
            selectReview()
        }
    }, [])

    if (!state.activeSubmissionLink) {
        return <div>None</div>
    }

    if (!state.activeSubmissionLink.hasSubmission() || !state.activeSubmissionLink.hasAssignment()) {
        return <div>No Submission</div>
    }

    const countReadyReviews = (submission: Submission) => {
        let total = 0
        for (const review of submission.getReviewsList()) {
            total = review.getReady() ? total++ : total
        }
        return total
    }

    const canTakeOver = (submission: Submission, assignment: Assignment) => {
        if (submission.getStatus() === Submission.Status.REVISION || countReadyReviews(submission) < assignment.getReviewers()) {
            return true
        }
        return false
    }

    const isAuthor = (reviews: Review[]) => {
        for (const review of reviews) {
            if (state.self.getId() === review.getReviewerid()) {
                return true
            }
        }
        return false
    }

    const reviews = state.review.reviews[courseID][state.activeSubmission]

    const selectReviewButton: JSX.Element[] = []
    for (let i = 0; i < (state.activeSubmissionLink.getAssignment()?.getReviewers() ?? 0); i++) {
        if (reviews && reviews[i]) {
            selectReviewButton.push(
                <Button key={i} onclick={() => { actions.review.setSelectedReview(i) }}
                    classname={`mr-1 ${state.review.selectedReview === i ? "active border border-dark" : ""}`}
                    text={reviews[i].getReady() ? "Ready" : "In Progress"}
                    color={reviews[i].getReady() ? Color.GREEN : Color.YELLOW}
                    type={ButtonType.BUTTON}
                />
            )
        } else {
            selectReviewButton.push(<Button classname="mr-1" key={i} type={ButtonType.BUTTON} color={Color.BLUE} text="Create a new review" onclick={() => actions.review.setSelectedReview(i)} />)
        }
    }

    if (!isManuallyGraded(state.activeSubmissionLink.getAssignment() as Assignment)) {
        return <div>This assignment is not for manual grading.</div>
    } else {
        return (
            <div className="col reviewLab">
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

import React, { useEffect } from "react"
import { Assignment, Review, Submission } from "../../../proto/ag/ag_pb"
import { getCourseID, isManuallyGraded, Color } from "../../Helpers"
import { useActions, useAppState } from "../../overmind"
import Button, { ButtonType } from "../admin/Button"
import ReviewInfo from "../ReviewInfo"
import ReviewResult from "../ReviewResult"


// TODO: Ensure all criteria are graded before setting ready

const ReviewForm = (): JSX.Element => {
    const state = useAppState()
    const actions = useActions()
    const courseID = getCourseID()

    const selectReview = () => {
        if (state.activeSubmission) {
            const reviews = state.review.reviews[courseID][state.activeSubmission]
            if (reviews) {
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

    if (!isManuallyGraded(state.activeSubmissionLink.getAssignment() as Assignment)) {
        return <div>This assignment is not for manual grading.</div>
    } else {
        return (
            <div className="col reviewLab">
                {reviews?.length == 0 &&
                    <Button type={ButtonType.BUTTON} color={Color.GREEN} text="Create a new review" onclick={() => actions.review.createReview()} />
                }
                {state.review.currentReview ?
                    <>
                        <ReviewInfo />
                        <ReviewResult review={state.review.currentReview} />
                    </> : null
                }
            </div>
        )
    }
}

export default ReviewForm

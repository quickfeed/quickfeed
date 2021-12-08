import React, { useEffect, useState } from "react"
import { Assignment, Review, Submission, SubmissionLink } from "../../../proto/ag/ag_pb"
import { useActions, useAppState } from "../../overmind"
import ReviewResult from "../ReviewResult"


// TODO: Ensure all criteria are graded before setting ready

const ReviewForm = ({submissionLink, setSelected}: {submissionLink: SubmissionLink, setSelected: React.Dispatch<React.SetStateAction<SubmissionLink | undefined>>}): JSX.Element => {

    const state = useAppState()
    const actions = useActions()

    const [selectedReview, setSelectedReview] = useState<Review | undefined>(undefined)

    useEffect(() => {
        return actions.setActiveReview(undefined)
    }, [submissionLink])

    const countReadyReviews = (submission: Submission) => {
        let total = 0
        for (const review of submission.getReviewsList()) {
            total = review.getReady() ? total++ : total
        }
        return total
    }

    const canTakeOver = (submission: Submission, assignment: Assignment) => {
        if (submission.getStatus() === Submission.Status.REVISION || countReadyReviews(submission) < assignment.getReviewers()) {
            return true;
        }
        return false;
    }

    const isAuthor = (reviews: Review[]) => {
        for (const review of reviews) {
            if (state.self.getId() === review.getReviewerid()) {
                return true
            }
        }
        return false
    }

    const Benchmarks = () => {
        if (state.activeReview) {
            return state.activeReview.getGradingbenchmarksList().map((bm, index) => {
                const list: JSX.Element[] = []

                const criteria = bm.getCriteriaList().map(gc => { return <li key={gc.getId()}>{gc.getDescription()}</li>})
                
                return <li key={bm.getId()}>{bm.getHeading()}</li>
            })
        }
        return <li>No selected review</li>
    }

    const Reviews = () => {
        if (submissionLink.hasSubmission()) {
            return (submissionLink.getSubmission() as Submission).getReviewsList().map(rw => {
                return <li key={rw.getId()} onClick={() => actions.setActiveReview(rw)}>{rw.getReviewerid()}</li>
            })
        }
    }


    if (submissionLink.hasAssignment() && submissionLink.getAssignment()?.getReviewers() == 0) {
        return <div>This assignment is not for manual grading.</div>
    }
    return (
        <div className="col">
            <ul className="list-group">
                {state.activeReview &&
                    <ReviewResult rev={state.activeReview} />
                }
                {Reviews()}
                {Benchmarks()}
            </ul>
        </div>
    )
}

export default ReviewForm
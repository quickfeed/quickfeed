import React, { useEffect, useState } from "react"
import { Assignment, Review, Submission, SubmissionLink } from "../../../proto/ag/ag_pb"
import { useAppState } from "../../overmind"


// TODO: Ensure all criteria are graded before setting ready

const ReviewForm = ({submissionLink, setSelected}: {submissionLink: SubmissionLink, setSelected: React.Dispatch<React.SetStateAction<SubmissionLink | undefined>>}): JSX.Element => {

    const state = useAppState()

    const [selectedReview, setSelectedReview] = useState<Review | undefined>(undefined)

    useEffect(() => {
        return setSelectedReview(undefined)
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
        if (selectedReview) {
            return selectedReview.getGradingbenchmarksList().map(bm => {
                return <li key={bm.getId()}>{bm.getHeading()}</li>
            })
        }
        return <li>No selected review</li>
    }

    const Reviews = () => {
        if (submissionLink.hasSubmission()) {
            return (submissionLink.getSubmission() as Submission).getReviewsList().map(rw => {
                return <li key={rw.getId()} onClick={() => setSelectedReview(rw)}>{rw.getReviewerid()}</li>
            })
        }
    }


    if (submissionLink.hasAssignment() && submissionLink.getAssignment()?.getReviewers() == 0) {
        return <div>This assignment is not for manual grading.</div>
    }
    return (
        <ul>
            {Reviews()}
            {Benchmarks()}
        </ul>
    )
}

export default ReviewForm
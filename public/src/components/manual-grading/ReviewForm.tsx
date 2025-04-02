import React from "react"
import { isManuallyGraded } from "../../Helpers"
import { useAppState } from "../../overmind"
import ReviewInfo from "./ReviewInfo"
import ReviewResult from "../ReviewResult"
import ReviewBanner from "./ReviewBanner"


const ReviewForm = () => {
    const state = useAppState()

    const submission = state.selectedSubmission
    if (!submission) {
        return <div>No submission selected</div>
    }
    const assignment = state.selectedAssignment
    if (!assignment) {
        return <div>No Submission</div>
    }
    const review = state.review.currentReview

    if (!isManuallyGraded(assignment)) {
        return <div>This assignment is not for manual grading.</div>
    } else {
        return (
            <div className="col">
                <ReviewBanner assignment={assignment} submission={submission} review={review} />
                <div className="reviewLabResult">
                    {review ? (
                        <>
                            <ReviewInfo review={review} />
                            <ReviewResult review={review} />
                        </>
                    ) : null}
                </div>
            </div >
        )
    }
}

export default ReviewForm

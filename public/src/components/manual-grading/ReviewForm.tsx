import React from "react"
import { isManuallyGraded } from "../../Helpers"
import { useAppState } from "../../overmind"
import ReviewInfo from "./ReviewInfo"
import ReviewResult from "../ReviewResult"
import ReviewBanner from "./ReviewBanner"
import { CenteredMessage, KnownMessage } from "../CenteredMessage"


const ReviewForm = () => {
    const state = useAppState()

    const selectedSubmission = state.selectedSubmission
    if (!selectedSubmission) {
        return <CenteredMessage message={KnownMessage.NoSubmission} />
    }
    const selectedAssignment = state.selectedAssignment
    if (!selectedAssignment) {
        return <CenteredMessage message={KnownMessage.NoAssignment} />
    }

    const review = state.review.currentReview

    if (!isManuallyGraded(selectedAssignment.reviewers)) {
        return <div>This assignment is not for manual grading.</div>
    } else {
        return (
            <div className="lab-sticky reviewLabResult">
                <ReviewBanner assignment={selectedAssignment} submission={selectedSubmission} review={review} />
                <ReviewInfo
                    courseID={selectedAssignment.CourseID.toString()}
                    assignmentName={selectedAssignment.name}
                    reviewers={selectedAssignment.reviewers}
                    submission={selectedSubmission}
                    review={review}
                />
                {review ? <ReviewResult review={review} /> : null}
            </div>
        )
    }
}

export default ReviewForm

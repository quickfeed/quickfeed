import { json } from 'overmind'
import { Context } from '../../'
import { Review } from '../../../../proto/ag/ag_pb'
import { success } from '../../actions'

/* Set the index of the selected review */
export const setSelectedReview = ({state}: Context, index: number): void => {
    state.review.selectedReview = index
}

/* Update the selected review */
export const updateReview = async ({state, actions, effects}: Context): Promise<void> => {
    // If canUpdate is false, the review cannot be updated
    if (state.review.canUpdate) {
        const response = await effects.grpcMan.updateReview(json((state.review.currentReview as Review)), state.activeCourse)
        if (success(response) && response.data) {
            // Updates the currently selected review with the new data from the server
            state.review.reviews[state.activeCourse][state.activeSubmission][state.review.selectedReview] = response.data
        } else {
            actions.alertHandler(response)
        } 
    }
}

/* createReview creates a new review for the current submission and course */
export const createReview = async ({state, effects}: Context): Promise<void> => {
    const submission = state.activeSubmissionLink?.getSubmission()
    
    // If there is no submission or active course, we cannot create a review
    if (submission && !state.review.currentReview && state.activeCourse) {
        const review = new Review
        // Set the current user as the reviewer
        review.setReviewerid(state.self.getId())
        // Set the submission id
        review.setSubmissionid(submission.getId())
        const response = await effects.grpcMan.createReview(review, state.activeCourse)
        if (response.data) {
            // Adds the new review to the reviews list if the server responded with a review
            state.review.reviews[state.activeCourse][submission.getId()].push(response.data)
        }
    }
}

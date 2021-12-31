import { json } from 'overmind'
import { Context } from '../../'
import { Review } from '../../../../proto/ag/ag_pb'
import { success } from '../../actions'

export const setSelectedReview = ({state}: Context, index: number): void => {
    state.review.selectedReview = index
}

export const updateReview = async ({state, actions, effects}: Context): Promise<void> => {
    if (state.review.canUpdate) {
        const response = await effects.grpcMan.updateReview(json((state.review.currentReview as Review)), state.activeCourse)
        if (success(response) && response.data) {
            state.review.reviews[state.activeCourse][state.activeSubmission][state.review.selectedReview] = response.data
        } else {
            actions.alertHandler(response)
        }
        
        
    }
}

export const createReview = async ({state, effects}: Context): Promise<void> => {
    const submission = state.activeSubmissionLink?.getSubmission()
    if (submission && !state.review.currentReview && state.activeCourse) {
        const review = new Review
        review.setReviewerid(state.self.getId())
        review.setSubmissionid(submission.getId())
        const response = await effects.grpcMan.createReview(review, state.activeCourse)
        if (response.data) {
            state.review.reviews[state.activeCourse][submission.getId()].push(response.data)
        }
    }
}

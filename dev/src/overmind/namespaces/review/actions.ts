import { json } from 'overmind'
import { Context } from '../../'
import { Review } from '../../../../proto/ag/ag_pb'

export const setSelectedReview = ({state}: Context, index: number): void => {
    state.review.selectedReview = index
}

export const updateReview = async ({state, effects}: Context): Promise<void> => {
    if (state.review.currentReview && state.activeSubmission && state.activeCourse) {
        const response = await effects.grpcMan.updateReview(json(state.review.currentReview), state.activeCourse)
        if (response.data) {
            state.review.reviews[state.activeCourse][state.activeSubmission][state.review.selectedReview] = response.data
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

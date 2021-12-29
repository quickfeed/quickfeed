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
    //effects.grpcMan.createReview()
    return
}

export const makeNewReview = ({state}: Context): void => {
    const submission = state.activeSubmissionLink?.getSubmission()
    const assignment = submission ? state.assignments[state.activeCourse].find(a => a.getId() == submission.getAssignmentid()) : undefined
    if (submission && assignment && !state.review.currentReview) {
        if (submission.getReviewsList().length >= assignment.getReviewers()) {
            return
        }
        const review = new Review
        review.setReviewerid(state.self.getId())
        review.setSubmissionid(submission.getId())
    
        review.setGradingbenchmarksList(assignment.getGradingbenchmarksList())
        state.review.reviews[state.activeCourse][submission.getId()].push(review)
    }
}
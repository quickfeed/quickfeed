import { Context } from '../../'
import { GradingBenchmark, GradingCriterion, Review, Void } from '../../../../proto/ag/ag_pb'
import { IGrpcResponse } from '../../../GRPCManager'
import { success } from '../../actions'


/* Set the index of the selected review */
export const setSelectedReview = ({ state }: Context, index: number): void => {
    state.review.selectedReview = index
}

/* Update the selected review */
export const updateReview = async ({ state, actions, effects }: Context): Promise<void> => {
    // If canUpdate is false, the review cannot be updated
    if (state.review.canUpdate) {
        const response = await effects.grpcMan.updateReview((state.review.currentReview as Review), state.activeCourse)
        if (success(response) && response.data) {
            // Updates the currently selected review with the new data from the server
            state.review.reviews[state.activeCourse][state.activeSubmission][state.review.selectedReview] = response.data
        } else {
            actions.alertHandler(response)
        }
    }
}

export const updateReady = async ({ state, actions }: Context, ready: boolean): Promise<void> => {
    if (state.review.currentReview) {
        state.review.currentReview.setReady(ready)
        await actions.review.updateReview()
    }
}

export const updateComment = async ({ effects }: Context, { grade, comment }: { grade: GradingBenchmark | GradingCriterion, comment: string }): Promise<void> => {
    let response: IGrpcResponse<Void> | undefined = undefined
    const oldComment = grade.getComment()
    grade.setComment(comment)
    if (grade instanceof GradingBenchmark) {
        response = await effects.grpcMan.updateBenchmark(grade)
    }
    if (grade instanceof GradingCriterion) {
        response = await effects.grpcMan.updateCriterion(grade)
    }
    if (!response || !success(response)) {
        grade.setComment(oldComment)
    }
}

export const setGrade = async ({ effects }: Context, { criterion, grade }: { criterion: GradingCriterion, grade: GradingCriterion.Grade }): Promise<void> => {
    const oldGrade = criterion.getGrade()
    criterion.setGrade(grade)
    const response = await effects.grpcMan.updateCriterion(criterion)
    if (!success(response)) {
        criterion.setGrade(oldGrade)
    }
}

/* createReview creates a new review for the current submission and course */
export const createReview = async ({ state, effects }: Context): Promise<void> => {
    const submission = state.activeSubmissionLink?.getSubmission()

    // If there is no submission or active course, we cannot create a review
    if (submission && !state.review.currentReview && state.activeCourse) {
        const review = new Review
        // Set the current user as the reviewer
        review.setReviewerid(state.self.getId())
        review.setSubmissionid(submission.getId())
        const response = await effects.grpcMan.createReview(review, state.activeCourse)
        if (response.data) {
            // Adds the new review to the reviews list if the server responded with a review
            state.review.reviews[state.activeCourse][submission.getId()].push(response.data)
        }
    }
}

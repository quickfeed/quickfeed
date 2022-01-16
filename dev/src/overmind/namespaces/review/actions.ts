import { Context } from '../../'
import { GradingBenchmark, GradingCriterion, Review } from '../../../../proto/ag/ag_pb'
import { success } from '../../actions'


/* Set the index of the selected review */
export const setSelectedReview = ({ state, actions }: Context, index: number): void => {
    state.review.selectedReview = index
    if (state.review.currentReview) {
        return
    }
    const reviews = state.review.reviews[state.activeCourse][state.activeSubmission]
    const reviewers = state.activeSubmissionLink?.getAssignment()?.getReviewers() ?? -1
    if (reviews.length < reviewers && !reviews.some(r => r.getReviewerid() === state.self.getId()) && confirm('Are you sure you want to review this submission?')) {
        actions.review.createReview()
    }
}

/* Update the selected review */
export const updateReview = async ({ state, actions, effects }: Context): Promise<boolean> => {
    // If canUpdate is false, the review cannot be updated
    if (state.review.canUpdate) {
        const response = await effects.grpcMan.updateReview((state.review.currentReview as Review), state.activeCourse)
        if (success(response) && response.data) {
            // Updates the currently selected review with the new data from the server
            state.review.reviews[state.activeCourse][state.activeSubmission][state.review.selectedReview] = response.data
            return true
        } else {
            actions.alertHandler(response)
        }
    }
    return false
}

export const updateReady = async ({ state, actions }: Context, ready: boolean): Promise<void> => {
    if (state.review.currentReview) {
        state.review.currentReview.setReady(ready)
        await actions.review.updateReview()
    }
}

export const updateComment = async ({ actions }: Context, { grade, comment }: { grade: GradingBenchmark | GradingCriterion, comment: string }): Promise<void> => {
    const oldComment = grade.getComment()
    grade.setComment(comment)
    const success = await actions.review.updateReview()
    if (!success) {
        grade.setComment(oldComment)
    }
}

export const setGrade = async ({ actions }: Context, { criterion, grade }: { criterion: GradingCriterion, grade: GradingCriterion.Grade }): Promise<void> => {
    const oldGrade = criterion.getGrade()
    criterion.setGrade(grade)
    const success = actions.review.updateReview()
    if (!success) {
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

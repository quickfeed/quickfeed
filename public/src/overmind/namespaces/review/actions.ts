import { clone, create } from "@bufbuild/protobuf"
import { Context } from '../..'
import { GradingBenchmark, GradingCriterion, GradingCriterion_Grade, ReviewSchema, Submission } from '../../../../proto/qf/types_pb'
import { isAuthor } from '../../../Helpers'


/* Set the index of the selected review */
export const setSelectedReview = ({ state }: Context, index: number): void => {
    const reviews = state.review.reviews.get(state.selectedSubmission?.ID ?? -1n)
    if (index < 0) {
        const idx = reviews?.findIndex(r => isAuthor(state.self, r) || state.isCourseCreator)
        state.review.selectedReview = idx && idx >= 0 ? idx : 0
    } else {
        state.review.selectedReview = index
    }
}

/* Update the selected review */
export const updateReview = async ({ state, effects }: Context): Promise<boolean> => {
    if (!(state.review.canUpdate && state.review.currentReview)) {
        // If canUpdate is false, the review cannot be updated
        return false
    }
    const submissionID = state.selectedSubmission?.ID ?? -1n
    const reviews = state.review.reviews.get(submissionID)
    if (!reviews) {
        // If there are no reviews, the review cannot be updated
        return false
    }

    const review = state.review.currentReview
    const response = await effects.global.api.client.updateReview({
        courseID: state.activeCourse,
        review
    })
    if (response.error) {
        return false
    }

    const idx = reviews.findIndex(r => r.ID === review.ID)
    if (idx === -1) {
        // If the review was not found, abort
        return false
    }
    reviews[idx] = response.message

    // Copy the review map and update the review
    const reviewMap = new Map(state.review.reviews)
    reviewMap.set(submissionID, reviews)
    state.review.reviews = reviewMap;

    (state.selectedSubmission as Submission).score = response.message.score
    return true
}

export const updateComment = async ({ actions }: Context, { grade, comment }: { grade: GradingBenchmark | GradingCriterion, comment: string }): Promise<void> => {
    const oldComment = grade.comment
    grade.comment = comment
    const ok = await actions.review.updateReview()
    if (!ok) {
        grade.comment = oldComment
    }
}

export const updateFeedback = async ({ state, actions }: Context, { feedback }: { feedback: string }): Promise<void> => {
    if (state.review.currentReview) {
        const oldFeedback = state.review.currentReview.feedback
        state.review.currentReview.feedback = feedback
        const ok = await actions.review.updateReview()
        if (!ok) {
            state.review.currentReview.feedback = oldFeedback
        }
    }
}

export const setGrade = async ({ actions }: Context, { criterion, grade }: { criterion: GradingCriterion, grade: GradingCriterion_Grade }): Promise<void> => {
    const oldGrade = criterion.grade
    criterion.grade = grade
    const ok = await actions.review.updateReview()
    if (!ok) {
        criterion.grade = oldGrade
    }
}

export const setAllGrade = async ({ state, actions }: Context, grade: GradingCriterion_Grade): Promise<void> => {
    if (!state.review.currentReview) {
        return
    }
    const review = state.review.currentReview
    const oldReview = clone(ReviewSchema, review)

    // Set all criteria to the specified grade
    review.gradingBenchmarks.forEach((bm: GradingBenchmark) => {
        bm.criteria.forEach((c: GradingCriterion) => {
            c.grade = grade
        })
    })
    const ok = await actions.review.updateReview()
    if (!ok) {
        // Revert to the old review if the update failed
        state.review.currentReview = oldReview
    }
}

/* createReview creates a new review for the current submission and course */
export const createReview = async ({ state, actions, effects }: Context): Promise<void> => {
    if (!confirm('Are you sure you want to create a new review?')) {
        return
    }

    const submission = state.selectedSubmission
    // If there is no submission or active course, we cannot create a review
    if (submission && state.activeCourse) {
        // Set the current user as the reviewer
        const review = create(ReviewSchema, {
            ReviewerID: state.self.ID,
            SubmissionID: submission.ID,
        })

        const response = await effects.global.api.client.createReview({
            courseID: state.activeCourse,
            review,
        })
        if (response.error) {
            return
        }
        // Adds the new review to the reviews list if the server responded with a review
        const reviews = new Map(state.review.reviews)
        const length = reviews.get(submission.ID)?.push(response.message) ?? 0
        state.review.reviews = reviews
        actions.review.setSelectedReview(length - 1)
    }
}


export const setAssignmentID = ({ state }: Context, aid: bigint): void => {
    const id = state.review.assignmentID > 0 ? BigInt(-1) : aid
    state.review.assignmentID = id
}

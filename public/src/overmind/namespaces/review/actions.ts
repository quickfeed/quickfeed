import { Context } from '../..'
import { GradingBenchmark, GradingCriterion, Review } from '../../../../proto/qf/types_pb'
import { Converter } from '../../../convert'
import { Color, isAuthor, isCourseCreator } from '../../../Helpers'
import { success } from '../../actions'


/* Set the index of the selected review */
export const setSelectedReview = ({ state }: Context, index: number): void => {
    const reviews = state.review.reviews[state.activeCourse][state.activeSubmission]
    if (index < 0) {
        const idx = reviews?.findIndex(r => isAuthor(state.self, r) || isCourseCreator(state.self, state.courses[state.activeCourse]))
        state.review.selectedReview = idx >= 0 ? idx : -1
    } else {
        state.review.selectedReview = index
    }
}

/* Update the selected review */
export const updateReview = async ({ state, actions, effects }: Context): Promise<boolean> => {
    // If canUpdate is false, the review cannot be updated
    if (state.review.canUpdate && state.review.currentReview) {
        const review = Converter.toReview(state.review.currentReview)
        const response = await effects.grpcMan.updateReview(review, state.activeCourse)
        if (success(response) && response.data) {
            // Updates the currently selected review with the new data from the server
            state.review.reviews[state.activeCourse][state.activeSubmission][state.review.selectedReview] = response.data.toObject()
            return true
        } else {
            actions.alertHandler(response)
        }
    }
    return false
}

export const updateReady = async ({ state, actions }: Context, ready: boolean): Promise<void> => {
    if (state.review.currentReview) {
        state.review.currentReview.ready = ready
        await actions.review.updateReview()
    }
}

export const updateComment = async ({ actions }: Context, { grade, comment }: { grade: GradingBenchmark.AsObject | GradingCriterion.AsObject, comment: string }): Promise<void> => {
    const oldComment = grade.comment
    grade.comment = comment
    const success = await actions.review.updateReview()
    if (!success) {
        grade.comment = oldComment
    }
}

export const updateFeedback = async ({ state, actions }: Context, { feedback }: { feedback: string }): Promise<void> => {
    if (state.review.currentReview) {
        const oldFeedback = state.review.currentReview.feedback
        state.review.currentReview.feedback = feedback
        if (!await actions.review.updateReview()) {
            state.review.currentReview.feedback = oldFeedback
        }
    }
}

export const setGrade = async ({ actions }: Context, { criterion, grade }: { criterion: GradingCriterion.AsObject, grade: GradingCriterion.Grade }): Promise<void> => {
    const oldGrade = criterion.grade
    criterion.grade = grade
    const success = actions.review.updateReview()
    if (!success) {
        criterion.grade = oldGrade
    }
}

/* createReview creates a new review for the current submission and course */
export const createReview = async ({ state, actions, effects }: Context): Promise<void> => {
    if (!confirm('Are you sure you want to create a new review?')) {
        return
    }

    const submission = state.activeSubmissionLink?.submission
    // If there is no submission or active course, we cannot create a review
    if (submission && state.activeCourse) {
        const review = new Review
        // Set the current user as the reviewer
        review.setReviewerid(state.self.id)
        review.setSubmissionid(submission.id)
        const response = await effects.grpcMan.createReview(review, state.activeCourse)
        if (response.data) {
            // Adds the new review to the reviews list if the server responded with a review
            const length = state.review.reviews[state.activeCourse][submission.id].push(response.data.toObject())
            actions.review.setSelectedReview(length - 1)
        }
    }
}

export const setAssignmentID = ({ state }: Context, aid: number): void => {
    const id = state.review.assignmentID > 0 ? -1 : aid
    state.review.assignmentID = id
}

export const setMinimumScore = ({ state }: Context, minimumScore: number): void => {
    state.review.minimumScore = minimumScore
}

export const releaseAll = async ({ state, actions, effects }: Context, { release, approve }: { release: boolean, approve: boolean }): Promise<void> => {
    const assignment = state.assignments[state.activeCourse].find(a => a.id === state.review.assignmentID)

    const releaseString = release && approve ? 'release and approve'
        : release ? 'release'
            : approve ? "approve"
                : ""
    const confirmText = `Are you sure you want to ${releaseString} all reviews for ${assignment?.name} above ${state.review.minimumScore} score?`
    const invalidMinimumScore = state.review.minimumScore < 0 || state.review.minimumScore > 100

    if (invalidMinimumScore || !confirm(confirmText)) {
        invalidMinimumScore && actions.alert({ text: 'Minimum score must be in range [0, 100]', color: Color.YELLOW })
        return
    }

    const response = await effects.grpcMan.updateSubmissions(state.review.assignmentID, state.activeCourse, state.review.minimumScore, release, approve)
    if (success(response)) {
        // Refresh submissions in state for the active course
        actions.getAllCourseSubmissions(state.activeCourse)
    } else {
        actions.alertHandler(response)
    }
}

export const release = async ({ state, actions, effects }: Context, release: boolean): Promise<void> => {
    const submission = state.activeSubmissionLink?.submission
    if (submission) {
        submission.released = release
        const response = await effects.grpcMan.updateSubmission(state.activeCourse, Converter.toSubmission(submission))
        if (!success(response)) {
            submission.released = !release
            actions.alertHandler(response)
        }
    }
}

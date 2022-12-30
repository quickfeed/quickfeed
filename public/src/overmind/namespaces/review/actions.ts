import { Context } from '../..'
import { GradingBenchmark, GradingCriterion, GradingCriterion_Grade, Review, Submission } from '../../../../proto/qf/types_pb'
import { Color, isAuthor, isCourseCreator } from '../../../Helpers'
import { success } from '../../actions'


/* Set the index of the selected review */
export const setSelectedReview = ({ state }: Context, index: number): void => {
    const reviews = state.review.reviews.get(state.currentSubmission?.ID ?? -1n)
    if (index < 0) {
        const idx = reviews?.findIndex(r => isAuthor(state.self, r) || isCourseCreator(state.self, state.courses[Number(state.activeCourse)]))
        state.review.selectedReview = idx && idx >= 0 ? idx : -1
    } else {
        state.review.selectedReview = index
    }
}

/* Update the selected review */
export const updateReview = async ({ state, actions, effects }: Context): Promise<boolean> => {
    if (!(state.review.canUpdate && state.review.currentReview)) {
        // If canUpdate is false, the review cannot be updated
        return false
    }
    const submissionID = state.currentSubmission?.ID ?? -1n
    const reviews = state.review.reviews.get(submissionID)
    if (!reviews) {
        // If there are no reviews, the review cannot be updated
        return false
    }

    const review = state.review.currentReview
    const response = await effects.grpcMan.updateReview(review, state.activeCourse)
    if (!(success(response) && response.data)) {
        // If the update was not successful, alert the user and abort
        actions.alertHandler(response)
        return false
    }

    const idx = reviews.findIndex(r => r.ID === review.ID)
    if (idx === -1) {
        // If the review was not found, abort
        return false
    }
    reviews[idx] = response.data

    // Copy the review map and update the review
    const reviewMap = new Map(state.review.reviews)
    reviewMap.set(submissionID, reviews)
    state.review.reviews = reviewMap;

    (state.currentSubmission as Submission).score = response.data.score
    return true
}

export const updateReady = async ({ state, actions }: Context, ready: boolean): Promise<void> => {
    if (state.review.currentReview) {
        state.review.currentReview.ready = ready
        await actions.review.updateReview()
    }
}

export const updateComment = async ({ actions }: Context, { grade, comment }: { grade: GradingBenchmark | GradingCriterion, comment: string }): Promise<void> => {
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
        const successful = await actions.review.updateReview()
        if (!successful) {
            state.review.currentReview.feedback = oldFeedback
        }
    }
}

export const setGrade = async ({ actions }: Context, { criterion, grade }: { criterion: GradingCriterion, grade: GradingCriterion_Grade }): Promise<void> => {
    const oldGrade = criterion.grade
    criterion.grade = grade
    const successful = await actions.review.updateReview()
    if (!successful) {
        criterion.grade = oldGrade
    }
}

/* createReview creates a new review for the current submission and course */
export const createReview = async ({ state, actions, effects }: Context): Promise<void> => {
    if (!confirm('Are you sure you want to create a new review?')) {
        return
    }

    const submission = state.currentSubmission
    // If there is no submission or active course, we cannot create a review
    if (submission && state.activeCourse) {
        // Set the current user as the reviewer
        const review = new Review({
            ReviewerID: state.self.ID,
            SubmissionID: submission.ID,
        })

        const response = await effects.grpcMan.createReview(review, state.activeCourse)
        if (response.data) {
            // Adds the new review to the reviews list if the server responded with a review
            const reviews = new Map(state.review.reviews)
            const length = reviews.get(submission.ID)?.push(response.data as Review) ?? 0
            state.review.reviews = reviews
            actions.review.setSelectedReview(length - 1)
        }
    }
}

export const setAssignmentID = ({ state }: Context, aid: bigint): void => {
    const id = state.review.assignmentID > 0 ? BigInt(-1) : aid
    state.review.assignmentID = id
}

export const setMinimumScore = ({ state }: Context, minimumScore: number): void => {
    state.review.minimumScore = minimumScore
}

export const releaseAll = async ({ state, actions, effects }: Context, { release, approve }: { release: boolean, approve: boolean }): Promise<void> => {
    const assignment = state.assignments[state.activeCourse.toString()].find(a => a.ID === state.review.assignmentID)

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

export const release = async ({ state, actions, effects }: Context, released: boolean): Promise<void> => {
    const submission = state.currentSubmission
    if (submission) {
        submission.released = released
        const response = await effects.grpcMan.updateSubmission(state.activeCourse, submission)
        if (!success(response)) {
            submission.released = !released
            actions.alertHandler(response)
        }
    }
}

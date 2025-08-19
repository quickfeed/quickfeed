import { clone, create } from "@bufbuild/protobuf";
import { ReviewSchema, SubmissionSchema } from '../../../../proto/qf/types_pb';
import { Color, isAuthor } from '../../../Helpers';
export const setSelectedReview = ({ state }, index) => {
    const reviews = state.review.reviews.get(state.selectedSubmission?.ID ?? -1n);
    if (index < 0) {
        const idx = reviews?.findIndex(r => isAuthor(state.self, r) || state.isCourseCreator);
        state.review.selectedReview = idx && idx >= 0 ? idx : 0;
    }
    else {
        state.review.selectedReview = index;
    }
};
export const updateReview = async ({ state, effects }) => {
    if (!(state.review.canUpdate && state.review.currentReview)) {
        return false;
    }
    const submissionID = state.selectedSubmission?.ID ?? -1n;
    const reviews = state.review.reviews.get(submissionID);
    if (!reviews) {
        return false;
    }
    const review = state.review.currentReview;
    const response = await effects.global.api.client.updateReview({
        courseID: state.activeCourse,
        review
    });
    if (response.error) {
        return false;
    }
    const idx = reviews.findIndex(r => r.ID === review.ID);
    if (idx === -1) {
        return false;
    }
    reviews[idx] = response.message;
    const reviewMap = new Map(state.review.reviews);
    reviewMap.set(submissionID, reviews);
    state.review.reviews = reviewMap;
    state.selectedSubmission.score = response.message.score;
    return true;
};
export const updateReady = async ({ state, actions }, ready) => {
    if (state.review.currentReview) {
        state.review.currentReview.ready = ready;
        await actions.review.updateReview();
    }
};
export const updateComment = async ({ actions }, { grade, comment }) => {
    const oldComment = grade.comment;
    grade.comment = comment;
    const ok = await actions.review.updateReview();
    if (!ok) {
        grade.comment = oldComment;
    }
};
export const updateFeedback = async ({ state, actions }, { feedback }) => {
    if (state.review.currentReview) {
        const oldFeedback = state.review.currentReview.feedback;
        state.review.currentReview.feedback = feedback;
        const ok = await actions.review.updateReview();
        if (!ok) {
            state.review.currentReview.feedback = oldFeedback;
        }
    }
};
export const setGrade = async ({ actions }, { criterion, grade }) => {
    const oldGrade = criterion.grade;
    criterion.grade = grade;
    const ok = await actions.review.updateReview();
    if (!ok) {
        criterion.grade = oldGrade;
    }
};
export const createReview = async ({ state, actions, effects }) => {
    if (!confirm('Are you sure you want to create a new review?')) {
        return;
    }
    const submission = state.selectedSubmission;
    if (submission && state.activeCourse) {
        const review = create(ReviewSchema, {
            ReviewerID: state.self.ID,
            SubmissionID: submission.ID,
        });
        const response = await effects.global.api.client.createReview({
            courseID: state.activeCourse,
            review,
        });
        if (response.error) {
            return;
        }
        const reviews = new Map(state.review.reviews);
        const length = reviews.get(submission.ID)?.push(response.message) ?? 0;
        state.review.reviews = reviews;
        actions.review.setSelectedReview(length - 1);
    }
};
export const setAssignmentID = ({ state }, aid) => {
    const id = state.review.assignmentID > 0 ? BigInt(-1) : aid;
    state.review.assignmentID = id;
};
export const setMinimumScore = ({ state }, minimumScore) => {
    state.review.minimumScore = minimumScore;
};
export const releaseAll = async ({ state, actions, effects }, { release, approve }) => {
    const assignment = state.assignments[state.activeCourse.toString()].find(a => a.ID === state.review.assignmentID);
    const releaseString = () => {
        if (release && approve)
            return "release and approve";
        if (release)
            return "release";
        if (approve)
            return "approve";
        return "";
    };
    const confirmText = `Are you sure you want to ${releaseString} all reviews for ${assignment?.name} above ${state.review.minimumScore} score?`;
    const invalidMinimumScore = state.review.minimumScore < 0 || state.review.minimumScore > 100;
    if (invalidMinimumScore || !confirm(confirmText)) {
        invalidMinimumScore && actions.global.alert({ text: 'Minimum score must be in range [0, 100]', color: Color.YELLOW });
        return;
    }
    const response = await effects.global.api.client.updateSubmissions({
        courseID: state.activeCourse,
        assignmentID: state.review.assignmentID,
        scoreLimit: state.review.minimumScore,
        release,
        approve,
    });
    if (response.error) {
        return;
    }
    await actions.global.refreshCourseSubmissions(state.activeCourse);
};
export const release = async ({ state, effects }, { submission, owner }) => {
    if (!submission) {
        return;
    }
    const clonedSubmission = clone(SubmissionSchema, submission);
    clonedSubmission.released = !submission.released;
    const response = await effects.global.api.client.updateSubmission({
        courseID: state.activeCourse,
        submissionID: submission.ID,
        grades: submission.Grades,
        released: clonedSubmission.released,
        score: submission.score,
    });
    if (response.error) {
        return;
    }
    submission.released = clonedSubmission.released;
    state.submissionsForCourse.update(owner, submission);
};

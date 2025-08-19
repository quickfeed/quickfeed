import React from "react";
import { isManuallyGraded, Color } from "../../Helpers";
import { useActions, useAppState } from "../../overmind";
import Button, { ButtonType } from "../admin/Button";
import ReviewInfo from "./ReviewInfo";
import ReviewResult from "../ReviewResult";
import { CenteredMessage, KnownMessage } from "../CenteredMessage";
const ReviewForm = () => {
    const state = useAppState();
    const actions = useActions();
    const selectedSubmission = state.selectedSubmission;
    if (!selectedSubmission) {
        return React.createElement(CenteredMessage, { message: KnownMessage.TeacherNoSubmission });
    }
    const selectedAssignment = state.selectedAssignment;
    if (!selectedAssignment) {
        return React.createElement(CenteredMessage, { message: KnownMessage.TeacherNoAssignment });
    }
    const isAuthor = (review) => {
        return review?.ReviewerID === state.self.ID;
    };
    const reviews = state.review.reviews.get(selectedSubmission.ID) ?? [];
    const selectReviewButton = [];
    reviews.forEach((review, index) => {
        const buttonText = review.ready ? "Ready" : "In Progress";
        const buttonColor = review.ready ? Color.GREEN : Color.YELLOW;
        const className = state.review.selectedReview === index ? "active border border-dark" : "";
        selectReviewButton.push(React.createElement(Button, { key: review.ID.toString(), text: buttonText, color: buttonColor, type: ButtonType.BUTTON, className: `mr-1 ${className}`, onClick: () => { actions.review.setSelectedReview(index); } }));
    });
    if ((reviews.length === 0 || reviews.some(review => !isAuthor(review))) && (selectedAssignment.reviewers - reviews.length) > 0) {
        selectReviewButton.push(React.createElement(Button, { key: "add", text: "Add Review", color: Color.BLUE, type: ButtonType.BUTTON, className: "mr-1", onClick: () => { actions.review.createReview(); } }));
    }
    if (!isManuallyGraded(selectedAssignment.reviewers)) {
        return React.createElement("div", null, "This assignment is not for manual grading.");
    }
    else {
        return (React.createElement("div", { className: "col lab-sticky reviewLabResult" },
            React.createElement("div", { className: "mb-1" }, selectReviewButton),
            state.review.currentReview ? (React.createElement(React.Fragment, null,
                React.createElement(ReviewInfo, { courseID: selectedAssignment.CourseID.toString(), assignmentName: selectedAssignment.name, reviewers: selectedAssignment.reviewers, submission: selectedSubmission, review: state.review.currentReview }),
                React.createElement(ReviewResult, { review: state.review.currentReview }))) : null));
    }
};
export default ReviewForm;

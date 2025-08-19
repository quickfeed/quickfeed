import React, { useCallback } from "react";
import { Submission_Status } from "../../../proto/qf/types_pb";
import { NoSubmission } from "../../consts";
import { Color, getFormattedTime, getStatusByUser, SubmissionStatus } from "../../Helpers";
import { useActions, useAppState } from "../../overmind";
import { ButtonType } from "../admin/Button";
import DynamicButton from "../DynamicButton";
import ManageSubmissionStatus from "../ManageSubmissionStatus";
import MarkReadyButton from "./MarkReadyButton";
const ReviewInfo = ({ courseID, assignmentName, reviewers, submission, review }) => {
    const state = useAppState();
    const actions = useActions();
    const handleRelease = useCallback(() => actions.review.release({ submission, owner: state.submissionOwner }), [actions, submission, state.submissionOwner]);
    const ready = review.ready;
    const markReadyButton = React.createElement(MarkReadyButton, { review: review });
    const user = state.selectedEnrollment?.user;
    let status = Submission_Status.NONE;
    let userLi = null;
    if (user) {
        status = getStatusByUser(submission, user.ID);
        userLi = (React.createElement("li", { className: "list-group-item" },
            React.createElement("span", { className: "w-25 mr-5 float-left" }, "User: "),
            user.Name));
    }
    const setReadyOrGradeButton = ready
        ? React.createElement(ManageSubmissionStatus, { courseID: courseID, reviewers: reviewers })
        : markReadyButton;
    const buttonText = submission.released ? "Released" : "Release";
    const buttonColor = submission.released ? Color.WHITE : Color.YELLOW;
    const releaseButton = (React.createElement(DynamicButton, { text: buttonText, color: buttonColor, type: ButtonType.BUTTON, className: `float-right ${!state.isCourseCreator && "disabled"} `, onClick: handleRelease }));
    const submissionStatus = submission ? SubmissionStatus[status] : NoSubmission;
    const reviewStatus = ready ? "Ready" : "In progress";
    return (React.createElement("ul", { className: "list-group" },
        React.createElement("li", { className: "list-group-item active" },
            React.createElement("span", { className: "align-middle" },
                React.createElement("span", { style: { display: "inline-block" }, className: "w-25 mr-5 p-3" }, assignmentName),
                releaseButton)),
        userLi,
        React.createElement("li", { className: "list-group-item" },
            React.createElement("span", { className: "w-25 mr-5 float-left" }, "Reviewer: "),
            state.review.reviewer?.Name),
        React.createElement("li", { className: "list-group-item" },
            React.createElement("span", { className: "w-25 mr-5 float-left" }, "Submission Status: "),
            submissionStatus),
        React.createElement("li", { className: "list-group-item" },
            React.createElement("span", { className: "w-25 mr-5 float-left" }, "Review Status: "),
            React.createElement("span", null, reviewStatus),
            ready && markReadyButton),
        React.createElement("li", { className: "list-group-item" },
            React.createElement("span", { className: "w-25 mr-5 float-left" }, "Score: "),
            review.score),
        React.createElement("li", { className: "list-group-item" },
            React.createElement("span", { className: "w-25 mr-5 float-left" }, "Updated: "),
            getFormattedTime(review.edited)),
        React.createElement("li", { className: "list-group-item" },
            React.createElement("span", { className: "w-25 mr-5 float-left" }, "Graded: "),
            state.review.graded,
            "/",
            state.review.criteriaTotal),
        React.createElement("li", { className: "list-group-item" }, setReadyOrGradeButton)));
};
export default ReviewInfo;

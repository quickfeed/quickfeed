import React, { useCallback } from 'react';
import { useNavigate } from 'react-router';
import { getFormattedTime, isValidSubmissionForAssignment } from "../../Helpers";
import { DefaultProgressBar } from '../ProgressBar';
import SubmissionRow from './SubmissionRow';
const AssignmentCard = ({ assignment, submissions, courseID, selfID }) => {
    const navigate = useNavigate();
    const redirectTo = useCallback((submission) => {
        if (submission.groupID !== 0n) {
            navigate(`/course/${courseID}/group-lab/${submission.AssignmentID.toString()}`);
        }
        else {
            navigate(`/course/${courseID}/lab/${submission.AssignmentID.toString()}`);
        }
    }, [courseID, navigate]);
    const validSubmissions = submissions.filter((submission) => isValidSubmissionForAssignment(submission, assignment));
    const hasSubmissions = validSubmissions.length > 0;
    const redirectToSubmission = () => {
        if (hasSubmissions) {
            redirectTo(validSubmissions[0]);
        }
    };
    const buttonRole = hasSubmissions ? "button" : "";
    const ariaHidden = hasSubmissions ? "true" : "false";
    const hover = hasSubmissions ? "hover-effect" : "";
    return (React.createElement("div", { key: assignment.ID.toString(), className: `card mb-4 shadow-sm ${hover}`, onClick: redirectToSubmission, role: buttonRole, "aria-hidden": ariaHidden },
        " ",
        React.createElement("div", { className: "card-header" },
            React.createElement("div", { className: "d-flex justify-content-between align-items-center" },
                React.createElement("div", { className: "d-flex align-items-center" },
                    React.createElement("h5", { className: "card-title mb-0" }, assignment.name),
                    assignment.isGroupLab && (React.createElement("span", { className: "badge badge-secondary ml-2 p-2" }, "Group"))),
                React.createElement("div", null,
                    React.createElement("i", { className: "fa fa-calendar mr-2" }),
                    " ",
                    getFormattedTime(assignment.deadline, true)))),
        React.createElement("div", { className: "card-body" },
            validSubmissions.map((submission) => (React.createElement(SubmissionRow, { key: submission.ID.toString(), submission: submission, assignment: assignment, courseID: courseID, selfID: selfID, redirectTo: redirectTo }))),
            submissions.length === 0 && React.createElement(DefaultProgressBar, { scoreLimit: assignment.scoreLimit, isGroupLab: assignment.isGroupLab }))));
};
export default AssignmentCard;

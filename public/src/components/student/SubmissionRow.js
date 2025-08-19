import React from 'react';
import { assignmentStatusText, getStatusByUser, isGroupSubmission } from "../../Helpers";
import ProgressBar, { Progress } from "../ProgressBar";
import SubmissionTypeIcon from './SubmissionTypeIcon';
const SubmissionRow = ({ submission, assignment, courseID, selfID, redirectTo }) => {
    const hoverSubmission = assignment.isGroupLab && !isGroupSubmission(submission) ? "hover-effect" : "";
    return (React.createElement("div", { key: submission.ID.toString(), className: `row clickable mb-1 py-2 align-items-center text-left ${hoverSubmission}`, onClick: (e) => { e.stopPropagation(); redirectTo(submission); }, role: "button", "aria-hidden": "true" },
        React.createElement("div", { className: "col-8" },
            React.createElement(ProgressBar, { courseID: courseID, submission: submission, type: Progress.LAB })),
        React.createElement(SubmissionTypeIcon, { solo: !isGroupSubmission(submission) }),
        React.createElement("div", { className: "col-3" }, assignmentStatusText(assignment, submission, getStatusByUser(submission, selfID)))));
};
export default SubmissionRow;

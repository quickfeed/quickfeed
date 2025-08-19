import React from "react";
import ProgressBar, { Progress } from "./ProgressBar";
import SubmissionInfo from "./submissions/SubmissionInfo";
import SubmissionScores from "./submissions/SubmissionScores";
const LabResultTable = ({ submission, assignment }) => {
    if (submission && assignment) {
        return (React.createElement("div", { className: "pb-2" },
            React.createElement("div", { className: "pb-2" },
                React.createElement(ProgressBar, { key: "progress-bar", courseID: assignment.CourseID.toString(), submission: submission, type: Progress.LAB })),
            React.createElement(SubmissionInfo, { submission: submission, assignment: assignment }),
            React.createElement(SubmissionScores, { submission: submission })));
    }
    return React.createElement("div", { className: "container" }, " No Submission ");
};
export default LabResultTable;

import React from "react";
import { useAppState } from "../overmind";
import { Submission_Status } from "../../proto/qf/types_pb";
import { getStatusByUser } from "../Helpers";
import SubmissionTypeIcon from "./student/SubmissionTypeIcon";
export var Progress;
(function (Progress) {
    Progress[Progress["NAV"] = 0] = "NAV";
    Progress[Progress["LAB"] = 1] = "LAB";
    Progress[Progress["OVERVIEW"] = 2] = "OVERVIEW";
})(Progress || (Progress = {}));
const ProgressBar = ({ courseID, submission, type }) => {
    const state = useAppState();
    const assignment = state.assignments[courseID]?.find(assignment => assignment.ID === submission.AssignmentID);
    const score = submission.score ?? 0;
    const scorelimit = assignment?.scoreLimit ?? 0;
    const status = getStatusByUser(submission, state.self.ID);
    const secondaryProgress = scorelimit - score;
    if (type === Progress.NAV) {
        const percentage = 100 - score;
        const color = score >= scorelimit ? "green" : "yellow";
        return (React.createElement("div", { style: {
                position: "absolute",
                borderBottom: `2px solid ${color}`,
                bottom: 0,
                left: 0,
                right: `${percentage}%`,
                opacity: 0.3
            } }));
    }
    let text = "";
    let secondaryText = "";
    if (type === Progress.LAB) {
        text = `${score} %`;
        secondaryText = `${secondaryProgress} %`;
    }
    let color = "";
    if (type > Progress.NAV) {
        switch (status) {
            case Submission_Status.NONE:
                color = "bg-primary";
                break;
            case Submission_Status.APPROVED:
                color = "bg-success";
                break;
            case Submission_Status.REJECTED:
                color = "bg-danger";
                break;
            case Submission_Status.REVISION:
                color = "bg-warning text-dark";
                break;
        }
    }
    return (React.createElement("div", { className: "progress" },
        React.createElement(PrimaryProgressBar, { color: color, score: score, text: text }),
        secondaryProgress > 0 &&
            React.createElement(SecondaryProgressBar, { progress: secondaryProgress, text: secondaryText })));
};
export default ProgressBar;
export const DefaultProgressBar = ({ scoreLimit, isGroupLab }) => {
    return (React.createElement("div", { className: "row mb-1 py-2 align-items-center text-left" },
        React.createElement("div", { className: "col-8" },
            React.createElement("div", { className: "progress" },
                React.createElement(PrimaryProgressBar, { score: 0, text: "0 %" }),
                React.createElement(SecondaryProgressBar, { progress: scoreLimit, text: `${scoreLimit} %` }))),
        React.createElement(SubmissionTypeIcon, { solo: !isGroupLab }),
        React.createElement("div", { className: "col-3" }, "No submission")));
};
const PrimaryProgressBar = ({ color, score, text }) => {
    return (React.createElement("div", { className: `progress-bar ${color}`, role: "progressbar", style: { width: `${score}%`, transitionDelay: "0.5s" }, "aria-valuenow": score, "aria-valuemin": 0, "aria-valuemax": 100 }, text));
};
const SecondaryProgressBar = ({ progress, text }) => {
    return (React.createElement("div", { className: "progress-bar progressbar-secondary bg-secondary", role: "progressbar", style: { width: `${progress}%` }, "aria-valuemax": 100 }, text));
};

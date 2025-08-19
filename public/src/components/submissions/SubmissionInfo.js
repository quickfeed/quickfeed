import React from "react";
import { assignmentStatusText, getFormattedTime, getPassedTestsCount, getStatusByUser, isAllApproved, isManuallyGraded } from "../../Helpers";
import { useAppState } from "../../overmind";
const SubmissionInfo = ({ submission, assignment }) => {
    const state = useAppState();
    const enrollment = state.selectedEnrollment ?? state.enrollmentsByCourseID[assignment.CourseID.toString()];
    const buildInfo = submission.BuildInfo;
    const delivered = getFormattedTime(buildInfo?.SubmissionDate);
    const built = getFormattedTime(buildInfo?.BuildDate);
    const executionTime = buildInfo ? `${buildInfo.ExecTime / BigInt(1000)} seconds` : "";
    const status = getStatusByUser(submission, enrollment.userID);
    const className = isAllApproved(submission) ? "passed" : "failed";
    return (React.createElement("table", { className: "table table-curved table-striped" },
        React.createElement("thead", { className: "thead-dark" },
            React.createElement("tr", null,
                React.createElement("th", { colSpan: 2 }, "Lab information"),
                React.createElement("th", null, assignment.name))),
        React.createElement("tbody", null,
            React.createElement("tr", null,
                React.createElement("td", { colSpan: 2, className: className }, "Status"),
                React.createElement("td", null, assignmentStatusText(assignment, submission, status))),
            React.createElement("tr", null,
                React.createElement("td", { colSpan: 2 }, "Delivered"),
                React.createElement("td", null, delivered)),
            React.createElement("tr", null,
                React.createElement("td", { colSpan: 2 }, "Built"),
                React.createElement("td", null, built)),
            submission.approvedDate ? (React.createElement("tr", null,
                React.createElement("td", { colSpan: 2 }, "Approved"),
                React.createElement("td", null, getFormattedTime(submission.approvedDate)))) : null,
            React.createElement("tr", null,
                React.createElement("td", { colSpan: 2 }, "Deadline"),
                React.createElement("td", null, getFormattedTime(assignment.deadline, true))),
            !isManuallyGraded(assignment.reviewers) ? (React.createElement("tr", null,
                React.createElement("td", { colSpan: 2 }, "Tests Passed"),
                React.createElement("td", null, getPassedTestsCount(submission.Scores)))) : null,
            React.createElement("tr", null,
                React.createElement("td", { colSpan: 2 }, "Execution time"),
                React.createElement("td", null, executionTime)),
            React.createElement("tr", null,
                React.createElement("td", { colSpan: 2 }, "Slip days"),
                React.createElement("td", null, enrollment.slipDaysRemaining)))));
};
export default SubmissionInfo;

import React from "react";
import { useNavigate } from "react-router";
import { assignmentStatusText, getStatusByUser, Icon, isApproved, isExpired, SubmissionStatus, deadlineFormatter } from "../../Helpers";
import { useAppState } from "../../overmind";
import { Enrollment_UserStatus, SubmissionSchema } from "../../../proto/qf/types_pb";
import ProgressBar, { Progress } from "../ProgressBar";
import { create } from "@bufbuild/protobuf";
import { timestampDate } from "@bufbuild/protobuf/wkt";
const SubmissionsTable = () => {
    const state = useAppState();
    const navigate = useNavigate();
    const sortedAssignments = () => {
        const assignments = [];
        for (const courseID in state.assignments) {
            assignments.push(...state.assignments[courseID]);
        }
        assignments.sort((a, b) => {
            if (a.deadline && b.deadline) {
                return timestampDate(a.deadline).getTime() - timestampDate(b.deadline).getTime();
            }
            return 0;
        });
        return assignments;
    };
    const table = [];
    sortedAssignments().forEach(assignment => {
        const deadline = assignment.deadline;
        if (!deadline || isExpired(deadline)) {
            return;
        }
        const courseID = assignment.CourseID;
        const submissions = state.submissions.ForAssignment(assignment);
        if (submissions.length === 0) {
            return;
        }
        if (state.enrollmentsByCourseID[courseID.toString()]?.status !== Enrollment_UserStatus.STUDENT) {
            return;
        }
        const submission = submissions.find(sub => sub.AssignmentID === assignment.ID) ?? create(SubmissionSchema);
        const status = getStatusByUser(submission, state.self.ID);
        if (!isApproved(status)) {
            const deadlineInfo = deadlineFormatter(deadline, assignment.scoreLimit, submission.score);
            const course = state.courses.find(c => c.ID === courseID);
            table.push(React.createElement("tr", { key: assignment.ID.toString(), className: `clickable-row ${deadlineInfo.className}`, onClick: () => navigate(`/course/${courseID}/lab/${assignment.ID}`) },
                React.createElement("th", { scope: "row" }, course?.code),
                React.createElement("td", null,
                    assignment.name,
                    assignment.isGroupLab ?
                        React.createElement("span", { className: "badge ml-2 float-right" },
                            React.createElement("i", { className: Icon.GROUP, title: "Group Assignment" })) : null),
                React.createElement("td", null,
                    React.createElement(ProgressBar, { courseID: courseID.toString(), submission: submission, type: Progress.OVERVIEW })),
                React.createElement("td", null, deadlineInfo.time),
                React.createElement("td", null, deadlineInfo.message),
                React.createElement("td", { className: SubmissionStatus[status] }, assignmentStatusText(assignment, submission, status))));
        }
    });
    if (table.length === 0) {
        return null;
    }
    return (React.createElement("div", null,
        React.createElement("h2", null, " Assignment Deadlines "),
        React.createElement("table", { className: "table rounded-lg table-bordered table-hover", id: "LandingPageTable" },
            React.createElement("thead", null,
                React.createElement("tr", null,
                    React.createElement("th", { scope: "col" }, "Course"),
                    React.createElement("th", { scope: "col" }, "Assignment"),
                    React.createElement("th", { scope: "col" }, "Progress"),
                    React.createElement("th", { scope: "col" }, "Deadline"),
                    React.createElement("th", { scope: "col" }, "Due in"),
                    React.createElement("th", { scope: "col" }, "Status"))),
            React.createElement("tbody", null, table))));
};
export default SubmissionsTable;

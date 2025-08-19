import React, { useEffect } from 'react';
import { useLocation, useParams } from 'react-router';
import { hasReviews, isManuallyGraded } from '../Helpers';
import { useActions, useAppState } from '../overmind';
import { CenteredMessage, KnownMessage } from './CenteredMessage';
import CourseLinks from "./CourseLinks";
import LabResultTable from "./LabResultTable";
import ReviewResult from './ReviewResult';
import AssignmentFeedbackForm from './student/AssignmentFeedbackForm';
const Lab = () => {
    const state = useAppState();
    const actions = useActions().global;
    const { id, lab } = useParams();
    const courseID = id ?? "";
    const assignmentID = lab ? BigInt(lab) : BigInt(-1);
    const location = useLocation();
    const isGroupLab = location.pathname.includes("group-lab");
    useEffect(() => {
        if (!state.isTeacher) {
            actions.setSelectedAssignmentID(Number(lab));
        }
    }, [actions, lab, state.isTeacher]);
    const InternalLab = () => {
        let submission;
        let assignment;
        if (state.isTeacher) {
            submission = state.selectedSubmission;
            assignment = state.assignments[courseID].find(a => a.ID === submission?.AssignmentID) ?? null;
        }
        else {
            assignment = state.assignments[courseID]?.find(a => a.ID === assignmentID) ?? null;
            if (!assignment) {
                return React.createElement(CenteredMessage, { message: KnownMessage.StudentNoAssignment });
            }
            const submissions = state.submissions.ForAssignment(assignment);
            if (submissions.length === 0) {
                return React.createElement(CenteredMessage, { message: KnownMessage.StudentNoSubmission });
            }
            const query = (s) => isGroupLab
                ? s.groupID > 0n
                : s.userID === state.self.ID && s.groupID === 0n;
            submission = submissions.find(s => query(s)) ?? null;
        }
        if (assignment && submission) {
            const review = hasReviews(submission) ? submission.reviews : [];
            let buildLog = [];
            const buildLogRaw = submission.BuildInfo?.BuildLog;
            if (buildLogRaw) {
                buildLog = buildLogRaw.split("\n").map((logLine, idx) => React.createElement("span", { key: idx },
                    logLine,
                    React.createElement("br", null)));
            }
            return (React.createElement("div", { key: submission.ID.toString(), className: "mb-4" },
                React.createElement(LabResultTable, { submission: submission, assignment: assignment }),
                isManuallyGraded(assignment.reviewers) && submission.released ? React.createElement(ReviewResult, { review: review[0] }) : null,
                React.createElement("div", { className: "card bg-light" },
                    React.createElement("code", { className: "card-body", style: { color: "#c7254e", wordBreak: "break-word" } }, buildLog)),
                !state.isTeacher && (React.createElement(AssignmentFeedbackForm, { assignment: assignment, courseID: courseID }))));
        }
        return React.createElement(CenteredMessage, { message: state.isTeacher ? KnownMessage.TeacherNoSubmission : KnownMessage.StudentNoSubmission });
    };
    return (React.createElement("div", { className: state.isTeacher ? "" : "row" },
        React.createElement("div", { className: state.isTeacher ? "" : "col-md-9" },
            React.createElement(InternalLab, null)),
        state.isTeacher ? null : React.createElement(CourseLinks, null)));
};
export default Lab;

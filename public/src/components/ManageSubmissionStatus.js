import React, { useEffect, useCallback } from "react";
import { Submission_Status } from "../../proto/qf/types_pb";
import { Color, hasAllStatus, isManuallyGraded } from "../Helpers";
import { useActions, useAppState } from "../overmind";
import { ButtonType } from "./admin/Button";
import DynamicButton from "./DynamicButton";
const ManageSubmissionStatus = ({ courseID, reviewers }) => {
    const actions = useActions().global;
    const state = useAppState();
    const [rebuilding, setRebuilding] = React.useState(false);
    const [updating, setUpdating] = React.useState(Submission_Status.NONE);
    const [viewIndividualGrades, setViewIndividualGrades] = React.useState(false);
    useEffect(() => {
        return () => {
            setViewIndividualGrades(false);
        };
    }, [state.selectedSubmission]);
    const handleRebuild = useCallback(async () => {
        if (rebuilding) {
            return;
        }
        setRebuilding(true);
        await actions.rebuildSubmission({ owner: state.submissionOwner, submission: state.selectedSubmission });
        setRebuilding(false);
    }, [rebuilding, actions, state.submissionOwner, state.selectedSubmission]);
    const handleSetStatusOrGrade = useCallback(async (status, grade) => {
        if (updating !== Submission_Status.NONE) {
            return;
        }
        setUpdating(status);
        if (grade) {
            await actions.updateGrade({ grade, status });
        }
        else {
            await actions.updateSubmission({ owner: state.submissionOwner, submission: state.selectedSubmission, status });
        }
        setUpdating(Submission_Status.NONE);
    }, [updating, actions, state.submissionOwner, state.selectedSubmission]);
    const getButtonType = (status, grade) => {
        const submission = state.selectedSubmission;
        if (grade?.Status === status || (submission?.Grades && hasAllStatus(submission, status))) {
            return ButtonType.BUTTON;
        }
        return ButtonType.OUTLINE;
    };
    const StatusButtons = ({ grade }) => {
        const buttonsInfo = [
            { text: "Approve", color: Color.GREEN, status: Submission_Status.APPROVED },
            { text: "Revision", color: Color.YELLOW, status: Submission_Status.REVISION },
            { text: "Reject", color: Color.RED, status: Submission_Status.REJECTED }
        ];
        const dynamicButtons = buttonsInfo.map(({ text, color, status }) => (React.createElement(DynamicButton, { key: text, text: text, color: color, type: getButtonType(status, grade), className: `mr-2 ${viewIndividualGrades ? "" : "col"}`, onClick: () => handleSetStatusOrGrade(status, grade) })));
        if (grade) {
            return dynamicButtons;
        }
        return dynamicButtons.map((button, index) => (React.createElement("div", { key: `${buttonsInfo[index].text}-divButton`, className: "col" }, button)));
    };
    const getUserName = (userID) => state.courseEnrollments[courseID].find(enrollment => enrollment.userID === userID)?.user?.Name ?? "";
    return (React.createElement(React.Fragment, null,
        React.createElement("div", { className: "row mb-1 ml-auto mr-auto" },
            state.selectedSubmission?.Grades && state.selectedSubmission.Grades.length > 1 && (React.createElement(DynamicButton, { text: viewIndividualGrades ? "All Grades" : "Individual Grades", color: Color.GRAY, type: ButtonType.OUTLINE, className: "col mr-2", onClick: () => Promise.resolve(setViewIndividualGrades(!viewIndividualGrades)) })),
            !isManuallyGraded(reviewers) && (React.createElement(DynamicButton, { text: rebuilding ? "Rebuilding..." : "Rebuild", color: Color.BLUE, type: ButtonType.OUTLINE, className: "col mr-2", onClick: handleRebuild }))),
        !viewIndividualGrades && (React.createElement("div", { className: "row m-auto" },
            React.createElement(StatusButtons, null))),
        viewIndividualGrades &&
            React.createElement("table", { className: "table" },
                React.createElement("tbody", null, state.selectedSubmission?.Grades.map((grade) => (React.createElement("tr", { key: grade.UserID.toString() },
                    React.createElement("td", { className: "td-center word-break" }, getUserName(grade.UserID)),
                    React.createElement(StatusButtons, { grade: grade }))))))));
};
export default ManageSubmissionStatus;

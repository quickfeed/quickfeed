import React from "react";
import { useAppState } from "../overmind";
import Lab from "./Lab";
import ManageSubmissionStatus from "./ManageSubmissionStatus";
import { CenteredMessage, KnownMessage } from "./CenteredMessage";
const LabResult = () => {
    const state = useAppState();
    if (!state.selectedSubmission) {
        return React.createElement(CenteredMessage, { message: KnownMessage.TeacherNoSubmission });
    }
    const assignment = state.selectedAssignment;
    if (!assignment) {
        return React.createElement(CenteredMessage, { message: KnownMessage.TeacherNoAssignment });
    }
    return (React.createElement("div", { className: "lab-resize lab-sticky" },
        React.createElement(ManageSubmissionStatus, { courseID: assignment.CourseID.toString(), reviewers: assignment.reviewers }),
        React.createElement("div", { className: "reviewLabResult mt-2" },
            React.createElement(Lab, null))));
};
export default LabResult;

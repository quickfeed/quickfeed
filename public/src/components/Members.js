import React, { useState, useCallback } from "react";
import { Color, EnrollmentSort, EnrollmentStatus, EnrollmentStatusBadge, getFormattedTime, isPending, sortEnrollments } from "../Helpers";
import { useAppState, useActions } from "../overmind";
import { Enrollment_UserStatus } from "../../proto/qf/types_pb";
import Search from "./Search";
import DynamicTable from "./DynamicTable";
import DynamicButton from "./DynamicButton";
import Button, { ButtonType } from "./admin/Button";
import { useCourseID } from "../hooks/useCourseID";
const Members = () => {
    const state = useAppState();
    const actions = useActions().global;
    const courseID = useCourseID();
    const [sortBy, setSortBy] = useState(EnrollmentSort.Status);
    const [descending, setDescending] = useState(false);
    const [edit, setEditing] = useState(false);
    const setSort = (sort) => {
        if (sortBy === sort) {
            setDescending(!descending);
        }
        setSortBy(sort);
    };
    let enrollments = [];
    if (state.courseEnrollments[courseID.toString()]) {
        enrollments = state.courseEnrollments[courseID.toString()].slice();
    }
    const pending = state.pendingEnrollments;
    const header = [
        { value: "Name", onClick: () => setSort(EnrollmentSort.Name) },
        { value: "Email", onClick: () => setSort(EnrollmentSort.Email) },
        { value: "Student ID", onClick: () => setSort(EnrollmentSort.StudentID) },
        { value: "Activity", onClick: () => setSort(EnrollmentSort.Activity) },
        { value: "Approved", onClick: () => setSort(EnrollmentSort.Approved) },
        { value: "Slipdays", onClick: () => { setSort(EnrollmentSort.Slipdays); } },
        { value: "Role", onClick: () => { setSort(EnrollmentSort.Status); } },
    ];
    const handleMemberChange = useCallback((enrollment, status) => (() => actions.updateEnrollment({ enrollment, status })), [actions]);
    const handleApprovePendingEnrollments = useCallback(() => actions.approvePendingEnrollments(), [actions]);
    const members = sortEnrollments(enrollments, sortBy, descending).map(enrollment => {
        let buttonColor = Color.GREEN;
        let enrollmentButtonText = "";
        let role = Enrollment_UserStatus.STUDENT;
        switch (enrollment.status) {
            case Enrollment_UserStatus.PENDING:
                role = Enrollment_UserStatus.STUDENT;
                enrollmentButtonText = "Accept";
                buttonColor = Color.GREEN;
                break;
            case Enrollment_UserStatus.STUDENT:
                role = Enrollment_UserStatus.TEACHER;
                enrollmentButtonText = "Promote";
                buttonColor = Color.BLUE;
                break;
            case Enrollment_UserStatus.TEACHER:
                role = Enrollment_UserStatus.STUDENT;
                enrollmentButtonText = "Demote";
                buttonColor = Color.YELLOW;
                break;
            default:
                role = enrollment.status;
                break;
        }
        const buttons = (React.createElement("div", { className: "d-flex" },
            React.createElement(DynamicButton, { text: enrollmentButtonText, color: buttonColor, type: ButtonType.BADGE, className: "mr-2", onClick: handleMemberChange(enrollment, role) }),
            React.createElement(DynamicButton, { text: "Reject", color: Color.RED, type: ButtonType.BADGE, onClick: handleMemberChange(enrollment, Enrollment_UserStatus.NONE) })));
        const enrollmentBadgeIcon = (React.createElement("i", { className: EnrollmentStatusBadge[enrollment.status] }, EnrollmentStatus[enrollment.status]));
        const roleButtons = isPending(enrollment) || edit ? buttons : enrollmentBadgeIcon;
        const { Name = "", Email = "", StudentID = "" } = enrollment.user || {};
        return [
            Name, Email, StudentID,
            getFormattedTime(enrollment.lastActivityDate),
            enrollment.totalApproved.toString(),
            enrollment.slipDaysRemaining.toString(),
            roleButtons,
        ];
    });
    return (React.createElement(React.Fragment, null,
        React.createElement("div", { className: "row no-gutters pb-2" },
            React.createElement("div", { className: "col-md-6" },
                React.createElement(Search, null)),
            React.createElement("div", { className: "ml-auto" },
                React.createElement(Button, { text: edit ? "Done" : "Edit", color: edit ? Color.RED : Color.BLUE, type: ButtonType.BUTTON, onClick: () => setEditing(!edit) })),
            pending?.length > 0 ?
                React.createElement("div", { style: { marginLeft: "10px" } },
                    React.createElement(DynamicButton, { text: "Approve All", color: Color.GREEN, type: ButtonType.BUTTON, onClick: handleApprovePendingEnrollments })) : null),
        React.createElement("div", null,
            React.createElement(DynamicTable, { header: header, data: members }))));
};
export default Members;

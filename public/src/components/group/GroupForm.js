import { clone, create } from "@bufbuild/protobuf";
import React, { useEffect, useState } from "react";
import { Enrollment_UserStatus, EnrollmentSchema, GroupSchema, UserSchema } from "../../../proto/qf/types_pb";
import { Color, hasTeacher, isApprovedGroup, isHidden, isPending, isStudent } from "../../Helpers";
import { useCourseID } from "../../hooks/useCourseID";
import { useActions, useAppState } from "../../overmind";
import Button, { ButtonType } from "../admin/Button";
import DynamicButton from "../DynamicButton";
import Search from "../Search";
const GroupForm = () => {
    const state = useAppState();
    const actions = useActions().global;
    const [query, setQuery] = useState("");
    const [enrollmentType, setEnrollmentType] = useState(Enrollment_UserStatus.STUDENT);
    const courseID = useCourseID();
    const group = state.activeGroup;
    useEffect(() => {
        if (isStudent(state.enrollmentsByCourseID[courseID.toString()])) {
            actions.setActiveGroup(create(GroupSchema));
            actions.updateGroupUsers(clone(UserSchema, state.self));
        }
        return () => {
            actions.setActiveGroup(null);
        };
    }, [actions, courseID, state.enrollmentsByCourseID, state.self]);
    if (!group) {
        return null;
    }
    const userIds = group.users.map(user => user.ID);
    const search = (enrollment) => {
        if (userIds.includes(enrollment.userID) || enrollment.group && enrollment.groupID !== group.ID) {
            return true;
        }
        if (enrollment.user) {
            return isHidden(enrollment.user.Name, query);
        }
        return false;
    };
    const enrollments = state.courseEnrollments[courseID.toString()].map(enrollment => clone(EnrollmentSchema, enrollment));
    const isTeacher = hasTeacher(state.status[courseID.toString()]);
    const enrollmentFilter = (enrollment) => {
        if (isTeacher) {
            return enrollment.status === enrollmentType;
        }
        return enrollment.status === Enrollment_UserStatus.STUDENT;
    };
    const groupFilter = (enrollment) => {
        if (group && group.ID) {
            return enrollment.groupID === group.ID || enrollment.groupID === BigInt(0);
        }
        return enrollment.groupID === BigInt(0);
    };
    const sortedAndFilteredEnrollments = enrollments
        .filter(enrollment => enrollmentFilter(enrollment) && groupFilter(enrollment))
        .sort((a, b) => (a.user?.Name ?? "").localeCompare((b.user?.Name ?? "")));
    const AvailableUser = ({ enrollment }) => {
        const id = enrollment.userID;
        if (isPending(enrollment)) {
            return null;
        }
        if (id !== state.self.ID && !userIds.includes(id)) {
            return (React.createElement("li", { hidden: search(enrollment), key: id.toString(), className: "list-group-item" },
                enrollment.user?.Name,
                React.createElement(Button, { text: "+", color: Color.GREEN, type: ButtonType.BADGE, className: "ml-2 float-right", onClick: () => actions.updateGroupUsers(enrollment.user) })));
        }
        return null;
    };
    const groupMembers = group.users.map(user => {
        return (React.createElement("li", { key: user.ID.toString(), className: "list-group-item" },
            React.createElement("img", { id: "group-image", src: user.AvatarURL, alt: "" }),
            user.Name,
            React.createElement(Button, { text: "-", color: Color.RED, type: ButtonType.BADGE, className: "float-right", onClick: () => actions.updateGroupUsers(user) })));
    });
    const toggleEnrollmentType = () => {
        if (hasTeacher(enrollmentType)) {
            setEnrollmentType(Enrollment_UserStatus.STUDENT);
        }
        else {
            setEnrollmentType(Enrollment_UserStatus.TEACHER);
        }
    };
    const EnrollmentTypeButton = () => {
        if (!isTeacher) {
            return React.createElement("div", null, "Students");
        }
        return (React.createElement("button", { className: "btn btn-primary w-100", type: "button", onClick: toggleEnrollmentType }, enrollmentType === Enrollment_UserStatus.STUDENT ? "Students" : "Teachers"));
    };
    const GroupNameBanner = React.createElement("div", { className: "card-header", style: { textAlign: "center" } }, group.name);
    const GroupNameInput = group && isApprovedGroup(group)
        ? null
        : React.createElement("input", { placeholder: "Group Name:", onKeyUp: e => actions.updateGroupName(e.currentTarget.value) });
    return (React.createElement("div", { className: "container" },
        React.createElement("div", { className: "row" },
            React.createElement("div", { className: "card well col-md-offset-2" },
                React.createElement("div", { className: "card-header", style: { textAlign: "center" } },
                    React.createElement(EnrollmentTypeButton, null)),
                React.createElement(Search, { placeholder: "Search", setQuery: setQuery }),
                React.createElement("ul", { className: "list-group list-group-flush" }, sortedAndFilteredEnrollments.map((enrollment) => {
                    return React.createElement(AvailableUser, { key: enrollment.ID, enrollment: enrollment });
                }))),
            React.createElement("div", { className: 'col' },
                React.createElement("div", { className: "card well col-md-offset-2" },
                    GroupNameBanner,
                    GroupNameInput,
                    groupMembers,
                    group && group.ID ?
                        React.createElement("div", { className: "row justify-content-md-center" },
                            React.createElement(DynamicButton, { text: "Update", color: Color.BLUE, type: ButtonType.BUTTON, className: "ml-2", onClick: () => actions.updateGroup(group) }),
                            React.createElement(Button, { text: "Cancel", color: Color.RED, type: ButtonType.OUTLINE, className: "ml-2", onClick: () => actions.setActiveGroup(null) }))
                        :
                            React.createElement(DynamicButton, { text: "Create Group", color: Color.GREEN, type: ButtonType.BUTTON, onClick: () => actions.createGroup({ courseID, users: userIds, name: group.name }) }))))));
};
export default GroupForm;

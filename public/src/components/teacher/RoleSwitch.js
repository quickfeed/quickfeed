import React from "react";
import { hasTeacher, isTeacher } from "../../Helpers";
import { useActions, useAppState } from "../../overmind";
const RoleSwitch = ({ enrollment }) => {
    const state = useAppState();
    const actions = useActions().global;
    if (hasTeacher(state.status[enrollment.courseID.toString()])) {
        return (React.createElement("span", { className: "clickable", onClick: () => actions.changeView() }, isTeacher(enrollment) ? "Switch to Student View" : "Switch to Teacher View"));
    }
    return null;
};
export default RoleSwitch;

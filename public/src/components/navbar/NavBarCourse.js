import React from "react";
import { useNavigate } from "react-router";
import { Status } from "../../consts";
import { isStudent, isTeacher } from "../../Helpers";
import { useActions, useAppState } from "../../overmind";
import NavBarLabs from "./NavBarLabs";
import NavBarTeacher from "./NavBarTeacher";
const NavBarCourse = ({ enrollment }) => {
    const state = useAppState();
    const actions = useActions().global;
    const navigate = useNavigate();
    const active = state.activeCourse === enrollment.courseID;
    const course = state.courses.find(c => c.ID === enrollment.courseID);
    const navigateTo = (courseID) => {
        if (active) {
            actions.setActiveCourse(BigInt(0));
            navigate("/");
        }
        else {
            actions.setActiveCourse(courseID);
            navigate(`/course/${courseID}`);
        }
    };
    return (React.createElement(React.Fragment, null,
        React.createElement("div", { role: "button", onClick: () => navigateTo(enrollment.courseID), "aria-hidden": "true" },
            React.createElement("li", { className: "activeClass" },
                React.createElement("div", { className: "col", id: "title" }, course?.code),
                React.createElement("div", { className: "col", title: "icon" },
                    React.createElement("i", { className: `icon fa fa-caret-down fa-lg float-right ${active ? "" : "fa-rotate-90"}` })))),
        React.createElement("div", { className: active ? Status.ActiveLab : Status.Inactive },
            active && isStudent(enrollment) ? React.createElement(NavBarLabs, null) : null,
            active && isTeacher(enrollment) ? React.createElement(NavBarTeacher, null) : null)));
};
export default NavBarCourse;

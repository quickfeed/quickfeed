import React, { useState } from "react";
import { useActions, useAppState } from "../../overmind";
import CourseForm from "../forms/CourseForm";
import CourseCreationInfo from "./CourseCreationInfo";
import { Color } from "../../Helpers";
const CreateCourse = () => {
    const state = useAppState();
    const actions = useActions().global;
    const [course, setCourse] = useState();
    const [orgName, setOrgName] = useState("");
    const refresh = React.useCallback(async () => {
        await actions.getCourses();
        const c = state.courses.find(c => c.ScmOrganizationName === orgName);
        if (c) {
            await actions.getEnrollmentsByUser();
            setCourse(c);
        }
        else {
            actions.alert({ text: "Course not found. Make sure the organization name is correct and that you have installed the GitHub App.", color: Color.YELLOW, delay: 10000 });
        }
    }, [actions, orgName, state.courses]);
    const buttonClass = course ? "btn btn-success disabled" : "btn btn-primary";
    const findTextOrIcon = course ? React.createElement("i", { className: "fa fa-check" }) : "Find";
    const refreshIfNoCourse = course ? undefined : refresh;
    return (React.createElement("div", { className: "container" },
        React.createElement(CourseCreationInfo, null),
        React.createElement("div", { className: "row" },
            React.createElement("div", { className: "col input-group mb-3" },
                React.createElement("div", { className: "input-group-prepend" },
                    React.createElement("div", { className: "input-group-text" }, "Get Course")),
                React.createElement("input", { className: "form-control", disabled: course ? true : false, onKeyUp: e => setOrgName(e.currentTarget.value) }),
                React.createElement("span", { role: "button", "aria-hidden": "true", className: buttonClass, onClick: refreshIfNoCourse }, findTextOrIcon))),
        course ? React.createElement(CourseForm, { courseToEdit: course }) : null));
};
export default CreateCourse;

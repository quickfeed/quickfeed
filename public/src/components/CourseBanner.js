import React from "react";
import { useAppState } from "../overmind";
import CourseFavoriteButton from "./CourseFavoriteButton";
import RoleSwitch from "./teacher/RoleSwitch";
import { useCourseID } from "../hooks/useCourseID";
const CourseBanner = () => {
    const state = useAppState();
    const enrollment = state.enrollmentsByCourseID[useCourseID().toString()];
    return (React.createElement("div", { className: "jumbotron" },
        React.createElement("div", { className: "centerblock container" },
            React.createElement("h1", null,
                enrollment.course?.name,
                React.createElement(CourseFavoriteButton, { enrollment: enrollment, style: { "padding": "20px" } })),
            React.createElement(RoleSwitch, { enrollment: enrollment }))));
};
export default CourseBanner;

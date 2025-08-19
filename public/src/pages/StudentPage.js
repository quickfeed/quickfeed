import React from "react";
import { Route, Routes, useLocation } from "react-router";
import CourseLabs from "../components/student/CourseLabs";
import CourseLinks from "../components/CourseLinks";
import GroupPage from "./GroupPage";
import Lab from "../components/Lab";
import RedirectButton from "../components/RedirectButton";
import Alerts from "../components/alerts/Alerts";
import { useCourseID } from "../hooks/useCourseID";
const StudentPage = () => {
    const courseID = useCourseID();
    const location = useLocation();
    const root = `/course/${courseID}`;
    return (React.createElement("div", { className: "box" },
        React.createElement(RedirectButton, { to: root }),
        React.createElement(Alerts, null),
        React.createElement("div", { className: "row", hidden: location.pathname !== root },
            React.createElement("div", { className: "col-md-9" },
                React.createElement(CourseLabs, null)),
            React.createElement(CourseLinks, null)),
        React.createElement(Routes, null,
            React.createElement(Route, { path: "/group", element: React.createElement(GroupPage, null) }),
            React.createElement(Route, { path: "/lab/:lab", element: React.createElement(Lab, null) }),
            React.createElement(Route, { path: "/group-lab/:lab", element: React.createElement(Lab, null) }))));
};
export default StudentPage;

import React from "react";
import { useAppState } from "../overmind";
import { Link } from "react-router-dom";
import NavBarCourse from "./navbar/NavBarCourse";
import { isEnrolled, isVisible } from "../Helpers";
const NavFavorites = () => {
    const state = useAppState();
    const visible = state.enrollments.filter(enrollment => isEnrolled(enrollment) && isVisible(enrollment));
    const courses = visible.map((enrollment) => {
        return React.createElement(NavBarCourse, { key: enrollment.ID.toString(), enrollment: enrollment });
    });
    return (React.createElement("nav", { className: `navigator ${state.showFavorites ? "" : "hidden"}` },
        React.createElement("ul", { key: "list", className: "sidebarList" },
            courses,
            state.isLoggedIn &&
                React.createElement("li", { key: "all", className: "courseLink" },
                    React.createElement(Link, { to: "/courses", className: "sidebar-items-link" }, "View all courses")))));
};
export default NavFavorites;

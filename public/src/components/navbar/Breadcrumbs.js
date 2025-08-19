import React, { useEffect, useState } from 'react';
import { Link, useLocation } from 'react-router-dom';
import { ScreenSize } from "../../consts";
import useWindowSize from "../../hooks/windowsSize";
import { useActions, useAppState } from '../../overmind';
const Breadcrumbs = () => {
    const state = useAppState();
    const actions = useActions().global;
    const location = useLocation();
    const { width } = useWindowSize();
    const [courseName, setCourseName] = useState(null);
    const [assignmentName, setAssignmentName] = useState(null);
    const pathnames = location.pathname.split('/').filter(x => x);
    const handleDashboard = () => {
        actions.setActiveCourse(0n);
    };
    const resolveCourseName = (courses, courseId, width) => {
        const course = courses.find(c => c.ID.toString() === courseId);
        if (!course)
            return null;
        return width < ScreenSize.ExtraLarge ? course.code : course.name;
    };
    const resolveAssignmentName = (assignments, assignmentId) => {
        const assignment = assignments.find(a => a.ID.toString() === assignmentId);
        return assignment?.name ?? null;
    };
    useEffect(() => {
        const [prefix, courseId, section, assignmentId] = pathnames;
        if (prefix === 'course' && courseId) {
            setCourseName(resolveCourseName(state.courses, courseId, width));
            if (section === 'lab' && assignmentId) {
                const courseAssignments = state.assignments?.[courseId] ?? [];
                setAssignmentName(resolveAssignmentName(courseAssignments, assignmentId));
            }
        }
    }, [pathnames, state.courses, state.assignments, width]);
    return (React.createElement("nav", { "aria-label": "breadcrumb" },
        React.createElement("ol", { className: "breadcrumb m-0 bg-transparent" },
            React.createElement("li", { className: "breadcrumb-item" },
                React.createElement(Link, { to: "/", onClick: handleDashboard }, "Dashboard")),
            pathnames.map((value, index) => {
                const last = index === pathnames.length - 1;
                const to = `/${pathnames.slice(0, index + 1).join('/')}`;
                let breadcrumbName = decodeURIComponent(value.charAt(0).toUpperCase() + value.slice(1));
                if (index === 0 && value === 'course') {
                    return null;
                }
                if (index === 2 && value === 'lab') {
                    return null;
                }
                if (index === 1 && courseName && pathnames[0] === 'course') {
                    breadcrumbName = courseName;
                }
                if (index === 3 && assignmentName && pathnames[2] === 'lab') {
                    breadcrumbName = assignmentName;
                }
                return last ? (React.createElement("li", { key: to, className: "breadcrumb-item active", "aria-current": "page" }, breadcrumbName)) : (React.createElement("li", { key: to, className: "breadcrumb-item" },
                    React.createElement(Link, { to: to }, breadcrumbName)));
            }))));
};
export default Breadcrumbs;

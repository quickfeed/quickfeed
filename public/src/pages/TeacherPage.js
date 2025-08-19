import React, { useCallback } from "react";
import { Route, Routes, useLocation } from "react-router";
import { Color, isManuallyGraded } from "../Helpers";
import { useActions, useAppState } from "../overmind";
import Card from "../components/Card";
import GroupPage from "./GroupPage";
import Members from "../components/Members";
import RedirectButton from "../components/RedirectButton";
import Results from "../components/Results";
import Assignments from "../components/teacher/Assignments";
import Alerts from "../components/alerts/Alerts";
import { useCourseID } from "../hooks/useCourseID";
const ReviewResults = () => React.createElement(Results, { review: true });
const RegularResults = () => React.createElement(Results, { review: false });
const TeacherPage = () => {
    const state = useAppState();
    const actions = useActions().global;
    const courseID = useCourseID();
    const location = useLocation();
    const root = `/course/${courseID}`;
    const courseHasManualGrading = state.assignments[courseID.toString()]?.some(assignment => isManuallyGraded(assignment.reviewers));
    const members = {
        title: "View Members",
        notification: state.pendingEnrollments.length > 0 ? { color: Color.YELLOW, text: "Pending enrollments" } : undefined,
        text: "View all students, and approve new enrollments.",
        buttonText: "Members", to: `${root}/members`
    };
    const groups = {
        title: "Manage Groups",
        notification: state.pendingGroups.length > 0 ? { color: Color.YELLOW, text: "Pending groups" } : undefined,
        text: "View, edit or delete course groups.",
        buttonText: "Groups", to: `${root}/groups`
    };
    const results = { title: "View results", text: "View results for all students in the course.", buttonText: "Results", to: `${root}/results` };
    const assignments = { title: "Manage Assignments", text: "View and edit assignments.", buttonText: "Assignments", to: `${root}/assignments` };
    const handleUpdateAssignments = useCallback(() => actions.updateAssignments(courseID), [actions, courseID]);
    const updateAssignments = {
        title: "Update Course Assignments",
        text: "Fetch assignments from GitHub.",
        buttonText: "Update Assignments",
        onclick: handleUpdateAssignments
    };
    const review = { title: "Review Assignments", text: "Review assignments for students.", buttonText: "Review", to: `${root}/review` };
    return (React.createElement("div", { className: "box" },
        React.createElement(RedirectButton, { to: root }),
        React.createElement(Alerts, null),
        React.createElement("div", { className: "row", hidden: location.pathname !== root },
            courseHasManualGrading && React.createElement(Card, { ...review }),
            React.createElement(Card, { ...results }),
            React.createElement(Card, { ...groups }),
            React.createElement(Card, { ...members }),
            React.createElement(Card, { ...assignments }),
            React.createElement(Card, { ...updateAssignments })),
        React.createElement(Routes, null,
            React.createElement(Route, { path: "/groups", element: React.createElement(GroupPage, null) }),
            React.createElement(Route, { path: "/members", element: React.createElement(Members, null) }),
            React.createElement(Route, { path: "/review", element: React.createElement(ReviewResults, null) }),
            React.createElement(Route, { path: "/results", element: React.createElement(RegularResults, null) }),
            React.createElement(Route, { path: "/assignments", element: React.createElement(Assignments, null) }))));
};
export default TeacherPage;

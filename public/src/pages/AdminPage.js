import React from "react";
import { Route, Routes, useNavigate, useLocation } from "react-router";
import { useAppState } from "../overmind";
import EditCourse from "../components/admin/EditCourse";
import Users from "../components/admin/Users";
import Card from "../components/Card";
import RedirectButton from "../components/RedirectButton";
import Alerts from "../components/alerts/Alerts";
import CreateCourse from "../components/admin/CreateCourse";
const AdminPage = () => {
    const state = useAppState();
    const navigate = useNavigate();
    const location = useLocation();
    const manageUsers = { title: "Manage Users", text: "View and manage all users.", buttonText: "Manage Users", to: "/admin/manage" };
    const createCourse = { title: "Create Course", text: "Create a new course.", buttonText: "Create Course", to: "/admin/create" };
    const editCourse = { title: "Edit Course", text: "Edit an existing course.", buttonText: "Edit Course", to: "/admin/edit" };
    if (!state.self.IsAdmin) {
        navigate("/");
    }
    const root = "/admin";
    return (React.createElement("div", { className: "box" },
        React.createElement(RedirectButton, { to: root }),
        React.createElement(Alerts, null),
        React.createElement("div", { className: "row", hidden: location.pathname !== root },
            React.createElement(Card, { ...manageUsers }),
            React.createElement(Card, { ...createCourse }),
            React.createElement(Card, { ...editCourse })),
        React.createElement(Routes, null,
            React.createElement(Route, { path: "/manage", element: React.createElement(Users, null) }),
            React.createElement(Route, { path: "/create", element: React.createElement(CreateCourse, null) }),
            React.createElement(Route, { path: "/edit", element: React.createElement(EditCourse, null) }))));
};
export default AdminPage;

import React from "react";
import { Navigate } from "react-router";
import { hasEnrollment } from "../Helpers";
import { useAppState } from "../overmind";
import Alerts from "./alerts/Alerts";
import Courses from "./Courses";
import SubmissionsTable from "./dashboard/SubmissionsTable";
const Dashboard = () => {
    const state = useAppState();
    if (!hasEnrollment(state.enrollments)) {
        return React.createElement(Navigate, { to: "/courses" });
    }
    return (React.createElement("div", { className: "mt-5" },
        React.createElement(Alerts, null),
        React.createElement(Courses, { home: true }),
        React.createElement(SubmissionsTable, null)));
};
export default Dashboard;

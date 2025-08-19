import React from "react";
import { useGrpc } from "../../overmind";
import { EnrollmentStatus, EnrollmentStatusBadge } from "../../Helpers";
const User = ({ user }) => {
    const { api } = useGrpc().global;
    const [enrollments, setEnrollments] = React.useState([]);
    const [showEnrollments, setShowEnrollments] = React.useState(false);
    const toggleEnrollments = () => {
        setShowEnrollments(!showEnrollments);
        if (!enrollments.length) {
            getEnrollments();
        }
    };
    const getEnrollments = () => {
        api.client
            .getEnrollments({
            FetchMode: { case: "userID", value: user.ID },
        })
            .then((response) => {
            setEnrollments(response.message.enrollments);
        });
    };
    const enrollmentsList = enrollments.length ? (React.createElement("div", null, enrollments.map((enrollment) => (React.createElement("div", { key: enrollment.ID.toString() },
        React.createElement("span", { className: "badge badge-secondary" }, enrollment.course?.name),
        " ",
        React.createElement("span", { className: EnrollmentStatusBadge[enrollment.status] }, EnrollmentStatus[enrollment.status])))))) : (React.createElement("div", null,
        React.createElement("span", { className: "badge badge-secondary" }, "No enrollments")));
    return (React.createElement("div", { role: "button", "aria-hidden": "true", className: "clickable", onClick: toggleEnrollments },
        user.Name,
        user.IsAdmin ? (React.createElement("span", { className: "badge badge-primary ml-2" }, "Admin")) : null,
        showEnrollments ? enrollmentsList : null));
};
export default User;

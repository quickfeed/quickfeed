import React, { useCallback } from 'react';
import { useNavigate } from 'react-router';
import { EnrollmentStatus, hasEnrolled, hasNone, hasPending } from '../Helpers';
import { useActions } from '../overmind';
import CourseFavoriteButton from './CourseFavoriteButton';
const CardColor = [
    "info",
    "secondary",
    "primary",
    "success"
];
const CourseCard = ({ course, enrollment }) => {
    const actions = useActions().global;
    const navigate = useNavigate();
    const status = enrollment.status;
    const handleEnroll = useCallback(() => actions.enroll(course.ID), [actions, course.ID]);
    const CourseEnrollmentButton = () => {
        if (hasNone(status)) {
            return React.createElement("button", { className: "btn btn-primary course-button", onClick: handleEnroll }, "Enroll");
        }
        else if (hasPending(status)) {
            return React.createElement("button", { className: "btn btn-secondary course-button disabled" }, "Pending");
        }
        return React.createElement("button", { className: "btn btn-primary course-button", onClick: () => navigate(`/course/${enrollment.courseID}`) }, "Go to Course");
    };
    const CourseEnrollmentStatus = () => {
        if (!hasEnrolled(status)) {
            return null;
        }
        return (React.createElement("div", { className: "d-flex align-items-center" },
            React.createElement(CourseFavoriteButton, { enrollment: enrollment, style: { marginLeft: 'auto' } }),
            React.createElement("p", { className: "mb-0 ml-2 text-white" }, EnrollmentStatus[status])));
    };
    return (React.createElement("div", { className: "card course-card mb-4 shadow-sm" },
        React.createElement("div", { className: `card-header bg-${CardColor[status]} text-white d-flex justify-content-between align-items-center` },
            React.createElement("span", null, course.code),
            React.createElement(CourseEnrollmentStatus, null)),
        React.createElement("div", { className: "card-body" },
            React.createElement("h5", { className: "card-title" }, course.name),
            React.createElement("p", { className: "card-text text-muted" },
                course.tag,
                " ",
                course.year),
            React.createElement(CourseEnrollmentButton, null))));
};
export default CourseCard;

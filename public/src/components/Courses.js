import React from "react";
import { useAppState } from "../overmind";
import { Enrollment_UserStatus, EnrollmentSchema } from "../../proto/qf/types_pb";
import CourseCard from "./CourseCard";
import Button, { ButtonType } from "./admin/Button";
import { useNavigate } from "react-router";
import { Color, isVisible } from "../Helpers";
import { create } from "@bufbuild/protobuf";
const Courses = (overview) => {
    const state = useAppState();
    const navigate = useNavigate();
    if (state.courses.length === 0) {
        return (React.createElement("div", { className: "container centered" },
            React.createElement("h3", null, "There are currently no available courses."),
            state.self.IsAdmin ?
                React.createElement("div", null,
                    React.createElement(Button, { text: "Go to course creation", color: Color.GREEN, type: ButtonType.BUTTON, className: "mr-3", onClick: () => navigate("/admin/create") }),
                    React.createElement(Button, { text: "Manage users", color: Color.BLUE, type: ButtonType.BUTTON, onClick: () => navigate("/admin/manage") }))
                : null));
    }
    const courses = () => {
        const favorite = [];
        const student = [];
        const teacher = [];
        const pending = [];
        const availableCourses = [];
        state.courses.forEach(course => {
            const enrol = state.enrollmentsByCourseID[course.ID.toString()];
            if (enrol) {
                const courseCard = React.createElement(CourseCard, { key: course.ID.toString(), course: course, enrollment: enrol });
                if (isVisible(enrol)) {
                    favorite.push(courseCard);
                }
                else {
                    switch (enrol.status) {
                        case Enrollment_UserStatus.PENDING:
                            pending.push(courseCard);
                            break;
                        case Enrollment_UserStatus.STUDENT:
                            student.push(courseCard);
                            break;
                        case Enrollment_UserStatus.TEACHER:
                            teacher.push(courseCard);
                            break;
                    }
                }
            }
            else {
                availableCourses.push(React.createElement(CourseCard, { key: course.ID.toString(), course: course, enrollment: create(EnrollmentSchema) }));
            }
        });
        if (overview.home) {
            return (React.createElement(React.Fragment, null, favorite.length > 0 &&
                React.createElement("div", { className: "container-fluid" },
                    React.createElement("div", { className: "card-deck course-card-row favorite-row" }, favorite))));
        }
        return (React.createElement("div", { className: "box container-fluid" },
            favorite.length > 0 &&
                React.createElement("div", { className: "container-fluid" },
                    React.createElement("h2", null, "Favorites"),
                    React.createElement("div", { className: "card-deck course-card-row favorite-row" }, favorite)),
            (student.length > 0 || teacher.length > 0) &&
                React.createElement("div", { className: "container-fluid myCourses" },
                    React.createElement("h2", null, "My Courses"),
                    React.createElement("div", { className: "card-deck course-card-row" },
                        teacher,
                        student)),
            pending.length > 0 &&
                React.createElement("div", { className: "container-fluid" },
                    (student.length === 0 && teacher.length === 0) &&
                        React.createElement("h2", null, "My Courses"),
                    React.createElement("div", { className: "card-deck" }, pending)),
            availableCourses.length > 0 &&
                React.createElement(React.Fragment, null,
                    React.createElement("h2", null, "Available Courses"),
                    React.createElement("div", { className: "card-deck course-card-row" }, availableCourses))));
    };
    return courses();
};
export default Courses;

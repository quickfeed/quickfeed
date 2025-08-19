import React, { useEffect } from "react";
import { Enrollment_UserStatus } from "../../../proto/qf/types_pb";
import { useActions, useAppState } from "../../overmind";
import { hasTeacher } from "../../Helpers";
import { useNavigate } from "react-router";
const ToggleSwitch = () => {
    const { activeCourse, enrollmentsByCourseID, status } = useAppState();
    const actions = useActions().global;
    const navigate = useNavigate();
    const [enrollmentStatus, setEnrollmentStatus] = React.useState(false);
    const [text, setText] = React.useState("");
    useEffect(() => {
        if (activeCourse && enrollmentsByCourseID[activeCourse.toString()]) {
            updateStatus(isTeacher());
        }
    });
    const isTeacher = () => {
        return (enrollmentsByCourseID[activeCourse.toString()].status ===
            Enrollment_UserStatus.TEACHER);
    };
    const updateStatus = (isTeacher) => {
        setEnrollmentStatus(isTeacher);
        setText(isTeacher ? "T" : "S");
    };
    const switchView = () => {
        actions.changeView().then(() => {
            updateStatus(isTeacher());
            navigate(`/course/${activeCourse}`);
        });
    };
    if (!hasTeacher(status[activeCourse.toString()])) {
        return null;
    }
    return (React.createElement("label", { className: "switch", "data-toggle": "tooltip", title: "Toggle between student and teacher view" },
        React.createElement("input", { type: "checkbox", readOnly: true, checked: enrollmentStatus }),
        React.createElement("span", { className: "slider round", onClick: switchView },
            React.createElement("span", { className: "toggle" }, text))));
};
export default ToggleSwitch;

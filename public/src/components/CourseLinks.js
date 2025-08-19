import React from "react";
import { Link } from "react-router-dom";
import { Repository_Type } from "../../proto/qf/types_pb";
import { useAppState } from "../overmind";
import { useCourseID } from "../hooks/useCourseID";
const CourseLinks = () => {
    const state = useAppState();
    const courseID = useCourseID();
    const enrollment = state.enrollmentsByCourseID[courseID.toString()];
    const repo = state.repositories[courseID.toString()];
    const hasGroup = state.hasGroup(courseID.toString());
    const groupName = enrollment.group ? `(${enrollment.group?.name})` : "";
    const links = [
        { type: Repository_Type.USER, text: "User Repository" },
        { type: Repository_Type.GROUP, text: `Group Repository ${groupName}`, style: { textAlign: "left" }, className: "overflow-ellipses" },
        { type: Repository_Type.ASSIGNMENTS, text: "Assignments" },
        { type: Repository_Type.INFO, text: "Course Info" }
    ];
    const LinkElement = ({ link }) => {
        if (repo[link.type] === undefined) {
            return null;
        }
        return React.createElement("a", { href: repo[link.type], target: "_blank", rel: "noopener noreferrer", className: `list-group-item list-group-item-action ${link.className ?? ""}`, style: link.style }, link.text);
    };
    return (React.createElement("div", { className: "col-lg-3" },
        React.createElement("div", { className: "list-group width-resize" },
            React.createElement("div", { className: "list-group-item list-group-item-action active text-center" },
                React.createElement("h6", null,
                    React.createElement("strong", null, "Links"))),
            links.map(link => { return React.createElement(LinkElement, { key: link.type, link: link }); }),
            React.createElement(Link, { to: `/course/${courseID}/group`, className: `list-group-item list-group-item-action ${hasGroup ? "" : "list-group-item-success"}` }, hasGroup ? "View Group" : "Create a Group"))));
};
export default CourseLinks;

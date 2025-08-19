import React, { useState } from "react";
import { GradingCriterion_Grade } from "../../../proto/qf/types_pb";
import { useAppState } from "../../overmind";
import GradeComment from "./GradeComment";
import CriteriaStatus from "./CriteriaStatus";
import CriterionComment from "./Comment";
import UnstyledButton from "../UnstyledButton";
const Criteria = ({ criteria }) => {
    const [editing, setEditing] = useState(false);
    const [showComment, setShowComment] = useState(true);
    const { isTeacher } = useAppState();
    let className;
    switch (criteria.grade) {
        case GradingCriterion_Grade.PASSED:
            className = "passed";
            break;
        case GradingCriterion_Grade.FAILED:
            className = "failed";
            break;
        case GradingCriterion_Grade.NONE:
            className = "not-graded";
            break;
    }
    const passed = criteria.grade == GradingCriterion_Grade.PASSED;
    const criteriaStatusOrPassFailIcon = isTeacher
        ? React.createElement(CriteriaStatus, { criterion: criteria })
        : React.createElement("i", { className: passed ? "fa fa-check" : "fa fa-exclamation-circle" });
    let comment = null;
    let button = null;
    if (isTeacher) {
        button = React.createElement(UnstyledButton, { onClick: () => setEditing(true) },
            React.createElement("i", { className: "fa fa-pencil-square-o", "aria-hidden": "true" }));
        if (criteria.comment.length > 0) {
            comment = React.createElement(CriterionComment, { comment: criteria.comment });
        }
    }
    else {
        comment = React.createElement(CriterionComment, { comment: criteria.comment });
        button = React.createElement(UnstyledButton, { onClick: () => setShowComment(!showComment) },
            React.createElement("i", { className: `fa fa-comment${!showComment ? "-o" : ""}` }));
    }
    const displayComment = criteria.comment.length > 0;
    return (React.createElement(React.Fragment, null,
        React.createElement("tr", { className: "align-items-center" },
            React.createElement("td", { className: className }, criteria.description),
            React.createElement("td", null, criteriaStatusOrPassFailIcon),
            React.createElement("td", null, (displayComment || isTeacher) ? button : null)),
        displayComment ?
            React.createElement("tr", { className: `comment comment-${className}${!showComment ? " hidden" : ""} ` },
                React.createElement("td", { onClick: () => setEditing(true), colSpan: 3 }, comment)) : null,
        React.createElement(GradeComment, { grade: criteria, editing: editing, setEditing: setEditing })));
};
export default Criteria;

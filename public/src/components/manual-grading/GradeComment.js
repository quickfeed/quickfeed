import React from "react";
import { useActions, useAppState } from "../../overmind";
const GradeComment = ({ grade, editing, setEditing }) => {
    const actions = useActions();
    const state = useAppState();
    if (!state.isTeacher || !editing) {
        return null;
    }
    const updateComment = (value) => {
        setEditing(false);
        if (value === grade.comment) {
            return;
        }
        actions.review.updateComment({ grade: grade, comment: value });
    };
    const handleBlur = (event) => {
        const { value } = event.currentTarget;
        updateComment(value);
    };
    const handleKeyUp = (event) => {
        if (event.key === "Escape") {
            setEditing(false);
            return;
        }
        if ((event.key === "Enter" || event.key === "q" || event.key === "Q") && (event.ctrlKey || event.metaKey)) {
            const { value } = event.currentTarget;
            updateComment(value);
        }
    };
    return (React.createElement("tr", null,
        React.createElement("th", { colSpan: 3 },
            React.createElement("textarea", { rows: 20, autoFocus: true, onBlur: handleBlur, onKeyUp: handleKeyUp, defaultValue: grade.comment, className: "form-control" }),
            " ")));
};
export default GradeComment;

import React, { useState } from "react";
import { useActions, useAppState } from "../../overmind";
import CriterionComment from "./Comment";
const SummaryFeedback = ({ review }) => {
    const state = useAppState();
    const actions = useActions();
    const [editing, setEditing] = useState(false);
    const summaryFeedback = (React.createElement("td", { colSpan: 3 },
        React.createElement(CriterionComment, { comment: review.feedback.length > 0 ? review.feedback : "No summary feedback" })));
    if (!state.isTeacher) {
        return React.createElement("tr", null, summaryFeedback);
    }
    const handleChange = (event) => {
        const { value } = event.currentTarget;
        setEditing(false);
        if (value === review.feedback) {
            return;
        }
        actions.review.updateFeedback({ feedback: value });
    };
    return (React.createElement(React.Fragment, null,
        React.createElement("tr", { onClick: () => setEditing(true) }, summaryFeedback),
        editing &&
            React.createElement("tr", null,
                React.createElement("td", { colSpan: 3 },
                    React.createElement("textarea", { rows: 20, autoFocus: true, onBlur: handleChange, defaultValue: review.feedback, className: "form-control" }),
                    " "))));
};
export default SummaryFeedback;

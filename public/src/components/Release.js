import React, { useCallback, useEffect } from "react";
import { Color } from "../Helpers";
import { useActions, useAppState } from "../overmind";
import { ButtonType } from "./admin/Button";
import DynamicButton from "./DynamicButton";
import FormInput from "./forms/FormInput";
const Release = () => {
    const state = useAppState();
    const actions = useActions();
    const canRelease = state.review.assignmentID > -1;
    useEffect(() => {
        return () => actions.review.setMinimumScore(0);
    }, [actions.review]);
    const handleMinimumScore = useCallback((event) => {
        event.preventDefault();
        actions.review.setMinimumScore(parseInt(event.currentTarget.value));
    }, [actions.review]);
    const handleRelease = useCallback((approve, release) => () => {
        return actions.review.releaseAll({ approve, release });
    }, [actions.review]);
    if (!canRelease) {
        return null;
    }
    return (React.createElement("div", { className: "input-group" },
        React.createElement(FormInput, { type: "number", prepend: "Set minimum score", name: "score", onChange: handleMinimumScore },
            React.createElement("div", { className: "input-group-append" },
                React.createElement(DynamicButton, { text: "Approve all", color: Color.GRAY, type: ButtonType.OUTLINE, onClick: handleRelease(true, false) })),
            React.createElement("div", { className: "input-group-append" },
                React.createElement(DynamicButton, { text: "Release all", color: Color.GRAY, type: ButtonType.OUTLINE, onClick: handleRelease(false, true) })))));
};
export default Release;

import React from "react";
import { Color } from "../../Helpers";
import { useActions, useAppState } from "../../overmind";
import Button, { ButtonType } from "../admin/Button";
const MarkReadyButton = ({ review }) => {
    const allCriteriaGraded = useAppState((state) => state.review.graded === state.review.criteriaTotal);
    const actions = useActions();
    const ready = review.ready;
    const handleMarkReady = React.useCallback(() => {
        if (allCriteriaGraded || ready) {
            actions.review.updateReady(!ready);
        }
    }, [actions.review, allCriteriaGraded, ready]);
    if (ready) {
        return (React.createElement(Button, { text: "Mark in Progress", color: Color.YELLOW, type: ButtonType.BADGE, className: "float-right", onClick: handleMarkReady }));
    }
    return (React.createElement(Button, { text: "Mark Ready", color: Color.GREEN, type: ButtonType.BUTTON, className: allCriteriaGraded ? "" : "disabled", onClick: handleMarkReady }));
};
export default MarkReadyButton;

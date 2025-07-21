import React from "react"
import { Review } from "../../../proto/qf/types_pb"
import { Color } from "../../Helpers"
import { useActions, useAppState } from "../../overmind"
import Button, { ButtonType } from "../admin/Button"


const MarkReadyButton = ({ review }: { review: Review }) => {
    const allCriteriaGraded = useAppState((state) => state.review.graded === state.review.criteriaTotal)
    const actions = useActions()
    const ready = review.ready

    const handleMarkReady = React.useCallback(() => {
        if (allCriteriaGraded || ready) {
            actions.review.updateReady(!ready)
        }
    }, [actions.review, allCriteriaGraded, ready])

    if (ready) {
        return (
            <Button
                text="Mark in Progress"
                color={Color.YELLOW}
                type={ButtonType.BADGE}
                className="float-right"
                onClick={handleMarkReady}
            />
        )
    }

    return (
        <Button
            text="Mark Ready"
            color={Color.GREEN}
            type={ButtonType.BUTTON}
            className={allCriteriaGraded ? "" : "disabled"}
            onClick={handleMarkReady}
        />
    )
}

export default MarkReadyButton

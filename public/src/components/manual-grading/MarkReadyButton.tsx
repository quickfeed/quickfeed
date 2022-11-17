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
    }, [allCriteriaGraded, ready])

    return (
        <Button onclick={handleMarkReady}
            classname={ready ? "float-right" : allCriteriaGraded ? "" : "disabled"}
            text={ready ? "Mark In progress" : "Mark Ready"}
            color={ready ? Color.YELLOW : Color.GREEN}
            type={ready ? ButtonType.BADGE : ButtonType.BUTTON}
        />
    )
}

export default MarkReadyButton

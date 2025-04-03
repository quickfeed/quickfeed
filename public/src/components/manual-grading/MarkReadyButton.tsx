import React from "react"
import { Color } from "../../Helpers"
import { useActions } from "../../overmind"
import Button, { ButtonType } from "../admin/Button"


const MarkReadyButton = ({ allCriteriaGraded, ready }: { allCriteriaGraded: boolean, ready: boolean }) => {
    const actions = useActions()

    const handleMarkReady = React.useCallback(() => {
        if (allCriteriaGraded || ready) {
            actions.review.updateReady(!ready)
        }
    }, [allCriteriaGraded, ready])

    return (
        <Button
            text={ready ? "Mark In progress" : "Mark Ready"}
            color={ready ? Color.YELLOW : Color.GREEN}
            type={ButtonType.BUTTON}
            className={allCriteriaGraded ? "" : "disabled"}
            onClick={handleMarkReady}
        />
    )
}

export default MarkReadyButton

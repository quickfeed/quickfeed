import React, { useCallback, useEffect } from "react"
import { Color } from "../Helpers"
import { useActions, useAppState } from "../overmind"
import { ButtonType } from "./admin/Button"
import DynamicButton from "./DynamicButton"
import FormInput from "./forms/FormInput"

const Release = () => {
    const state = useAppState()
    const actions = useActions()
    const canRelease = state.review.assignmentID > -1

    useEffect(() => {
        return () => actions.review.setMinimumScore(0)
    }, [actions.review])

    const handleMinimumScore = useCallback((event: React.FormEvent<HTMLInputElement>) => {
        event.preventDefault()
        actions.review.setMinimumScore(parseInt(event.currentTarget.value))
    }, [actions.review])

    const handleRelease = useCallback((approve: boolean, release: boolean) => () => {
        return actions.review.releaseAll({ approve, release })
    }, [actions.review])

    if (!canRelease) {
        return null
    }

    return (
        <div className="input-group">
            <FormInput type="number" prepend="Set minimum score" name="score" onChange={handleMinimumScore}>
                <div className="input-group-append">
                    <DynamicButton
                        text="Approve all"
                        color={Color.GRAY}
                        type={ButtonType.OUTLINE}
                        onClick={handleRelease(true, false)}
                    />
                </div>
                <div className="input-group-append">
                    <DynamicButton
                        text="Release all"
                        color={Color.GRAY}
                        type={ButtonType.OUTLINE}
                        onClick={handleRelease(false, true)}
                    />
                </div>
            </FormInput>
        </div>
    )
}

export default Release

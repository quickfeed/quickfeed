import React, { useEffect } from "react"
import { useActions, useAppState } from "../overmind"
import FormInput from "./forms/FormInput"
import DynamicButton from "./DynamicButton"
import { Color } from "../Helpers"
import { ButtonType } from "./admin/Button"

const Release = () => {
    const state = useAppState()
    const actions = useActions()
    const canRelease = state.review.assignmentID > -1

    useEffect(() => {
        return () => actions.review.setMinimumScore(0)
    }, [state.review.assignmentID])

    const handleMinimumScore = (event: React.FormEvent<HTMLInputElement>) => {
        event.preventDefault()
        actions.review.setMinimumScore(parseInt(event.currentTarget.value))
    }

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
                        onClick={() => actions.review.releaseAll({ approve: true, release: false })}
                    />
                </div>
                <div className="input-group-append">
                    <DynamicButton
                        text="Release all"
                        color={Color.GRAY}
                        type={ButtonType.OUTLINE}
                        onClick={() => actions.review.releaseAll({ approve: false, release: true })}
                    />
                </div>
            </FormInput>
        </div>
    )
}

export default Release

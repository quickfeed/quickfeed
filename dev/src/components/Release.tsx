import React, { useEffect } from "react"
import { useActions, useAppState } from "../overmind"
import FormInput from "./forms/FormInput"


const Release = (): JSX.Element => {
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

    return (
        <div className="col">
            {canRelease ?
                <div className="input-group">
                    <FormInput type="number" prepend="Set minimum score" name="score" onChange={handleMinimumScore}>
                        <div className="input-group-append">
                            <button className="btn btn-outline-secondary" onClick={() => actions.review.releaseAll({ approve: true, release: false })}>Approve all</button>
                        </div>
                        <div className="input-group-append">
                            <button className="btn btn-outline-secondary" onClick={() => actions.review.releaseAll({ approve: false, release: true })}>Release all</button>
                        </div>
                    </FormInput>
                </div>
                : "Select an assignment by clicking in the table header to release submissions."}
        </div>
    )
}

export default Release

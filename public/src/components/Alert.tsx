import React, { useEffect } from "react"
import { useOvermind } from "../overmind"

/* This component displays all alerts found in state.alerts */
export const Alert = () => {
    const { state, actions } = useOvermind()
    useEffect(() => {

    }, [state.alerts])

    const alerts = state.alerts.map((alert, index) => {
        return (
            <div className="alert alert-danger" role="alert" style={{marginTop: "20px"}} onClick={() => actions.popAlert(index)}>{alert}</div>
        )
    })

    return (
        <div>
            {alerts}
        </div>
    )
}

export default Alert
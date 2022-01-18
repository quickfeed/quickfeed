import React from "react"
import { useAppState, useActions } from "../overmind"


/* This component displays all alerts found in state.alerts */
const Alert = (): JSX.Element => {
    const state = useAppState()
    const actions = useActions()

    /* Index is used to remove the alert from the state.alerts array */
    const alerts = state.alerts.map((alert, index) => {
        return (
            <div
                key={index}
                className={`alert alert-${alert.color}`}
                role="alert" style={{ marginTop: "20px" }}
                onClick={() => actions.popAlert(index)}>
                {alert.text}
            </div>
        )
    })
    return <div>{alerts}</div>
}

export default Alert

import React from "react"
import { useAppState } from "../../overmind"
import Alert from "./Alert"


/* This component displays all alerts found in state.alerts */
const Alerts = (): JSX.Element => {
    const state = useAppState()
    /* Index is used to remove the alert from the state.alerts array */
    const alerts = state.alerts.map((alert) => {
        return <Alert alert={alert} key={alert.id} />
    })

    return <div>{alerts}</div>
}

export default Alerts

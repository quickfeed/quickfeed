import React from "react"
import { useAppState } from "../../overmind"
import Alert from "./Alert"


/* This component displays all alerts found in state.alerts */
const Alerts = () => {
    const state = useAppState()
    const alerts = state.alerts.map((alert) => {
        return <Alert alert={alert} key={alert.id} />
    })

    return <div className="mb-3">{alerts}</div>
}

export default Alerts

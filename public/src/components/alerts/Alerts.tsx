import React from "react"
import { useAppState } from "../../overmind"
import Alert from "./Alert"


/* This component displays all alerts as toast-style notifications.
 * Alerts appear in a fixed position at the bottom-right of the viewport.
 * This component should be rendered once at the app root level. */
const Alerts = () => {
    const state = useAppState()
    const alerts = state.alerts.map((alert) => {
        return <Alert alert={alert} key={alert.id} />
    })

    return <div className="alerts-container">{alerts}</div>
}

export default Alerts

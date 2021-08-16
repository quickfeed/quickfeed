import React, { useEffect } from "react"
import { AlertType } from "../Helpers"
import { useAppState, useActions } from "../overmind"

/* This component displays all alerts found in state.alerts */
export const Alert = () => {
    const state = useAppState()
    const actions = useActions()
    const alerts = state.alerts.map((alert, index) => {
        return  <div
                    key={index} 
                    className={`alert alert-${AlertType[alert.type].toLowerCase()}`} 
                    role="alert" style={{marginTop: "20px"}} 
                    onClick={() => actions.popAlert(index)}>
                    {alert.text}
                </div>
    })
    return <div>{alerts}</div>
}

export default Alert
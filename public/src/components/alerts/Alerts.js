import React from "react";
import { useAppState } from "../../overmind";
import Alert from "./Alert";
const Alerts = () => {
    const state = useAppState();
    const alerts = state.alerts.map((alert) => {
        return React.createElement(Alert, { alert: alert, key: alert.id });
    });
    return React.createElement("div", null, alerts);
};
export default Alerts;

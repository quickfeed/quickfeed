import React, { useState } from "react";
import { ConnStatus } from "../../Helpers";
import { useActions, useAppState } from "../../overmind";
const StreamStatus = () => {
    const status = useAppState((state) => state.connectionStatus);
    const isLoggedIn = useAppState((state) => state.isLoggedIn);
    const reconnect = useActions().global.startSubmissionStream;
    const [open, setOpen] = useState(false);
    if (!isLoggedIn) {
        return null;
    }
    const handleMouseEnter = () => {
        setOpen(true);
    };
    const handleMouseLeave = () => {
        setOpen(false);
    };
    const handleOnClick = () => {
        reconnect();
    };
    const streamStarter = open ? React.createElement("i", { className: "fa fa-repeat fa-stack-1x ", onMouseLeave: handleMouseLeave, onClick: handleOnClick }) : null;
    switch (status) {
        case ConnStatus.CONNECTED:
            return React.createElement("i", { className: "fa fa-circle text-success", title: "Connected" });
        case ConnStatus.RECONNECTING:
            return (React.createElement("span", { className: "fa-stack fa-lg2" },
                React.createElement("i", { className: "fa fa-circle fa-stack-1x text-warning", title: "Attempting to re-establish stream connection", onMouseEnter: handleMouseEnter }),
                streamStarter));
        default:
            return (React.createElement("span", { className: "fa-stack" },
                React.createElement("i", { className: "fa fa-circle fa-stack-1x text-danger", title: "No stream connection", onMouseEnter: handleMouseEnter }),
                streamStarter));
    }
};
export default StreamStatus;

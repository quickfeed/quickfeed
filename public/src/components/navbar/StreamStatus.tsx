import React, { useState } from "react"
import { ConnStatus } from "../../Helpers"
import { useActions, useAppState } from "../../overmind"


const StreamStatus = () => {
    const status = useAppState((state) => state.connectionStatus)
    const isLoggedIn = useAppState((state) => state.isLoggedIn)
    const reconnect = useActions().startSubmissionStream
    const [open, setOpen] = useState<boolean>(false)

    if (!isLoggedIn) {
        // Don't show stream status if not logged in
        return null
    }

    const handleMouseEnter = () => {
        setOpen(true)
    }

    const handleMouseLeave = () => {
        setOpen(false)
    }

    const handleOnClick = () => {
        // Attempt to reconnect
        reconnect()
    }

    const streamStarter = open ? <i className="fa fa-repeat fa-stack-1x " onMouseLeave={handleMouseLeave} onClick={handleOnClick} /> : null

    // Show stream status based on connection status
    switch (status) {
        case ConnStatus.CONNECTED:
            return <i className="fa fa-circle text-success pl-2" title="Connected" />
        case ConnStatus.RECONNECTING:
            return (
                <span className="fa-stack fa-lg pl-2">
                    <i className="fa fa-circle fa-stack-1x text-warning" title="Attempting to re-establish stream connection" onMouseEnter={handleMouseEnter} />
                    {streamStarter}
                </span>
            )
        default:
            return (
                <span className="fa-stack pl-2">
                    <i className="fa fa-circle fa-stack-1x text-danger" title="No stream connection" onMouseEnter={handleMouseEnter} />
                    {streamStarter}
                </span>
            )
    }
}

export default StreamStatus

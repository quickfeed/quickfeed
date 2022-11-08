import React from "react"
import { ConnStatus } from "../../Helpers"
import { useAppState } from "../../overmind"


const StreamStatus = () => {
    const status = useAppState((state) => state.connectionStatus)

    switch (status) {
        case ConnStatus.CONNECTED:
            return <i className="fa fa-circle text-success pl-2" title="Connected" />
        case ConnStatus.RECONNECTING:
            return <i className="fa fa-circle text-warning pl-2" title="Attempting to re-establish stream connection" />
        default:
            return <i className="fa fa-circle text-danger pl-2" title="No stream connection" />
    }
}

export default StreamStatus
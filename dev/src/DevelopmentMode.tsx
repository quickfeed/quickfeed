import React from "react"
import { MockGrpcManager } from "./MockGRPCManager"
import { useActions, useGrpc } from "./overmind"


// DevelopmentButtons contain functionality to save the current state
// and to switch between different signed in users.
// This is only used for development purposes.
// NOTE: this only works with the mocked gRPC manager.
export const DevelopmentMode = () => {
    const actions = useActions()
    const effects = useGrpc()

    if (!(effects.grpcMan instanceof MockGrpcManager)) {
        return null
    }

    const setUser = async (id: string) => {
        (effects.grpcMan as MockGrpcManager).setCurrentUser(Number(id))
        actions.resetState()
        await actions.fetchUserData()
    }

    return (
        <div style={{
            position: "fixed",
            bottom: "0",
            zIndex: "1000",
            width: "15.6rem",
        }}>
            <span className="badge badge-danger" style={{ width: "100%" }}>
                Development Mode
            </span>
            <select className="form-control" onChange={(e) => setUser(e.target.value)}>
                {effects.grpcMan.getMockedUsers().getUsersList().map((user) => (
                    <option key={user.getId()} value={user.getId()}>{user.getName()}</option>
                ))}
            </select>
        </div>
    )
}

export default DevelopmentMode

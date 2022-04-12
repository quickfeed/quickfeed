import React from "react"
import { useActions, useGrpc } from "./overmind"


// DevelopmentButtons contain functionality to save the current state
// and to switch between different signed in users.
// This is only used for development purposes.
// NOTE: this only works with the mocked gRPC manager.
export const DevelopmentButtons = () => {
    const actions = useActions()
    const effects = useGrpc()

    const setUser = async (id: string) => {
        effects.grpcMan.setCurrentUser(Number(id))
        actions.resetState()
        await actions.fetchUserData()
    }

    return (
        <div>
            {/*<button className="btn btn-primary" onClick={() => actions.saveState()}>Save State</button>*/}
            <select className="form-control" onChange={(e) => setUser(e.target.value)}>
                {effects.grpcMan.getMockedUsers().getUsersList().map((user) => (
                    <option key={user.getId()} value={user.getId()}>{user.getName()}</option>
                ))}
            </select>
        </div>
    )
}

export default DevelopmentButtons

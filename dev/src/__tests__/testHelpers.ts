import { MockGrpcManager } from "../MockGRPCManager"
import { config } from "../overmind"
import { State } from "../overmind/state"
import { ReviewState } from "../overmind/namespaces/review/state"
import { createOvermindMock } from "overmind"


// initializeOvermind creates a mock overmind instance with the given state.
// The returned overmind instance also has the grpcMan set to a MockGrpcManager
// NOTE: Directly setting derived values in the state is not supported.
export const initializeOvermind = (state: Partial<State>, reviewState?: Partial<ReviewState>, userID?: number) => {
    const overmind = createOvermindMock(config, {
        grpcMan: userID ? new MockGrpcManager(userID) : new MockGrpcManager(),
    }, initialState => {
        Object.assign(initialState, state)
        if (reviewState) {
            Object.assign(initialState.review, reviewState)
        }
    })
    return overmind
}

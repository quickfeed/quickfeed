import { MockGrpcManager } from "../MockGRPCManager"
import { GrpcManager } from "../GRPCManager"
import { StreamService } from "../streamService"


// Effects should contain all impure functions used to manage state.
export const grpcMan: GrpcManager | MockGrpcManager = (() => {
    let grpcMan: GrpcManager | MockGrpcManager
    if ((process.env.NODE_ENV === "development" || process.env.NODE_ENV === "test") && window.location.hostname === "localhost") {
        // If in development or test mode, and the hostname is localhost, use the MockGrpcManager.
        grpcMan = new MockGrpcManager()
    } else {
        // Otherwise, use the real gRPC manager.
        grpcMan = new GrpcManager()
    }
    return grpcMan
})()

export const streamService = new StreamService()
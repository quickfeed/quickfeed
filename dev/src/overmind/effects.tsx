import { GrpcManager } from "../GRPCManager"
import { MockGrpcManager } from "../MockGRPCManager"

// Effects should contain all impure functions used to manage state.
//export const grpcMan = new GrpcManager()
export const grpcMan = new MockGrpcManager()

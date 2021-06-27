import { Self } from "./state";
import { GrpcManager } from "../GRPCManager";

// Effects should contain all impure functions used to manage state.

export const grpcMan = new GrpcManager()

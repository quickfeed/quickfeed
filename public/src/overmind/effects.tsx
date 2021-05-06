import {Self} from "./state";

import { AutograderServiceClient } from "../proto/AgServiceClientPb";
// import { Void, User, Course, Submissions, SubmissionRequest, Enrollments, EnrollmentRequest, EnrollmentStatusRequest } from "../proto/ag_pb";
import { GrpcManager } from "../GRPCManager";

const AgService = new AutograderServiceClient('https://' + window.location.hostname, null, null)
// Effects should contain all impure functions used to manage state.

export const grpcMan = new GrpcManager()

export const api = {
    // getUser requests your user data (session key sent in request) and returns a User object if you are logged in.
    getUser: async (): Promise<Self> => {
        const resp = await fetch("https://" + window.location.host + "/api/v1/user")
        return resp.json()
    },
}
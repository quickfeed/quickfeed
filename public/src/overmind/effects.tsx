import {IUser, State, state} from './state'

import { AutograderServiceClient } from '../proto/AgServiceClientPb'
import { Void, User, Course, Submissions, SubmissionRequest, Enrollments, EnrollmentRequest, EnrollmentStatusRequest } from '../proto/ag_pb'
import * as grpcWeb from 'grpc-web'
import { GrpcManager } from '../GRPCManager'

const AgService = new AutograderServiceClient('https://' + window.location.hostname, null, null)
// Effects should contain all impure functions used to manage state.

export const grpcMan = new GrpcManager()

export const api = {
    
    // TODO:
    // Could structure this into either separate exports, ex. 'export const course_api' and 'export const user_api'
    // or 'export const api { course: { ... functions related to course ... }, user: { ... functions related to user ... }}'

    // getUser requests your user data (session key sent in request) and returns a User object if you are logged in.
    getUser: async (): Promise<IUser> => {
        const resp = await fetch('https://' + window.location.host + '/api/v1/user')
        return resp.json()
    },

    // Returns all users from the server
    getUsers: async (state: State): Promise<Array<User>> => {
            return (await AgService.getUsers(new Void(), {'user': state.Metadata.user})).getUsersList()       
    },

    // Returns all courses from the server
    getCourses: async (state: State): Promise<Course[]> => {
        return (await AgService.getCourses(new Void(), {'user': state.Metadata.user})).getCoursesList()
    },
    getSubmissions: async (state: State, courseID: number, userID: number): Promise<Submissions> => {
        const request = new SubmissionRequest()
        request.setUserid(userID)
        request.setCourseid(courseID)
        return (await AgService.getSubmissions(request, {'user': state.Metadata.user}))
    },
    updateUser: async (state: State, user: User): Promise<Void> => {
        return (await AgService.updateUser(user, {'user': state.user.id.toString()}))
    },
    getEnrollmentsByUser: async (state: State, courseId: number): Promise<Enrollments> => {
        const request = new EnrollmentStatusRequest()
        request.setUserid(state.user.id)
        return (await AgService.getEnrollmentsByUser(request, {'user': state.Metadata.user}))
    }
}
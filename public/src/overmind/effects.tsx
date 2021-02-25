import { Context} from 'overmind'
import {Todo, IUser, State, state} from "./state";

import axios from "axios";
import { AutograderServiceClient } from "../proto/AgServiceClientPb";
import { Void, User, Users, Course } from "../proto/ag_pb";
import * as grpcWeb from "grpc-web"

const AgService = new AutograderServiceClient("https://" + window.location.hostname, null, null);
// Effects should contain all impure functions used to manage state.

export const api = {

    // TODO:
    // Could structure this into either separate exports, ex. 'export const course_api' and 'export const user_api'
    // or 'export const api { course: { ... functions related to course ... }, user: { ... functions related to user ... }}'

    getTodos: async (): Promise<Object> => {
        const response = await axios.get('https://jsonplaceholder.typicode.com/todos')
        return response.data
    },
    // getUser requests your user data (session key sent in request) and returns a User object if you are logged in.
    getUser: async (): Promise<IUser> => {
        const resp = await fetch("https://" + window.location.host + "/api/v1/user")
        return resp.json()
    },

    // Returns all users from the server
    getUsers: async (state: State): Promise<Array<User>> => {
            return (await AgService.getUsers(new Void(), {'user': state.Metadata.user})).getUsersList()       
    },

    // Returns all courses from the server
    getCourses: async (state: State): Promise<Course[]> => {
        return (await AgService.getCourses(new Void(), {'user': state.Metadata.user})).getCoursesList()
    }
}
import { Context, Action } from "overmind";
import { Courses, Course, User, EnrollmentStatusRequest, Enrollment, Status, Submissions } from "../proto/ag_pb";
import { useEffects } from ".";
import { state } from "./state";
import { useEffect } from "react";
import { resolve } from "url";


export const getUser: Action<void, Promise<boolean>> = ({state, effects}) => {
    return effects.api.getUser()
    .then((user) => {
        if (user.id === undefined) {
            return false
        }
        state.user = user;
        effects.grpcMan.setUserid(state.user.id.toString())
        return true
    })
    
}
export const getUsers: Action<void> = ({state, effects}) => {
    state.users = []
    effects.grpcMan.getUsers().then(res => {
        if (res.data) {
            state.users = res.data.getUsersList()
        }
    })
}

export const getCourses: Action<void, Promise<boolean>> = ({state, effects}) => {
    state.courses = []
    return effects.grpcMan.getCourses().then(res => {
        if (res.data) {
            state.courses = res.data.getCoursesList()
            return true
        }
        return false
    })
}

export const setTheme: Action<void> = ({state}) => {
    let theme = window.localStorage.getItem("theme")
    state.theme = (theme === null) ? "light" : theme

}

export const changeTheme: Action<void> = ({state}) => {
    state.theme = (state.theme === "light") ? "dark" : "light"
}

export const getSubmissions: Action<number> = ({state, effects}, courseID) => {
    effects.grpcMan.getSubmissions(courseID, state.user.id).then(res => {
        console.log(state.user.id, courseID)
        if (res.data) {
            state.submissions = res.data.getSubmissionsList()
        }

    })
}

export const getEnrollmentsByUser: Action<void, Promise<boolean>> = ({state, effects}) => {
    return effects.grpcMan.getEnrollmentsByUser(state.user.id)
    .then(res => {
        if (res.data) {
            state.enrollments = res.data.getEnrollmentsList()
            return true
        }
        return false
    })
}

export const changeUser: Action<User> = ({state, actions, effects}, user) => {
    user.setIsadmin(state.user.isadmin)
    user.setAvatarurl(state.user.avatarurl)
    effects.api.updateUser(state, user).then(response => {
        console.log(response)
        actions.getUser()
    })
}

export const getEnrollmentByCourseId: Action<number, Enrollment | null> = ({state}, courseID) => {
    let enrol: Enrollment | null = null
    state.enrollments.forEach(enrollment => {
        if (enrollment.getCourseid() === courseID) {
            enrol = enrollment
        }
    })
    return enrol
}

export const getAssignments: Action<void> = ({state, effects}) => {
    state.enrollments.forEach(enrollment => {
        console.log(enrollment.getCourseid())
        effects.grpcMan.getAssignments(enrollment.getCourseid()).then(res => {
            if (res.data) {
                state.assignments = res.data.getAssignmentsList()
                console.log(state.assignments)
            }
        })
    })
}
import { Context, Action } from "overmind";
import { Courses, Course, User, EnrollmentStatusRequest, Enrollment, Status, Submissions, Assignment } from "../proto/ag_pb";
import { useEffects } from ".";
import { state } from "./state";
import { useEffect } from "react";
import { resolve } from "url";

/** Fetches and stores an authenticated user in state */
export const getUser: Action<void, Promise<boolean>> = ({state, effects}) => {
    return effects.api.getUser()
    .then((user) => {
        console.log("Fetching.")
        if (user.id === undefined) {
            return false
        }
        state.user = user;
        effects.grpcMan.setUserid(state.user.AccessToken)
        return true
    })
    
}

/** Fetches all users */
export const getUsers: Action<void> = ({state, effects}) => {
    state.users = []
    effects.grpcMan.getUsers().then(res => {
        if (res.data) {
            state.users = res.data.getUsersList()
        }
    })
}

/** Fetches all courses */
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

/** Tries to get saved theme setting from localStorage, else sets theme to Light by default */
export const setTheme: Action<void> = ({state}) => {
    let theme = window.localStorage.getItem("theme")
    state.theme = (theme === null) ? "light" : theme

}

/** Changes between Light and Dark theme */
export const changeTheme: Action<void> = ({state}) => {
    state.theme = (state.theme === "light") ? "dark" : "light"
}


/** Gets all submission for the current user by Course ID and stores them in state */
export const getSubmissions: Action<number> = ({state, effects}, courseID) => {
    effects.grpcMan.getSubmissions(courseID, state.user.id).then(res => {
        console.log(state.user.id, courseID)
        if (res.data) {
            state.submissions[courseID] = res.data.getSubmissionsList()
            console.log("Hey submissions is happening")
        }
        state.submissions[courseID]

    })
    /* TODO implement getting submission from grouplabs
    effects.grpcMan.getGroupByUserAndCourse(courseID,state.user.id).then(res =>{
        if (res.data){
            state.submissions[courseID]
        }
    })*/
}


/** Gets all enrollments for the current user and stores them in state */
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

/** Changes user information server-side */
export const changeUser: Action<User> = ({state, actions, effects}, user) => {
    user.setIsadmin(state.user.isadmin)
    user.setAvatarurl(state.user.avatarurl)
    effects.api.updateUser(state, user).then(response => {
        console.log(response)
        actions.getUser()
    })
}

/** Gets a specific enrollment for a given course by the course ID if the user has an enrollment for that course. Returns null if none found */
export const getEnrollmentByCourseId: Action<number, Enrollment | null> = ({state}, courseID) => {
    let enrol: Enrollment | null = null
    state.enrollments.forEach(enrollment => {
        if (enrollment.getCourseid() === courseID) {
            enrol = enrollment
        }
    })
    return enrol
}

/** TODO: Either store assignments for all courses, or get assignments by course ID. Currently sets state.assignments to the assignments in the last enrollment in state.enrollments */
export const getAssignments: Action<void> = ({state, effects}) => {
    let assignments: { [courseID: number] : Assignment[]} = {}
    state.enrollments.forEach(enrollment => {
        //console.log(enrollment.getCourseid())
        effects.grpcMan.getAssignments(enrollment.getCourseid()).then(res => {
            if (res.data) {
                assignments[enrollment.getCourseid()] = res.data.getAssignmentsList()
                //console.log(state.assignments)
            }
            
        })
        .finally(() => {
            state.assignments = assignments
        })
    })
}
/** Gets the assignments from a course by the course id. Meant to be used in places where you want only 1 assignment list. */
export const getAssignmentsByCourse: Action<number, Promise<boolean>> = ({state, effects}, courseid:number) => {
    return effects.grpcMan.getAssignments(courseid).then(res => {
        if (res.data){
            state.assignments[courseid] = res.data.getAssignmentsList()
            return true
        }
        return false
    })
}
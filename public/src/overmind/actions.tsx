import { Action, AsyncAction } from "overmind";
import { useActions } from ".";
import { IGrpcResponse } from "../GRPCManager";
import {  User, Enrollment, Assignment, Submission, Repository } from "../proto/ag_pb";
import { CourseGroup } from "./state";


/** Fetches and stores an authenticated user in state */
export const getUser: AsyncAction<void, boolean> = async ({state, effects}) => {
    return await effects.api.getUser()
    .then(async (user) => {
        console.log('Fetching.')
        if (user.id === undefined) {
            return false
        }
        state.user = user
        await effects.grpcMan.setUserid(state.user.AccessToken)
        return true
    })
    
}

export const getPerson: AsyncAction<void, boolean> = async ({ state, effects }) => {
    const user = await effects.api.getUser()
    if (user.id === undefined) {
        console.log("No user ID")
        return false
    }
    console.log("Fetched user")
    state.user = user
    effects.grpcMan.setUserid(state.user.AccessToken)
    return true
}



/** Fetches all users */
export const getUsers: AsyncAction<void> = async ({state, effects}) => {
    state.users = []
    await effects.grpcMan.getUsers().then(res => {
        if (res.data) {
            console.log(res.data)
        }
    })
}

/** Fetches all courses */
export const getCourses: AsyncAction<void, boolean> = async ({state, effects}) => {
    state.courses = []
    return await effects.grpcMan.getCourses().then(res => {
        if (res.data) {
            state.courses = res.data.getCoursesList()
            console.log("Fetched courses")
            return true
        }
        return false
    })
}

/**  */
export const getCoursesByUser: AsyncAction<void> = async ({state, effects}) => {
    let statuses: Enrollment.UserStatus[] = []
    await effects.grpcMan.getCoursesByUser(state.user.id, statuses).then(res => {
        if (res.data) {
            console.log(res.data)
            state.userCourses = res.data.getCoursesList()
        }
    })

}

/** Gets all submission for the current user by Course ID and stores them in state */
export const getSubmissions: AsyncAction<number, Boolean> = async ({state, effects}, courseID) => {
    return await effects.grpcMan.getSubmissions(courseID, state.user.id).then(res => {
        console.log(state.user.id, courseID)
        if (res.data) {
            state.submissions[courseID] = res.data.getSubmissionsList()
        }
        return false
        
    })
}


/** Gets all enrollments for the current user and stores them in state */
export const getEnrollmentsByUser: AsyncAction<void, boolean> = async ({state, effects}) => {
    return await effects.grpcMan.getEnrollmentsByUser(state.user.id)
    .then(res => {
        if (res.data) {
            const enrollments = res.data.getEnrollmentsList()
            state.enrollments = enrollments
            console.log("Fetched enrollments")
            return true
        }
        return false
    })
}


/** Changes user information server-side */
export const changeUser: AsyncAction<User> = async ({state, actions, effects}, user) => {
    user.setAvatarurl(state.user.avatarurl)
    await effects.grpcMan.updateUser(user).then(response => {
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

export const getEnrollmentsByCourse: AsyncAction<{courseID: number, statuses: Enrollment.UserStatus[]}> = async ({state, effects}, value) => {

    state.users = []
    state.courseEnrollments[value.courseID] = []
    await effects.grpcMan.getEnrollmentsByCourse(value.courseID, undefined, undefined, value.statuses).then(res => {
        if (res.data) {
            state.users = res.data.getEnrollmentsList()
            state.courseEnrollments[value.courseID] = res.data.getEnrollmentsList()
            console.log(state.courseEnrollments)
            return true
        }
        return false
    })
}


export const setEnrollmentState: AsyncAction<Enrollment> = async ({state, effects}, enrollment) => {
    let e = new Enrollment()
    e.setCourseid(enrollment.getCourseid())
    e.setUserid(enrollment.getUserid())
    e.setState(enrollment.getState() == Enrollment.DisplayState.VISIBLE ? Enrollment.DisplayState.FAVORITE : Enrollment.DisplayState.VISIBLE)
    state.enrollments.find(e => e.getId() === enrollment.getId())?.setState(e.getState())
    if (e) {
        await effects.grpcMan.updateCourseVisibility(e).then(res => {
            console.log(res)
        })
        .catch(res => {
            console.log(res)
        })
    }

}

/** TODO: Either store assignments for all courses, or get assignments by course ID. Currently sets state.assignments to the assignments in the last enrollment in state.enrollments */
// export const getAssignments: AsyncAction<void> = async ({state, effects}) => {
//         let assignments: { [courseID: number] : Assignment[]} = {}
//         state.enrollments.forEach(async enrollment => {
//         //console.log(enrollment.getCourseid())
//          await effects.grpcMan.getAssignments(enrollment.getCourseid()).then(res => {
//             if (res.data) {
//                 // console.log(enrollment, "load enrolls")
//                 enrollment.getCourse()?.setAssignmentsList(res.data.getAssignmentsList())
//                 assignments[enrollment.getCourseid()] = res.data.getAssignmentsList()
//                 //console.log(state.assignments)
//             }
            
//         }).finally(() => {
//             state.assignments = assignments
//             console.log("Fetched assignments")
//         })
        

//     })
    
// }

export const getAssignments: AsyncAction<number> = async ({ state, effects }, courseID) => {
    await effects.grpcMan.getAssignments(courseID).then(res => {
        if(res.data) {
            state.assignments[courseID] = res.data.getAssignmentsList()
            console.log("Fetched assignment")
        }
    })
}

/** Gets the assignments from a course by the course id. Meant to be used in places where you want only 1 assignment list. */
export const getAssignmentsByCourse: AsyncAction<number, boolean> = async ({state, effects}, courseid) => {
    return await effects.grpcMan.getAssignments(courseid).then(res => {
        if (res.data){
            state.assignments[courseid] = res.data.getAssignmentsList()
            return true
        }
        return false
    })
}

export const getRepository: AsyncAction<void> = async ({state, effects}) => {
    
    state.enrollments.forEach(async enrollment => {
        state.repositories[enrollment.getCourseid()] = {};

        await effects.grpcMan.getRepositories(enrollment.getCourseid(), [Repository.Type.USER, Repository.Type.GROUP, Repository.Type.COURSEINFO, Repository.Type.ASSIGNMENTS]).then(res => {
            if (res.data) {
                const repoMap = res.data.toObject().urlsMap;
                repoMap.forEach(repo => {
                    state.repositories[enrollment.getCourseid()][(Repository.Type as any)[repo[0]]] = repo[1];
                    console.log("Fetched repo")
                });
            }
        })
            .finally();
    });

}



export const updateCourseGroup: Action<CourseGroup> = ({state}, cg) => {
    state.cg = cg
}

export const alertHandler: Action<IGrpcResponse<any>> = ({state}, response) => {
    if (response.status.getCode() >= 0) {
        state.alerts.push(response.status.getError())
    }
}

export const createGroup: Action<number> = ({state, actions, effects}, courseID) => {
    let users: number[] = []
    state.cg.users.forEach(user => {
        users.push(user.getId())
    })
    effects.grpcMan.createGroup(courseID, state.cg.groupName, users)
    .then(res => {
        if (res.data) {
            console.log("Group Creation Success", res.data)
        }
        actions.alertHandler(res)
    })
}

export const popAlert: Action<number> = ({state}, index) => {
    state.alerts = state.alerts.filter((s, i) => i != index)
}



export const loading: Action<void> = ({state}) => {
    state.isLoading = !state.isLoading
}

// Attempt at getting all submissions at once
export const getCourseSubmissions: AsyncAction<number> = async ({state, effects}, courseID) => {
    let userSubmissions: Submission[] = []
    let groupSubmissions: Submission[] = []
    const groupID: number | undefined = state.enrollments.find(enrollment => enrollment.getCourseid() == courseID)?.getGroupid()


    await effects.grpcMan.getGroupSubmissions(courseID, groupID !== undefined ? groupID : -1)
    .then(res => {
        if (res.data) {
            groupSubmissions = res.data.getSubmissionsList()
        }
    })
    .then(() =>
    effects.grpcMan.getSubmissions(courseID, state.user.id)
    .then(res => {
        if (res.data) {
            userSubmissions = res.data.getSubmissionsList()
        }
    })    
    )
    .then(() =>
        // Make magic happen
        {
        state.submissions[courseID] = []
        state.assignments[courseID].forEach(assignment => {
            let submission: Submission | undefined = undefined
            if (assignment.getIsgrouplab()) {
                submission = groupSubmissions.find(submission => submission.getAssignmentid() == assignment.getId())
            }
            else {
                submission = userSubmissions.find(submission => submission.getAssignmentid() == assignment.getId())
            }
            if(submission) {
                state.submissions[courseID][assignment.getOrder() - 1] = submission
            }
        })
        console.log("Fetched submission")
    })
}

export const setActiveCourse: Action<number> = ({state}, courseID) => {
    if(state.activeCourse === courseID) {
        state.activeCourse = -1
    } else {
        state.activeCourse = courseID
    }
}

export const enroll: AsyncAction<number> = async ({state, effects}, courseID) => {
    await effects.grpcMan.createEnrollment(courseID, state.user.id).then(res => {
        console.log(res.status)
    })
    .catch(res => {
        console.log("catch")
    })
}

export const updateSearch: Action<string> = ({state}, search) => {
    state.search = search
}

// EXPERIMENTS BELOW
/** Initializes a student user with all required data */
/** //TODO: Figure out how to await this monster  */
export const setupUser: AsyncAction<void, boolean> = async ({state, actions}) => {
    actions.loading()
    const check = await actions.getPerson()
    .then(async success => {
        console.log("Loading enrollments", success)
        if (success) {
            return await actions.getEnrollmentsByUser()
        }
        return false
    })
    .then(async success => {
        console.log("Loading assignments", success)
        if (success) {
            state.enrollments.forEach(async enrollment => {
                await actions.getAssignments(enrollment.getCourseid())
            })
            console.log("Fetched assignments")
            return true
        }
        return false
    })
    .then(success => {
        console.log("Loading submissions", success)
        if (success) {
            state.enrollments.forEach(async enrollment => {
                await actions.getCourseSubmissions(enrollment.getCourseid())
            });
            return true
        }
        return false
    })
    .then(async success => {
        console.log("Loading repositories", success)
        if (success) {
            await actions.getRepository()
            return true
        }
        return false
    }).then(async success => {
        console.log("Loading courses", success)
        if (success) {
            return await actions.getCourses().then(success => {
                return success
            })

        }
        return false
        
    }).then(async success => {
        if (success) {
            state.enrollments.forEach(async enrollment => {
                let statuses = enrollment.getStatus() === Enrollment.UserStatus.STUDENT ? [Enrollment.UserStatus.STUDENT, Enrollment.UserStatus.TEACHER ] : []
                await actions.getEnrollmentsByCourse({courseID: enrollment.getCourseid(), statuses: statuses})
            })
            return true
        }
        return false
    })
    .finally(() => {
        console.log("Complete")
    })
    return check
}

/* START UTILITY ACTIONS */

/** Tries to get saved theme setting from localStorage, else sets theme to Light by default */
export const setTheme: Action<void> = ({state}) => {
    let theme = window.localStorage.getItem("theme")
    state.theme = (theme === null) ? "light" : theme
    document.body.className = state.theme
}

/** Changes between Light and Dark theme */
export const changeTheme: Action<void> = ({state}) => {
    state.theme = (state.theme === "light") ? "dark" : "light"
    document.body.className = state.theme
    window.localStorage.setItem("theme", state.theme)
}

/** Sets the time to now. */
export const setTimeNow: Action<void> = ({state}) =>{
    state.timeNow = new Date()
}
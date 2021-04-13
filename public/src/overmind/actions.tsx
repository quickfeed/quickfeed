import { Action } from "overmind";
import {  User, Enrollment, Assignment, Submission, Repository } from "../proto/ag_pb";


/** Fetches and stores an authenticated user in state */
export const getUser: Action<void, Promise<boolean>> = ({state, effects}) => {
    return effects.api.getUser()
    .then((user) => {
        console.log('Fetching.')
        if (user.id === undefined) {
            return false
        }
        state.user = user
        effects.grpcMan.setUserid(state.user.AccessToken)
        return true
    })
    
}

/** Fetches all users */
export const getUsers: Action<void> = ({state, effects}) => {
    state.users = []
    effects.grpcMan.getUsers().then(res => {
        if (res.data) {
            console.log(res.data)
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

/**  */
export const getCoursesByUser: Action<void> = ({state, effects}) => {
    let statuses: Enrollment.UserStatus[] = []
    effects.grpcMan.getCoursesByUser(state.user.id, statuses).then(res => {
        if (res.data) {
            console.log(res.data)
            state.userCourses = res.data.getCoursesList()
        }
    })

}

/** Gets all submission for the current user by Course ID and stores them in state */
export const getSubmissions: Action<number, Promise<Boolean>> = ({state, effects}, courseID) => {
    return effects.grpcMan.getSubmissions(courseID, state.user.id).then(res => {
        console.log(state.user.id, courseID)
        if (res.data) {
            state.submissions[courseID] = res.data.getSubmissionsList()
        }
        return false
        
    })
}


/** Gets all enrollments for the current user and stores them in state */
export const getEnrollmentsByUser: Action<void, Promise<boolean>> = async ({state, effects}) => {
    return await effects.grpcMan.getEnrollmentsByUser(state.user.id)
    .then(res => {
        if (res.data) {
            const enrollments = res.data.getEnrollmentsList()
            state.enrollments = enrollments
            return true
        }
        return false
    })
}

/** Changes user information server-side */
export const changeUser: Action<User> = ({state, actions, effects}, user) => {
    user.setAvatarurl(state.user.avatarurl)
    effects.grpcMan.updateUser(user).then(response => {
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

export const getEnrollmentsByCourse: Action<number> = ({state, effects}, courseID) => {
    state.users = []
    effects.grpcMan.getEnrollmentsByCourse(courseID, undefined, undefined, [Enrollment.UserStatus.STUDENT]).then(res => {
        if (res.data) {
            state.users = res.data.getEnrollmentsList()
        }
    })
}

export const setEnrollmentState: Action<Enrollment> = ({state, effects}, enrollment) => {
    let e = new Enrollment()
    e.setCourseid(enrollment.getCourseid())
    e.setUserid(enrollment.getUserid())
    e.setState(enrollment.getState() == Enrollment.DisplayState.VISIBLE ? Enrollment.DisplayState.FAVORITE : Enrollment.DisplayState.VISIBLE)
    state.enrollments.find(e => e.getId() === enrollment.getId())?.setState(e.getState())
    if (e) {
        effects.grpcMan.updateCourseVisibility(e).then(res => {
            console.log(res)
        })
        .catch(res => {
            console.log(res)
        })
    }

}

/** TODO: Either store assignments for all courses, or get assignments by course ID. Currently sets state.assignments to the assignments in the last enrollment in state.enrollments */
export const getAssignments: Action<void> = ({state, effects}) => {
        let assignments: { [courseID: number] : Assignment[]} = {}
        state.enrollments.forEach( enrollment => {
        //console.log(enrollment.getCourseid())
         effects.grpcMan.getAssignments(enrollment.getCourseid()).then(res => {
            if (res.data) {
                console.log(enrollment, "load enrolls")
                enrollment.getCourse()?.setAssignmentsList(res.data.getAssignmentsList())
                assignments[enrollment.getCourseid()] = res.data.getAssignmentsList()
                //console.log(state.assignments)
            }
            
        }).finally(() => {
            state.assignments = assignments
        })
        

    })
    
}
/** Gets the assignments from a course by the course id. Meant to be used in places where you want only 1 assignment list. */
export const getAssignmentsByCourse: Action<number, Promise<boolean>> = ({state, effects}, courseid) => {
    return effects.grpcMan.getAssignments(courseid).then(res => {
        if (res.data){
            state.assignments[courseid] = res.data.getAssignmentsList()
            return true
        }
        return false
    })
}

export const getRepository: Action<void> = ({state, effects}) => {
    
    state.enrollments.forEach(enrollment => {
        state.repositories[enrollment.getCourseid()] = {}    
    
    effects.grpcMan.getRepositories(enrollment.getCourseid(), [Repository.Type.USER, Repository.Type.GROUP, Repository.Type.COURSEINFO, Repository.Type.ASSIGNMENTS]).then(res => {
            if(res.data) {
                const repoMap = res.data.toObject().urlsMap
                repoMap.forEach(repo => {
                    state.repositories[enrollment.getCourseid()][(Repository.Type as any)[repo[0]]] = repo[1]
                })
            }
        })
        .finally(
        )
    });

}

export const loading: Action<void> = ({state}) => {
    state.isLoading = !state.isLoading
}

// Attempt at getting all submissions at once
export const getCourseSubmissions: Action<number> = ({state, effects}, courseID) => {
    let userSubmissions: Submission[] = []
    let groupSubmissions: Submission[] = []
    const groupID: number | undefined = state.enrollments.find(enrollment => enrollment.getCourseid() == courseID)?.getGroupid()


    effects.grpcMan.getGroupSubmissions(courseID, groupID !== undefined ? groupID : -1)
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
    })
}

export const setActiveCourse: Action<number> = ({state}, courseID) => {
    if(state.activeCourse === courseID) {
        state.activeCourse = -1
    } else {
        state.activeCourse = courseID
    }
}

export const enroll: Action<number> = ({state, effects}, courseID) => {
    effects.grpcMan.createEnrollment(courseID, state.user.id).then(res => {
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
export const setupUser: Action<void, Promise<boolean>> = ({state, actions}) => {
    return actions.getUser()
    .then(success => {
        console.log("Loading enrollments", success)
        if (success) {
            return actions.getEnrollmentsByUser()
        }
        return false
    })
    .then(success => {
        console.log("Loading assignments", success)
        if (success) {
            
            actions.getAssignments()
            return true
        }
        return false
    })
    .then(success => {
        console.log("Loading submissions", success)
        if (success) {
            state.enrollments.forEach(enrollment => {
                actions.getCourseSubmissions(enrollment.getCourseid())
            });
            return true
        }
        return false
    })
    .then(success => {
        console.log("Loading repositories", success)
        if (success) {
            actions.getRepository()
            return true
        }
        return false
    }).then(success => {
        console.log("Loading courses", success)
        if (success) {
            return actions.getCourses().then(success => {
                return success
            })

        }
        
        return false
        
    })
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
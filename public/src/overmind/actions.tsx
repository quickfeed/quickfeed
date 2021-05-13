import { Action, AsyncAction } from "overmind";
import { IGrpcResponse } from "../GRPCManager";
import {  User, Enrollment, Assignment, Submission, Repository, Organization, Course, SubmissionsForCourseRequest } from "../../proto/ag_pb";
import { CourseGroup, state } from "./state";


/** Fetches and stores an authenticated user in state */
export const getUser: AsyncAction<void, boolean> = async ({state, effects}) => {
    const user = await effects.api.getUser()
    
    if (user.id) {
        state.user = user
        effects.grpcMan.setUserid(state.user.Token)
        return true
    }
    return false
}

export const getSubmissionCommitHash: Action<{assignmentID: number, courseID: number}> = ({state, actions, effects}, value) => {
    const submission = state.submissions[value.courseID].find(s => s.getAssignmentid() === value.assignmentID)
    if (submission) {
        effects.grpcMan.getSubmissionCommitHash(submission.getId()).then(async res => {
            if (res.data) {
                if (submission.getCommithash() !== res.data.getCommithash()) {
                    await actions.refreshSubmissions({courseID: value.courseID, submissionID: submission.getId()})
                }
            }
        })  
    }
}

export const getPerson: AsyncAction<void, boolean> = async ({ state, effects }) => {
    const user = await effects.api.getUser()
    if (user.id === undefined) {
        return false
    }
    state.user = user
    effects.grpcMan.setUserid(state.user.Token)
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
    const result = await effects.grpcMan.getCourses()
    if (result.data) {
        state.courses = result.data.getCoursesList()
        return true
    }
    return false
}

/**  */
export const getCoursesByUser: AsyncAction<void, boolean> = async ({state, effects}) => {
    let statuses: Enrollment.UserStatus[] = []
    let courses = await effects.grpcMan.getCoursesByUser(state.user.id, statuses)
    if (courses.data) {
        courses.data.getCoursesList().forEach(course => {
            state.userCourses[course.getId()] = course
        })
        return true
    }
    return false
}


/** Gets all submission for the current user by Course ID and stores them in state */
export const getSubmissions: AsyncAction<number, boolean> = async ({state, effects}, courseID) => {
    const result = await effects.grpcMan.getSubmissions(courseID, state.user.id)
    if (result.data) {
        state.submissions[courseID] = result.data.getSubmissionsList()
        return true
    }
    return false
}


/** Gets all enrollments for the current user and stores them in state */
export const getEnrollmentsByUser: AsyncAction<void, boolean> = async ({state, effects}) => {
    const result = await effects.grpcMan.getEnrollmentsByUser(state.user.id)
    if (result.data) {
            const enrollments = result.data.getEnrollmentsList()
            state.enrollments = enrollments
            return true
        }
    return false
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

export const getEnrollmentsByCourse: AsyncAction<{courseID: number, statuses: Enrollment.UserStatus[]}, boolean> = async ({state, effects}, value) => {

    state.users = []
    state.courseEnrollments[value.courseID] = []
    const result = await effects.grpcMan.getEnrollmentsByCourse(value.courseID, undefined, undefined, value.statuses)
    if (result.data) {
            state.users = result.data.getEnrollmentsList()
            state.courseEnrollments[value.courseID] = result.data.getEnrollmentsList()
            return true
    }
    return false
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

export const updateSubmission: AsyncAction<{courseID: number, submission: Submission}> = async ({actions, effects}, value) => {
    const result = await effects.grpcMan.updateSubmission(value.courseID, value.submission)
    if (result.status.getCode() > 0) {
        actions.alertHandler(result)
    }
}

export const test: Action<void, Enrollment.AsObject> = ({state}) => {
    return state.enrollmentsByCourseId[1].toObject()
}

export const updateEnrollment: Action<{enrollment: Enrollment, status: Enrollment.UserStatus}> = ({state, actions, effects}, update) => {
    let e = new Enrollment()
    e.setId(update.enrollment.getId())
    e.setStatus(update.status)
    e.setUserid(update.enrollment.getUserid())
    e.setCourseid(update.enrollment.getCourseid())

    effects.grpcMan.updateEnrollment(e).then(res => {
        if (res.data) {
            // Good
        }
        actions.alertHandler(res)
    })
}

export const getAssignments: AsyncAction<void, boolean> = async ({ state, effects }) => {
    let success = false
    for (const enrollment of state.enrollments) {
        const result = await effects.grpcMan.getAssignments(enrollment.getCourseid())
        if(result.data) {
            state.assignments[enrollment.getCourseid()] = result.data.getAssignmentsList()
            success = true
        }
        success = false
    }
    return success
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

export const getRepositories: AsyncAction<void, boolean> = async ({state, effects}) => {
    let success = true
    for (const enrollment of state.enrollments) {
        state.repositories[enrollment.getCourseid()] = {};

        const result = await effects.grpcMan.getRepositories(enrollment.getCourseid(), [Repository.Type.USER, Repository.Type.GROUP, Repository.Type.COURSEINFO, Repository.Type.ASSIGNMENTS])
        if (result.data) {
            const repoMap = result.data.toObject().urlsMap;
            repoMap.forEach(repo => {
                    state.repositories[enrollment.getCourseid()][(Repository.Type as any)[repo[0]]] = repo[1];
            });
            success = success === false ? false : true
        }
        success = false
    }
    return success

}



export const updateCourseGroup: Action<CourseGroup> = ({state}, cg) => {
    state.courseGroup = cg
}

export const alertHandler: Action<IGrpcResponse<any>> = ({state}, response) => {
    if (response.status.getCode() >= 0) {
        state.alerts.push(response.status.getError())
    }
}

export const alert: Action<string> = ({state}, alertString) => {
    state.alerts.push(alertString)
}

export const getGroupByUserAndCourse: AsyncAction<number> = async ({state, effects}, courseID) => {
    const result = await effects.grpcMan.getGroupByUserAndCourse(courseID, state.user.id)
    if (result.data) {
        state.userGroup[courseID] = result.data
    }
}

export const createGroup: Action<number> = ({state, actions, effects}, courseID) => {
    let users: number[] = []
    state.courseGroup.users.forEach(user => {
        users.push(user.getId())
    })
    effects.grpcMan.createGroup(courseID, state.courseGroup.groupName, users)
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


export const getOrganization: Action<string> = ({actions, effects}, orgName) => {
    effects.grpcMan.getOrganization(orgName).then(res => {
        if (res.data) {
            console.log(res.data)
        }
        actions.alertHandler(res)
    })
}

export const createCourse: Action<{course: Course, orgName: string}> = ({state, actions, effects}, value) => {
    let course = new Course()
    effects.grpcMan.getOrganization(value.orgName).then(res => {
        if (res.data) {
            // TODO: Is there a more elegant way to do this?
            course.setOrganizationid(res.data.getId())
            course.setOrganizationpath(res.data.getPath())
            course.setSlipdays(value.course.getSlipdays())
            course.setTag(value.course.getTag())
            course.setCode(value.course.getCode())
            course.setYear(value.course.getYear())
            course.setName(value.course.getName())
            course.setProvider("github")
            course.setCoursecreatorid(state.user.id)
            effects.grpcMan.createCourse(course).then(res => {
                if (res.data) {
                    state.courses.push(res.data)

                    // success
                }
                actions.alertHandler(res)
            })
        }
        actions.alertHandler(res)
    })
}

export const loading: Action<void> = ({state}) => {
    state.isLoading = !state.isLoading
}

export const refreshSubmissions: AsyncAction<{courseID: number, submissionID: number}> = async ({state, effects}, input) => {
    const result = await effects.grpcMan.getSubmissions(input.courseID, state.user.id)
    if (result.data) {
        let submissions = result.data.getSubmissionsList()
        submissions.forEach(submission => {
            let assignment = state.assignments[input.courseID].find(a => a.getId() === submission.getAssignmentid())
            if (assignment) {
                if (state.submissions[input.courseID][assignment.getOrder() - 1].getCommithash() !== submission.getCommithash()) {
                    state.submissions[input.courseID][assignment.getOrder() - 1] = submission
                }
                    
            }
        })
    }
}

export const getAllCourseSubmissions: AsyncAction<number, boolean> = async ({state, effects}, courseID) => {
    state.courseSubmissions[courseID] = []
    state.isLoading = true
    const result =  await effects.grpcMan.getSubmissionsByCourse(courseID, SubmissionsForCourseRequest.Type.ALL)
    if (result.data) {
            state.courseSubmissions[courseID] = result.data.getLinksList()
            state.isLoading = false
            return true
            
    }
    state.isLoading = false
    return false
}

export const getGroupsByCourse: AsyncAction<number> = async ({state, effects}, courseID) => {
    await effects.grpcMan.getGroupsByCourse(courseID).then(res => {
        if (res.data) {
            state.groups[courseID] = res.data.getGroupsList()
        }
    })
}

// Attempt at getting all submissions at once
export const getCourseSubmissions: AsyncAction<number, boolean> = async ({state, effects}, courseID) => {
    let userSubmissions: Submission[] = []
    let groupSubmissions: Submission[] = []
    const groupID: number | undefined = state.enrollments.find(enrollment => enrollment.getCourseid() == courseID)?.getGroupid()


    const groupResult = await effects.grpcMan.getGroupSubmissions(courseID, groupID !== undefined ? groupID : -1)
    
    if (groupResult.data) {
        groupSubmissions = groupResult.data.getSubmissionsList()
    }

    const userResult = await effects.grpcMan.getSubmissions(courseID, state.user.id)
    if (userResult.data) {
            userSubmissions = userResult.data.getSubmissionsList()
    }    
    if (groupResult.status.getCode() > 0 || userResult.status.getCode() > 0) {
        return false
    }
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
        if(submission == undefined){
            submission = new Submission()
            state.submissions[courseID][assignment.getOrder() - 1] = submission
        }
    })
    return true
}

export const setActiveCourse: Action<number> = ({state}, courseID) => {
    state.activeCourse = courseID
}

export const setActiveLab: Action<number> = ({state}, assignmentID) => {
    state.activeLab = assignmentID
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

export const fetchUserData: AsyncAction<void, boolean> = async ({state, actions}) =>  {
    let success = await actions.getUser()
    if (!success) { state.isLoading = false; return false;}
    while (state.isLoading) {
        success = await actions.getEnrollmentsByUser()
        success = await actions.getAssignments()
        for (const enrollment of state.enrollments) {
            if (enrollment.getStatus() === Enrollment.UserStatus.STUDENT) {
                success = await actions.getCourseSubmissions(enrollment.getCourseid())
            }
            if (enrollment.getStatus() >= 2) {
                let statuses = enrollment.getStatus() === Enrollment.UserStatus.STUDENT ? [Enrollment.UserStatus.STUDENT, Enrollment.UserStatus.TEACHER ] : []
                success = await actions.getEnrollmentsByCourse({courseID: enrollment.getCourseid(), statuses: statuses})
            }
        }
        success = await actions.getRepositories()
        success = await actions.getCourses()
        state.isLoading = false

    }
    return success
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
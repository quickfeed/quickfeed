import { Context } from "overmind";
import { IGrpcResponse } from "../GRPCManager";
import {  User, Enrollment, Submission, Repository, Course, SubmissionsForCourseRequest, Status, CourseSubmissions, SubmissionLink } from "../../proto/ag/ag_pb";
import { CourseGroup, ParsedCourseSubmissions } from "./state";
import { AlertType } from "../Helpers";
import { StatusCode } from "grpc-web";


/** Fetches and stores an authenticated user in state */
export const getSelf = async ({state, effects}: Context) => {
        const user = await effects.grpcMan.getUser()
    
        if (user.data) {
            state.self = user.data
            return true
        } 
    
    return false
}

/**  */
/*
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
*/

/** Fetches all users (requires admin priveleges) */
export const getUsers = async ({state, effects}: Context) => {
    state.users = []
    state.isLoading = true
    const users = await effects.grpcMan.getUsers()
    if (users.data) {
        // Insert users sorted by admin priveleges
        state.allUsers = users.data.getUsersList().sort((a, b) => {
            if(a.getIsadmin() > b.getIsadmin()) { return -1 }
            if(a.getIsadmin() < b.getIsadmin()) { return 1 }
            return 0
        })
    }
    state.isLoading = false
}

/** Fetches all courses */
export const getCourses= async ({state, effects}: Context) => {
    state.courses = []
    const result = await effects.grpcMan.getCourses()
    if (result.data) {
        state.courses = result.data.getCoursesList()
        return true
    }
    return false
}

/**  */
export const getCoursesByUser= async ({state, effects}: Context) => {
    let statuses: Enrollment.UserStatus[] = []
    let courses = await effects.grpcMan.getCoursesByUser(state.self.getId(), statuses)
    if (courses.data) {
        courses.data.getCoursesList().forEach(course => {
            state.userCourses[course.getId()] = course
        })
        return true
    }
    return false
}


/** Gets all submission for the current user by Course ID and stores them in state */
export const getSubmissions= async ({state, effects}: Context, courseID: number) => {
    const result = await effects.grpcMan.getSubmissions(courseID, state.self.getId())
    if (result.data) {
        state.submissions[courseID] = result.data.getSubmissionsList()
        return true
    }
    return false
}


/** Gets all enrollments for the current user and stores them in state */
export const getEnrollmentsByUser = async ({state, effects}: Context) => {
    const result = await effects.grpcMan.getEnrollmentsByUser(state.self.getId())
    if (result.data) {
            const enrollments = result.data.getEnrollmentsList()
            state.enrollments = enrollments
            return true
        }
    return false
}


/** Changes user information server-side */
export const changeUser = async ({actions, effects}: Context, user: User) => {
    const result = await effects.grpcMan.updateUser(user)
    if (result.status.getCode() == 0) {
        actions.getSelf()
    }
}

export const updateAdmin = async ({ effects }: Context, user: User) => {
    let u = new User
    u.setIsadmin(!user.getIsadmin())
    u.setId(user.getId())
    const result = await effects.grpcMan.updateUser(u) 
    if (result.status.getCode() == 0) {
        user.setIsadmin(!user.getIsadmin())
    }
}

/** Gets a specific enrollment for a given course by the course ID if the user has an enrollment for that course. Returns null if none found */
export const getEnrollmentByCourseId = ({state}: Context, courseID: number) => {
    let enrol: Enrollment | null = null
    state.enrollments.forEach(enrollment => {
        if (enrollment.getCourseid() === courseID) {
            enrol = enrollment
        }
    })
    return enrol
}

export const getEnrollmentsByCourse = async ({state, effects}: Context, value: {courseID: number, statuses: Enrollment.UserStatus[]}) => {
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


export const setEnrollmentState = async ({state, effects}: Context, enrollment: Enrollment) => {
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

export const updateSubmission = async ({state, actions, effects}: Context, value: {courseID: number, submission: Submission, userIndex: number,  submissionIndex: number}) => {
    const result = await effects.grpcMan.updateSubmission(value.courseID, value.submission)
    if (result.status.getCode() > 0) {
        actions.alertHandler(result)
    } else {
        let c = state.courseSubmissions[value.courseID][value.userIndex].submissions?.find((e, i) => i === value.submissionIndex)
        if (c) { c.getSubmission()?.setStatus(value.submission.getStatus())}
    }
}

export const updateEnrollment = ({actions, effects}: Context, update: {enrollment: Enrollment, status: Enrollment.UserStatus}) => {
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

export const getAssignments = async ({ state, effects }: Context) => {
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
export const getAssignmentsByCourse = async ({state, effects}: Context, courseID: number) => {
    const res = await effects.grpcMan.getAssignments(courseID)
    if (res.data){
        state.assignments[courseID] = res.data.getAssignmentsList()
        return true
    }
    return false
}

export const getRepositories = async ({state, effects}: Context) => {
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



export const updateCourseGroup = ({state}: Context, cg: CourseGroup) => {
    state.courseGroup = cg
}

export const alertHandler = ({state}: Context, response: IGrpcResponse<any>) => {
    if (response.status.getCode() >= 0) {
        state.alerts.push({text: response.status.getError(), type: AlertType.DANGER})
    }
}

export const alert = ({state}: Context, alert: {text: string, type: AlertType}) => {
    state.alerts.push(alert)
}

export const getGroupByUserAndCourse = async ({state, effects}: Context, courseID: number) => {
    const result = await effects.grpcMan.getGroupByUserAndCourse(courseID, state.self.getId())
    if (result.data) {
        state.userGroup[courseID] = result.data
    }
}

export const createGroup = async ({actions, effects}: Context, group: {courseID: number, users: number[], name: string}) => {
    const res = await effects.grpcMan.createGroup(group.courseID, group.name, group.users)
    actions.alertHandler(res)
}

export const popAlert = ({state}: Context, index: number) => {
    state.alerts = state.alerts.filter((s, i) => i != index)
}


export const getOrganization = async ({actions, effects}: Context, orgName: string) => {
    const res = await effects.grpcMan.getOrganization(orgName)
    actions.alertHandler(res)
}

export const createCourse = async ({state, actions, effects}: Context, value: {course: Course, orgName: string}) => {
    let course = new Course()
    const result = await effects.grpcMan.getOrganization(value.orgName)
    if (result.data) {
        // TODO: Is there a more elegant way to do this?
        course.setOrganizationid(result.data.getId())
        course.setOrganizationpath(result.data.getPath())
        course.setSlipdays(value.course.getSlipdays())
        course.setTag(value.course.getTag())
        course.setCode(value.course.getCode())
        course.setYear(value.course.getYear())
        course.setName(value.course.getName())
        course.setProvider("github")
        course.setCoursecreatorid(state.self.getId())
        const response =  await effects.grpcMan.createCourse(course)
        if (response.data) {
            state.courses.push(response.data)
        }
        actions.alertHandler(response)
    }
    actions.alertHandler(result)
}

export const loading = ({state}: Context) => {
    state.isLoading = !state.isLoading
}

export const refreshSubmissions = async ({state, effects}: Context, input: {courseID: number, submissionID: number}) => {
    const result = await effects.grpcMan.getSubmissions(input.courseID, state.self.getId())
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

const convertCourseSubmission = (data: CourseSubmissions) => {
    // TODO: {[index]: {Parsed}}
    const courseSubmissions: ParsedCourseSubmissions[] = []
    data.getLinksList().forEach((l, index) => {
        courseSubmissions.push({enrollment: l.getEnrollment(), submissions: l.getSubmissionsList(), user: l.getEnrollment()?.getUser()}) 
    })
    return courseSubmissions
}

export const getAllCourseSubmissions = async ({state, effects}: Context, courseID: number) => {
    let success = false

    state.courseSubmissions[courseID] = []
    state.isLoading = true
    const result =  await effects.grpcMan.getSubmissionsByCourse(courseID, SubmissionsForCourseRequest.Type.ALL)
    if (result.data) {
            state.courseSubmissions[courseID] = convertCourseSubmission(result.data)
            state.isLoading = false
            success = true
            
    }
    state.isLoading = false
    return false
}

export const getGroupsByCourse = async ({state, effects}: Context, courseID: number) => {
    /** TODO:
     *  An interesting interaction between gRPC and Overmind
     *  It seems that as the server returns a Group that has no enrollments or users
     *  that an empty array is not created until you try to retrieve it.
     * 
     *  This was experienced when a Group was returned and stored in the state.
     *  When attempting to access Group.getEnrollmentsList for a Group that does not have any enrollments
     *  we get an error implying that a third party is attempting to modify the state.
     * 
     *  Simply calling Group.getEnrollmentsList from an action before saving to the state enables
     *  us to call Group.getEnrollmentsList from components without issue.
     */
    state.groups[courseID] = []
    const res = await effects.grpcMan.getGroupsByCourse(courseID)
    if (res.data) {
        res.data.getGroupsList().forEach(element => {
            if (element.getEnrollmentsList() || element.getUsersList()) {
                // TODO: This serves as a temporary fix for the issue described above.
            }
        });
        state.groups[courseID] = res.data.getGroupsList()
    }

}

export const getUserSubmissions = async ({state, effects}: Context, courseID: number) => {
    state.submissions[courseID] = []
    const res = await effects.grpcMan.getSubmissions(courseID, state.self.getId())
    if (res.data) {
        state.assignments[courseID].forEach(assignment => {
            let submission = res.data?.getSubmissionsList().find(s => s.getAssignmentid() === assignment.getId())
            if (submission) {
                state.submissions[courseID][assignment.getOrder() - 1] = submission
            }
            else {
                state.submissions[courseID][assignment.getOrder() - 1] = new Submission()
            }
        })
        return true
    }
    return false
}

// Attempt at getting all submissions at once
export const getCourseSubmissions = async ({state, effects}: Context, courseID: number) => {
    console.log("Subs get")
    let userSubmissions: Submission[] = []
    let groupSubmissions: Submission[] = []

    const groupID: number = state.enrollmentsByCourseId[courseID].getGroupid()
    //const groupID: number | undefined = state.enrollments.find(enrollment => enrollment.getCourseid() == courseID)?.getGroupid()
    console.log("GRP: ", groupID)

    const groupResult = await effects.grpcMan.getGroupSubmissions(courseID, 0) 
    if (groupResult.status.getCode() > 0 || !groupResult.data) {
        groupSubmissions = []
    }
    else {
        groupSubmissions = groupResult.data.getSubmissionsList()
    }

    const userResult = await effects.grpcMan.getSubmissions(courseID, state.self.getId())
    if (userResult.data) {
        userSubmissions = userResult.data.getSubmissionsList()
    }
    console.log(userResult.status.getCode(), groupResult.status.getCode())
    if (groupResult.status.getCode() > 0 || userResult.status.getCode() > 0) {
        return false
    }
    console.log(userSubmissions, groupSubmissions, courseID)
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
            console.log("Sub")
            state.submissions[courseID][assignment.getOrder() - 1] = submission
        }
        if(submission == undefined){
            submission = new Submission()
            state.submissions[courseID][assignment.getOrder() - 1] = submission
        }
    })
    return true
}

export const setActiveCourse = ({state}: Context, courseID: number) => {
    state.activeCourse = courseID
}

export const setActiveLab = ({state}: Context, assignmentID: number) => {
    state.activeLab = assignmentID
}

export const enroll = async ({state, effects}: Context, courseID: number) => {
    const res = await effects.grpcMan.createEnrollment(courseID, state.self.getId())
}

export const logout = ({state}: Context) => {
    state.self = new User()
}

/** Initializes a student user with all required data */
export const fetchUserData = async ({state, actions}: Context) =>  {
    let success = await actions.getSelf()
    if (!success) { state.isLoading = false; return false;}
    while (state.isLoading) {
        success = await actions.getEnrollmentsByUser()
        success = await actions.getAssignments()
        for (const enrollment of state.enrollments) {
            if (enrollment.getStatus() === Enrollment.UserStatus.STUDENT) {
                //success = await actions.getCourseSubmissions(enrollment.getCourseid())
                success = await actions.getUserSubmissions(enrollment.getCourseid())
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
export const setTheme = ({state}: Context) => {
    let theme = window.localStorage.getItem("theme")
    state.theme = (theme === null) ? "light" : theme
    document.body.className = state.theme
}

/** Changes between Light and Dark theme */
export const changeTheme = ({state}: Context) => {
    state.theme = (state.theme === "light") ? "dark" : "light"
    document.body.className = state.theme
    window.localStorage.setItem("theme", state.theme)
}

/** Sets the time to now. */
export const setTimeNow = ({state}: Context) =>{
    state.timeNow = new Date()
}
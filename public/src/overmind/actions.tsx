import { json, Overmind } from 'overmind'
import { Context } from "./";
import { IGrpcResponse } from "../GRPCManager";
import {  User, Enrollment, Submission, Repository, Course, SubmissionsForCourseRequest, CourseSubmissions, Group, GradingCriterion, Assignment } from "../../proto/ag/ag_pb";
import { CourseGroup, ParsedCourseSubmissions } from "./state";
import { AlertType } from "../Helpers";

/** Fetches and stores an authenticated user in state */
export const getSelf = async ({state, effects}: Context) => {
        const user = await effects.grpcMan.getUser()
    
        if (user.data) {
            state.self = user.data
            return true
        } 
    return false
};

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
};

/** Fetches all courses */
export const getCourses = async ({state, effects}: Context) => {
    state.courses = []
    const result = await effects.grpcMan.getCourses()
    if (result.data) {
        state.courses = result.data.getCoursesList()
        return true
    }
    return false
};

/**  */
export const getCoursesByUser = async ({state, effects}: Context) => {
    let statuses: Enrollment.UserStatus[] = []
    let courses = await effects.grpcMan.getCoursesByUser(state.self.getId(), statuses)
    if (courses.data) {
        courses.data.getCoursesList().forEach(course => {
            state.userCourses[course.getId()] = course
        })
        return true
    }
    return false
};


/** Gets all submission for the current user by Course ID and stores them in state */
export const getSubmissions = async ({state, effects}: Context, courseID: number) => {
    const result = await effects.grpcMan.getSubmissions(courseID, state.self.getId())
    if (result.data) {
        state.submissions[courseID] = result.data.getSubmissionsList()
        return true
    }
    return false
};


/** Gets all enrollments for the current user and stores them in state */
export const getEnrollmentsByUser = async ({state, effects}: Context) => {
    const result = await effects.grpcMan.getEnrollmentsByUser(state.self.getId())
    if (result.data) {
            state.enrollments = result.data.getEnrollmentsList()
            for (const enrollment of state.enrollments) {
                state.status[enrollment.getCourseid()] = enrollment.getStatus()
            }
            return true
        }
    return false
};


/** Changes user information server-side */
export const changeUser = async ({actions, effects}: Context, user: User) => {
    const result = await effects.grpcMan.updateUser(user)
    if (result.status.getCode() == 0) {
        actions.getSelf()
    }
};

export const updateAdmin = async ({ effects }: Context, user: User) => {
    let u = new User
    u.setIsadmin(!user.getIsadmin())
    u.setId(user.getId())
    const result = await effects.grpcMan.updateUser(u) 
    if (result.status.getCode() == 0) {
        user.setIsadmin(!user.getIsadmin())
    }
};

/** Gets a specific enrollment for a given course by the course ID if the user has an enrollment for that course. Returns null if none found */
export const getEnrollmentByCourseId = ({state}: Context, courseID: number) => {
    let enrol: Enrollment | null = null
    state.enrollments.forEach(enrollment => {
        if (enrollment.getCourseid() === courseID) {
            enrol = enrollment
        }
    })
    return enrol
};

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
};


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
};

export const updateSubmission = async ({state, actions, effects}: Context, value: {courseID: number, submission: Submission, status: Submission.Status}) => {
    value.submission.setStatus(value.status)
    const result = await effects.grpcMan.updateSubmission(value.courseID, value.submission)
    if (result.status.getCode() > 0) {
        actions.alertHandler(result)
    } else {
        value.submission.setStatus(value.status)
    }
};

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
};

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
};

/** Gets the assignments from a course by the course id. Meant to be used in places where you want only 1 assignment list. */
export const getAssignmentsByCourse = async ({state, effects}: Context, courseID: number) => {
    const res = await effects.grpcMan.getAssignments(courseID)
    if (res.data){
        state.assignments[courseID] = res.data.getAssignmentsList()
        return true
    }
    return false
};

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
            success = true
        } else {
        success = false
        }
    }
    return success
};



export const updateCourseGroup = ({state}: Context, cg: CourseGroup) => {
    state.courseGroup = cg
};

export const getGroupByUserAndCourse = async ({state, effects}: Context, courseID: number) => {
    const result = await effects.grpcMan.getGroupByUserAndCourse(courseID, state.self.getId())
    if (result.data) {
        state.userGroup[courseID] = result.data
    }
};

export const createGroup = async ({actions, effects}: Context, group: {courseID: number, users: number[], name: string}) => {
    const res = await effects.grpcMan.createGroup(group.courseID, group.name, group.users)
    actions.alertHandler(res)
};

export const popAlert = ({state}: Context, index: number) => {
    state.alerts = state.alerts.filter((_, i) => i != index)
};


export const getOrganization = async ({actions, effects}: Context, orgName: string) => {
    const res = await effects.grpcMan.getOrganization(orgName)
    actions.alertHandler(res)
};

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
};

export const editCourse = async ({effects}: Context, {course}: {course: Course} ) => {
    const response = await effects.grpcMan.updateCourse(course)
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
};

const convertCourseSubmission = (data: CourseSubmissions) => {
    const courseSubmissions: ParsedCourseSubmissions[] = []
    data.getLinksList().forEach((l, index) => {
        courseSubmissions.push({enrollment: l.getEnrollment(), submissions: l.getSubmissionsList(), user: l.getEnrollment()?.getUser()}) 
    })
    return courseSubmissions
};

/** Fetches and stores all submissions of a given course into state */
export const getAllCourseSubmissions = async ({state, effects}: Context, courseID: number) => {
    state.courseSubmissions[courseID] = []
    state.isLoading = true
    const result =  await effects.grpcMan.getSubmissionsByCourse(courseID, SubmissionsForCourseRequest.Type.ALL)
    if (result.data) {
            state.courseSubmissions[courseID] = convertCourseSubmission(result.data)
            state.isLoading = false    
    }
    state.isLoading = false
};

export const getGroupsByCourse = async ({state, effects}: Context, courseID: number) => {
    state.groups[courseID] = []
    const res = await effects.grpcMan.getGroupsByCourse(courseID)
    if (res.data) {
        state.groups[courseID] = res.data.getGroupsList()
    }
};

export const getUserSubmissions = async ({state, effects}: Context, courseID: number) => {
    state.submissions[courseID] = []
    const submissions = await effects.grpcMan.getSubmissions(courseID, state.self.getId())
    if (submissions.data) {
        for (const assignment of state.assignments[courseID]) {
            let submission = submissions.data.getSubmissionsList().find(s => s.getAssignmentid() === assignment.getId())
            state.submissions[courseID][assignment.getOrder() - 1] = submission ? submission : new Submission()
        }
        return true
    }
    return false
};

export const getGroupSubmissions = async ({state, effects}: Context, courseID: number) => {
    const enrollment = state.enrollmentsByCourseId[courseID]
    if (enrollment.hasGroup()) {
        const submissions = await effects.grpcMan.getGroupSubmissions(enrollment.getCourseid(), enrollment.getGroupid())
        for (const assignment of state.assignments[enrollment.getCourseid()]) {
            const submission = submissions.data?.getSubmissionsList().find(submission => submission.getAssignmentid() === assignment.getId())
            if (submission && assignment.getIsgrouplab()) {
                state.submissions[enrollment.getCourseid()][assignment.getOrder() - 1] = submission
            }
        }
    }
};

export const setActiveCourse = ({state}: Context, courseID: number) => {
    state.activeCourse = courseID
};

export const setActiveLab = ({state}: Context, assignmentID: number) => {
    state.activeLab = assignmentID
};

/** Enrolls a user (self) in a course given by courseID. Refreshes enrollments in state if enroll is sucessful. */
export const enroll = async ({state, effects}: Context, courseID: number) => {
    const res = await effects.grpcMan.createEnrollment(courseID, state.self.getId())

    if (res.status.getCode() == 0) {
        const enrollments = await effects.grpcMan.getEnrollmentsByUser(state.self.getId())
        if (enrollments.data) {
            state.enrollments = enrollments.data.getEnrollmentsList()
        }
    }
};

export const updateGroupStatus = async ({effects}: Context, {group, status}: {group: Group, status: Group.GroupStatus}) => {
    group.setStatus(status)
    const response = await effects.grpcMan.updateGroup(json(group))
    if (response.status.getCode() != 0) {
        group.setStatus(status == 1 ? 0 : 1)
    }
}

export const deleteGroup = async ({state, effects}: Context, group: Group) => {
    const response = await effects.grpcMan.deleteGroup(group.getCourseid(), group.getId())
    if (response.status.getCode() == 0) {
        state.groups[group.getCourseid()] = state.groups[group.getCourseid()].filter(g => g.getId() !== group.getId())
    }
}

export const updateGroup = async ({effects}: Context, group: Group) => {
    const response = await effects.grpcMan.updateGroup(group)
}

export const createCriterion = async ({state, effects}: Context, {criterion, assignment}: {criterion: GradingCriterion, assignment: Assignment}) => {
    for (const bm of assignment.getGradingbenchmarksList()) {
        if (bm.getId() === criterion.getBenchmarkid()) {
            bm.getCriteriaList().push(criterion)
        }
    }
}

export const logout = ({state}: Context) => {
    state.self = new User()
};

/** Initializes a student user with all required data */
export const fetchUserData = async ({state, actions}: Context) =>  {
    let success = await actions.getSelf()
    if (!success) { state.isLoading = false; return false;}
    while (state.isLoading) {
        success = await actions.getEnrollmentsByUser()
        success = await actions.getAssignments()
        for (const enrollment of state.enrollments) {
            if (enrollment.getStatus() >= Enrollment.UserStatus.STUDENT) {
                success = await actions.getUserSubmissions(enrollment.getCourseid())
                await actions.getGroupSubmissions(enrollment.getCourseid())
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
};

/** Switches between teacher and student view. */
export const changeView = async ({state, effects}: Context, courseID: number) => {
    const status = state.enrollmentsByCourseId[courseID].getStatus()
    if (status === Enrollment.UserStatus.STUDENT) {
        const status = await effects.grpcMan.getEnrollmentsByUser(state.self.getId(), [Enrollment.UserStatus.TEACHER])
        if (status.data?.getEnrollmentsList().find(enrollment => enrollment.getCourseid() == courseID)) {
            state.enrollmentsByCourseId[courseID].setStatus(Enrollment.UserStatus.TEACHER)
        }
    }
    if (status === Enrollment.UserStatus.TEACHER) {
        state.enrollmentsByCourseId[courseID].setStatus(Enrollment.UserStatus.STUDENT)
    }
    
}

/* START UTILITY ACTIONS */

export const loading = ({state}: Context) => {
    state.isLoading = !state.isLoading
};

/** Sets the time to now. */
export const setTimeNow = ({state}: Context) =>{
    state.timeNow = new Date()
};

/** Sets a query string in state. */
export const setQuery = ({state}: Context, query: string) => {
    state.query = query
};

export const enableRedirect = ({state}: Context, bool: boolean) => {
    state.enableRedirect = bool
};

export const setActiveSubmission = ({state}: Context, submission: Submission | undefined) => {
    state.activeSubmission = submission
}

export const isAuthorizedTeacher = async ({state, effects}: Context) => {
    const response = await effects.grpcMan.isAuthorizedTeacher()
    if (response.data) { return response.data.getIsauthorized() }
    return false
}

export const alertHandler = ({state}: Context, response: IGrpcResponse<any>) => {
    if (response.status.getCode() >= 0) {
        state.alerts.push({text: response.status.getError(), type: AlertType.DANGER})
    }
};

export const alert = ({state}: Context, alert: {text: string, type: AlertType}) => {
    state.alerts.push(alert)
};
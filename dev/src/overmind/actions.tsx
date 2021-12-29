import { json } from 'overmind'
import { Context } from ".";
import { IGrpcResponse } from "../GRPCManager";
import { User, Enrollment, Submission, Repository, Course, SubmissionsForCourseRequest, CourseSubmissions, Group, GradingCriterion, Assignment, SubmissionLink } from "../../proto/ag/ag_pb";
import { CourseGroup } from "./state";
import { AlertType } from "../Helpers";

/** 
 *      START CURRENT USER ACTIONS
 */

/** Fetches and stores an authenticated user in state */
export const getSelf = async ({state, effects}: Context): Promise<boolean> => {
        const user = await effects.grpcMan.getUser()
        if (user.data) {
            state.self = user.data
            return true
        } 
    return false
};

export const setSelf = ({state}: Context, user: User): void => {
    state.self.setName(user.getName())
    state.self.setEmail(user.getEmail())
    state.self.setStudentid(user.getStudentid())
}

/** Fetches all users (requires admin priveleges) */
export const getUsers = async ({state, effects}: Context): Promise<void> => {
    const users = await effects.grpcMan.getUsers()
    if (users.data) {
        for (const user of users.data.getUsersList()) {
            state.users[user.getId()] = user
        }
        // Insert users sorted by admin priveleges
        state.allUsers = users.data.getUsersList().sort((a, b) => {
            if(a.getIsadmin() > b.getIsadmin()) { return -1 }
            if(a.getIsadmin() < b.getIsadmin()) { return 1 }
            return 0
        })
    }
};


/** Fetches the courses */
export const getCoursesByUser = async ({state, effects}: Context, status?: Enrollment.UserStatus[]): Promise<boolean> => {
    const statuses: Enrollment.UserStatus[] = status ? status : []
    const courses = await effects.grpcMan.getCoursesByUser(state.self.getId(), statuses)
    if (courses.data) {
        courses.data.getCoursesList().forEach(course => {
            state.userCourses[course.getId()] = course
        })
        return true
    }
    return false
};


/** getSubmissions fetches all submission for the current user by Course ID and stores them in state */
export const getSubmissions = async ({state, effects}: Context, courseID: number): Promise<boolean> => {
    const result = await effects.grpcMan.getSubmissions(courseID, state.self.getId())
    if (result.data) {
        state.submissions[courseID] = result.data.getSubmissionsList()
        return true
    }
    return false
};


/** Gets all enrollments for the current user and stores them in state */
export const getEnrollmentsByUser = async ({state, effects}: Context): Promise<boolean> => {
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
export const updateUser = async ({actions, effects}: Context, user: User): Promise<void> => {
    const result = await effects.grpcMan.updateUser(user)
    if (result.status.getCode() == 0) {
        actions.getSelf()
    }
};

/** 
 *      END CURRENT USER ACTIONS 
*/

/** Fetches all courses */
export const getCourses = async ({state, effects}: Context): Promise<boolean> => {
    state.courses = []
    const result = await effects.grpcMan.getCourses()
    if (result.data) {
        state.courses = result.data.getCoursesList()
        return true
    }
    return false
};

export const updateAdmin = async ({ state, effects }: Context, user: User): Promise<void> => {
    if (confirm(`Are you sure you want to ${user.getIsadmin() ? "demote" : "promote"} ${user.getName()}?`)) {
        const u = json(user)
        u.setIsadmin(!user.getIsadmin())
        const result = await effects.grpcMan.updateUser(u) 
        if (result.status.getCode() == 0) {
            const found = state.allUsers.findIndex(s => s.getId() == user.getId())
            if (found) {
                state.allUsers[found] = u
            }
        }
    }
};

export const getEnrollmentsByCourse = async ({state, effects}: Context, value: {courseID: number, statuses: Enrollment.UserStatus[]}) : Promise<boolean> => {
    state.users = []
    state.courseEnrollments[value.courseID] = []
    const result = await effects.grpcMan.getEnrollmentsByCourse(value.courseID, undefined, undefined, value.statuses)
    if (result.data) {
            state.courseEnrollments[value.courseID] = result.data.getEnrollmentsList()
            return true
    }
    return false
};

export const setEnrollmentState = async ({actions, effects}: Context, enrollment: Enrollment): Promise<void> => {
    enrollment.setState(enrollment.getState() == Enrollment.DisplayState.VISIBLE ? Enrollment.DisplayState.FAVORITE : Enrollment.DisplayState.VISIBLE)
    const response = await effects.grpcMan.updateCourseVisibility(json(enrollment))
    if (!success(response)) {
        actions.alertHandler(response)
    }
};

/** Updates a given submission with a new status. This updates the given submission, as well as all other occurences of the given submission in state. */
export const updateSubmission = async ({state, effects}: Context, status: Submission.Status): Promise<void> => {
    if (!state.currentSubmission) {
        return
    }
    const submission = state.currentSubmission.setStatus(status)
    const result = await effects.grpcMan.updateSubmission(state.activeCourse, submission)
    if (!success(result)) {
        return
    }
    state.currentSubmission = submission
    for (const link of state.courseSubmissionsList[state.activeCourse]) {
        if (!link.submissions) { 
            continue 
        }
        for (const submission of link.submissions) {
            if (!submission.hasSubmission()) { 
                continue 
            }
            if ((submission.getSubmission() as Submission).getId() == state.activeSubmission) {
                (submission.getSubmission() as Submission).setStatus(status)
            }
        }
    }
};

export const updateEnrollment = async ({actions, effects}: Context, {enrollment, status}: {enrollment: Enrollment, status: Enrollment.UserStatus}): Promise<void> => {
    if (status > 0 || status === Enrollment.UserStatus.NONE && confirm("WARNNG! Rejecting a student is irreversible. Are you sure?")) {
        const tempEnrollment = json(enrollment)
        tempEnrollment.setStatus(status)
        const response = await effects.grpcMan.updateEnrollment(tempEnrollment)
        if (success(response)) {
            enrollment = tempEnrollment
        }
        else {
            actions.alertHandler(response)
        }
    }
};

/** Get assignments for all the courses the current logged in user is enrolled in */
export const getAssignments = async ({ state, effects }: Context): Promise<boolean> => {
    let success = false
    for (const enrollment of state.enrollments) {
        const response = await effects.grpcMan.getAssignments(enrollment.getCourseid())
        if(response.data) {
            state.assignments[enrollment.getCourseid()] = response.data.getAssignmentsList()
            success = true
        }
        success = false
    }
    return success
};

/** Get assignments for a single course, given by courseID */
export const getAssignmentsByCourse = async ({state, effects}: Context, courseID: number): Promise<boolean> => {
    const response = await effects.grpcMan.getAssignments(courseID)
    if (response.data){
        state.assignments[courseID] = response.data.getAssignmentsList()
        return true
    }
    return false
};


type RepoKey = keyof typeof Repository.Type
export const getRepositories = async ({state, effects}: Context): Promise<boolean> => {
    let success = true
    for (const enrollment of state.enrollments) {
        state.repositories[enrollment.getCourseid()] = {};
        
        const response = await effects.grpcMan.getRepositories(enrollment.getCourseid(), generateRepositoryList(enrollment))
        if (response.data) {
            response.data.getUrlsMap().forEach((entry, key) => {
                state.repositories[enrollment.getCourseid()][Repository.Type[key as RepoKey]] = entry
            })
        } else {
            success = false
        }
    }
    return success
};



export const updateCourseGroup = ({state}: Context, cg: CourseGroup): void => {
    state.courseGroup = cg
};

export const getGroupByUserAndCourse = async ({state, effects}: Context, courseID: number): Promise<void> => {
    const response = await effects.grpcMan.getGroupByUserAndCourse(courseID, state.self.getId())
    if (response.data) {
        state.userGroup[courseID] = response.data
    }
};

// TODO: Add group to state
export const createGroup = async ({actions, effects}: Context, group: {courseID: number, users: number[], name: string}): Promise<void> => {
    const response = await effects.grpcMan.createGroup(group.courseID, group.name, group.users)
    if (!success(response)) {
        actions.alertHandler(response)
    }
};



export const getOrganization = async ({actions, effects}: Context, orgName: string): Promise<boolean> => {
    const response = await effects.grpcMan.getOrganization(orgName)
    if (!success(response)) {
        actions.alertHandler(response)
        return false
    }
    return true
};

export const createCourse = async ({state, actions, effects}: Context, value: {course: Course, orgName: string}): Promise<void> => {
    const course = json(value.course)
    const orgResponse = await effects.grpcMan.getOrganization(value.orgName)
    if (orgResponse.data) {
        course.setOrganizationid(orgResponse.data.getId())
        course.setOrganizationpath(orgResponse.data.getPath())
        course.setProvider("github")
        course.setCoursecreatorid(state.self.getId())
        const response =  await effects.grpcMan.createCourse(course)
        if (response.data) {
            state.courses.push(response.data)
            actions.getEnrollmentsByUser()
        }
        actions.alertHandler(response)
    }
    actions.alertHandler(orgResponse)
};

/** Updates a given course and refreshes courses in state if successful  */
export const editCourse = async ({actions, effects}: Context, {course}: {course: Course}): Promise<void> => {
    const response = await effects.grpcMan.updateCourse(course)
    if (success(response)) {
        actions.getCourses()
    }
}

/** Updates all submissions in state where the fetched submission commit hash differs from the one in state. */
export const refreshSubmissions = async ({state, effects}: Context, input: {courseID: number, submissionID: number}): Promise<void> => {
    const response = await effects.grpcMan.getSubmissions(input.courseID, state.self.getId())
    if (!response.data || !success(response)) {
        return
    }
    const submissions = response.data.getSubmissionsList()
    for (const submission of submissions) {
        const assignment = state.assignments[input.courseID].find(a => a.getId() === submission.getAssignmentid())
        if (!assignment) {
            continue
        }
        if (state.submissions[input.courseID][assignment.getOrder() - 1].getCommithash() !== submission.getCommithash()) {
            state.submissions[input.courseID][assignment.getOrder() - 1] = submission
        }    
    }

};


export const convertCourseSubmission = ({state}: Context, {courseID, data}: {courseID: number, data: CourseSubmissions}): void => {
    state.review.reviews[courseID] = {}
    for (const link of data.getLinksList()) {
        if (link.hasEnrollment()) {
            link.getSubmissionsList().forEach(submissions => {
                if (submissions.hasSubmission()) {
                    state.review.reviews[courseID][(submissions.getSubmission() as Submission).getId()] = (submissions.getSubmission() as Submission).getReviewsList()
                }
            })
            state.courseSubmissions[courseID][(link.getEnrollment() as Enrollment).getId()] = {enrollment: link.getEnrollment(), submissions: link.getSubmissionsList(), user: link.getEnrollment()?.getUser()}
        }
    }
};

/** Fetches and stores all submissions of a given course into state */
export const getAllCourseSubmissions = async ({state, actions, effects}: Context, courseID: number): Promise<void> => {
    state.courseSubmissions[courseID] = {}
    state.isLoading = true
    const result =  await effects.grpcMan.getSubmissionsByCourse(courseID, SubmissionsForCourseRequest.Type.ALL, true)
    if (result.data) {
        actions.convertCourseSubmission({courseID: courseID, data: result.data})
        state.isLoading = false    
    }
    state.isLoading = false
};

export const getGroupsByCourse = async ({state, effects}: Context, courseID: number): Promise<void> => {
    state.groups[courseID] = []
    const res = await effects.grpcMan.getGroupsByCourse(courseID)
    if (res.data) {
        state.groups[courseID] = res.data.getGroupsList()
    }
};

export const getUserSubmissions = async ({state, effects}: Context, courseID: number): Promise<boolean> => {
    state.submissions[courseID] = []
    const submissions = await effects.grpcMan.getSubmissions(courseID, state.self.getId())
    if (submissions.data) {
        for (const assignment of state.assignments[courseID]) {
            const submission = submissions.data.getSubmissionsList().find(s => s.getAssignmentid() === assignment.getId())
            state.submissions[courseID][assignment.getOrder() - 1] = submission ? submission : new Submission()
        }
        return true
    }
    return false
};

export const getGroupSubmissions = async ({state, effects}: Context, courseID: number): Promise<void> => {
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

export const setActiveCourse = ({state}: Context, courseID: number): void => {
    state.activeCourse = courseID
};

export const setActiveLab = ({state}: Context, assignmentID: number): void => {
    state.activeLab = assignmentID
};

/** Rebuilds the currently active submission */
export const rebuildSubmission = async ({state, effects}: Context): Promise<void> => {
    if (state.currentSubmission && state.selectedAssignment) {
        const response = await effects.grpcMan.rebuildSubmission(state.selectedAssignment.getId(), state.activeSubmission)
        if (response.data) {
            state.currentSubmission = response.data
        }
    }
}

/** Enrolls a user (self) in a course given by courseID. Refreshes enrollments in state if enroll is sucessful. */
export const enroll = async ({state, effects}: Context, courseID: number): Promise<void> => {
    const response = await effects.grpcMan.createEnrollment(courseID, state.self.getId())
    if (response.status.getCode() == 0) {
        const enrollments = await effects.grpcMan.getEnrollmentsByUser(state.self.getId())
        if (enrollments.data) {
            state.enrollments = enrollments.data.getEnrollmentsList()
        }
    }
};

export const updateGroupStatus = async ({effects}: Context, {group, status}: {group: Group, status: Group.GroupStatus}): Promise<void> => {
    const grp = json(group).setStatus(status)
    const response = await effects.grpcMan.updateGroup(grp)
    if (success(response)) {
        group = grp
    }
}

export const deleteGroup = async ({state, effects}: Context, group: Group): Promise<void> => {
    if (confirm("Deleting a group is an irreversible action. Are you sure?")) {
        const isRepoEmpty = await effects.grpcMan.isEmptyRepo(group.getCourseid(), 0, group.getId())
        if (isRepoEmpty || confirm(`Warning! Group repository is not empty! Do you still want to delete group, github team and group repository?`)) {
            const response = await effects.grpcMan.deleteGroup(group.getCourseid(), group.getId())
            if (response.status.getCode() == 0) {
                state.groups[group.getCourseid()] = state.groups[group.getCourseid()].filter(g => g.getId() !== group.getId())
            }
        }
    }
}

export const updateGroup = async ({actions, effects}: Context, group: Group): Promise<void> => {
    const response = await effects.grpcMan.updateGroup(group)
    actions.alertHandler(response)
}

export const createCriterion = async ({effects}: Context, {criterion, assignment}: {criterion: GradingCriterion, assignment: Assignment}): Promise<void> => {
    for (const bm of assignment.getGradingbenchmarksList()) {
        if (bm.getId() === criterion.getBenchmarkid()) {
            bm.getCriteriaList().push(criterion)
            effects.grpcMan.createCriterion(criterion)
        }
    }
}

export const setActiveSubmissionLink = ({state}: Context, link: SubmissionLink): void => {
    state.activeSubmissionLink = json(link)
}

export const setGrade = async ({effects}: Context, {criterion, grade}: {criterion: GradingCriterion, grade: GradingCriterion.Grade}): Promise<void> => {
    const oldGrade = criterion.getGrade()
    criterion.setGrade(grade)
    const response = await effects.grpcMan.updateCriterion(criterion)
    if (!success(response)) {
        criterion.setGrade(oldGrade)
    }
}

export const setSelectedEnrollment = ({state}: Context, enrollmentId: number): void => {
    state.selectedEnrollment = enrollmentId
}

/** Initializes a student user with all required data */
export const fetchUserData = async ({state, actions}: Context): Promise<boolean> =>  {
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
            if (enrollment.getStatus() >= Enrollment.UserStatus.STUDENT) {
                const statuses = enrollment.getStatus() === Enrollment.UserStatus.STUDENT ? [Enrollment.UserStatus.STUDENT, Enrollment.UserStatus.TEACHER ] : []
                success = await actions.getEnrollmentsByCourse({courseID: enrollment.getCourseid(), statuses: statuses})
            }
            if (enrollment.getStatus() == Enrollment.UserStatus.TEACHER) {
                actions.getGroupsByCourse(enrollment.getCourseid())
            }
        }
        if (state.self.getIsadmin()) {
            actions.getUsers()
        }
        success = await actions.getRepositories()
        success = await actions.getCourses()
        state.isLoading = false
        
    }
    return success
};


/* Utility Actions */

/** Switches between teacher and student view. */
export const changeView = async ({state, effects}: Context, courseID: number): Promise<void> => {
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

export const loading = ({state}: Context): void => {
    state.isLoading = !state.isLoading
};

/** Sets the time to now. */
export const setTimeNow = ({state}: Context): void =>{
    state.timeNow = new Date()
};

/** Sets a query string in state. */
export const setQuery = ({state}: Context, query: string): void => {
    state.query = query
};

export const enableRedirect = ({state}: Context, bool: boolean): void => {
    state.enableRedirect = bool
};

export const setActiveSubmission = ({state}: Context, submission: number): void => {
    state.activeSubmission = submission 
};


export const setSelectedUser = ({state}: Context, user: User | undefined): void => {
    state.activeUser = user
};

/** Returns whether or not the current user is an authorized teacher with teacher scopes */
export const isAuthorizedTeacher = async ({effects}: Context): Promise<boolean> => {
    const response = await effects.grpcMan.isAuthorizedTeacher()
    if (response.data) { 
        return response.data.getIsauthorized() 
    }
    return false
};

export const alertHandler = ({state}: Context, response: IGrpcResponse<unknown>): void => {
    if (response.status.getCode() >= 0) {
        state.alerts.push({text: response.status.getError(), type: AlertType.DANGER})
    }
};

export const alert = ({state}: Context, alert: {text: string, type: AlertType}): void => {
    state.alerts.push(alert)
};

export const popAlert = ({state}: Context, index: number): void => {
    state.alerts = state.alerts.filter((_, i) => i != index)
};

export const logout = ({state}: Context): void => {
    state.self = new User()
};

const generateRepositoryList = (enrollment: Enrollment): Repository.Type[] => {
    switch (enrollment.getStatus()) {
        case Enrollment.UserStatus.TEACHER:
            return [Repository.Type.ASSIGNMENTS, Repository.Type.COURSEINFO, Repository.Type.GROUP, Repository.Type.TESTS, Repository.Type.USER]
        case Enrollment.UserStatus.STUDENT:
            return [Repository.Type.ASSIGNMENTS, Repository.Type.COURSEINFO, Repository.Type.GROUP, Repository.Type.USER]
        default:
            return [Repository.Type.NONE]
    }
}

/** Use this to verify that a gRPC request completed without an error code */
const success = (response: IGrpcResponse<unknown>): boolean => {
    return response.status.getCode() === 0
}
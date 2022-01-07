import { json } from 'overmind'
import { Context } from ".";
import { IGrpcResponse } from "../GRPCManager";
import { User, Enrollment, Submission, Repository, Course, SubmissionsForCourseRequest, CourseSubmissions, Group, GradingCriterion, Assignment, SubmissionLink, Organization, GradingBenchmark } from "../../proto/ag/ag_pb";
import { CourseGroup } from "./state";
import { AlertType, SubmissionStatus } from "../Helpers";


/** 
 *      START CURRENT USER ACTIONS
 */

/** Fetches and stores an authenticated user in state */
export const getSelf = async ({ state, effects }: Context): Promise<boolean> => {
    const user = await effects.grpcMan.getUser()
    if (user.data) {
        state.self = user.data
        return true
    }
    return false
};

/** setSelf is used to update the state with modified user data */
export const setSelf = ({ state }: Context, user: User): void => {
    state.self.setName(user.getName())
    state.self.setEmail(user.getEmail())
    state.self.setStudentid(user.getStudentid())
}

/** Gets all enrollments for the current user and stores them in state */
export const getEnrollmentsByUser = async ({ state, effects }: Context): Promise<boolean> => {
    const result = await effects.grpcMan.getEnrollmentsByUser(state.self.getId())
    if (result.data) {
        state.enrollments = result.data.getEnrollmentsList()
        for (const enrollment of state.enrollments) {
            state.status[enrollment.getCourseid()] = enrollment.getStatus()
        }
        console.log(state.enrollments)
        return true
    }
    return false
};

/** Fetches all users (requires admin priveleges) */
export const getUsers = async ({ state, effects }: Context): Promise<void> => {
    const users = await effects.grpcMan.getUsers()
    if (users.data) {
        for (const user of users.data.getUsersList()) {
            state.users[user.getId()] = user
        }
        // Insert users sorted by admin priveleges
        state.allUsers = users.data.getUsersList().sort((a, b) => {
            if (a.getIsadmin() > b.getIsadmin()) { return -1 }
            if (a.getIsadmin() < b.getIsadmin()) { return 1 }
            return 0
        })
    }
};





/** Changes user information server-side */
export const updateUser = async ({ actions, effects }: Context, user: User): Promise<void> => {
    const result = await effects.grpcMan.updateUser(user)
    if (result.status.getCode() == 0) {
        await actions.getSelf()
    }
};

/**
 *      END CURRENT USER ACTIONS 
 */

/** Fetches all courses */
export const getCourses = async ({ state, effects }: Context): Promise<boolean> => {
    state.courses = []
    const result = await effects.grpcMan.getCourses()
    if (result.data) {
        state.courses = result.data.getCoursesList()
        return true
    }
    return false
};

/** updateAdmin is used to update the admin priveleges of a user. Admin status toggles between true and false */
export const updateAdmin = async ({ state, effects }: Context, user: User): Promise<void> => {
    // Confirm that user really wants to change admin status
    if (confirm(`Are you sure you want to ${user.getIsadmin() ? "demote" : "promote"} ${user.getName()}?`)) {
        // Copy user object and change admin status
        const u = json(user)
        u.setIsadmin(!user.getIsadmin())

        // Send updated user to server
        const result = await effects.grpcMan.updateUser(u)
        if (result.status.getCode() == 0) {
            // If successful, update user in state with new admin status
            const found = state.allUsers.findIndex(s => s.getId() == user.getId())
            if (found) {
                state.allUsers[found] = u
            }
        }
    }
};


export const getEnrollmentsByCourse = async ({ state, effects }: Context, value: { courseID: number, statuses: Enrollment.UserStatus[] }): Promise<boolean> => {
    state.courseEnrollments[value.courseID] = []
    const result = await effects.grpcMan.getEnrollmentsByCourse(value.courseID, undefined, undefined, value.statuses)
    if (result.data) {
        state.courseEnrollments[value.courseID] = result.data.getEnrollmentsList()
        return true
    }
    return false
};

/**  setEnrollmentState toggles the state of an enrollment between favorite and visible */
export const setEnrollmentState = async ({ actions, effects }: Context, enrollment: Enrollment): Promise<void> => {
    enrollment.setState(enrollment.getState() == Enrollment.DisplayState.VISIBLE ? Enrollment.DisplayState.FAVORITE : Enrollment.DisplayState.VISIBLE)
    const response = await effects.grpcMan.updateCourseVisibility(json(enrollment))
    if (!success(response)) {
        actions.alertHandler(response)
    }
};

/** Updates a given submission with a new status. This updates the given submission, as well as all other occurences of the given submission in state. */
export const updateSubmission = async ({ state, effects }: Context, status: Submission.Status): Promise<void> => {
    /* Do not update if the status is already the same or if there is no selected submission */
    if (!state.currentSubmission || state.currentSubmission.getStatus() == status) {
        return
    }

    /* Confirm that user really wants to change submission status */
    if (!confirm(`Are you sure you want to set status ${SubmissionStatus[status]} on this submission?`)) {
        return
    }

    /* Store the previous submission status */
    const previousStatus = state.currentSubmission.getStatus()

    /* Update the submission status */
    state.currentSubmission.setStatus(status)
    const result = await effects.grpcMan.updateSubmission(state.activeCourse, state.currentSubmission)
    if (!success(result)) {
        /* If the update failed, revert the submission status */
        state.currentSubmission.setStatus(previousStatus)
        return
    }

    /* Loop through all submissions for the current course and update the status if it matches the current submission ID */
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

/** updateEnrollment updates an enrollment status with the given status */
export const updateEnrollment = async ({ actions, effects }: Context, { enrollment, status }: { enrollment: Enrollment, status: Enrollment.UserStatus }): Promise<void> => {

    // If rejecing an enrollment, confirm that user really wants to do so. If approving, no confirmation is required.
    if (status > 0 || status === Enrollment.UserStatus.NONE && confirm("WARNNG! Rejecting a student is irreversible. Are you sure?")) {
        // Copy enrollment object and change status
        const temp = json(enrollment).setStatus(status)
        // Send updated enrollment to server
        const response = await effects.grpcMan.updateEnrollment(temp)
        if (success(response)) {
            // If successful, update enrollment in state with new status
            enrollment.setStatus(status)
        }
        else {
            // If unsuccessful, alert user
            actions.alertHandler(response)
        }
    }
};

/** Get assignments for all the courses the current user is enrolled in */
export const getAssignments = async ({ state, effects }: Context): Promise<boolean> => {
    let success = false
    for (const enrollment of state.enrollments) {
        const response = await effects.grpcMan.getAssignments(enrollment.getCourseid())
        if (response.data) {
            // Store assignments in state by course ID
            state.assignments[enrollment.getCourseid()] = response.data.getAssignmentsList()
            success = true
        } else {
            success = false
        }
    }
    //TODO: Might return true even if some assignment requests failed. Revisit this later.
    return success
};

/** Get assignments for a single course, given by courseID */
export const getAssignmentsByCourse = async ({ state, effects }: Context, courseID: number): Promise<boolean> => {
    const response = await effects.grpcMan.getAssignments(courseID)
    if (response.data) {
        state.assignments[courseID] = response.data.getAssignmentsList()
        return true
    }
    return false
};


type RepoKey = keyof typeof Repository.Type
export const getRepositories = async ({ state, effects }: Context): Promise<boolean> => {
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



export const updateCourseGroup = ({ state }: Context, cg: CourseGroup): void => {
    state.courseGroup = cg
};

export const getGroupByUserAndCourse = async ({ state, effects }: Context, courseID: number): Promise<void> => {
    const response = await effects.grpcMan.getGroupByUserAndCourse(courseID, state.self.getId())
    if (response.data) {
        state.userGroup[courseID] = response.data
    }
};

// TODO: Add group to state
export const createGroup = async ({ actions, effects }: Context, group: { courseID: number, users: number[], name: string }): Promise<void> => {
    const response = await effects.grpcMan.createGroup(group.courseID, group.name, group.users)
    if (!success(response)) {
        actions.alertHandler(response)
    }
};


/** getOrganization ask the server if orgName is a valid organization name. If so, it returns the organization object. */
// TODO: Currently returns a boolean. 
export const getOrganization = async ({ actions, effects }: Context, orgName: string): Promise<boolean> => {
    const response = await effects.grpcMan.getOrganization(orgName)
    if (!success(response)) {
        actions.alertHandler(response)
        return false
    }
    return true
};

/* createCourse creates a new course */
export const createCourse = async ({ state, actions, effects }: Context, value: { course: Course, orgName: string }): Promise<boolean> => {
    const course = json(value.course)

    /* Check if the organization exists */
    const orgResponse = await effects.grpcMan.getOrganization(value.orgName)
    if (orgResponse.data) {
        /* Fill in required fields */
        course.setOrganizationid(orgResponse.data.getId())
        course.setOrganizationpath(orgResponse.data.getPath())
        course.setProvider("github")
        course.setCoursecreatorid(state.self.getId())
        /* Send the course to the server */
        const response = await effects.grpcMan.createCourse(course)
        if (response.data) {
            /* If successful, add the course to the state */
            state.courses.push(response.data)
            /* User that created the course is automatically enrolled in the course. Refresh the enrollment list */
            actions.getEnrollmentsByUser()
            return true
        }
        actions.alertHandler(response)
    }
    actions.alertHandler(orgResponse)
    return false
};

/** Updates a given course and refreshes courses in state if successful  */
export const editCourse = async ({ actions, effects }: Context, { course }: { course: Course }): Promise<void> => {
    const response = await effects.grpcMan.updateCourse(course)
    if (success(response)) {
        actions.getCourses()
    }
}


/** getSubmissions fetches all submission for the current user by Course ID and stores them in state */
// TODO: Currently not used, see refreshSubmissions. 
export const getSubmissions = async ({ state, effects }: Context, courseID: number): Promise<boolean> => {
    const result = await effects.grpcMan.getSubmissions(courseID, state.self.getId())
    if (result.data) {
        state.submissions[courseID] = result.data.getSubmissionsList()
        return true
    }
    return false
};

// TODO: Currently not in use. Requires gRPC streaming to be implemented. Intended to be used to update submissions in state when a new commit is pushed to a repository.
// TODO: A workaround to not use gRPC streaming is to ping the server at set intervals to check for new commits. This functionality was removed pending gRPC streaming implementation.
/** Updates all submissions in state where the fetched submission commit hash differs from the one in state. */
export const refreshSubmissions = async ({ state, effects }: Context, input: { courseID: number, submissionID: number }): Promise<void> => {
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


export const convertCourseSubmission = ({ state }: Context, { courseID, data }: { courseID: number, data: CourseSubmissions }): void => {
    state.review.reviews[courseID] = {}
    for (const link of data.getLinksList()) {
        if (link.hasEnrollment()) {
            link.getSubmissionsList().forEach(submissions => {
                if (submissions.hasSubmission()) {
                    state.review.reviews[courseID][(submissions.getSubmission() as Submission).getId()] = (submissions.getSubmission() as Submission).getReviewsList()
                }
            })
            state.courseSubmissions[courseID][(link.getEnrollment() as Enrollment).getId()] = { enrollment: link.getEnrollment(), submissions: link.getSubmissionsList(), user: link.getEnrollment()?.getUser() }
        }
    }
};

/** Fetches and stores all submissions of a given course into state */
export const getAllCourseSubmissions = async ({ state, actions, effects }: Context, courseID: number): Promise<void> => {
    state.courseSubmissions[courseID] = {}
    state.isLoading = true
    const result = await effects.grpcMan.getSubmissionsByCourse(courseID, SubmissionsForCourseRequest.Type.ALL, true)
    if (result.data) {
        actions.convertCourseSubmission({ courseID: courseID, data: result.data })
        state.isLoading = false
    }
    state.isLoading = false
};

export const getGroupsByCourse = async ({ state, effects }: Context, courseID: number): Promise<void> => {
    state.groups[courseID] = []
    const res = await effects.grpcMan.getGroupsByCourse(courseID)
    if (res.data) {
        state.groups[courseID] = res.data.getGroupsList()
    }
};

export const getUserSubmissions = async ({ state, effects }: Context, courseID: number): Promise<boolean> => {
    state.submissions[courseID] = []
    const submissions = await effects.grpcMan.getSubmissions(courseID, state.self.getId())
    if (submissions.data) {
        // Insert submissions into state.submissions by the assignment order
        for (const assignment of state.assignments[courseID]) {
            const submission = submissions.data.getSubmissionsList().find(s => s.getAssignmentid() === assignment.getId())
            state.submissions[courseID][assignment.getOrder() - 1] = submission ? submission : new Submission()
        }
        return true
    }
    return false
};

export const getGroupSubmissions = async ({ state, effects }: Context, courseID: number): Promise<void> => {
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

export const setActiveCourse = ({ state }: Context, courseID: number): void => {
    state.activeCourse = courseID
};

export const setActiveLab = ({ state }: Context, assignmentID: number): void => {
    state.activeLab = assignmentID
};

/** Rebuilds the currently active submission */
export const rebuildSubmission = async ({ state, effects }: Context): Promise<void> => {
    if (state.currentSubmission && state.selectedAssignment) {
        const response = await effects.grpcMan.rebuildSubmission(state.selectedAssignment.getId(), state.activeSubmission)
        if (response.data) {
            state.currentSubmission = response.data
        }
    }
}

/* rebuildAllSubmissions rebuilds all submissions for a given assignment */
export const rebuildAllSubmissions = async ({ effects }: Context, { courseID, assignmentID }: { courseID: number, assignmentID: number }): Promise<boolean> => {
    const response = await effects.grpcMan.rebuildSubmissions(assignmentID, courseID)
    if (success(response)) {
        return true
    }
    return false
}

/** Enrolls a user (self) in a course given by courseID. Refreshes enrollments in state if enroll is sucessful. */
export const enroll = async ({ state, effects }: Context, courseID: number): Promise<void> => {
    const response = await effects.grpcMan.createEnrollment(courseID, state.self.getId())
    if (response.status.getCode() == 0) {
        const enrollments = await effects.grpcMan.getEnrollmentsByUser(state.self.getId())
        if (enrollments.data) {
            state.enrollments = enrollments.data.getEnrollmentsList()
        }
    }
};

export const updateGroupStatus = async ({ effects }: Context, { group, status }: { group: Group, status: Group.GroupStatus }): Promise<void> => {
    const grp = json(group).setStatus(status)
    const response = await effects.grpcMan.updateGroup(grp)
    if (success(response)) {
        group = grp
    }
}

export const deleteGroup = async ({ state, effects }: Context, group: Group): Promise<void> => {
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

export const updateGroup = async ({ actions, effects }: Context, group: Group): Promise<void> => {
    const response = await effects.grpcMan.updateGroup(group)
    actions.alertHandler(response)
}

export const createCriterion = async ({ effects }: Context, { criterion, assignment }: { criterion: GradingCriterion, assignment: Assignment }): Promise<void> => {
    for (const bm of assignment.getGradingbenchmarksList()) {
        if (bm.getId() === criterion.getBenchmarkid()) {
            bm.getCriteriaList().push(criterion)
            effects.grpcMan.createCriterion(criterion)
        }
    }
}

export const createBenchmark = async ({ effects }: Context, { benchmark, assignment }: { benchmark: GradingBenchmark, assignment: Assignment }): Promise<void> => {
    benchmark.setAssignmentid(assignment.getId())
    const response = await effects.grpcMan.createBenchmark(benchmark)
    if (success(response)) {
        assignment.getGradingbenchmarksList().push(benchmark)
    }
}

export const setActiveSubmissionLink = ({ state }: Context, link: SubmissionLink): void => {
    state.activeSubmissionLink = json(link)
}

export const setGrade = async ({ effects }: Context, { criterion, grade }: { criterion: GradingCriterion, grade: GradingCriterion.Grade }): Promise<void> => {
    const oldGrade = criterion.getGrade()
    criterion.setGrade(grade)
    const response = await effects.grpcMan.updateCriterion(criterion)
    if (!success(response)) {
        criterion.setGrade(oldGrade)
    }
}

export const setSelectedEnrollment = ({ state }: Context, enrollmentId: number): void => {
    state.selectedEnrollment = enrollmentId
}

/* fetchUserData is called when the user enters the app. It fetches all data that is needed for the user to be able to use the app. */
/* If the user is not logged in, i.e does not have a valid token, the process is aborted. */
export const fetchUserData = async ({ state, actions }: Context): Promise<boolean> => {
    let success = await actions.getSelf()
    // If getSelf returns false, the user is not logged in. Abort.
    if (!success) { state.isLoading = false; return false; }

    // Start fetching all data. Loading screen will be shown until all data is fetched, i.e state.isLoading is set to false.
    while (state.isLoading) {
        // Order matters here. Some data is dependent on other data. Ex. fetching submissions depends on enrollments.
        success = await actions.getEnrollmentsByUser()
        success = await actions.getAssignments()
        for (const enrollment of state.enrollments) {
            if (enrollment.getStatus() >= Enrollment.UserStatus.STUDENT) {
                success = await actions.getUserSubmissions(enrollment.getCourseid())
                await actions.getGroupSubmissions(enrollment.getCourseid())
            }
            if (enrollment.getStatus() >= Enrollment.UserStatus.STUDENT) {
                const statuses = enrollment.getStatus() === Enrollment.UserStatus.STUDENT ? [Enrollment.UserStatus.STUDENT, Enrollment.UserStatus.TEACHER] : []
                success = await actions.getEnrollmentsByCourse({ courseID: enrollment.getCourseid(), statuses: statuses })
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
        // End loading screen.
        state.isLoading = false

    }
    // The value of success is unreliable. The intention is to return true if the user is logged in and all data was fetched.
    // However, if one of the above calls fail, it could still be the case that success returns true.
    return success
};


/* Utility Actions */

/** Switches between teacher and student view. */
export const changeView = async ({ state, effects }: Context, courseID: number): Promise<void> => {
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

export const loading = ({ state }: Context): void => {
    state.isLoading = !state.isLoading
};

/** Sets the time to now. */
export const setTimeNow = ({ state }: Context): void => {
    state.timeNow = new Date()
};

/** Sets a query string in state. */
export const setQuery = ({ state }: Context, query: string): void => {
    state.query = query
};


export const setActiveSubmission = ({ state }: Context, submission: number): void => {
    state.activeSubmission = submission
};


export const setSelectedUser = ({ state }: Context, user: User | undefined): void => {
    state.activeUser = user
};

/** Returns whether or not the current user is an authorized teacher with teacher scopes */
export const isAuthorizedTeacher = async ({ effects }: Context): Promise<boolean> => {
    const response = await effects.grpcMan.isAuthorizedTeacher()
    if (response.data) {
        return response.data.getIsauthorized()
    }
    return false
};

export const alertHandler = ({ state }: Context, response: IGrpcResponse<unknown>): void => {
    if (response.status.getCode() >= 0) {
        state.alerts.push({ text: response.status.getError(), type: AlertType.DANGER })
    }
};

export const alert = ({ state }: Context, alert: { text: string, type: AlertType }): void => {
    state.alerts.push(alert)
};

export const popAlert = ({ state }: Context, index: number): void => {
    state.alerts = state.alerts.filter((_, i) => i != index)
};

export const logout = ({ state }: Context): void => {
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
export const success = (response: IGrpcResponse<unknown>): boolean => {
    return response.status.getCode() === 0
}

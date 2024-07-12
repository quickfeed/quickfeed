import { Code, ConnectError } from "@bufbuild/connect"
import { Context } from "."
import { Organization, SubmissionRequest_SubmissionType, } from "../../proto/qf/requests_pb"
import {
    Assignment,
    Course,
    Enrollment,
    Enrollment_DisplayState,
    Enrollment_UserStatus,
    Grade,
    GradingBenchmark,
    GradingCriterion,
    Group,
    Group_GroupStatus,
    Submission,
    Submission_Status,
    User
} from "../../proto/qf/types_pb"
import { Response } from "../client"
import { Color, ConnStatus, getStatusByUser, hasAllStatus, hasStudent, hasTeacher, isPending, isStudent, isTeacher, isVisible, newID, setStatusAll, setStatusByUser, SubmissionSort, SubmissionStatus, validateGroup } from "../Helpers"
import * as internalActions from "./internalActions"
import { Alert, CourseGroup, SubmissionOwner } from "./state"

export const internal = internalActions

export const onInitializeOvermind = async ({ actions, effects }: Context) => {
    // Initialize the API client. *Must* be done before accessing the client.
    effects.api.init(actions.errorHandler)
    await actions.fetchUserData()
    // Currently this only alerts the user if they are not logged in after a page refresh
    const alert = localStorage.getItem("alert")
    if (alert) {
        actions.alert({ text: alert, color: Color.RED })
        localStorage.removeItem("alert")
    }
}

export const handleStreamError = (context: Context, error: Error): void => {
    context.state.connectionStatus = ConnStatus.DISCONNECTED
    context.actions.alert({ text: error.message, color: Color.RED, delay: 10000 })
}

export const receiveSubmission = ({ state }: Context, submission: Submission): void => {
    let courseID = 0n
    let assignmentOrder = 0
    Object.entries(state.assignments).forEach(
        ([, assignments]) => {
            const assignment = assignments.find(a => a.ID === submission.AssignmentID)
            if (assignment && assignment.CourseID !== 0n) {
                assignmentOrder = assignment.order
                courseID = assignment.CourseID
                return
            }
        }
    )
    if (courseID === 0n) {
        return
    }
    Object.assign(state.submissions[courseID.toString()][assignmentOrder - 1], submission)
}

/**
 *      START CURRENT USER ACTIONS
 */

/** Fetches and stores an authenticated user in state */
export const getSelf = async ({ state, effects }: Context): Promise<boolean> => {
    const response = await effects.api.client.getUser({})
    if (response.error) {
        return false
    }
    state.self = response.message
    return true
}

/** Gets all enrollments for the current user and stores them in state */
export const getEnrollmentsByUser = async ({ state, effects }: Context): Promise<void> => {
    const response = await effects.api.client.getEnrollments({
        FetchMode: {
            case: "userID",
            value: state.self.ID,
        }
    })
    if (response.error) {
        return
    }
    state.enrollments = response.message.enrollments
    for (const enrollment of state.enrollments) {
        state.status[enrollment.courseID.toString()] = enrollment.status
    }
}

/** Fetches all users (requires admin privileges) */
export const getUsers = async ({ state, effects }: Context): Promise<void> => {
    const response = await effects.api.client.getUsers({})
    if (response.error) {
        return
    }
    for (const user of response.message.users) {
        state.users[user.ID.toString()] = user
    }
    // Insert users sorted by admin privileges
    state.allUsers = response.message.users.sort((a, b) => {
        if (a.IsAdmin > b.IsAdmin) { return -1 }
        if (a.IsAdmin < b.IsAdmin) { return 1 }
        return 0
    })
}

/** Changes user information server-side */
export const updateUser = async ({ actions, effects }: Context, user: User): Promise<void> => {
    const response = await effects.api.client.updateUser(user)
    if (response.error) {
        return
    }
    await actions.getSelf()
}

/**
 *      END CURRENT USER ACTIONS
 */

/** Fetches all courses */
export const getCourses = async ({ state, effects }: Context): Promise<void> => {
    state.courses = []
    const response = await effects.api.client.getCourses({})
    if (response.error) {
        return
    }
    state.courses = response.message.courses
}

/** updateAdmin is used to update the admin privileges of a user. Admin status toggles between true and false */
export const updateAdmin = async ({ state, effects }: Context, user: User): Promise<void> => {
    // Confirm that user really wants to change admin status
    if (confirm(`Are you sure you want to ${user.IsAdmin ? "demote" : "promote"} ${user.Name}?`)) {
        // Convert to proto object and change admin status
        const req = new User(user)
        req.IsAdmin = !user.IsAdmin
        // Send updated user to server
        const response = await effects.api.client.updateUser(req)
        if (response.error) {
            return
        }
        // If successful, update user in state with new admin status
        const found = state.allUsers.findIndex(s => s.ID === user.ID)
        if (found > -1) {
            state.allUsers[found].IsAdmin = req.IsAdmin
        }
    }
}

export const getEnrollmentsByCourse = async ({ state, effects }: Context, value: { courseID: bigint, statuses: Enrollment_UserStatus[] }): Promise<void> => {
    const response = await effects.api.client.getEnrollments({
        FetchMode: {
            case: "courseID",
            value: value.courseID,
        },
        statuses: value.statuses,
    })
    if (response.error) {
        return
    }
    state.courseEnrollments[value.courseID.toString()] = response.message.enrollments
}

/**  setEnrollmentState toggles the state of an enrollment between favorite and visible */
export const setEnrollmentState = async ({ effects }: Context, enrollment: Enrollment): Promise<void> => {
    enrollment.state = isVisible(enrollment)
        ? Enrollment_DisplayState.HIDDEN
        : Enrollment_DisplayState.VISIBLE

    await effects.api.client.updateCourseVisibility(enrollment)
}

/** Updates a given submission with a new status. This updates the given submission, as well as all other occurrences of the given submission in state. */
export const updateSubmission = async ({ state, effects }: Context, { owner, submission, status }: { owner: SubmissionOwner, submission: Submission | null, status: Submission_Status }): Promise<void> => {
    /* Do not update if the status is already the same or if there is no selected submission */
    if (!submission) {
        return
    }

    switch (owner.type) {
        // Take no action if there is no change in status
        case "ENROLLMENT":
            if (getStatusByUser(submission, submission.userID) === status) {
                return
            }
            break
        case "GROUP":
            if (hasAllStatus(submission, status)) {
                return
            }
            break
    }

    /* Confirm that user really wants to change submission status */
    if (!confirm(`Are you sure you want to set status ${SubmissionStatus[status]} on this submission?`)) {
        return
    }

    let clone = submission.clone()
    switch (owner.type) {
        case "ENROLLMENT":
            clone = setStatusByUser(clone, submission.userID, status)
            break
        case "GROUP":
            clone = setStatusAll(clone, status)
            break
    }
    /* Update the submission status */
    const response = await effects.api.client.updateSubmission({
        courseID: state.activeCourse,
        submissionID: submission.ID,
        grades: clone.Grades,
        released: submission.released,
        score: submission.score,
    })
    if (response.error) {
        return
    }
    submission.Grades = clone.Grades
    state.submissionsForCourse.update(owner, submission)
}

export const updateGrade = async ({ state, effects }: Context, { grade, status }: { grade: Grade, status: Submission_Status }): Promise<void> => {
    if (grade.Status === status || !state.selectedSubmission) {
        return
    }

    if (!confirm(`Are you sure you want to set status ${SubmissionStatus[status]} on this grade?`)) {
        return
    }

    const clone = state.selectedSubmission.clone()
    clone.Grades = clone.Grades.map(g => {
        if (g.UserID === grade.UserID) {
            g.Status = status
        }
        return g
    })
    const response = await effects.api.client.updateSubmission({
        courseID: state.activeCourse,
        submissionID: state.selectedSubmission.ID,
        grades: clone.Grades,
        released: state.selectedSubmission.released,
        score: state.selectedSubmission.score,
    })
    if (response.error) {
        return
    }

    state.selectedSubmission.Grades = clone.Grades
    const type = clone.userID ? "ENROLLMENT" : "GROUP"
    switch (type) {
        case "ENROLLMENT":
            state.submissionsForCourse.update({ type, id: clone.userID }, clone)
            break
        case "GROUP":
            state.submissionsForCourse.update({ type, id: clone.groupID }, clone)
            break
    }
}

/** updateEnrollment updates an enrollment status with the given status */
export const updateEnrollment = async ({ state, actions, effects }: Context, { enrollment, status }: { enrollment: Enrollment, status: Enrollment_UserStatus }): Promise<void> => {
    if (!enrollment.user) {
        // user name is required
        return
    }

    if (status === Enrollment_UserStatus.NONE) {
        const proceed = await actions.internal.isEmptyRepo({ userID: enrollment.userID, courseID: enrollment.courseID })
        if (!proceed) {
            return
        }
    }

    // Confirm that user really wants to change enrollment status
    let confirmed = false
    switch (status) {
        case Enrollment_UserStatus.NONE:
            confirmed = confirm("WARNING! Rejecting a student is irreversible. Are you sure?")
            break
        case Enrollment_UserStatus.STUDENT:
            // If the enrollment is pending, don't ask for confirmation
            confirmed = isPending(enrollment) || confirm(`Warning! ${enrollment.user.Name} is a teacher. Are sure you want to demote?`)
            break
        case Enrollment_UserStatus.TEACHER:
            confirmed = confirm(`Are you sure you want to promote ${enrollment.user.Name} to teacher status?`)
            break
        case Enrollment_UserStatus.PENDING:
            // Status pending should never be set by this function.
            // If the intent is to accept a pending enrollment, status should be set to student.
            return
    }
    if (!confirmed) {
        return
    }

    // Lookup the enrollment
    // The enrollment should be in state, if it is not, do nothing
    const enrollments = state.courseEnrollments[state.activeCourse.toString()] ?? []
    const found = enrollments.findIndex(e => e.ID === enrollment.ID)
    if (found === -1) {
        return
    }

    // Clone enrollment object and change status
    const temp = enrollment.clone()
    temp.status = status

    // Send updated enrollment to server
    const response = await effects.api.client.updateEnrollments({ enrollments: [temp] })
    if (response.error) {
        return
    }
    // If successful, update enrollment in state with new status
    if (status === Enrollment_UserStatus.NONE) {
        // If the enrollment is rejected, remove it from state
        enrollments.splice(found, 1)
    } else {
        // If the enrollment is accepted, update the enrollment in state
        enrollments[found].status = status
    }
}

/** approvePendingEnrollments approves all pending enrollments for the current course */
export const approvePendingEnrollments = async ({ state, actions, effects }: Context): Promise<void> => {
    if (!confirm("Please confirm that you want to approve all students")) {
        return
    }

    // Clone and set status to student for all pending enrollments.
    // We need to clone the enrollments to avoid modifying the state directly.
    // We do not want to update set the enrollment status before the update is successful.
    const enrollments = state.pendingEnrollments.map(e => {
        const temp = e.clone()
        temp.status = Enrollment_UserStatus.STUDENT
        return temp
    })

    // Send updated enrollments to server
    const response = await effects.api.client.updateEnrollments({ enrollments })
    if (response.error) {
        // Fetch enrollments again if update failed in case the user was able to approve some enrollments
        await actions.getEnrollmentsByCourse({ courseID: state.activeCourse, statuses: [Enrollment_UserStatus.PENDING] })
        return
    }
    for (const enrollment of state.pendingEnrollments) {
        enrollment.status = Enrollment_UserStatus.STUDENT
    }
}

/** Get assignments for all the courses the current user is enrolled in */
export const getAssignments = async ({ state, actions }: Context): Promise<void> => {
    await Promise.all(state.enrollments.map(async enrollment => {
        if (isPending(enrollment)) {
            // No need to get assignments for pending enrollments
            return
        }
        await actions.getAssignmentsByCourse(enrollment.courseID)
    }))
}

/** Get assignments for a single course, given by courseID */
export const getAssignmentsByCourse = async ({ state, effects }: Context, courseID: bigint): Promise<void> => {
    const response = await effects.api.client.getAssignments({ courseID })
    if (response.error) {
        return
    }
    state.assignments[courseID.toString()] = response.message.assignments
}

export const getRepositories = async ({ state, effects }: Context): Promise<void> => {
    await Promise.all(state.enrollments.map(async enrollment => {
        if (isPending(enrollment)) {
            // No need to get repositories for pending enrollments
            return
        }
        const courseID = enrollment.courseID
        state.repositories[courseID.toString()] = {}
        const response = await effects.api.client.getRepositories({ courseID })
        if (response.error) {
            return
        }
        state.repositories[courseID.toString()] = response.message.URLs
    }))
}

export const getGroup = async ({ state, effects }: Context, enrollment: Enrollment): Promise<void> => {
    const response = await effects.api.client.getGroup({ courseID: enrollment.courseID, groupID: enrollment.groupID })
    if (response.error) {
        return
    }
    state.userGroup[enrollment.courseID.toString()] = response.message
}

export const createGroup = async ({ state, actions, effects }: Context, group: CourseGroup): Promise<void> => {
    const check = validateGroup(group)
    if (!check.valid) {
        actions.alert({ text: check.message, color: Color.RED, delay: 10000 })
        return
    }

    const response = await effects.api.client.createGroup({
        courseID: group.courseID,
        name: group.name,
        users: group.users.map(userID => new User({ ID: userID }))
    })

    if (response.error) {
        return
    }

    state.userGroup[group.courseID.toString()] = response.message
    state.activeGroup = null
}

/** getOrganization returns the organization object for orgName retrieved from the server. */
export const getOrganization = async ({ effects }: Context, orgName: string): Promise<Response<Organization>> => {
    return await effects.api.client.getOrganization({ ScmOrganizationName: orgName })
}

/** Updates a given course and refreshes courses in state if successful  */
export const editCourse = async ({ actions, effects }: Context, { course }: { course: Course }): Promise<void> => {
    const response = await effects.api.client.updateCourse(course)
    if (response.error) {
        return
    }
    await actions.getCourses()
}

/** Fetches and stores all submissions of a given course into state. Triggers the loading spinner. */
export const loadCourseSubmissions = async ({ state, actions }: Context, courseID: bigint): Promise<void> => {
    state.isLoading = true
    await actions.refreshCourseSubmissions(courseID)
    state.loadedCourse[courseID.toString()] = true
    state.isLoading = false
}

/** Refreshes all submissions for a given course. Calling this action directly will not trigger the loading spinner.
 *  Use `loadCourseSubmissions` instead if you want to trigger the loading spinner, such as on page load. */
export const refreshCourseSubmissions = async ({ state, effects }: Context, courseID: bigint): Promise<void> => {
    // None of these should fail independently.
    const userResponse = await effects.api.client.getSubmissionsByCourse({
        CourseID: courseID,
        FetchMode: {
            case: "Type",
            value: SubmissionRequest_SubmissionType.ALL
        }
    })
    const groupResponse = await effects.api.client.getSubmissionsByCourse({
        CourseID: courseID,
        FetchMode: {
            case: "Type",
            value: SubmissionRequest_SubmissionType.GROUP
        }
    })
    if (userResponse.error || groupResponse.error) {
        return
    }

    state.submissionsForCourse.setSubmissions("USER", userResponse.message)
    state.submissionsForCourse.setSubmissions("GROUP", groupResponse.message)

    for (const submissions of Object.values(userResponse.message.submissions)) {
        for (const submission of submissions.submissions) {
            state.review.reviews.set(submission.ID, submission.reviews)
        }
    }
}

export const getGroupsByCourse = async ({ state, effects }: Context, courseID: bigint): Promise<void> => {
    state.groups[courseID.toString()] = []
    const response = await effects.api.client.getGroupsByCourse({ courseID })
    if (response.error) {
        return
    }
    state.groups[courseID.toString()] = response.message.groups
}

export const getUserSubmissions = async ({ state, effects }: Context, courseID: bigint): Promise<void> => {
    const id = courseID.toString()
    if (!state.submissions[id]) {
        state.submissions[id] = []
    }
    const response = await effects.api.client.getSubmissions({
        CourseID: courseID,
        FetchMode: {
            case: "UserID",
            value: state.self.ID,
        },
    })
    if (response.error) {
        return
    }
    // Insert submissions into state.submissions by the assignment order
    state.assignments[id]?.forEach(assignment => {
        const submission = response.message.submissions.find(s => s.AssignmentID === assignment.ID)
        if (!state.submissions[id][assignment.order - 1]) {
            state.submissions[id][assignment.order - 1] = submission ? submission : new Submission()
        }
    })
}

export const getGroupSubmissions = async ({ state, effects }: Context, courseID: bigint): Promise<void> => {
    const enrollment = state.enrollmentsByCourseID[courseID.toString()]
    if (!(enrollment && enrollment.group)) {
        return
    }
    const response = await effects.api.client.getSubmissions({
        CourseID: courseID,
        FetchMode: {
            case: "GroupID",
            value: enrollment.groupID,
        },
    })
    if (response.error) {
        return
    }
    state.assignments[courseID.toString()]?.forEach(assignment => {
        const submission = response.message.submissions.find(sbm => sbm.AssignmentID === assignment.ID)
        if (submission && assignment.isGroupLab) {
            state.submissions[courseID.toString()][assignment.order - 1] = submission
        }
    })
}

export const setActiveCourse = ({ state }: Context, courseID: bigint): void => {
    state.activeCourse = courseID
}

export const toggleFavorites = ({ state }: Context): void => {
    state.showFavorites = !state.showFavorites
}

export const setSelectedAssignmentID = ({ state }: Context, assignmentID: number): void => {
    state.selectedAssignmentID = assignmentID
}

export const setSelectedSubmission = ({ state }: Context, submission: Submission): void => {
    state.selectedSubmission = submission.clone()
}

export const getSubmission = async ({ state, effects }: Context, { courseID, owner, submission }: { courseID: bigint, owner: SubmissionOwner, submission: Submission }): Promise<void> => {
    const response = await effects.api.client.getSubmission({
        CourseID: courseID,
        FetchMode: {
            case: "SubmissionID",
            value: submission.ID,
        },
    })
    if (response.error) {
        return
    }
    state.submissionsForCourse.update(owner, response.message)
    if (state.selectedSubmission && state.selectedSubmission.ID === submission.ID) {
        // Only update the selected submission if it is the same as the one we just fetched.
        // This is to avoid overwriting the selected submission with a different one.
        // This can happen when the user clicks on a submission in the submission list, and then
        // selects a different submission in the submission list before the first request has finished.
        state.selectedSubmission = response.message
    }
}

/** Rebuilds the currently active submission */
export const rebuildSubmission = async ({ state, actions, effects }: Context, { owner, submission }: { owner: SubmissionOwner, submission: Submission | null }): Promise<void> => {
    if (!(submission && state.selectedAssignment && state.activeCourse)) {
        return
    }
    const response = await effects.api.client.rebuildSubmissions({
        courseID: state.activeCourse,
        assignmentID: state.selectedAssignment.ID,
        submissionID: submission.ID,
    })
    if (response.error) {
        return
    }
    // TODO: Alerting is temporary due to the fact that the server no longer returns the updated submission.
    // TODO: gRPC streaming should be implemented to send the updated submission to the api.client.
    await actions.getSubmission({ courseID: state.activeCourse, submission, owner })
    actions.alert({ color: Color.GREEN, text: 'Submission rebuilt successfully' })
}

/* rebuildAllSubmissions rebuilds all submissions for a given assignment */
export const rebuildAllSubmissions = async ({ effects }: Context, { courseID, assignmentID }: { courseID: bigint, assignmentID: bigint }): Promise<boolean> => {
    const response = await effects.api.client.rebuildSubmissions({
        courseID,
        assignmentID,
    })
    return !response.error
}

/** Enrolls a user (self) in a course given by courseID. Refreshes enrollments in state if enroll is successful. */
export const enroll = async ({ state, effects }: Context, courseID: bigint): Promise<void> => {
    const response = await effects.api.client.createEnrollment({
        courseID,
        userID: state.self.ID,
    })
    if (response.error) {
        return
    }
    const enrolsResponse = await effects.api.client.getEnrollments({
        FetchMode: {
            case: "userID",
            value: state.self.ID,
        }
    })

    if (enrolsResponse.error) {
        return
    }
    state.enrollments = enrolsResponse.message.enrollments

}

export const updateGroupStatus = async ({ effects }: Context, { group, status }: { group: Group, status: Group_GroupStatus }): Promise<void> => {
    const oldStatus = group.status
    group.status = status
    const response = await effects.api.client.updateGroup(group)
    if (response.error) {
        group.status = oldStatus
    }
}

export const deleteGroup = async ({ state, actions, effects }: Context, group: Group): Promise<void> => {
    if (!confirm("Deleting a group is an irreversible action. Are you sure?")) {
        return
    }
    const proceed = await actions.internal.isEmptyRepo({ groupID: group.ID, courseID: group.courseID })
    if (!proceed) {
        return
    }

    const deleteResponse = await effects.api.client.deleteGroup({
        courseID: group.courseID,
        groupID: group.ID,
    })
    if (deleteResponse.error) {
        return
    }
    state.groups[group.courseID.toString()] = state.groups[group.courseID.toString()].filter(g => g.ID !== group.ID)
}

export const updateGroup = async ({ state, actions, effects }: Context, group: Group): Promise<void> => {
    const response = await effects.api.client.updateGroup(group)
    if (response.error) {
        return
    }
    const found = state.groups[group.courseID.toString()].find(g => g.ID === group.ID)
    if (found && response.message) {
        Object.assign(found, response.message)
        actions.setActiveGroup(null)
    }
}

export const createOrUpdateCriterion = async ({ effects }: Context, { criterion, assignment }: { criterion: GradingCriterion, assignment: Assignment }): Promise<void> => {
    const benchmark = assignment.gradingBenchmarks.find(bm => bm.ID === criterion.ID)
    if (!benchmark) {
        // If a benchmark is not found, the criterion is invalid.
        return
    }

    // Existing criteria have a criteria id > 0, new criteria have a criteria id of 0
    if (criterion.ID) {
        const response = await effects.api.client.updateCriterion(criterion)
        if (response.error) {
            return
        }
        const index = benchmark.criteria.findIndex(c => c.ID === criterion.ID)
        if (index > -1) {
            benchmark.criteria[index] = criterion
        }
    } else {
        const response = await effects.api.client.createCriterion(criterion)
        if (response.error) {
            return
        }
        benchmark.criteria.push(response.message)
    }
}

export const createOrUpdateBenchmark = async ({ effects }: Context, { benchmark, assignment }: { benchmark: GradingBenchmark, assignment: Assignment }): Promise<void> => {
    // Check if this need cloning
    const bm = benchmark.clone()
    if (benchmark.ID) {
        const response = await effects.api.client.updateBenchmark(bm)
        if (response.error) {
            return
        }
        const index = assignment.gradingBenchmarks.indexOf(benchmark)
        if (index > -1) {
            assignment.gradingBenchmarks[index] = benchmark
        }
    } else {
        const response = await effects.api.client.createBenchmark(benchmark)
        if (response.error) {
            return
        }
        assignment.gradingBenchmarks.push(response.message)
    }
}

export const createBenchmark = async ({ effects }: Context, { benchmark, assignment }: { benchmark: GradingBenchmark, assignment: Assignment }): Promise<void> => {
    benchmark.AssignmentID = assignment.ID
    const response = await effects.api.client.createBenchmark(benchmark)
    if (response.error) {
        return
    }
    assignment.gradingBenchmarks.push(benchmark)
}

export const deleteCriterion = async ({ effects }: Context, { criterion, assignment }: { criterion?: GradingCriterion, assignment: Assignment }): Promise<void> => {
    if (!criterion) {
        // Criterion is invalid
        return
    }

    const benchmark = assignment.gradingBenchmarks.find(bm => bm.ID === criterion?.ID)
    if (!benchmark) {
        // Criterion has no parent benchmark
        return
    }

    if (!confirm("Do you really want to delete this criterion?")) {
        // Do nothing if user cancels
        return
    }

    // Delete criterion
    const response = await effects.api.client.deleteCriterion(criterion)
    if (response.error) {
        return
    }

    // Remove criterion from benchmark in state if request was successful
    const index = assignment.gradingBenchmarks.indexOf(benchmark)
    if (index > -1) {
        assignment.gradingBenchmarks.splice(index, 1)
    }

}

export const deleteBenchmark = async ({ effects }: Context, { benchmark, assignment }: { benchmark?: GradingBenchmark, assignment: Assignment }): Promise<void> => {
    if (benchmark && confirm("Do you really want to delete this benchmark?")) {
        const response = await effects.api.client.deleteBenchmark(benchmark)
        if (response.error) {
            return
        }
        const index = assignment.gradingBenchmarks.indexOf(benchmark)
        if (index > -1) {
            assignment.gradingBenchmarks.splice(index, 1)
        }
    }
}

export const setActiveEnrollment = ({ state }: Context, enrollment: Enrollment | null): void => {
    state.selectedEnrollment = enrollment ? enrollment : null
}

export const startSubmissionStream = ({ actions, effects }: Context) => {
    effects.streamService.submissionStream({
        onStatusChange: actions.setConnectionStatus,
        onMessage: actions.receiveSubmission,
        onError: actions.handleStreamError,
    })
}

export const updateAssignments = async ({ actions, effects }: Context, courseID: bigint): Promise<void> => {
    const response = await effects.api.client.updateAssignments({ courseID })
    if (response.error) {
        return
    }
    actions.alert({ text: "Assignments updated", color: Color.GREEN })
}

/* fetchUserData is called when the user enters the app. It fetches all data that is needed for the user to be able to use the app. */
/* If the user is not logged in, i.e does not have a valid token, the process is aborted. */
export const fetchUserData = async ({ state, actions }: Context): Promise<boolean> => {
    const successful = await actions.getSelf()
    // If getSelf returns false, the user is not logged in. Abort.
    if (!successful) {
        state.isLoading = false
        return false
    }
    // Order matters here. Some data is dependent on other data. Ex. fetching submissions depends on enrollments.
    await actions.getEnrollmentsByUser()
    await actions.getAssignments()
    await actions.getCourses()
    const results = []
    for (const enrollment of state.enrollments) {
        const courseID = enrollment.courseID
        if (isStudent(enrollment) || isTeacher(enrollment)) {
            results.push(actions.getUserSubmissions(courseID))
            results.push(actions.getGroupSubmissions(courseID))
            const statuses = isStudent(enrollment) ? [Enrollment_UserStatus.STUDENT, Enrollment_UserStatus.TEACHER] : []
            results.push(actions.getEnrollmentsByCourse({ courseID, statuses }))
            if (enrollment.groupID > 0) {
                results.push(actions.getGroup(enrollment))
            }
        }
        if (isTeacher(enrollment)) {
            results.push(actions.getGroupsByCourse(courseID))
        }
    }
    await Promise.all(results)
    if (state.self.IsAdmin) {
        await actions.getUsers()
    }
    await actions.getRepositories()
    actions.startSubmissionStream()
    // End loading screen.
    state.isLoading = false
    return true
}

/* Utility Actions */

/** Switches between teacher and student view. */
export const changeView = async ({ state, effects }: Context, courseID: bigint): Promise<void> => {
    const enrollment = state.enrollmentsByCourseID[courseID.toString()]
    if (hasStudent(enrollment.status)) {
        const response = await effects.api.client.getEnrollments({
            FetchMode: {
                case: "userID",
                value: state.self.ID,
            },
            statuses: [Enrollment_UserStatus.TEACHER],
        })
        if (response.error) {
            return
        }
        if (response.message.enrollments.find(enrol => enrol.courseID === courseID && hasTeacher(enrol.status))) {
            enrollment.status = Enrollment_UserStatus.TEACHER
        }
    } else if (hasTeacher(enrollment.status)) {
        enrollment.status = Enrollment_UserStatus.STUDENT
    }
}

export const loading = ({ state }: Context): void => {
    state.isLoading = !state.isLoading
}

/** Sets a query string in state. */
export const setQuery = ({ state }: Context, query: string): void => {
    state.query = query
}

export const errorHandler = (context: Context, { method, error }: { method: string, error: ConnectError }): void => {
    if (!error) {
        return
    }

    // TODO(jostein): Currently all errors are handled the same way.
    // We could handle each method individually, and assign a log level to each method.
    // The log level could be determined based on user role.

    if (error.code === Code.Unauthenticated) {
        // If we end up here, the user session has expired.
        if (method === "GetUser") {
            return // Do not show alert if the user is not logged in.
        }
        context.actions.alert({
            text: "Your session has expired. Please log in again.",
            color: Color.RED
        })
        // Store an alert message in localStorage that will be displayed after reloading the page.
        localStorage.setItem("alert", "Your session has expired. Please log in again.")
    } else {
        // The error message includes the error code, while the rawMessage only includes the error message.
        //
        // error.message:     "[not_found] failed to create github application: ..."
        // error.rawMessage:  "failed to create github application: ..."
        //
        // If the current user is an admin, the method name is included along with the error code.
        // e.g. "GetOrganization: [not_found] failed to create github application: ..."
        const message = context.state.self.IsAdmin ? `${method}: ${error.message}` : error.rawMessage
        context.actions.alert({
            text: message,
            color: Color.RED
        })
    }
}

export const alert = ({ state }: Context, a: Pick<Alert, "text" | "color" | "delay">): void => {
    state.alerts.push({ id: newID(), ...a })
}

export const popAlert = ({ state }: Context, alert: Alert): void => {
    state.alerts = state.alerts.filter(a => a.id !== alert.id)
}

export const logout = ({ state }: Context): void => {
    // This does not empty the state.
    state.self = new User()
}

export const setAscending = ({ state }: Context, ascending: boolean): void => {
    state.sortAscending = ascending
}

export const setSubmissionSort = ({ state }: Context, sort: SubmissionSort): void => {
    if (state.sortSubmissionsBy !== sort) {
        state.sortSubmissionsBy = sort
    } else {
        state.sortAscending = !state.sortAscending
    }
}

export const clearSubmissionFilter = ({ state }: Context): void => {
    state.submissionFilters = []
}

export const setSubmissionFilter = ({ state }: Context, filter: string): void => {
    if (state.submissionFilters.includes(filter)) {
        state.submissionFilters = state.submissionFilters.filter(f => f !== filter)
    } else {
        state.submissionFilters.push(filter)
    }
}

export const setGroupView = ({ state }: Context, groupView: boolean): void => {
    state.groupView = groupView
}

export const setActiveGroup = ({ state }: Context, group: Group | null): void => {
    state.activeGroup = group?.clone() ?? null
}

export const updateGroupUsers = ({ state }: Context, user: User): void => {
    if (!state.activeGroup) {
        return
    }
    const group = state.activeGroup
    // Remove the user from the group if they are already in it.
    const index = group.users.findIndex(u => u.ID === user.ID)
    if (index >= 0) {
        group.users.splice(index, 1)
    } else {
        group.users.push(user)
    }
}

export const updateGroupName = ({ state }: Context, name: string): void => {
    if (!state.activeGroup) {
        return
    }
    state.activeGroup.name = name
}

export const setConnectionStatus = ({ state }: Context, status: ConnStatus) => {
    state.connectionStatus = status
}

// setSubmissionOwner sets the owner of the currently selected submission.
// The owner is either an enrollment or a group.
export const setSubmissionOwner = ({ state }: Context, owner: Enrollment | Group) => {
    if (owner instanceof Group) {
        state.submissionOwner = { type: "GROUP", id: owner.ID }
    } else {
        const groupID = state.selectedSubmission?.groupID ?? 0n
        if (groupID > 0) {
            state.submissionOwner = { type: "GROUP", id: groupID }
            return
        }
        state.submissionOwner = { type: "ENROLLMENT", id: owner.ID }
    }
}

export const updateSubmissionOwner = ({ state }: Context, owner: SubmissionOwner) => {
    state.submissionOwner = owner
}

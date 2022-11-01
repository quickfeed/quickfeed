import { Color, hasStudent, hasTeacher, isPending, isStudent, isTeacher, isVisible, SubmissionSort, SubmissionStatus } from "../Helpers"
import {
    User, Enrollment, Submission, Course, Group, GradingCriterion, Assignment, GradingBenchmark, SubmissionLink, Enrollment_UserStatus, Submission_Status, Enrollment_DisplayState, Group_GroupStatus, Repository_Type
} from "../../gen/qf/types_pb"
import { CourseSubmissions, Organization, SubmissionsForCourseRequest_Type, } from "../../gen/qf/requests_pb"
import { Alert, UserCourseSubmissions } from "./state"
import { IGrpcResponse } from "../GRPCManager"
import { Context } from "."
import { Code } from "@bufbuild/connect-web"


/** Use this to verify that a gRPC request completed without an error code */
export const success = (response: IGrpcResponse<unknown>): boolean => !(response.status.Code > 0)

export const onInitializeOvermind = ({ actions }: Context) => {
    // Currently this only alerts the user if they are not logged in after a page refresh
    const alert = localStorage.getItem("alert")
    if (alert) {
        actions.alert({ text: alert, color: Color.RED })
        localStorage.removeItem("alert")
    }
}

export const resetState = ({ state }: Context) => {
    Object.assign(state.review, {
        selectedReview: -1,
        reviews: {},
        minimumScore: 0,
        assignmentID: -1
    })

    const initialState = {
        activeAssignment: -1,
        activeCourse: -1,
        activeEnrollment: null,
        activeSubmissionLink: null,
        query: "",
        sortSubmissionsBy: SubmissionSort.Approved,
        sortAscending: true,
        submissionFilters: [],
        groupView: false,
        status: [],
        activeUser: null,
        assignments: {},
        repositories: {},

        courseGroup: { courseID: 0, enrollments: [], users: [], groupName: "" },
        alerts: [],
        isLoading: true,
        courseEnrollments: {},
        groups: {},
        users: {},
        allUsers: [],
        courses: [],
        courseSubmissions: [],
        courseGroupSubmissions: {},
        submissions: {},
        userGroup: {},
        enrollments: [],
    }

    Object.assign(state, initialState)
}

/**
 *      START CURRENT USER ACTIONS
 */

/** Fetches and stores an authenticated user in state */
export const getSelf = async ({ state, effects }: Context): Promise<boolean> => {
    console.log("getSelf")
    const user = await effects.grpcMan.getUser()
    console.log("getSelf", user)
    if (user.data) {
        state.self = user.data
        return true
    }
    return false
}

/** Gets all enrollments for the current user and stores them in state */
export const getEnrollmentsByUser = async ({ state, effects }: Context): Promise<boolean> => {
    const result = await effects.grpcMan.getEnrollmentsByUser(state.self.ID)
    if (result.data) {
        state.enrollments = result.data.enrollments
        for (const enrollment of state.enrollments) {
            state.status[enrollment.courseID.toString()] = enrollment.status
        }
        return true
    }
    return false
}

/** Fetches all users (requires admin privileges) */
export const getUsers = async ({ state, effects }: Context): Promise<void> => {
    const users = await effects.grpcMan.getUsers()
    if (users.data) {
        for (const user of users.data.users) {
            state.users[user.ID.toString()] = user
        }
        // Insert users sorted by admin privileges
        state.allUsers = users.data.users.sort((a, b) => {
            if (a.isAdmin > b.isAdmin) { return -1 }
            if (a.isAdmin < b.isAdmin) { return 1 }
            return 0
        })
    }
}

/** Changes user information server-side */
export const updateUser = async ({ actions, effects }: Context, user: User): Promise<void> => {
    const result = await effects.grpcMan.updateUser(user)
    if (success(result)) {
        await actions.getSelf()
    }
}

/**
 *      END CURRENT USER ACTIONS
 */

/** Fetches all courses */
export const getCourses = async ({ state, effects }: Context): Promise<boolean> => {
    state.courses = []
    const result = await effects.grpcMan.getCourses()
    if (result.data) {
        state.courses = result.data.courses
        return true
    }
    return false
}

/** updateAdmin is used to update the admin privileges of a user. Admin status toggles between true and false */
export const updateAdmin = async ({ state, effects }: Context, user: User): Promise<void> => {
    // Confirm that user really wants to change admin status
    if (confirm(`Are you sure you want to ${user.isAdmin ? "demote" : "promote"} ${user.name}?`)) {
        // Convert to proto object and change admin status
        const req = new User(user)
        req.isAdmin = !user.isAdmin
        // Send updated user to server
        const result = await effects.grpcMan.updateUser(req)
        if (success(result)) {
            // If successful, update user in state with new admin status
            const found = state.allUsers.findIndex(s => s.ID == user.ID)
            if (found > -1) {
                state.allUsers[found].isAdmin = req.isAdmin
            }
        }
    }
}

export const getEnrollmentsByCourse = async ({ state, effects }: Context, value: { courseID: bigint, statuses: Enrollment_UserStatus[] }): Promise<boolean> => {
    const result = await effects.grpcMan.getEnrollmentsByCourse(value.courseID, undefined, true, value.statuses)
    if (result.data) {
        state.courseEnrollments[value.courseID.toString()] = result.data.enrollments
        return true
    }
    return false
}

/**  setEnrollmentState toggles the state of an enrollment between favorite and visible */
export const setEnrollmentState = async ({ actions, effects }: Context, enrollment: Enrollment): Promise<void> => {
    enrollment.state = isVisible(enrollment) ? Enrollment_DisplayState.HIDDEN : Enrollment_DisplayState.VISIBLE
    const response = await effects.grpcMan.updateCourseVisibility(enrollment)
    if (!success(response)) {
        actions.alertHandler(response)
    }
}

/** Updates a given submission with a new status. This updates the given submission, as well as all other occurrences of the given submission in state. */
export const updateSubmission = async ({ state, actions, effects }: Context, status: Submission_Status): Promise<void> => {
    /* Do not update if the status is already the same or if there is no selected submission */
    if (!state.currentSubmission || state.currentSubmission.status == status) {
        return
    }

    /* Confirm that user really wants to change submission status */
    if (!confirm(`Are you sure you want to set status ${SubmissionStatus[status]} on this submission?`)) {
        return
    }

    /* Store the previous submission status */
    const previousStatus = state.currentSubmission.status

    /* Update the submission status */
    state.currentSubmission.status = status
    const result = await effects.grpcMan.updateSubmission(state.activeCourse, state.currentSubmission)
    if (!success(result)) {
        /* If the update failed, revert the submission status */
        state.currentSubmission.status = previousStatus
        return
    }

    if (state.activeSubmissionLink?.assignment?.isGroupLab) {
        actions.updateCurrentSubmissionStatus({ links: state.courseGroupSubmissions[state.activeCourse.toString()], status: status })
    }
    actions.updateCurrentSubmissionStatus({ links: state.courseSubmissions[state.activeCourse.toString()], status: status })
}

export const updateCurrentSubmissionStatus = ({ state }: Context, { links, status }: { links: UserCourseSubmissions[], status: Submission_Status }): void => {
    /* Loop through all submissions for the current course and update the status if it matches the current submission ID */
    for (const link of links) {
        if (!link.submissions) {
            continue
        }
        for (const submission of link.submissions) {
            if (!submission.submission) {
                continue
            }
            if (submission.submission.ID == BigInt(state.activeSubmission)) {
                submission.submission.status = status
            }
        }
    }
}

/** updateEnrollment updates an enrollment status with the given status */
export const updateEnrollment = async ({ state, effects }: Context, { enrollment, status }: { enrollment: Enrollment, status: Enrollment_UserStatus }): Promise<void> => {
    // Confirm that user really wants to change enrollment status
    let confirmed = false
    switch (status) {
        case Enrollment_UserStatus.NONE:
            confirmed = confirm("WARNING! Rejecting a student is irreversible. Are you sure?")
            break
        case Enrollment_UserStatus.STUDENT:
            // If the enrollment is pending, don't ask for confirmation
            confirmed = isPending(enrollment) || confirm(`Warning! ${enrollment.user?.name} is a teacher. Are sure you want to demote?`)
            break
        case Enrollment_UserStatus.TEACHER:
            confirmed = confirm(`Are you sure you want to promote ${enrollment.user?.name} to teacher status?`)
            break
    }

    if (confirmed) {
        // Lookup the enrollment
        // The enrollment should be in state, if it is not, do nothing
        const enrollments = state.courseEnrollments[state.activeCourse.toString()] ?? []
        const found = enrollments.findIndex(e => e.ID == enrollment.ID)
        if (found === -1) {
            return
        }

        // Clone enrollment object and change status
        const temp = enrollment.clone()
        temp.status = status

        // Send updated enrollment to server
        const response = await effects.grpcMan.updateEnrollments([temp])
        if (success(response)) {
            // If successful, update enrollment in state with new status
            if (status == Enrollment_UserStatus.NONE) {
                // If the enrollment is rejected, remove it from state
                enrollments.splice(found, 1)
            } else {
                // If the enrollment is accepted, update the enrollment in state
                enrollments[found].status = status
            }
        }
    }
}

/** approvePendingEnrollments approves all pending enrollments for the current course */
export const approvePendingEnrollments = async ({ state, actions, effects }: Context): Promise<void> => {
    if (confirm("Please confirm that you want to approve all students")) {
        // Clone and set status to student for all pending enrollments.
        // We need to clone the enrollments to avoid modifying the state directly.
        // We do not want to update set the enrollment status before the update is successful.
        const enrollments = Object.assign({}, state.pendingEnrollments)
        enrollments.forEach(e => e.status = Enrollment_UserStatus.STUDENT)

        // Send updated enrollments to server
        const response = await effects.grpcMan.updateEnrollments(enrollments)
        if (success(response)) {
            for (const enrollment of state.pendingEnrollments) {
                enrollment.status = Enrollment_UserStatus.STUDENT
            }
        } else {
            // Fetch enrollments again if update failed in case the user was able to approve some enrollments
            await actions.getEnrollmentsByCourse({ courseID: state.activeCourse, statuses: [Enrollment_UserStatus.PENDING] })
            actions.alertHandler(response)
        }
    }
}

/** Get assignments for all the courses the current user is enrolled in */
export const getAssignments = async ({ state, effects }: Context): Promise<boolean> => {
    let success = true
    for (const enrollment of state.enrollments) {
        if (isPending(enrollment)) {
            // No need to get assignments for pending enrollments
            continue
        }
        const response = await effects.grpcMan.getAssignments(enrollment.courseID)
        if (response.data) {
            // Store assignments in state by course ID
            state.assignments[enrollment.courseID.toString()] = response.data.assignments
        } else {
            success = false
        }
    }
    return success
}

/** Get assignments for a single course, given by courseID */
export const getAssignmentsByCourse = async ({ state, effects }: Context, courseID: bigint): Promise<boolean> => {
    const response = await effects.grpcMan.getAssignments(courseID)
    if (response.data) {
        state.assignments[courseID.toString()] = response.data.assignments
        return true
    }
    return false
}

export const getRepositories = async ({ state, effects }: Context): Promise<boolean> => {
    let success = true
    for (const enrollment of state.enrollments) {
        if (isPending(enrollment)) {
            // No need to get repositories for pending enrollments
            continue
        }
        const courseID = enrollment.courseID
        state.repositories[courseID.toString()] = {}

        const response = await effects.grpcMan.getRepositories(courseID, generateRepositoryList(enrollment))
        if (response.data) {
                state.repositories[courseID.toString()] = response.data.URLs
        } else {
            success = false
        }
    }
    return success
}

export const getGroupByUserAndCourse = async ({ state, effects }: Context, courseID: bigint): Promise<void> => {
    const response = await effects.grpcMan.getGroupByUserAndCourse(courseID, state.self.ID)
    if (response.data) {
        state.userGroup[courseID.toString()] = response.data
    }
}

export const createGroup = async ({ state, actions, effects }: Context, group: { courseID: bigint, users: bigint[], name: string }): Promise<void> => {
    const response = await effects.grpcMan.createGroup(group.courseID, group.name, group.users)
    if (success(response) && response.data) {
        state.userGroup[group.courseID.toString()] = response.data
        state.activeGroup = null
    } else {
        actions.alertHandler(response)
    }
}

/** getOrganization returns the organization object for orgName retrieved from the server. */
export const getOrganization = async ({ actions, effects }: Context, orgName: string): Promise<IGrpcResponse<Organization>> => {
    const response = await effects.grpcMan.getOrganization(orgName)
    if (!success(response)) {
        actions.alertHandler(response)
        return response
    }
    return response
}

/* createCourse creates a new course */
export const createCourse = async ({ state, actions, effects }: Context, value: { course: Course, org: Organization }): Promise<boolean> => {
    const course = Object.assign({}, value.course)
    /* Fill in required fields */
    course.organizationID = value.org.ID
    course.organizationName = value.org.name
    course.provider = "github"
    course.courseCreatorID = state.self.ID
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
    return false
}

/** Updates a given course and refreshes courses in state if successful  */
export const editCourse = async ({ actions, effects }: Context, { course }: { course: Course }): Promise<void> => {
    const response = await effects.grpcMan.updateCourse(course)
    if (success(response)) {
        actions.getCourses()
    } else {
        actions.alertHandler(response)
    }
}

/** getSubmissions fetches all submission for the current user by Course ID and stores them in state */
// TODO: Currently not used, see refreshSubmissions.
export const getSubmissions = async ({ state, effects }: Context, courseID: bigint): Promise<boolean> => {
    const result = await effects.grpcMan.getSubmissions(courseID, state.self.ID)
    if (result.data) {
        state.submissions[courseID.toString()] = result.data.submissions
        return true
    }
    return false
}

// TODO: Currently not in use. Requires gRPC streaming to be implemented. Intended to be used to update submissions in state when a new commit is pushed to a repository.
// TODO: A workaround to not use gRPC streaming is to ping the server at set intervals to check for new commits. This functionality was removed pending gRPC streaming implementation.
/** Updates all submissions in state where the fetched submission commit hash differs from the one in state. */
export const refreshSubmissions = async ({ state, effects }: Context, input: { courseID: number, submissionID: number }): Promise<void> => {
    const response = await effects.grpcMan.getSubmissions(BigInt(input.courseID), state.self.ID)
    if (!response.data || !success(response)) {
        return
    }
    const submissions = response.data.submissions
    for (const submission of submissions) {
        const assignment = state.assignments[input.courseID].find(a => a.ID === submission.AssignmentID)
        if (!assignment) {
            continue
        }
        if (state.submissions[input.courseID][assignment.order - 1].commitHash !== submission.commitHash) {
            state.submissions[input.courseID][assignment.order - 1] = submission
        }
    }
}

export const convertCourseSubmission = ({ state }: Context, { courseID, data }: { courseID: bigint, data: CourseSubmissions }): void => {
    state.review.reviews[courseID.toString()] = {}
    state.courseSubmissions[courseID.toString()] = []
    for (const link of data.links) {
        if (link.enrollment) {
            const submissionLinks = link.submissions
            submissionLinks.forEach(submissionLink => {
                if (submissionLink.submission) {
                    const submission = submissionLink.submission
                    state.review.reviews[courseID.toString()][Number(submission.ID)] = submission.reviews
                }
            })
            state.courseSubmissions[courseID.toString()].push({
                enrollment: link.enrollment,
                submissions: link.submissions,
                user: link.enrollment?.user
            })
        }
    }
    state.isLoading = false
}

/** Fetches and stores all submissions of a given course into state */
export const getAllCourseSubmissions = async ({ state, actions, effects }: Context, courseID: bigint): Promise<boolean> => {
    state.isLoading = true

    // None of these should fail independently.
    const result = await effects.grpcMan.getSubmissionsByCourse(courseID, SubmissionsForCourseRequest_Type.ALL)
    const groups = await effects.grpcMan.getSubmissionsByCourse(courseID, SubmissionsForCourseRequest_Type.GROUP)
    if (!success(result) || !success(groups)) {
        const failed = !success(result) ? result : groups
        actions.alertHandler(failed)
        state.isLoading = false
        return false
    }

    if (result.data) {
        actions.convertCourseSubmission({ courseID: courseID, data: result.data })
    }
    if (groups.data) {
        state.courseGroupSubmissions[courseID.toString()] = []
        groups.data.links.forEach(link => {
            if (!link.enrollment?.group) {
                return
            }
            state.courseGroupSubmissions[courseID.toString()].push({
                group: link.enrollment?.group,
                submissions: link.submissions
            })
        })
    }
    state.isLoading = false
    return true
}

export const getGroupsByCourse = async ({ state, effects }: Context, courseID: bigint): Promise<void> => {
    state.groups[courseID.toString()] = []
    const res = await effects.grpcMan.getGroupsByCourse(courseID)
    if (res.data) {
        state.groups[courseID.toString()] = res.data.groups
    }
}

export const getUserSubmissions = async ({ state, effects }: Context, courseID: bigint): Promise<boolean> => {
    state.submissions[courseID.toString()] = []
    const submissions = await effects.grpcMan.getSubmissions(courseID, state.self.ID)
    if (submissions.data) {
        // Insert submissions into state.submissions by the assignment order
        state.assignments[courseID.toString()]?.forEach(assignment => {
            const submission = submissions.data?.submissions.find(s => s.AssignmentID === assignment.ID)
            state.submissions[courseID.toString()][assignment.order - 1] = submission ? submission : new Submission()
        })
        return true
    }
    return false
}

export const getGroupSubmissions = async ({ state, effects }: Context, courseID: bigint): Promise<void> => {
    const enrollment = state.enrollmentsByCourseID[courseID.toString()]
    if (enrollment && enrollment.group) {
        const submissions = await effects.grpcMan.getGroupSubmissions(courseID, enrollment.groupID)
        state.assignments[courseID.toString()]?.forEach(assignment => {
            const submission = submissions.data?.submissions.find(submission => submission.AssignmentID === assignment.ID)
            if (submission && assignment.isGroupLab) {
                state.submissions[courseID.toString()][assignment.order - 1] = submission
            }
        })
    }
}

export const setActiveCourse = ({ state }: Context, courseID: bigint): void => {
    state.activeCourse = courseID
}

export const toggleFavorites = ({ state }: Context): void => {
    state.showFavorites = !state.showFavorites
}

export const setActiveAssignment = ({ state }: Context, assignmentID: number): void => {
    state.activeAssignment = assignmentID
}

export const getSubmission = async ({ state, effects }: Context, { courseID, submissionID }: { courseID: bigint, submissionID: bigint }): Promise<void> => {
    const response = await effects.grpcMan.getSubmission(courseID, submissionID)
    if (!response.data || !success(response)) {
        return
    }
    const submissions = state.groupView ? state.courseGroupSubmissions[courseID.toString()] : state.courseSubmissions[courseID.toString()]
    if (!submissions) {
        return
    }
    submissions.forEach(link => {
        const sub = link.submissions?.find(submission => submission.submission?.ID === submissionID)
        if (sub?.submission && response.data) {
            sub.submission = response.data
            if (state.activeSubmissionLink) {
                state.activeSubmissionLink.submission = response.data
            }
        }
    })
}

/** Rebuilds the currently active submission */
export const rebuildSubmission = async ({ state, actions, effects }: Context): Promise<boolean> => {
    if (state.currentSubmission && state.selectedAssignment && state.activeCourse) {
        const response = await effects.grpcMan.rebuildSubmission(state.selectedAssignment.ID, BigInt(state.activeSubmission), state.activeCourse)
        if (success(response)) {
            // TODO: Alerting is temporary due to the fact that the server no longer returns the updated submission.
            // TODO: gRPC streaming should be implemented to send the updated submission to the client.
            await actions.getSubmission({ courseID: state.activeCourse, submissionID: BigInt(state.activeSubmission) })
            actions.alert({ color: Color.GREEN, text: 'Submission rebuilt successfully' })
            return true
        }
    }
    return false
}

/* rebuildAllSubmissions rebuilds all submissions for a given assignment */
export const rebuildAllSubmissions = async ({ effects }: Context, { courseID, assignmentID }: { courseID: bigint, assignmentID: bigint }): Promise<boolean> => {
    const response = await effects.grpcMan.rebuildSubmissions(assignmentID, courseID)
    return success(response)
}

/** Enrolls a user (self) in a course given by courseID. Refreshes enrollments in state if enroll is successful. */
export const enroll = async ({ state, effects }: Context, courseID: bigint): Promise<void> => {
    const response = await effects.grpcMan.createEnrollment(courseID, state.self.ID)
    if (success(response)) {
        const enrollments = await effects.grpcMan.getEnrollmentsByUser(state.self.ID)
        if (enrollments.data) {
            state.enrollments = enrollments.data.enrollments
        }
    }
}

export const updateGroupStatus = async ({ effects }: Context, { group, status }: { group: Group, status: Group_GroupStatus }): Promise<void> => {
    const oldStatus = group.status
    group.status = status
    const response = await effects.grpcMan.updateGroup(group)
    if (!success(response)) {
        group.status = oldStatus
    }
}

export const deleteGroup = async ({ state, effects }: Context, group: Group): Promise<void> => {
    if (confirm("Deleting a group is an irreversible action. Are you sure?")) {
        const isRepoEmpty = await effects.grpcMan.isEmptyRepo(group.courseID, BigInt(0), group.ID)
        if (isRepoEmpty || confirm(`Warning! Group repository is not empty! Do you still want to delete group, github team and group repository?`)) {
            const response = await effects.grpcMan.deleteGroup(group.courseID, group.ID)
            if (success(response)) {
                state.groups[group.courseID.toString()] = state.groups[group.courseID.toString()].filter(g => g.ID !== group.ID)
            }
        }
    }
}

export const updateGroup = async ({ state, actions, effects }: Context, group: Group): Promise<void> => {
    const response = await effects.grpcMan.updateGroup(group)
    if (success(response)) {
        const found = state.groups[group.courseID.toString()].find(g => g.ID === group.ID)
        if (found && response.data) {
            Object.assign(found, response.data)
            actions.setActiveGroup(null)
        }
    } else {
        actions.alertHandler(response)
    }
}

export const createOrUpdateCriterion = async ({ effects }: Context, { criterion, assignment }: { criterion: GradingCriterion, assignment: Assignment }): Promise<void> => {
    for (const bm of assignment.gradingBenchmarks) {
        if (bm.ID === criterion.BenchmarkID) {
            // Existing criteria have a criteria id > 0, new criteria have a criteria id of 0
            if (criterion.ID && success(await effects.grpcMan.updateCriterion(criterion))) {
                const index = bm.criteria.findIndex(c => c.ID === criterion.ID)
                if (index > -1) {
                    bm.criteria[index] = criterion
                }
            } else {
                const response = await effects.grpcMan.createCriterion(criterion)
                if (success(response) && response.data) {
                    bm.criteria.push(response.data)
                }
            }
        }
    }
}

export const createOrUpdateBenchmark = async ({ effects }: Context, { benchmark, assignment }: { benchmark: GradingBenchmark, assignment: Assignment }): Promise<void> => {
    // Check if this need cloning
    const bm = benchmark.clone()
    if (benchmark.ID && success(await effects.grpcMan.updateBenchmark(bm))) {
        const index = assignment.gradingBenchmarks.indexOf(benchmark)
        if (index > -1) {
            assignment.gradingBenchmarks[index] = benchmark
        }
    } else {
        const response = await effects.grpcMan.createBenchmark(bm)
        if (success(response) && response.data) {
            assignment.gradingBenchmarks.push(response.data)
        }
    }
}

export const createBenchmark = async ({ effects }: Context, { benchmark, assignment }: { benchmark: GradingBenchmark, assignment: Assignment }): Promise<void> => {
    benchmark.AssignmentID = assignment.ID
    const response = await effects.grpcMan.createBenchmark(benchmark)
    if (success(response)) {
        assignment.gradingBenchmarks.push(benchmark)
    }
}

export const deleteCriterion = async ({ actions, effects }: Context, { criterion, assignment }: { criterion?: GradingCriterion, assignment: Assignment }): Promise<void> => {
    for (const benchmark of assignment.gradingBenchmarks) {
        if (benchmark.ID !== criterion?.BenchmarkID) {
            continue
        }
        if (!confirm("Do you really want to delete this criterion?")) {
            // Do nothing if user cancels
            return
        }

        // Delete criterion
        const response = await effects.grpcMan.deleteCriterion(criterion)
        if (success(response)) {
            // Remove criterion from benchmark in state if successful
            const index = assignment.gradingBenchmarks.indexOf(benchmark)
            if (index > -1) {
                assignment.gradingBenchmarks.splice(index, 1)
            }
        } else {
            actions.alertHandler(response)
        }
    }
}

export const deleteBenchmark = async ({ actions, effects }: Context, { benchmark, assignment }: { benchmark?: GradingBenchmark, assignment: Assignment }): Promise<void> => {
    if (benchmark && confirm("Do you really want to delete this benchmark?")) {
        const response = await effects.grpcMan.deleteBenchmark(benchmark)
        if (success(response)) {
            const index = assignment.gradingBenchmarks.indexOf(benchmark)
            if (index > -1) {
                assignment.gradingBenchmarks.splice(index, 1)
            }
        } else {
            actions.alertHandler(response)
        }
    }
}

export const refreshSubmission = async ({ effects }: Context, { link }: { link: SubmissionLink }): Promise<SubmissionLink> => {
    if (link.submission && link.assignment) {
        const response = await effects.grpcMan.getSubmission(link.assignment.CourseID, link.submission.ID)
        if (success(response) && response.data) {
            link.submission = response.data
        }
    }
    return link
}

export const setActiveSubmissionLink = async ({ state, actions }: Context, link: SubmissionLink): Promise<void> => {
    link = await actions.refreshSubmission({ link })
    state.activeSubmissionLink = link ? link : null
}

export const setActiveEnrollment = ({ state }: Context, enrollment: Enrollment): void => {
    state.activeEnrollment = enrollment ? enrollment : null
}

/* fetchUserData is called when the user enters the app. It fetches all data that is needed for the user to be able to use the app. */
/* If the user is not logged in, i.e does not have a valid token, the process is aborted. */
export const fetchUserData = async ({ state, actions }: Context): Promise<boolean> => {
    let success = await actions.getSelf()
    console.log("Hello")
    // If getSelf returns false, the user is not logged in. Abort.
    if (!success) { state.isLoading = false; return false }

    // Start fetching all data. Loading screen will be shown until all data is fetched, i.e state.isLoading is set to false.
    while (state.isLoading) {
        // Order matters here. Some data is dependent on other data. Ex. fetching submissions depends on enrollments.
        success = await actions.getEnrollmentsByUser()
        success = await actions.getAssignments()
        for (const enrollment of state.enrollments) {
            const courseID = enrollment.courseID
            if (isStudent(enrollment) || isTeacher(enrollment)) {
                success = await actions.getUserSubmissions(courseID)
                await actions.getGroupSubmissions(courseID)
                const statuses = isStudent(enrollment) ? [Enrollment_UserStatus.STUDENT, Enrollment_UserStatus.TEACHER] : []
                success = await actions.getEnrollmentsByCourse({ courseID: courseID, statuses: statuses })
                if (enrollment.groupID > 0) {
                    await actions.getGroupByUserAndCourse(courseID)
                }
            }
            if (isTeacher(enrollment)) {
                actions.getGroupsByCourse(courseID)
            }
        }
        if (state.self.isAdmin) {
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
}

/* Utility Actions */

/** Switches between teacher and student view. */
export const changeView = async ({ state, effects }: Context, courseID: bigint): Promise<void> => {
    const enrollment = state.enrollmentsByCourseID[courseID.toString()]
    if (hasStudent(enrollment.status)) {
        const status = await effects.grpcMan.getEnrollmentsByUser(state.self.ID, [Enrollment_UserStatus.TEACHER])
        if (status.data?.enrollments.find(enrollment => enrollment.courseID == BigInt(courseID) && hasTeacher(enrollment.status))) {
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

export const setSelectedUser = ({ state }: Context, user: User | null): void => {
    state.activeUser = user
}


export const alertHandler = ({ state }: Context, response: IGrpcResponse<unknown>): void => {
    if (response.status.Code === BigInt(Code.Unauthenticated)) {
        // If we end up here, the user session has expired.
        // Store an alert message in localStorage that will be displayed after reloading the page.
        localStorage.setItem("alert", "Your session has expired. Please log in again.")
        window.location.reload()
    } else if (response.status.Code >= BigInt(0)) {
        state.alerts.push({ text: response.status.Error, color: Color.RED })
    }
}

export const alert = ({ state }: Context, alert: Alert): void => {
    state.alerts.push(alert)
}

export const popAlert = ({ state }: Context, index: number): void => {
    state.alerts = state.alerts.filter((_, i) => i !== index)
}

export const logout = ({ state }: Context): void => {
    // This does not empty the state.
    state.self = new User()
}

const generateRepositoryList = (enrollment: Enrollment): Repository_Type[] => {
    switch (enrollment.status) {
        case Enrollment_UserStatus.TEACHER:
            return [Repository_Type.ASSIGNMENTS, Repository_Type.INFO, Repository_Type.GROUP, Repository_Type.TESTS, Repository_Type.USER]
        case Enrollment_UserStatus.STUDENT:
            return [Repository_Type.ASSIGNMENTS, Repository_Type.INFO, Repository_Type.GROUP, Repository_Type.USER]
        default:
            return [Repository_Type.NONE]
    }
}

export const setAscending = ({ state }: Context, ascending: boolean): void => {
    state.sortAscending = ascending
}

export const setSubmissionSort = ({ state }: Context, sort: SubmissionSort): void => {
    if (state.sortSubmissionsBy != sort) {
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
        state.submissionFilters = state.submissionFilters.filter(f => f != filter)
    } else {
        state.submissionFilters.push(filter)
    }
}

export const setGroupView = ({ state }: Context, groupView: boolean): void => {
    state.groupView = groupView
}

export const setActiveGroup = ({ state }: Context, group: Group | null): void => {
    state.activeGroup =  group?.clone() ?? null
}

export const updateGroupUsers = ({ state }: Context, user: User): void => {
    if (!state.activeGroup) {
        return
    }
    const group = state.activeGroup
    // Remove the user from the group if they are already in it.
    const index = group.users.findIndex(u => u.ID == user.ID)
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

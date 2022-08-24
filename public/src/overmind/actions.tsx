import { Color, hasStudent, hasTeacher, isPending, isStudent, isTeacher, isVisible, SubmissionSort, SubmissionStatus } from "../Helpers"
import {
    User, Enrollment, Submission, Repository, Course, Group, GradingCriterion, Assignment, GradingBenchmark, SubmissionLink
} from "../../proto/qf/types_pb"
import { SubmissionsForCourseRequest, CourseSubmissions, Organization, } from "../../proto/qf/requests_pb"
import { Alert, UserCourseSubmissions } from "./state"
import { IGrpcResponse } from "../GRPCManager"
import { StatusCode } from "grpc-web"
import { Context } from "."
import { Converter } from "../convert"


/** Use this to verify that a gRPC request completed without an error code */
export const success = (response: IGrpcResponse<unknown>): boolean => response.status.getCode() === 0

export const onInitializeOvermind = async ({ actions }: Context): Promise<void> => {
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
    const user = await effects.grpcMan.getUser()
    if (user.data) {
        state.self = user.data.toObject()
        return true
    }
    return false
}

/** Gets all enrollments for the current user and stores them in state */
export const getEnrollmentsByUser = async ({ state, effects }: Context): Promise<boolean> => {
    const result = await effects.grpcMan.getEnrollmentsByUser(state.self.id)
    if (result.data) {
        state.enrollments = result.data.getEnrollmentsList().map(e => e.toObject())
        for (const enrollment of state.enrollments) {
            state.status[enrollment.courseid] = enrollment.status
        }
        return true
    }
    return false
}

/** Fetches all users (requires admin privileges) */
export const getUsers = async ({ state, effects }: Context): Promise<void> => {
    const users = await effects.grpcMan.getUsers()
    if (users.data) {
        for (const user of users.data.getUsersList()) {
            state.users[user.getId()] = user.toObject()
        }
        // Insert users sorted by admin privileges
        state.allUsers = users.data.getUsersList().map(u => u.toObject()).sort((a, b) => {
            if (a.isadmin > b.isadmin) { return -1 }
            if (a.isadmin < b.isadmin) { return 1 }
            return 0
        })
    }
}

/** Changes user information server-side */
export const updateUser = async ({ actions, effects }: Context, user: User.AsObject): Promise<void> => {
    const result = await effects.grpcMan.updateUser(Converter.toUser(user))
    if (result.status.getCode() === 0) {
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
        state.courses = result.data.getCoursesList().map((course => course.toObject()))
        return true
    }
    return false
}

/** updateAdmin is used to update the admin privileges of a user. Admin status toggles between true and false */
export const updateAdmin = async ({ state, effects }: Context, user: User.AsObject): Promise<void> => {
    // Confirm that user really wants to change admin status
    if (confirm(`Are you sure you want to ${user.isadmin ? "demote" : "promote"} ${user.name}?`)) {
        // Convert to proto object and change admin status
        const protoUser = Converter.toUser(user)
        protoUser.setIsadmin(!user.isadmin)
        // Send updated user to server
        const result = await effects.grpcMan.updateUser(protoUser)
        if (success(result)) {
            // If successful, update user in state with new admin status
            const found = state.allUsers.findIndex(s => s.id == user.id)
            if (found > -1) {
                state.allUsers[found].isadmin = protoUser.getIsadmin()
            }
        }
    }
}

export const getEnrollmentsByCourse = async ({ state, effects }: Context, value: { courseID: number, statuses: Enrollment.UserStatus[] }): Promise<boolean> => {
    const result = await effects.grpcMan.getEnrollmentsByCourse(value.courseID, undefined, true, value.statuses)
    if (result.data) {
        state.courseEnrollments[value.courseID] = result.data.getEnrollmentsList().map((e) => e.toObject())
        return true
    }
    return false
}

/**  setEnrollmentState toggles the state of an enrollment between favorite and visible */
export const setEnrollmentState = async ({ actions, effects }: Context, enrollment: Enrollment.AsObject): Promise<void> => {
    enrollment.state = isVisible(enrollment) ? Enrollment.DisplayState.HIDDEN : Enrollment.DisplayState.VISIBLE
    const response = await effects.grpcMan.updateCourseVisibility(Converter.toEnrollment(enrollment))
    if (!success(response)) {
        actions.alertHandler(response)
    }
}

/** Updates a given submission with a new status. This updates the given submission, as well as all other occurrences of the given submission in state. */
export const updateSubmission = async ({ state, actions, effects }: Context, status: Submission.Status): Promise<void> => {
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
    const result = await effects.grpcMan.updateSubmission(state.activeCourse, Converter.toSubmission(state.currentSubmission))
    if (!success(result)) {
        /* If the update failed, revert the submission status */
        state.currentSubmission.status = previousStatus
        return
    }

    if (state.activeSubmissionLink?.assignment?.isgrouplab) {
        actions.updateCurrentSubmissionStatus({ links: state.courseGroupSubmissions[state.activeCourse], status: status })
    }
    actions.updateCurrentSubmissionStatus({ links: state.courseSubmissions[state.activeCourse], status: status })
}

export const updateCurrentSubmissionStatus = ({ state }: Context, { links, status }: { links: UserCourseSubmissions[], status: Submission.Status }): void => {
    /* Loop through all submissions for the current course and update the status if it matches the current submission ID */
    for (const link of links) {
        if (!link.submissions) {
            continue
        }
        for (const submission of link.submissions) {
            if (!submission.submission) {
                continue
            }
            if (submission.submission.id == state.activeSubmission) {
                submission.submission.status = status
            }
        }
    }
}

/** updateEnrollment updates an enrollment status with the given status */
export const updateEnrollment = async ({ state, effects }: Context, { enrollment, status }: { enrollment: Enrollment.AsObject, status: Enrollment.UserStatus }): Promise<void> => {
    // Confirm that user really wants to change enrollment status
    let confirmed = false
    switch (status) {
        case Enrollment.UserStatus.NONE:
            confirmed = confirm("WARNING! Rejecting a student is irreversible. Are you sure?")
            break
        case Enrollment.UserStatus.STUDENT:
            // If the enrollment is pending, don't ask for confirmation
            confirmed = isPending(enrollment) || confirm(`Warning! ${enrollment.user?.name} is a teacher. Are sure you want to demote?`)
            break
        case Enrollment.UserStatus.TEACHER:
            confirmed = confirm(`Are you sure you want to promote ${enrollment.user?.name} to teacher status?`)
            break
    }

    if (confirmed) {
        // Lookup the enrollment
        // The enrollment should be in state, if it is not, do nothing
        const enrollments = state.courseEnrollments[state.activeCourse] ?? []
        const found = enrollments.findIndex(e => e.id == enrollment.id)
        if (found === -1) {
            return
        }

        // Clone enrollment object and change status
        const temp = Converter.clone(enrollment)
        temp.status = status

        // Send updated enrollment to server
        const response = await effects.grpcMan.updateEnrollments([Converter.toEnrollment(temp)])
        if (success(response)) {
            // If successful, update enrollment in state with new status
            if (status == Enrollment.UserStatus.NONE) {
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
        // Clone and set status to student for all pending enrollments
        const enrollments = Object.assign({}, state.pendingEnrollments)
        enrollments.forEach(e => e.status = Enrollment.UserStatus.STUDENT)

        // Send updated enrollments to server
        const response = await effects.grpcMan.updateEnrollments(enrollments.map(e => Converter.toEnrollment(e)))
        if (success(response)) {
            for (const enrollment of state.pendingEnrollments) {
                enrollment.status = Enrollment.UserStatus.STUDENT
            }
        } else {
            // Fetch enrollments again if update failed in case the user was able to approve some enrollments
            await actions.getEnrollmentsByCourse({ courseID: state.activeCourse, statuses: [Enrollment.UserStatus.PENDING] })
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
        const response = await effects.grpcMan.getAssignments(enrollment.courseid)
        if (response.data) {
            // Store assignments in state by course ID
            state.assignments[enrollment.courseid] = response.data.getAssignmentsList().map(a => a.toObject())
        } else {
            success = false
        }
    }
    return success
}

/** Get assignments for a single course, given by courseID */
export const getAssignmentsByCourse = async ({ state, effects }: Context, courseID: number): Promise<boolean> => {
    const response = await effects.grpcMan.getAssignments(courseID)
    if (response.data) {
        state.assignments[courseID] = response.data.getAssignmentsList().map(a => a.toObject())
        return true
    }
    return false
}

type RepoKey = keyof typeof Repository.Type

export const getRepositories = async ({ state, effects }: Context): Promise<boolean> => {
    let success = true
    for (const enrollment of state.enrollments) {
        const courseID = enrollment.courseid
        state.repositories[courseID] = {}

        const response = await effects.grpcMan.getRepositories(courseID, generateRepositoryList(Converter.toEnrollment(enrollment)))
        if (response.data) {
            response.data.getUrlsMap().forEach((entry, key) => {
                state.repositories[courseID][Repository.Type[key as RepoKey]] = entry
            })
        } else {
            success = false
        }
    }
    return success
}

export const getGroupByUserAndCourse = async ({ state, effects }: Context, courseID: number): Promise<void> => {
    const response = await effects.grpcMan.getGroupByUserAndCourse(courseID, state.self.id)
    if (response.data) {
        state.userGroup[courseID] = response.data.toObject()
    }
}

export const createGroup = async ({ state, actions, effects }: Context, group: { courseID: number, users: number[], name: string }): Promise<void> => {
    const response = await effects.grpcMan.createGroup(group.courseID, group.name, group.users)
    if (success(response) && response.data) {
        state.userGroup[group.courseID] = response.data.toObject()
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
export const createCourse = async ({ state, actions, effects }: Context, value: { course: Course.AsObject, org: Organization }): Promise<boolean> => {
    const course = Object.assign({}, value.course)
    /* Fill in required fields */
    course.organizationid = value.org.getId()
    course.organizationpath = value.org.getPath()
    course.provider = "github"
    course.coursecreatorid = state.self.id
    /* Send the course to the server */
    const response = await effects.grpcMan.createCourse(Converter.toCourse(course))
    if (response.data) {
        /* If successful, add the course to the state */
        state.courses.push(response.data.toObject())
        /* User that created the course is automatically enrolled in the course. Refresh the enrollment list */
        actions.getEnrollmentsByUser()
        return true
    }
    actions.alertHandler(response)
    return false
}

/** Updates a given course and refreshes courses in state if successful  */
export const editCourse = async ({ actions, effects }: Context, { course }: { course: Course.AsObject }): Promise<void> => {
    const response = await effects.grpcMan.updateCourse(Converter.toCourse(course))
    if (success(response)) {
        actions.getCourses()
    } else {
        actions.alertHandler(response)
    }
}

/** getSubmissions fetches all submission for the current user by Course ID and stores them in state */
// TODO: Currently not used, see refreshSubmissions.
export const getSubmissions = async ({ state, effects }: Context, courseID: number): Promise<boolean> => {
    const result = await effects.grpcMan.getSubmissions(courseID, state.self.id)
    if (result.data) {
        state.submissions[courseID] = result.data.getSubmissionsList().map(s => s.toObject())
        return true
    }
    return false
}

// TODO: Currently not in use. Requires gRPC streaming to be implemented. Intended to be used to update submissions in state when a new commit is pushed to a repository.
// TODO: A workaround to not use gRPC streaming is to ping the server at set intervals to check for new commits. This functionality was removed pending gRPC streaming implementation.
/** Updates all submissions in state where the fetched submission commit hash differs from the one in state. */
export const refreshSubmissions = async ({ state, effects }: Context, input: { courseID: number, submissionID: number }): Promise<void> => {
    const response = await effects.grpcMan.getSubmissions(input.courseID, state.self.id)
    if (!response.data || !success(response)) {
        return
    }
    const submissions = response.data.getSubmissionsList()
    for (const submission of submissions) {
        const assignment = state.assignments[input.courseID].find(a => a.id === submission.getAssignmentid())
        if (!assignment) {
            continue
        }
        if (state.submissions[input.courseID][assignment.order - 1].commithash !== submission.getCommithash()) {
            state.submissions[input.courseID][assignment.order - 1] = submission.toObject()
        }
    }
}

export const convertCourseSubmission = ({ state }: Context, { courseID, data }: { courseID: number, data: CourseSubmissions }): void => {
    state.review.reviews[courseID] = {}
    state.courseSubmissions[courseID] = []
    for (const link of data.getLinksList()) {
        if (link.hasEnrollment()) {
            const submissionLinks = link.getSubmissionsList()
            submissionLinks.forEach(submissionLink => {
                if (submissionLink.hasSubmission()) {
                    const submission = submissionLink.getSubmission() as Submission
                    state.review.reviews[courseID][submission.getId()] = submission.getReviewsList().map(r => r.toObject())
                }
            })
            state.courseSubmissions[courseID].push({
                enrollment: link.getEnrollment()?.toObject(),
                submissions: link.getSubmissionsList().map((subl) => subl.toObject()),
                user: link.getEnrollment()?.getUser()?.toObject()
            })
        }
    }
    state.isLoading = false
}

/** Fetches and stores all submissions of a given course into state */
export const getAllCourseSubmissions = async ({ state, actions, effects }: Context, courseID: number): Promise<boolean> => {
    state.isLoading = true

    // None of these should fail independently.
    const result = await effects.grpcMan.getSubmissionsByCourse(courseID, SubmissionsForCourseRequest.Type.ALL, true)
    const groups = await effects.grpcMan.getSubmissionsByCourse(courseID, SubmissionsForCourseRequest.Type.GROUP, true)
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
        state.courseGroupSubmissions[courseID] = []
        groups.data.getLinksList().forEach(link => {
            if (!link.getEnrollment()?.hasGroup()) {
                return
            }
            state.courseGroupSubmissions[courseID].push({
                group: link.getEnrollment()?.getGroup()?.toObject(),
                submissions: link.getSubmissionsList().map((subl) => subl.toObject())
            })
        })
    }
    state.isLoading = false
    return true
}

export const getGroupsByCourse = async ({ state, effects }: Context, courseID: number): Promise<void> => {
    state.groups[courseID] = []
    const res = await effects.grpcMan.getGroupsByCourse(courseID)
    if (res.data) {
        state.groups[courseID] = res.data.getGroupsList().map(g => g.toObject())
    }
}

export const getUserSubmissions = async ({ state, effects }: Context, courseID: number): Promise<boolean> => {
    state.submissions[courseID] = []
    const submissions = await effects.grpcMan.getSubmissions(courseID, state.self.id)
    if (submissions.data) {
        // Insert submissions into state.submissions by the assignment order
        state.assignments[courseID]?.forEach(assignment => {
            const submission = submissions.data?.getSubmissionsList().find(s => s.getAssignmentid() === assignment.id)
            state.submissions[courseID][assignment.order - 1] = submission ? submission.toObject() : (new Submission()).toObject()
        })
        return true
    }
    return false
}

export const getGroupSubmissions = async ({ state, effects }: Context, courseID: number): Promise<void> => {
    const enrollment = state.enrollmentsByCourseID[courseID]
    if (enrollment && enrollment.group) {
        const submissions = await effects.grpcMan.getGroupSubmissions(courseID, enrollment.groupid)
        state.assignments[courseID]?.forEach(assignment => {
            const submission = submissions.data?.getSubmissionsList().find(submission => submission.getAssignmentid() === assignment.id)
            if (submission && assignment.isgrouplab) {
                state.submissions[courseID][assignment.order - 1] = submission.toObject()
            }
        })
    }
}

export const setActiveCourse = ({ state }: Context, courseID: number): void => {
    state.activeCourse = courseID
}

export const toggleFavorites = ({ state }: Context): void => {
    state.showFavorites = !state.showFavorites
}

export const setActiveAssignment = ({ state }: Context, assignmentID: number): void => {
    state.activeAssignment = assignmentID
}

/** Rebuilds the currently active submission */
export const rebuildSubmission = async ({ state, actions, effects }: Context): Promise<void> => {
    if (state.currentSubmission && state.selectedAssignment) {
        const response = await effects.grpcMan.rebuildSubmission(state.selectedAssignment.id, state.activeSubmission)
        if (success(response)) {
            // TODO: Alerting is temporary due to the fact that the server no longer returns the updated submission.
            // TODO: gRPC streaming should be implemented to send the updated submission to the client.
            actions.alert({ color: Color.GREEN, text: 'Submission rebuilt successfully' })
        }
    }
}

/* rebuildAllSubmissions rebuilds all submissions for a given assignment */
export const rebuildAllSubmissions = async ({ effects }: Context, { courseID, assignmentID }: { courseID: number, assignmentID: number }): Promise<boolean> => {
    const response = await effects.grpcMan.rebuildSubmissions(assignmentID, courseID)
    return success(response)
}

/** Enrolls a user (self) in a course given by courseID. Refreshes enrollments in state if enroll is successful. */
export const enroll = async ({ state, effects }: Context, courseID: number): Promise<void> => {
    const response = await effects.grpcMan.createEnrollment(courseID, state.self.id)
    if (success(response)) {
        const enrollments = await effects.grpcMan.getEnrollmentsByUser(state.self.id)
        if (enrollments.data) {
            state.enrollments = enrollments.data.getEnrollmentsList().map(e => e.toObject())
        }
    }
}

export const updateGroupStatus = async ({ effects }: Context, { group, status }: { group: Group.AsObject, status: Group.GroupStatus }): Promise<void> => {
    const oldStatus = group.status
    group.status = status
    const response = await effects.grpcMan.updateGroup(Converter.toGroup(group))
    if (!success(response)) {
        group.status = oldStatus
    }
}

export const deleteGroup = async ({ state, effects }: Context, group: Group.AsObject): Promise<void> => {
    if (confirm("Deleting a group is an irreversible action. Are you sure?")) {
        const isRepoEmpty = await effects.grpcMan.isEmptyRepo(group.courseid, 0, group.id)
        if (isRepoEmpty || confirm(`Warning! Group repository is not empty! Do you still want to delete group, github team and group repository?`)) {
            const response = await effects.grpcMan.deleteGroup(group.courseid, group.id)
            if (success(response)) {
                state.groups[group.courseid] = state.groups[group.courseid].filter(g => g.id !== group.id)
            }
        }
    }
}

export const updateGroup = async ({ state, actions, effects }: Context, group: Group.AsObject): Promise<void> => {
    const response = await effects.grpcMan.updateGroup(Converter.toGroup(group))
    if (success(response)) {
        const found = state.groups[group.courseid].find(g => g.id === group.id)
        if (found && response.data) {
            Object.assign(found, response.data.toObject())
            actions.setActiveGroup(null)
        }
    } else {
        actions.alertHandler(response)
    }
}

export const createOrUpdateCriterion = async ({ effects }: Context, { criterion, assignment }: { criterion: GradingCriterion.AsObject, assignment: Assignment.AsObject }): Promise<void> => {
    for (const bm of assignment.gradingbenchmarksList) {
        if (bm.id === criterion.benchmarkid) {
            // Existing criteria have a criteria id > 0, new criteria have a criteria id of 0
            if (criterion.id && success(await effects.grpcMan.updateCriterion(Converter.toGradingCriterion(criterion)))) {
                const index = bm.criteriaList.findIndex(c => c.id === criterion.id)
                if (index > -1) {
                    bm.criteriaList[index] = criterion
                }
            } else {
                const response = await effects.grpcMan.createCriterion(Converter.toGradingCriterion(criterion))
                if (success(response) && response.data) {
                    bm.criteriaList.push(response.data.toObject())
                }
            }
        }
    }
}

export const createOrUpdateBenchmark = async ({ effects }: Context, { benchmark, assignment }: { benchmark: GradingBenchmark.AsObject, assignment: Assignment.AsObject }): Promise<void> => {
    const bm = Converter.toGradingBenchmark(benchmark)
    if (benchmark.id && success(await effects.grpcMan.updateBenchmark(bm))) {
        const index = assignment.gradingbenchmarksList.indexOf(benchmark)
        if (index > -1) {
            assignment.gradingbenchmarksList[index] = benchmark
        }
    } else {
        const response = await effects.grpcMan.createBenchmark(bm)
        if (success(response) && response.data) {
            assignment.gradingbenchmarksList.push(response.data.toObject())
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

export const deleteCriterion = async ({ actions, effects }: Context, { criterion, assignment }: { criterion?: GradingCriterion.AsObject, assignment: Assignment.AsObject }): Promise<void> => {
    for (const benchmark of assignment.gradingbenchmarksList) {
        if (benchmark.id !== criterion?.benchmarkid) {
            continue
        }
        if (!confirm("Do you really want to delete this criterion?")) {
            // Do nothing if user cancels
            return
        }

        // Delete criterion
        const response = await effects.grpcMan.deleteCriterion(Converter.toGradingCriterion(criterion))
        if (success(response)) {
            // Remove criterion from benchmark in state if successful
            const index = assignment.gradingbenchmarksList.indexOf(benchmark)
            if (index > -1) {
                assignment.gradingbenchmarksList.splice(index, 1)
            }
        } else {
            actions.alertHandler(response)
        }
    }
}

export const deleteBenchmark = async ({ actions, effects }: Context, { benchmark, assignment }: { benchmark?: GradingBenchmark.AsObject, assignment: Assignment.AsObject }): Promise<void> => {
    if (benchmark && confirm("Do you really want to delete this benchmark?")) {
        const response = await effects.grpcMan.deleteBenchmark(Converter.toGradingBenchmark(benchmark))
        if (success(response)) {
            const index = assignment.gradingbenchmarksList.indexOf(benchmark)
            if (index > -1) {
                assignment.gradingbenchmarksList.splice(index, 1)
            }
        } else {
            actions.alertHandler(response)
        }
    }
}

export const setActiveSubmissionLink = ({ state }: Context, link: SubmissionLink.AsObject): void => {
    state.activeSubmissionLink = link ? Converter.clone(link) : null
}

export const setActiveEnrollment = ({ state }: Context, enrollment: Enrollment.AsObject): void => {
    state.activeEnrollment = enrollment ? Converter.clone(enrollment) : null
}

/* fetchUserData is called when the user enters the app. It fetches all data that is needed for the user to be able to use the app. */
/* If the user is not logged in, i.e does not have a valid token, the process is aborted. */
export const fetchUserData = async ({ state, actions }: Context): Promise<boolean> => {
    let success = await actions.getSelf()
    // If getSelf returns false, the user is not logged in. Abort.
    if (!success) { state.isLoading = false; return false }

    // Start fetching all data. Loading screen will be shown until all data is fetched, i.e state.isLoading is set to false.
    while (state.isLoading) {
        // Order matters here. Some data is dependent on other data. Ex. fetching submissions depends on enrollments.
        success = await actions.getEnrollmentsByUser()
        success = await actions.getAssignments()
        for (const enrollment of state.enrollments) {
            const courseID = enrollment.courseid
            if (isStudent(enrollment) || isTeacher(enrollment)) {
                success = await actions.getUserSubmissions(courseID)
                await actions.getGroupSubmissions(courseID)
                const statuses = isStudent(enrollment) ? [Enrollment.UserStatus.STUDENT, Enrollment.UserStatus.TEACHER] : []
                success = await actions.getEnrollmentsByCourse({ courseID: courseID, statuses: statuses })
                if (enrollment.groupid > 0) {
                    await actions.getGroupByUserAndCourse(courseID)
                }
            }
            if (isTeacher(enrollment)) {
                actions.getGroupsByCourse(courseID)
            }
        }
        if (state.self.isadmin) {
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
export const changeView = async ({ state, effects }: Context, courseID: number): Promise<void> => {
    const enrollment = state.enrollmentsByCourseID[courseID]
    if (hasStudent(enrollment.status)) {
        const status = await effects.grpcMan.getEnrollmentsByUser(state.self.id, [Enrollment.UserStatus.TEACHER])
        if (status.data?.getEnrollmentsList().find(enrollment => enrollment.getCourseid() == courseID && hasTeacher(enrollment.getStatus()))) {
            enrollment.status = Enrollment.UserStatus.TEACHER
        }
    } else if (hasTeacher(enrollment.status)) {
        enrollment.status = Enrollment.UserStatus.STUDENT
    }
}

export const loading = ({ state }: Context): void => {
    state.isLoading = !state.isLoading
}

/** Sets a query string in state. */
export const setQuery = ({ state }: Context, query: string): void => {
    state.query = query
}

export const setSelectedUser = ({ state }: Context, user: User.AsObject | null): void => {
    state.activeUser = user
}


export const alertHandler = ({ state }: Context, response: IGrpcResponse<unknown>): void => {
    if (response.status.getCode() === StatusCode.UNAUTHENTICATED) {
        // If we end up here, the user session has expired.
        // Store an alert message in localStorage that will be displayed after reloading the page.
        localStorage.setItem("alert", "Your session has expired. Please log in again.")
        window.location.reload()
    } else if (response.status.getCode() >= 0) {
        state.alerts.push({ text: response.status.getError(), color: Color.RED })
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
    state.self = {} as User.AsObject
}

const generateRepositoryList = (enrollment: Enrollment): Repository.Type[] => {
    switch (enrollment.getStatus()) {
        case Enrollment.UserStatus.TEACHER:
            return [Repository.Type.ASSIGNMENTS, Repository.Type.INFO, Repository.Type.GROUP, Repository.Type.TESTS, Repository.Type.USER]
        case Enrollment.UserStatus.STUDENT:
            return [Repository.Type.ASSIGNMENTS, Repository.Type.INFO, Repository.Type.GROUP, Repository.Type.USER]
        default:
            return [Repository.Type.NONE]
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

export const setActiveGroup = ({ state }: Context, group: Group.AsObject | null): void => {
    state.activeGroup = Converter.clone(group)
}

export const updateGroupUsers = ({ state }: Context, user: User.AsObject): void => {
    if (!state.activeGroup) {
        return
    }
    const group = state.activeGroup
    // Remove the user from the group if they are already in it.
    const index = group.usersList.findIndex(u => u.id == user.id)
    if (index >= 0) {
        group.usersList.splice(index, 1)
    } else {
        group.usersList.push(user)
    }
}

export const updateGroupName = ({ state }: Context, name: string): void => {
    if (!state.activeGroup) {
        return
    }
    state.activeGroup.name = name
}

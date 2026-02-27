import { create } from "@bufbuild/protobuf"
import { derived } from "overmind"
import { Context } from "."
import { Assignment, Course, Enrollment, Enrollment_UserStatus, Group, Submission, User, UserSchema } from "../../proto/qf/types_pb"
import { Color, ConnStatus, filterByApproval, filterByReleased, getApprovalSortValue, getScoreSortValue, getSubmissionData, isManuallyGraded, isPending, isPendingGroup, isTeacher, SubmissionsForCourse, SubmissionsForUser, SubmissionSort } from "../Helpers"

export interface CourseGroup {
    courseID: bigint
    // User IDs of all members of the group
    users: bigint[]
    name: string
}

export interface Alert {
    id: number
    text: string
    color: Color
    // The delay in milliseconds before the alert is removed
    delay?: number
}

type EnrollmentsByCourse = { [CourseID: string]: Enrollment }
export type SubmissionOwner = { type: "ENROLLMENT" | "GROUP", id: bigint }
export type AssignmentsMap = { [key: string]: boolean }
export type State = {

    /***************************************************************************
    *                   Data relating to the current user
    ***************************************************************************/

    /* This is the user that is currently logged in */
    self: User,

    /* Indicates if the user has valid data */
    // derived from self
    isValid: boolean,

    /* Indicates if the user is logged in */
    // derived from self
    isLoggedIn: boolean,

    /* Contains all the courses the user is enrolled in */
    enrollments: Enrollment[],

    /* Contains all the courses the user is enrolled in, indexed by course ID */
    // derived from enrollments
    enrollmentsByCourseID: { [courseID: string]: Enrollment },

    /* Contains all the groups the user is a member of, indexed by course ID */
    userGroup: { [courseID: string]: Group },

    /* Contains all submissions for the user, indexed by course ID */
    // The individual submissions for a given course are indexed by assignment order - 1
    submissions: SubmissionsForUser,

    /* Current enrollment status of the user for a given course */
    status: { [courseID: string]: Enrollment_UserStatus }

    /* Indicates if the user is a teacher of the current course */
    // derived from enrollmentsByCourseID
    isTeacher: boolean

    /* Indicates if the user is the course creator of the current course */
    // derived from courses
    isCourseCreator: boolean

    /* Contains links to all repositories for a given course */
    // Individual repository links are accessed by Repository.Type
    repositories: { [courseid: string]: { [repo: string]: string } },

    theme: "light" | "dark" | "cupcake" | "synthwave" | "cyberpunk" | "forest" | "aqua"

    /***************************************************************************
    *                              Public data
    ***************************************************************************/

    /* Contains all users of a given course */
    // Requires the user to be admin to get from backend
    users: { [userID: string]: User },

    /* Contains all courses */
    courses: Course[],

    /* Contains all assignments for a given course */
    assignments: { [courseID: string]: Assignment[] },


    /***************************************************************************
    *                         Course Specific Data
     ***************************************************************************/

    /* Contains all submissions for a given course */


    /** Filtered and sorted enrollments for the current course (individual view).
     *  The enrollments are filtered and sorted based on the current
     *  values of sortSubmissionsBy, sortAscending, and submissionFilters */
    filteredEnrollments: Enrollment[],

    /** Filtered and sorted groups for the current course (group view).
     *  The groups are filtered and sorted based on the current
     *  values of sortSubmissionsBy, sortAscending, and submissionFilters */
    filteredGroups: Group[],

    /** Contains all members of the current course.
     *  Derived from either filteredEnrollments or filteredGroups based on groupView. */
    courseMembers: Enrollment[] | Group[],

    /* Course teachers, indexed by user ID */
    /* Derived from enrollments for selected course */
    courseTeachers: { [userID: string]: User }

    /* Contains all enrollments for a given course */
    courseEnrollments: { [courseID: string]: Enrollment[] },

    /* Contains all groups for a given course */
    groups: { [courseID: string]: Group[] },

    /* Number of groups in the course */
    // derived from groups
    numGroups: number,

    /* Number of enrolled users */
    // derived from courseEnrollments
    numEnrolled: number,

    /* Contains all enrollments where the enrollment status is pending */
    // derived from courseEnrollments
    pendingEnrollments: Enrollment[],

    /* Contains all groups where the group status is pending */
    // derived from groups
    pendingGroups: Group[],

    /* Contains all users with admins sorted first */
    allUsers: User[],

    /* Indicates if the course has any assignment that is manually graded */
    isCourseManuallyGraded: boolean


    /***************************************************************************
     *                             Frontend Activity State
     ***************************************************************************/

    /* Indicates if the state is loading */
    isLoading: boolean,

    /* The current course ID */
    activeCourse: bigint,

    /* The currently selected assignment ID */
    selectedAssignmentID: number,

    /* The current assignment */
    selectedAssignment: Assignment | null,

    /* Contains a group in creation */
    courseGroup: CourseGroup,

    /* Contains alerts to be displayed to the user */
    alerts: Alert[],

    /* Current search query */
    query: string,

    /* Currently selected enrollment */
    selectedEnrollment: Enrollment | null,

    /* Currently selected submission */
    selectedSubmission: Submission | null,

    /* The value to sort submissions by */
    sortSubmissionsBy: SubmissionSort,

    /* Whether to sort by ascending or descending */
    sortAscending: boolean,

    /* Submission filters */
    submissionFilters: string[],

    /* Determine if all submissions should be displayed, or only group submissions */
    groupView: boolean,

    /* Can be used to determine whether or not to show only individual submissions */
    individualSubmissionView: boolean,

    showFavorites: boolean,


    /* Currently selected group */
    /* Contains either an existing group to edit, or a new group to create */
    activeGroup: Group | null,

    hasGroup: (courseID: string) => boolean,

    connectionStatus: ConnStatus,

    // ID of owner of the current submission
    // Must be either an enrollment ID or a group ID
    submissionOwner: SubmissionOwner,

    submissionsForCourse: SubmissionsForCourse,
    isManuallyGraded: (submission: Submission) => boolean,
    loadedCourse: { [courseID: string]: boolean },
    getAssignmentsMap: (courseID: bigint) => AssignmentsMap,
}


/* Initial State */
/* To add to state, extend the State type and initialize the variable below */
export const state: State = {
    self: create(UserSchema),
    isLoggedIn: derived(({ self }: State) => {
        return Number(self.ID) !== 0
    }),

    isValid: derived(({ self }: State) => {
        return self.Name.length > 0 && self.StudentID.length > 0 && self.Email.length > 0
    }),

    enrollments: [],
    enrollmentsByCourseID: derived(({ enrollments }: State) => {
        const enrollmentsByCourseID: EnrollmentsByCourse = {}
        for (const enrollment of enrollments) {
            enrollmentsByCourseID[enrollment.courseID.toString()] = enrollment
        }
        return enrollmentsByCourseID
    }),
    submissions: new SubmissionsForUser(),
    userGroup: derived(({ enrollments }: State) => {
        const userGroup: { [courseID: string]: Group } = {}
        for (const enrollment of enrollments) {
            if (enrollment.group) {
                userGroup[enrollment.courseID.toString()] = enrollment.group
            }
        }
        return userGroup
    }),
    theme: "light",
    isTeacher: derived(({ enrollmentsByCourseID, activeCourse }: State) => {
        if (activeCourse > 0 && enrollmentsByCourseID[activeCourse.toString()]) {
            return isTeacher(enrollmentsByCourseID[activeCourse.toString()])
        }
        return false
    }),
    isCourseCreator: derived(({ courses, activeCourse, self }: State) => {
        const course = courses.find(c => c.ID === activeCourse)
        return course !== undefined && course.courseCreatorID === self.ID
    }),
    status: {},

    users: {},
    allUsers: [],
    courses: [],
    courseTeachers: derived(({ courseEnrollments, activeCourse }: State) => {
        if (!activeCourse || !courseEnrollments[activeCourse.toString()]) {
            return {}
        }
        const teachersMap: { [userID: string]: User } = {}
        courseEnrollments[activeCourse.toString()].forEach(enrollment => {
            if (isTeacher(enrollment) && enrollment.user) {
                teachersMap[enrollment.userID.toString()] = enrollment.user
            }
        })
        return teachersMap
    }),
    filteredEnrollments: derived(({
        activeCourse, submissionsForCourse, assignments,
        courseEnrollments, submissionFilters, sortAscending, sortSubmissionsBy
    }: State, {
        review: { assignmentID }
    }: Context["state"]) => {
        if (!activeCourse) {
            return []
        }
        const submissions = submissionsForCourse.userSubmissions
        if (submissions.size === 0) {
            return []
        }

        const numAssignments = assignmentID > 0
            ? 1
            : assignments[activeCourse.toString()]?.length ?? 0

        let filtered = courseEnrollments[activeCourse.toString()] ?? []

        for (const filter of submissionFilters) {
            switch (filter) {
                case "teachers":
                    filtered = filtered.filter(e => e.status !== Enrollment_UserStatus.TEACHER)
                    break
                case "approved":
                    filtered = filtered.filter(e => {
                        const data = getSubmissionData(submissions, e.ID, assignmentID)
                        return filterByApproval(e, data, numAssignments)
                    })
                    break
                case "released":
                    filtered = filtered.filter(e => {
                        const data = getSubmissionData(submissions, e.ID, assignmentID)
                        return filterByReleased(e, data, numAssignments)
                    })
                    break
            }
        }

        const sortOrder = sortAscending ? -1 : 1
        return [...filtered].sort((a, b) => {
            const dataA = getSubmissionData(submissions, a.ID, assignmentID)
            const dataB = getSubmissionData(submissions, b.ID, assignmentID)

            switch (sortSubmissionsBy) {
                case SubmissionSort.ID:
                    return sortOrder * (Number(a.userID) - Number(b.userID))
                case SubmissionSort.Score:
                    return sortOrder * (getScoreSortValue(dataB) - getScoreSortValue(dataA))
                case SubmissionSort.Approved:
                    return sortOrder * (getApprovalSortValue(dataA) - getApprovalSortValue(dataB))
                case SubmissionSort.Name:
                    return sortOrder * ((a.user?.Name ?? "").localeCompare(b.user?.Name ?? ""))
                default:
                    return 0
            }
        })
    }),

    filteredGroups: derived(({
        activeCourse, submissionsForCourse, assignments, groups,
        submissionFilters, sortAscending, sortSubmissionsBy
    }: State, {
        review: { assignmentID }
    }: Context["state"]) => {
        if (!activeCourse) {
            return []
        }
        const submissions = submissionsForCourse.groupSubmissions
        if (submissions.size === 0) {
            return []
        }

        const numAssignments = assignmentID > 0
            ? 1
            : assignments[activeCourse.toString()]?.filter(a => a.isGroupLab).length ?? 0

        let filtered = groups[activeCourse.toString()] ?? []

        for (const filter of submissionFilters) {
            switch (filter) {
                case "teachers":
                    break
                case "approved":
                    filtered = filtered.filter(g => {
                        const data = getSubmissionData(submissions, g.ID, assignmentID)
                        return filterByApproval(g, data, numAssignments)
                    })
                    break
                case "released":
                    filtered = filtered.filter(g => {
                        const data = getSubmissionData(submissions, g.ID, assignmentID)
                        return filterByReleased(g, data, numAssignments)
                    })
                    break
            }
        }

        const sortOrder = sortAscending ? -1 : 1
        return [...filtered].sort((a, b) => {
            const dataA = getSubmissionData(submissions, a.ID, assignmentID)
            const dataB = getSubmissionData(submissions, b.ID, assignmentID)

            switch (sortSubmissionsBy) {
                case SubmissionSort.ID:
                    return sortOrder * (Number(a.ID) - Number(b.ID))
                case SubmissionSort.Score:
                    return sortOrder * (getScoreSortValue(dataB) - getScoreSortValue(dataA))
                case SubmissionSort.Approved:
                    return sortOrder * (getApprovalSortValue(dataA) - getApprovalSortValue(dataB))
                case SubmissionSort.Name:
                    return sortOrder * ((a.name ?? "").localeCompare(b.name ?? ""))
                default:
                    return 0
            }
        })
    }),

    courseMembers: derived(({ groupView, filteredEnrollments, filteredGroups }: State) => {
        return groupView ? filteredGroups : filteredEnrollments
    }),
    selectedEnrollment: null,
    selectedSubmission: null,
    selectedAssignment: derived(({ activeCourse, selectedSubmission, assignments }: State) => {
        return assignments[activeCourse.toString()]?.find(a => a.ID === selectedSubmission?.AssignmentID) ?? null
    }),
    assignments: {},
    repositories: {},

    courseGroup: { courseID: 0n, users: [], name: "" },
    alerts: [],
    isLoading: true,
    activeCourse: BigInt(-1),
    selectedAssignmentID: -1,
    courseEnrollments: {},
    groups: {},
    pendingGroups: derived(({ activeCourse, groups }: State) => {
        if (activeCourse > 0 && groups[activeCourse.toString()]) {
            return groups[activeCourse.toString()]?.filter(group => isPendingGroup(group))
        }
        return []
    }),
    pendingEnrollments: derived(({ activeCourse, courseEnrollments }: State) => {
        if (activeCourse > 0 && courseEnrollments[activeCourse.toString()]) {
            return courseEnrollments[activeCourse.toString()].filter(enrollment => isPending(enrollment))
        }
        return []
    }),
    numGroups: derived(({ groups, activeCourse }: State) => {
        if (activeCourse > 0 && groups[activeCourse.toString()]) {
            return groups[activeCourse.toString()]?.filter(group => !isPendingGroup(group)).length
        }
        return 0
    }),
    numEnrolled: derived(({ activeCourse, courseEnrollments }: State) => {
        if (activeCourse > 0 && courseEnrollments[activeCourse.toString()]) {
            return courseEnrollments[activeCourse.toString()]?.filter(enrollment => !isPending(enrollment)).length
        }
        return 0
    }),
    isCourseManuallyGraded: derived(({ activeCourse, assignments }: State) => {
        if (activeCourse > 0 && assignments[activeCourse.toString()]) {
            return assignments[activeCourse.toString()].some(a => isManuallyGraded(a.reviewers))
        }
        return false
    }),
    query: "",
    sortSubmissionsBy: SubmissionSort.Approved,
    sortAscending: true,
    submissionFilters: [],
    individualSubmissionView: false,
    groupView: false,
    activeGroup: null,
    hasGroup: derived(({ userGroup }: State) => courseID => {
        return userGroup[courseID] !== undefined
    }),
    showFavorites: false,

    connectionStatus: ConnStatus.DISCONNECTED,
    isManuallyGraded: derived(({ activeCourse, assignments }: State) => submission => {
        const assignment = assignments[activeCourse.toString()]?.find(a => a.ID === submission.AssignmentID)
        return assignment ? assignment.reviewers > 0 : false
    }),

    getAssignmentsMap: derived(({ assignments }: State, { review: { assignmentID } }: Context["state"]) => courseID => {
        const asgmts = assignments[courseID.toString()]?.filter(assignment => (assignmentID < 0) || assignment.ID === assignmentID) ?? []
        const assignmentsMap: AssignmentsMap = {}
        asgmts.forEach(assignment => {
            assignmentsMap[assignment.ID.toString()] = assignment.isGroupLab
        })
        return assignmentsMap
    }),

    submissionOwner: { type: "ENROLLMENT", id: 0n },
    loadedCourse: {},
    submissionsForCourse: new SubmissionsForCourse()
}

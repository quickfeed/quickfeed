import { derived } from "overmind"
import { Context } from "."
import { Assignment, Course, Enrollment, Enrollment_UserStatus, Group, Group_GroupStatus, Submission, User } from "../../proto/qf/types_pb"
import { Color, ConnStatus, getNumApproved, getSubmissionsScore, isApproved, isPending, isPendingGroup, isTeacher, SubmissionsForCourse, SubmissionSort } from "../Helpers"

export interface CourseGroup {
    courseID: number
    enrollments: number[]
    users: User[]
    groupName: string
}

export interface Alert {
    text: string
    color: Color
}

interface GroupOrEnrollment {
    ID: bigint,
    name?: string,
    user?: User,
    status?: Enrollment_UserStatus | Group_GroupStatus
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
    submissions: { [courseID: string]: Submission[] },

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


    /** Contains all members of the current course.
     *  Derived from either enrollments or groups based on groupView.
     *  The members are filtered and sorted based on the current
     *  values of sortSubmissionsBy, sortAscending, and submissionFilters */
    courseMembers: Enrollment[] | Group[],

    /* Contains all enrollments for a given course */
    courseEnrollments: { [courseID: string]: Enrollment[] },

    /* Contains all groups for a given course */
    groups: { [courseID: string]: Group[] },

    /* Currently selected submission ID */
    activeSubmission: bigint,

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


    /***************************************************************************
     *                             Frontend Activity State
     ***************************************************************************/

    /* Indicates if the state is loading */
    isLoading: boolean,

    /* The current course ID */
    activeCourse: bigint,

    /* The current assignment ID */
    activeAssignment: number,

    /* The current assignment */
    selectedAssignment: Assignment | null,

    /* Contains a group in creation */
    courseGroup: CourseGroup,

    /* Contains alerts to be displayed to the user */
    alerts: Alert[],

    /* Current search query */
    query: string,

    /* Current enrollment */
    activeEnrollment: Enrollment | null,

    /* Current submission */
    currentSubmission: Submission | null,

    /* The value to sort submissions by */
    sortSubmissionsBy: SubmissionSort,

    /* Whether to sort by ascending or descending */
    sortAscending: boolean,

    /* Submission filters */
    submissionFilters: string[],

    /* Determine if all submissions should be displayed, or only group submissions */
    groupView: boolean,
    showFavorites: boolean,


    /* Currently selected group */
    /* Contains either an existing group to edit, or a new group to create */
    activeGroup: Group | null,

    hasGroup: (courseID: number) => boolean,

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
    self: new User(),
    isLoggedIn: derived(({ self }: State) => {
        return Number(self.ID) !== 0
    }),

    isValid: derived(({ self }: State) => {
        return self.Name.length > 0 && self.StudentID.length > 0 && self.Email.length > 0
    }),

    enrollments: [],
    enrollmentsByCourseID: derived((state: State) => {
        const enrollmentsByCourseID: EnrollmentsByCourse = {}
        for (const enrollment of state.enrollments) {
            enrollmentsByCourseID[enrollment.courseID.toString()] = enrollment
        }
        return enrollmentsByCourseID
    }),
    submissions: {},
    userGroup: {},

    isTeacher: derived((state: State) => {
        if (state.activeCourse > 0 && state.enrollmentsByCourseID[state.activeCourse.toString()]) {
            return isTeacher(state.enrollmentsByCourseID[state.activeCourse.toString()])
        }
        return false
    }),
    isCourseCreator: derived((state: State) => {
        const course = state.courses.find(course => course.ID === state.activeCourse)
        if (course && course.courseCreatorID === state.self.ID) {
            return true
        }
        return false
    }),
    status: {},

    users: {},
    allUsers: [],
    courses: [],
    courseMembers: derived((state: State, rootState: Context["state"]) => {
        // Filter and sort course members based on the current state
        if (!state.activeCourse) {
            return []
        }
        const submissions = state.groupView
            ? state.submissionsForCourse.groupSubmissions
            : state.submissionsForCourse.userSubmissions

        if (Object.keys(submissions).length === 0) {
            return []
        }

        const assignmentsMap: AssignmentsMap = {}
        state.assignments[state.activeCourse.toString()].forEach(assignment => assignmentsMap[assignment.ID.toString()] = assignment.isGroupLab)
        // If a specific assignment is selected, filter by that assignment
        let numAssignments = 0
        if (rootState.review.assignmentID > 0) {
            numAssignments = 1
        } else if (state.groupView) {
            numAssignments = state.assignments[state.activeCourse.toString()].filter(a => a.isGroupLab).length || 0
        } else {
            numAssignments = state.assignments[state.activeCourse.toString()].length || 0
        }

        let filtered: GroupOrEnrollment[] = state.groupView ? state.groups[state.activeCourse.toString()] : state.courseEnrollments[state.activeCourse.toString()] ?? []
        for (const filter of state.submissionFilters) {
            switch (filter) {
                case "teachers":
                    filtered = filtered.filter(el => {
                        return el.status !== Enrollment_UserStatus.TEACHER
                    })
                    break
                case "approved":
                    // approved filters all entries where all assignments have been approved
                    filtered = filtered.filter(el => {
                        if (rootState.review.assignmentID > 0) {
                            // If a specific assignment is selected, filter by that assignment
                            const sub = submissions[el.ID.toString()].submissions?.find(sub => sub.AssignmentID === rootState.review.assignmentID)
                            return sub !== undefined && !isApproved(sub)
                        }
                        const numApproved = submissions[el.ID.toString()].submissions?.reduce((acc, cur) => {
                            return acc + ((cur &&
                                isApproved(cur)) ? 1 : 0)
                        }, 0) ?? 0
                        return numApproved < numAssignments
                    })
                    break
                default:
                    break
            }
        }

        const sortOrder = state.sortAscending ? -1 : 1
        const sortedSubmissions = Object.values(filtered).sort((a, b) => {
            let subA: Submission | undefined
            let subB: Submission | undefined
            if (rootState.review.assignmentID > 0) {
                // If a specific assignment is selected, sort by that assignment
                subA = submissions[a.ID.toString()]?.submissions.find(sub => sub.AssignmentID === rootState.review.assignmentID)
                subB = submissions[b.ID.toString()]?.submissions.find(sub => sub.AssignmentID === rootState.review.assignmentID)
            }

            const subsA = submissions[a.ID.toString()]?.submissions
            const subsB = submissions[b.ID.toString()]?.submissions

            switch (state.sortSubmissionsBy) {
                case SubmissionSort.Score: {
                    if (rootState.review.assignmentID > 0) {
                        const sA = subA?.score
                        const sB = subB?.score
                        if (sA !== undefined && sB !== undefined) {
                            return sortOrder * (sB - sA)
                        } else if (sA !== undefined) {
                            return -sortOrder
                        }
                        return sortOrder
                    }
                    const aSubs = subsA ? getSubmissionsScore(subsA) : 0
                    const bSubs = subsB ? getSubmissionsScore(subsB) : 0
                    return sortOrder * (aSubs - bSubs)
                }
                case SubmissionSort.Approved: {
                    if (rootState.review.assignmentID > 0) {
                        const sA = subA && isApproved(subA) ? 1 : 0
                        const sB = subB && isApproved(subB) ? 1 : 0
                        return sortOrder * (sA - sB)
                    }
                    const aApproved = subsA ? getNumApproved(subsA) : 0
                    const bApproved = subsB ? getNumApproved(subsB) : 0
                    return sortOrder * (aApproved - bApproved)
                }
                case SubmissionSort.Name: {
                    const nameA = state.groupView ? a.name ?? "" : a.user?.Name ?? ""
                    const nameB = state.groupView ? b.name ?? "" : b.user?.Name ?? ""
                    return sortOrder * (nameA.localeCompare(nameB))
                }
                default:
                    return 0
            }
        })
        return sortedSubmissions as Group[] | Enrollment[]
    }),
    activeSubmission: 0n,
    activeEnrollment: null,
    currentSubmission: derived((state: State) => {
        if (state.activeSubmission === 0n) {
            return null
        }
        const submissions = state.submissionsForCourse.getSubmissionsForOwner(state.submissionOwner)
        if (!submissions || submissions.length === 0) {
            return null
        }
        return submissions.find(submission => submission.ID === state.activeSubmission) ?? null
    }),
    selectedAssignment: derived(({ activeCourse, currentSubmission, assignments }: State) => {
        return assignments[activeCourse.toString()]?.find(a => a.ID === currentSubmission?.AssignmentID) ?? null
    }),
    assignments: {},
    repositories: {},

    courseGroup: { courseID: 0, enrollments: [], users: [], groupName: "" },
    alerts: [],
    isLoading: true,
    activeCourse: BigInt(-1),
    activeAssignment: -1,
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
    numEnrolled: derived(({ activeCourse, courseEnrollments }: State) => {
        if (activeCourse > 0 && courseEnrollments[activeCourse.toString()]) {
            return courseEnrollments[activeCourse.toString()]?.filter(enrollment => !isPending(enrollment)).length
        }
        return 0
    }),
    query: "",
    sortSubmissionsBy: SubmissionSort.Approved,
    sortAscending: true,
    submissionFilters: [],
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

    getAssignmentsMap: derived(({ assignments }: State, rootState: Context["state"]) => courseID => {
        const asgmts = assignments[courseID.toString()].filter(assignment => (rootState.review.assignmentID < 0) || assignment.ID === rootState.review.assignmentID)
        const assignmentsMap: AssignmentsMap = {}
        asgmts.forEach(assignment => assignmentsMap[assignment.ID.toString()] = assignment.isGroupLab)
        return assignmentsMap
    }),

    submissionOwner: { type: "ENROLLMENT", id: 0n },
    loadedCourse: {},
    submissionsForCourse: new SubmissionsForCourse()
}

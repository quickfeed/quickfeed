import { derived } from "overmind"
import { Context } from "."
import { Assignment, Course, Enrollment, Group, Submission, SubmissionLink, User } from "../../proto/qf/types_pb"
import { Color, getNumApproved, getSubmissionByAssignmentID, getSubmissionsScore, isApproved, isPending, isPendingGroup, isTeacher, SubmissionSort } from "../Helpers"

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

export interface UserCourseSubmissions {
    group?: Group.AsObject
    enrollment?: Enrollment.AsObject
    user?: User.AsObject
    submissions?: SubmissionLink.AsObject[]
}

type EnrollmentsByCourse = { [courseid: number]: Enrollment.AsObject }

export type State = {

    /***************************************************************************
    *                   Data relating to the current user
    ***************************************************************************/

    /* This is the user that is currently logged in */
    self: User.AsObject,

    /* Indicates if the user has valid data */
    // derived from self
    isValid: boolean,

    /* Indicates if the user is logged in */
    // derived from self
    isLoggedIn: boolean,

    /* Contains all the courses the user is enrolled in */
    enrollments: Enrollment.AsObject[],

    /* Contains all the courses the user is enrolled in, indexed by course ID */
    // derived from enrollments
    enrollmentsByCourseID: { [courseID: number]: Enrollment.AsObject },

    /* Contains all the groups the user is a member of, indexed by course ID */
    userGroup: { [courseID: number]: Group.AsObject },

    /* Contains all submissions for the user, indexed by course ID */
    // The individual submissions for a given course are indexed by assignment order - 1
    submissions: { [courseID: number]: Submission.AsObject[] },

    /* Current enrollment status of the user for a given course */
    status: { [courseID: number]: Enrollment.UserStatus }

    /* Indicates if the user is a teacher of the current course */
    // derived from enrollmentsByCourseID
    isTeacher: boolean

    /* Indicates if the user is the course creator of the current course */
    // derived from courses
    isCourseCreator: boolean

    /* Contains links to all repositories for a given course */
    // Individual repository links are accessed by Repository.Type
    repositories: { [courseid: number]: { [repo: string]: string } },

    /***************************************************************************
    *                              Public data
    ***************************************************************************/

    /* Contains all users of a given course */
    // Requires the user to be admin to get from backend
    users: { [userID: number]: User.AsObject },

    /* Contains all courses */
    courses: Course.AsObject[],

    /* Contains all assignments for a given course */
    assignments: { [courseID: number]: Assignment.AsObject[] },


    /***************************************************************************
    *                         Course Specific Data
     ***************************************************************************/

    /* Contains all submissions for a given course */
    courseSubmissions: { [courseID: number]: UserCourseSubmissions[] },

    courseGroupSubmissions: { [courseID: number]: UserCourseSubmissions[] },

    /** Contains all submissions for the current course.
     *  Derived from either courseSubmissions or courseGroupSubmissions based on groupView.
     *  The submissions are filtered and sorted based on the current
     *  values of sortSubmissionsBy, sortAscending, and submissionFilters */
    sortedAndFilteredSubmissions: UserCourseSubmissions[]

    /* Contains all enrollments for a given course */
    courseEnrollments: { [courseID: number]: Enrollment.AsObject[] },

    /* Contains all groups for a given course */
    groups: { [courseID: number]: Group.AsObject[] },

    /* Currently selected submission ID */
    activeSubmission: number,

    /* Currently selected user */
    activeUser: User.AsObject | null,

    /* Number of enrolled users */
    // derived from courseEnrollments
    numEnrolled: number,

    /* Contains all enrollments where the enrollment status is pending */
    // derived from courseEnrollments
    pendingEnrollments: Enrollment.AsObject[],

    /* Contains all groups where the group status is pending */
    // derived from groups
    pendingGroups: Group.AsObject[],

    /* Contains all users with admins sorted first */
    allUsers: User.AsObject[],


    /***************************************************************************
     *                             Frontend Activity State
     ***************************************************************************/

    /* Indicates if the state is loading */
    isLoading: boolean,

    /* The current course ID */
    activeCourse: number,

    /* The current assignment ID */
    activeAssignment: number,

    /* The current assignment */
    selectedAssignment: Assignment.AsObject | null,

    /* Contains a group in creation */
    courseGroup: CourseGroup,

    /* Contains alerts to be displayed to the user */
    alerts: Alert[],

    /* Current search query */
    query: string,

    /* Current submission link */
    activeSubmissionLink: SubmissionLink.AsObject | null,

    /* Current enrollment */
    activeEnrollment: Enrollment.AsObject | null,

    /* Current submission */
    currentSubmission: Submission.AsObject | null,

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
    activeGroup: Group.AsObject | null,

    hasGroup: (courseID: number) => boolean,
}


/* Initial State */
/* To add to state, extend the State type and initialize the variable below */
export const state: State = {
    self: User.toObject(false, new User()),
    isLoggedIn: derived(({ self }: State) => {
        return self.id !== 0
    }),

    isValid: derived(({ self }: State) => {
        return self.name.length > 0 && self.studentid.length > 0 && self.email.length > 0
    }),

    enrollments: [],
    enrollmentsByCourseID: derived((state: State) => {
        const enrollmentsByCourseID: EnrollmentsByCourse = {}
        for (const enrollment of state.enrollments) {
            enrollmentsByCourseID[enrollment.courseid] = enrollment
        }
        return enrollmentsByCourseID
    }),
    submissions: {},
    userGroup: {},

    isTeacher: derived((state: State) => {
        if (state.activeCourse > 0 && state.enrollmentsByCourseID[state.activeCourse]) {
            return isTeacher(state.enrollmentsByCourseID[state.activeCourse])
        }
        return false
    }),
    isCourseCreator: derived((state: State) => {
        const course = state.courses.find(course => course.id === state.activeCourse)
        if (course && course.coursecreatorid === state.self.id) {
            return true
        }
        return false
    }),
    status: [],

    users: {},
    allUsers: [],
    courses: [],
    courseSubmissions: [],
    courseGroupSubmissions: {},
    sortedAndFilteredSubmissions: derived((state: State, rootState: Context["state"]) => {
        // Filter and sort submissions based on the current state
        if (!state.activeCourse || !state.courseSubmissions[state.activeCourse]) {
            return []
        }
        let submissions = state.groupView
            ? state.courseGroupSubmissions[state.activeCourse]
            : state.courseSubmissions[state.activeCourse]

        if (!submissions) {
            return []
        }

        // If a specific assignment is selected, filter by that assignment
        const numAssignments = rootState.review.assignmentID > 0
            ? 1
            : state.assignments[state.activeCourse].length ?? 0

        let filteredSubmissions = submissions
        for (const filter of state.submissionFilters) {
            switch (filter) {
                case "teachers":
                    filteredSubmissions = filteredSubmissions.filter(submission => {
                        return submission.enrollment ? !isTeacher(submission.enrollment) : false
                    })
                    break
                case "approved":
                    // approved filters all entries where all assignments have been approved
                    filteredSubmissions = filteredSubmissions.filter(link => {
                        if (rootState.review.assignmentID > 0) {
                            // If a specific assignment is selected, filter by that assignment
                            const sub = getSubmissionByAssignmentID(link.submissions, rootState.review.assignmentID)
                            return sub !== undefined && !isApproved(sub)
                        }
                        const numApproved = link.submissions?.reduce((acc, cur) => {
                            return acc + ((cur.submission &&
                                isApproved(cur.submission)) ? 1 : 0)
                        }, 0) ?? 0
                        return numApproved < numAssignments
                    })
                    break
                default:
                    break
            }
        }

        const sortOrder = state.sortAscending ? -1 : 1
        const sortedSubmissions = Object.values(filteredSubmissions).sort((a, b) => {
            let subA: Submission.AsObject | undefined
            let subB: Submission.AsObject | undefined
            if (rootState.review.assignmentID > 0) {
                // If a specific assignment is selected, sort by that assignment
                subA = getSubmissionByAssignmentID(a.submissions, rootState.review.assignmentID)
                subB = getSubmissionByAssignmentID(b.submissions, rootState.review.assignmentID)
            }

            switch (state.sortSubmissionsBy) {
                case SubmissionSort.Score:
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
                    const aSubs = a.submissions ? getSubmissionsScore(a.submissions) : 0
                    const bSubs = b.submissions ? getSubmissionsScore(b.submissions) : 0
                    return sortOrder * (aSubs - bSubs)
                case SubmissionSort.Approved:
                    if (rootState.review.assignmentID > 0) {
                        const sA = subA && isApproved(subA) ? 1 : 0
                        const sB = subB && isApproved(subB) ? 1 : 0
                        return sortOrder * (sA - sB)
                    }
                    const aApproved = a.submissions ? getNumApproved(a.submissions) : 0
                    const bApproved = b.submissions ? getNumApproved(b.submissions) : 0
                    return sortOrder * (aApproved - bApproved)
                case SubmissionSort.Name:
                    const nameA = a.user?.name ?? ""
                    const nameB = b.user?.name ?? ""
                    return sortOrder * (nameA.localeCompare(nameB))
                default:
                    return 0
            }
        })
        return sortedSubmissions
    }),
    activeSubmission: derived((state: State) => {
        if (state.activeSubmissionLink) {
            return state.activeSubmissionLink.submission ? (state.activeSubmissionLink.submission).id : -1
        }
        return -1
    }),
    activeEnrollment: null,
    activeSubmissionLink: null,
    currentSubmission: derived(({ activeSubmissionLink }: State) => {
        return activeSubmissionLink?.submission ?? null
    }),
    selectedAssignment: derived(({ activeCourse, currentSubmission, assignments }: State) => {
        return assignments[activeCourse]?.find(a => a.id === currentSubmission?.assignmentid) ?? null
    }),
    activeUser: null,
    assignments: {},
    repositories: {},

    courseGroup: { courseID: 0, enrollments: [], users: [], groupName: "" },
    alerts: [],
    isLoading: true,
    activeCourse: -1,
    activeAssignment: -1,
    courseEnrollments: {},
    groups: {},
    pendingGroups: derived(({ activeCourse, groups }: State) => {
        if (activeCourse > 0 && groups[activeCourse]) {
            return groups[activeCourse]?.filter(group => isPendingGroup(group))
        }
        return []
    }),
    pendingEnrollments: derived(({ activeCourse, courseEnrollments }: State) => {
        if (activeCourse > 0 && courseEnrollments[activeCourse]) {
            return courseEnrollments[activeCourse].filter(enrollment => isPending(enrollment))
        }
        return []
    }),
    numEnrolled: derived(({ activeCourse, courseEnrollments }: State) => {
        if (activeCourse > 0 && courseEnrollments[activeCourse]) {
            return courseEnrollments[activeCourse]?.filter(enrollment => !isPending(enrollment)).length
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
}

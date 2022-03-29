import { derived } from "overmind"
import { Context } from "."
import { Assignment, Course, Enrollment, Group, Submission, SubmissionLink, User } from "../../proto/ag/ag_pb"
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
    group?: Group
    enrollment?: Enrollment
    user?: User
    submissions?: SubmissionLink[]
}

type EnrollmentsByCourse = { [courseid: number]: Enrollment }

type State = {

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
    enrollmentsByCourseID: { [courseID: number]: Enrollment },

    /* Contains all the groups the user is a member of, indexed by course ID */
    userGroup: { [courseID: number]: Group },

    /* Contains all submissions for the user, indexed by course ID */
    // The individual submissions for a given course are indexed by assignment order - 1
    submissions: { [courseID: number]: Submission[] },

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
    users: { [userID: number]: User },

    /* Contains all courses */
    courses: Course[],

    /* Contains all assignments for a given course */
    assignments: { [courseID: number]: Assignment[] },


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
    courseEnrollments: { [courseID: number]: Enrollment[] },

    /* Contains all groups for a given course */
    groups: { [courseID: number]: Group[] },

    /* Currently selected submission ID */
    activeSubmission: number,

    /* Currently selected user */
    activeUser: User | undefined,

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
    activeCourse: number,

    /* The current assignment ID */
    activeAssignment: number,

    /* The current assignment */
    selectedAssignment: Assignment | undefined,

    /* Contains a group in creation */
    courseGroup: CourseGroup,

    /* Contains alerts to be displayed to the user */
    alerts: Alert[],

    /* Current search query */
    query: string,

    /* Current submission link */
    activeSubmissionLink: SubmissionLink | undefined,

    /* Current enrollment */
    activeEnrollment: Enrollment | undefined,

    /* Current submission */
    currentSubmission: Submission | undefined,

    /* The value to sort submissions by */
    sortSubmissionsBy: SubmissionSort,

    /* Whether to sort by ascending or descending */
    sortAscending: boolean,

    /* Submission filters */
    submissionFilters: string[],

    /* Determine if all submissions should be displayed, or only group submissions */
    groupView: boolean,

    showFavorites: boolean,
}



/* Initial State */
/* To add to state, extend the State type and initialize the variable below */
export const state: State = {
    self: new User,
    isLoggedIn: derived(({ self }: State) => {
        return self.getId() !== 0
    }),

    isValid: derived(({ self }: State) => {
        return self.getName().length > 0 && self.getStudentid().length > 0 && self.getEmail().length > 0
    }),

    enrollments: [],
    enrollmentsByCourseID: derived((state: State) => {
        const enrollmentsByCourseID: EnrollmentsByCourse = {}
        for (const enrollment of state.enrollments) {
            enrollmentsByCourseID[enrollment.getCourseid()] = enrollment
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
        const course = state.courses.find(course => course.getId() === state.activeCourse)
        if (course && course.getCoursecreatorid() === state.self.getId()) {
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
                            return acc + ((cur.hasSubmission() &&
                                isApproved(cur.getSubmission() as Submission)) ? 1 : 0)
                        }, 0) ?? 0
                        return numApproved < numAssignments
                    })
                    break
                default:
                    break
            }
        }

        const m = state.sortAscending ? -1 : 1
        const sortedSubmissions = Object.values(filteredSubmissions).sort((a, b) => {
            let subA: Submission | undefined
            let subB: Submission | undefined
            if (rootState.review.assignmentID > 0) {
                // If a specific assignment is selected, sort by that assignment
                subA = getSubmissionByAssignmentID(a.submissions, rootState.review.assignmentID)
                subB = getSubmissionByAssignmentID(b.submissions, rootState.review.assignmentID)
            }

            switch (state.sortSubmissionsBy) {
                case SubmissionSort.Score:
                    if (rootState.review.assignmentID > 0) {
                        const sA = subA?.getScore()
                        const sB = subB?.getScore()
                        if (sA !== undefined && sB !== undefined) {
                            return m * (sB - sA)
                        } else if (sA !== undefined) {
                            return -m
                        }
                        return m
                    }
                    const aSubs = a.submissions ? getSubmissionsScore(a.submissions) : 0
                    const bSubs = b.submissions ? getSubmissionsScore(b.submissions) : 0
                    return m * (aSubs - bSubs)
                case SubmissionSort.Approved:
                    if (rootState.review.assignmentID > 0) {
                        const sA = subA && isApproved(subA) ? 1 : 0
                        const sB = subB && isApproved(subB) ? 1 : 0
                        return m * (sA - sB)
                    }
                    const aApproved = a.submissions ? getNumApproved(a.submissions) : 0
                    const bApproved = b.submissions ? getNumApproved(b.submissions) : 0
                    return m * (aApproved - bApproved)
                case SubmissionSort.Name:
                    const nameA = a.user?.getName() ?? ""
                    const nameB = b.user?.getName() ?? ""
                    return m * (nameA.localeCompare(nameB))
                default:
                    return 0
            }
        })
        return sortedSubmissions
    }),
    activeSubmission: derived((state: State) => {
        if (state.activeSubmissionLink) {
            return state.activeSubmissionLink.hasSubmission() ? (state.activeSubmissionLink.getSubmission() as Submission).getId() : -1
        }
        return -1
    }),
    activeEnrollment: undefined,
    activeSubmissionLink: undefined,
    currentSubmission: derived(({ activeSubmissionLink }: State) => {
        return activeSubmissionLink?.getSubmission()
    }),
    selectedAssignment: derived(({ activeCourse, currentSubmission, assignments }: State) => {
        return assignments[activeCourse]?.find(a => a.getId() === currentSubmission?.getAssignmentid())
    }),
    activeUser: undefined,
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
        return activeCourse > 0 ? groups[activeCourse]?.filter(group => isPendingGroup(group)) : []
    }),
    pendingEnrollments: derived(({ activeCourse, courseEnrollments }: State) => {
        return activeCourse > 0 ? courseEnrollments[activeCourse]?.filter(enrollment => isPending(enrollment)) : []
    }),
    numEnrolled: derived(({ activeCourse, courseEnrollments }: State) => {
        return activeCourse > 0 ? courseEnrollments[activeCourse]?.filter(enrollment => !isPending(enrollment)).length : 0
    }),
    query: "",
    sortSubmissionsBy: SubmissionSort.Approved,
    sortAscending: true,
    submissionFilters: [],
    groupView: false,
    showFavorites: false,
}

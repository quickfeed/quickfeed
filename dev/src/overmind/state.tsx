import { derived } from "overmind"
import { Assignment, Course, Enrollment, Group, Submission, SubmissionLink, User } from "../../proto/ag/ag_pb"
import { Color, isPending, isPendingGroup, isTeacher } from "../Helpers"

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

    /* Contains all submissions for a given course and enrollment */
    courseSubmissions: { [courseID: number]: UserCourseSubmissions[] },

    courseGroupSubmissions: { [courseID: number]: UserCourseSubmissions[] },

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
    status: [],

    users: {},
    allUsers: [],
    courses: [],
    courseSubmissions: [],
    courseGroupSubmissions: {},

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
}

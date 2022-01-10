import { derived } from "overmind";
import { Assignment, Course, Enrollment, Group, Submission, SubmissionLink, User } from "../../proto/ag/ag_pb";
import { Color } from "../Helpers";

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

export interface ParsedCourseSubmissions {
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
    enrollmentsByCourseId: { [courseid: number]: Enrollment },

    /* Contains all the groups the user is a member of, indexed by course ID */
    userGroup: { [courseid: number]: Group },

    /* Contains all submissions for the user, indexed by course ID */
    // The individual submissions for a given course are indexed by assignment order - 1
    submissions: { [courseid: number]: Submission[] },

    /* Current enrollment status of the user for a given course */
    status: { [courseid: number]: Enrollment.UserStatus }

    /* Indicates if the user is a teacher of the current course */
    // derived from enollmentsByCourseId
    isTeacher: boolean

    /* Contains links to all repositories for a given course */
    // Individual repository links are accessed by Reposiotry.Type
    repositories: { [courseid: number]: { [repo: string]: string } },

    /***************************************************************************
    *                              Public data
    ***************************************************************************/

    /* Contains all users of a given course */
    // Requires the user to be admin to get from backend
    users: { [userid: number]: User },

    /* Contains all courses */
    courses: Course[],

    /* Contains all assignments for a given course */
    assignments: { [courseid: number]: Assignment[] },


    /***************************************************************************
    *                         Course Specific Data 
    ***************************************************************************/

    /* Contains all submissions for a given course and enrollment */
    courseSubmissions: { [courseid: number]: { [enrollmentId: number]: ParsedCourseSubmissions } },

    /* Contains all submissions for a given course */
    // derived from courseSubmissions
    courseSubmissionsList: { [courseid: number]: ParsedCourseSubmissions[] }

    /* Contains all enrollments for a given course */
    courseEnrollments: { [courseid: number]: Enrollment[] },

    /* Contains all groups for a given course */
    groups: { [courseid: number]: Group[] },

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

    /* Contains all users sorted by admin status */
    allUsers: User[],


    /* Utility */

    /* Indicates if the state is loading */
    isLoading: boolean,

    /* The current course ID */
    activeCourse: number,

    /* The current submission ID */
    activeLab: number,

    /* The current assignment */
    selectedAssignment: Assignment | undefined,

    // TODO: Figure out if it is needed to store and continuously update this, or if it could be fetched on demand
    /* Current time */
    timeNow: Date,

    /* Contains a group in creation */
    courseGroup: CourseGroup,

    /* Contains alerts to be displayed to the user */
    alerts: Alert[],

    /* Current search query */
    query: string,

    /* Current enrollment ID */
    selectedEnrollment: number,

    /* Current submission link */
    activeSubmissionLink: SubmissionLink | undefined,

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
    enrollmentsByCourseId: derived((state: State) => {
        const enrollmentsByCourseId: EnrollmentsByCourse = {}
        for (const enrollment of state.enrollments) {
            enrollmentsByCourseId[enrollment.getCourseid()] = enrollment
        }
        return enrollmentsByCourseId
    }),
    submissions: {},
    userGroup: {},

    isTeacher: derived((state: State) => {
        if (state.activeCourse > 0 && state.enrollmentsByCourseId[state.activeCourse]) {
            return state.enrollmentsByCourseId[state.activeCourse].getStatus() === Enrollment.UserStatus.TEACHER
        }
        return false
    }),
    status: [],

    users: {},
    allUsers: [],
    courses: [],
    courseSubmissions: {},
    courseSubmissionsList: derived(({ courseSubmissions }: State) => {
        const courseSubmissionsList: { [courseid: number]: ParsedCourseSubmissions[] } = {}
        for (const courseid of Object.keys(courseSubmissions)) {
            courseSubmissionsList[Number(courseid)] = Object.values(courseSubmissions[Number(courseid)])
        }
        return courseSubmissionsList
    }),
    activeSubmission: -1,
    activeSubmissionLink: undefined,
    currentSubmission: derived(({ activeSubmissionLink }: State) => {
        return activeSubmissionLink?.getSubmission()
    }),
    selectedAssignment: derived(({ activeCourse, currentSubmission, assignments }: State) => {
        return assignments[activeCourse]?.find(a => currentSubmission && a.getId() === currentSubmission?.getAssignmentid())
    }),
    selectedEnrollment: -1,
    activeUser: undefined,
    assignments: {},
    repositories: {},

    courseGroup: { courseID: 0, enrollments: [], users: [], groupName: "" },
    timeNow: new Date(),
    alerts: [],
    isLoading: true,
    activeCourse: -1,
    activeLab: -1,
    courseEnrollments: {},
    groups: {},
    pendingGroups: derived(({ activeCourse, groups }: State) => { return activeCourse > 0 ? groups[activeCourse]?.filter((group) => group.getStatus() === Group.GroupStatus.PENDING) : [] }),
    pendingEnrollments: derived(({ activeCourse, courseEnrollments }: State) => {
        return activeCourse > 0 ? courseEnrollments[activeCourse]?.filter(enrollment => enrollment.getStatus() === Enrollment.UserStatus.PENDING) : []
    }),
    numEnrolled: derived(({ activeCourse, courseEnrollments }: State) => {
        return activeCourse > 0 ? courseEnrollments[activeCourse]?.filter(enrollment => enrollment.getStatus() !== Enrollment.UserStatus.PENDING).length : 0
    }),
    query: "",
};

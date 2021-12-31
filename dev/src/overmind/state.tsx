import { derived } from "overmind";
import { Assignment, Course, Enrollment, EnrollmentLink, Group, Submission, SubmissionLink, User } from "../../proto/ag/ag_pb";

export interface CourseGroup {
    courseID: number
    enrollments: number[]
    users: User[]
    groupName: string
}

export interface ParsedCourseSubmissions {
    enrollment?: Enrollment
    user?: User
    submissions?: SubmissionLink[]
}

type EnrollmentsByCourse = {[courseid: number]: Enrollment}

type State = {

    /* Data relating to self */
    self: User,
    isValid: boolean,
    isLoggedIn: boolean,
    enrollments: Enrollment[],
    enrollmentsByCourseId: {
        [courseid: number]: Enrollment
    },
    userCourses: {
        [courseid: number]: Course
    },
    userGroup: {
        [courseid: number]: Group
    },
    submissions:{
        [courseid: number]:Submission[]
    },
    status: {
        [courseid: number]: Enrollment.UserStatus
    }

    /* Public Data */
    users: {
        [userid: number]: User
    },
    courses: Course[],
    assignments: {
        [courseid:number]:Assignment[]
    },

    /* Course Specific Data */
    courseSubmissions: {
        [courseid:number]: {[enrollmentId: number]: ParsedCourseSubmissions}
    },
    courseSubmissionsList: {
        [courseid: number]: ParsedCourseSubmissions[]
    }
    activeSubmission: number,
    activeUser: User | undefined,


    allUsers: User[],
    courseGroupSubmissions: {
        [courseid: number]: EnrollmentLink[]
    },
    repositories: {
        [courseid:number]: { [repo: string]: string }
    },
    courseEnrollments: {
        [courseid: number]: Enrollment[]
    },
    groups: {
        [courseid: number]: Group[]
    },
    pendingGroups: Group[],
    pendingEnrollments: Enrollment[],
    numEnrolled: number,
    /* Utility */
    isLoading: boolean,
    activeCourse: number,
    activeLab: number,
    selectedAssignment: Assignment | undefined,
    timeNow: Date,
    // Used to create new group
    courseGroup: CourseGroup,
    alerts: {text: string, type: number}[],
    query: string,
    enableRedirect: boolean,
    selectedEnrollment: number,
    activeSubmissionLink: SubmissionLink | undefined,
    currentSubmission: Submission | undefined,
    isTeacher: boolean
}



/* Initial State */
/* To add to state, extend the State type and initialize the variable below */
export const state: State = {
    self: new User,
    isLoggedIn: derived(({self}: State) => {
        return self.getId() !== 0
    }),
    isValid: derived(({self}: State) => {
        return self.getName().length > 0 && self.getStudentid().length > 0 && self.getEmail().length > 0
    }),
    users: {},
    allUsers: [],
    enrollments: [],
    enrollmentsByCourseId: derived((state: State) => {
        const obj: EnrollmentsByCourse = {}
        for (const enrollment of state.enrollments) {
            obj[enrollment.getCourseid()] = enrollment
        }
        return obj
    }),
    isTeacher: derived((state: State) => {
        if (state.activeCourse > 0 && state.enrollmentsByCourseId[state.activeCourse]) {
            return state.enrollmentsByCourseId[state.activeCourse].getStatus() === Enrollment.UserStatus.TEACHER
        }
        return false
    })
    status: [],
    courses: [],
    userCourses: {},
    userGroup: {},
    submissions: {},
    courseSubmissions: {},
    courseSubmissionsList: derived(({courseSubmissions}: State) => {
        const obj: {[courseid: number]: ParsedCourseSubmissions[]} = {}
        for (const key of Object.keys(courseSubmissions)) {
            obj[Number(key)] = Object.values(courseSubmissions[Number(key)])
        }
        return obj
    }),
    activeSubmission: -1,
    activeSubmissionLink: undefined,
    currentSubmission: derived(({activeSubmissionLink}: State) => {
        return activeSubmissionLink?.getSubmission()
    }),
    selectedAssignment: derived(({activeCourse, currentSubmission, assignments}: State) => {
        return assignments[activeCourse]?.find(a => currentSubmission && a.getId() === currentSubmission?.getAssignmentid())
    }),
    selectedEnrollment: -1,
    activeUser: undefined,
    courseGroupSubmissions: {},
    assignments: {},
    repositories: {},

    courseGroup: {courseID: 0, enrollments: [], users: [], groupName: ""},
    timeNow : new Date(),
    alerts: [],
    isLoading: true,
    activeCourse: -1,
    activeLab: -1,
    courseEnrollments: {},
    groups: {},
    pendingGroups: derived(({activeCourse, groups}: State) => { return activeCourse > 0 ? groups[activeCourse].filter((group) => group.getStatus() === Group.GroupStatus.PENDING) : []}),
    pendingEnrollments: derived(({activeCourse, courseEnrollments}: State) => { 
        return activeCourse > 0 ? courseEnrollments[activeCourse].filter(enrollment => enrollment.getStatus() === Enrollment.UserStatus.PENDING) : []
    }),
    numEnrolled: derived(({activeCourse, courseEnrollments}: State) => { 
        return activeCourse > 0 ? courseEnrollments[activeCourse].filter(enrollment => enrollment.getStatus() !== Enrollment.UserStatus.PENDING).length : 0
    }),
    query: "",
    enableRedirect: true
};
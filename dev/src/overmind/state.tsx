import { derived } from "overmind";

import { Assignment, Course, Enrollment, EnrollmentLink, Group, Review, Submission, SubmissionLink, User } from "../../proto/ag/ag_pb";

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
    users: Enrollment[],
    courses: Course[],
    assignments: {
        [courseid:number]:Assignment[]
    },

    /* Course Specific Data */
    courseSubmissions: {
        [courseid:number]: ParsedCourseSubmissions[]
    },
    activeSubmission: Submission | undefined,
    activeSubmissionLink: SubmissionLink | undefined,
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
    theme: string,
    isLoading: boolean,
    activeCourse: number,
    activeLab: number,
    activeReview: Review | undefined
    timeNow: Date,
    // Used to create new group
    courseGroup: CourseGroup,
    alerts: {text: string, type: number}[],
    query: string,
    enableRedirect: boolean
}



/* Initial State */
/* To add to state, extend the State type and initialize the variable below */
export const state: State = {
    self: new User,
    isLoggedIn: derived(({self}: State) => {
        return self.getId() !== 0
    }),
    users: [],
    allUsers: [],
    enrollments: [],
    enrollmentsByCourseId: derived((state: State) => {
        const obj: EnrollmentsByCourse = {}
        for (const enrollment of state.enrollments) {
            obj[enrollment.getCourseid()] = enrollment
        }
        return obj
    }),
    status: [],
    courses: [],
    userCourses: {},
    userGroup: {},
    submissions: {},
    courseSubmissions: {},
    activeSubmission: undefined,
    activeSubmissionLink: undefined,
    activeReview: undefined,
    activeUser: undefined,
    courseGroupSubmissions: {},
    assignments: {},
    repositories: {},

    courseGroup: {courseID: 0, enrollments: [], users: [], groupName: ""},
    timeNow : new Date(),
    alerts: [],
    theme: "light",
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

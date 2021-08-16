import { derived } from "overmind";

import { Assignment, Course, Enrollment, EnrollmentLink, Group, Submission, SubmissionLink, User } from "../../proto/ag/ag_pb";



// TODO Style for members of Self should be camelCase. The JSON from /api/v1/user does not return an object with camelCase. Rewrite return on backend to comply with camelCase
export interface Self {
    remoteID: number;
    avatarurl: string;
    email: string;
    id: number;
    isadmin: boolean;
    name: string;
    studentid: number;
    Token: string;
}

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
    /*cSubs: {
        [courseid:number]: ParsedCourseSubmissions[]
    },*/

    allUsers: User[],
    courseGroupSubmissions: {
        [courseid: number]: EnrollmentLink[]
    },
    repositories: {
        [courseid:number]: { [repoType: number]: string }
    },
    courseEnrollments: {
        [courseid: number]: Enrollment[]
    },
    groups: {
        [courseid: number]: Group[]
    },

    /* Utility */
    theme: string,
    isLoading: boolean,
    activeCourse: number,
    activeLab: number,
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
    users: [],
    allUsers: [],
    enrollments: [],
    enrollmentsByCourseId: derived((state: State) => {
        let obj: EnrollmentsByCourse = {}
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
    query: "",
    enableRedirect: true
};

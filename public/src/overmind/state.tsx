import { derived } from "overmind";

import { Assignment, Course, Enrollment, EnrollmentLink, Group, Submission, User } from "../../proto/ag_pb";



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
    enrollments: number[]
    users: User[]
    groupName: string
}

type EnrollmentsByCourse = {[courseid: number]: Enrollment}

export type State = {
    /*  */
    user: Self,
    users: Enrollment[],
    enrollments: Enrollment[],
    enrollmentsByCourseId: {
        [courseid: number]: Enrollment
    }
    courses: Course[],
    userCourses: Course[],
    userGroup: {
        [courseid: number]: Group
    }
    submissions:{
        [courseid:number]:Submission[]
    },
    courseSubmissions:{
        [courseid:number]: EnrollmentLink[]
    }
    assignments: {
        [courseid:number]:Assignment[]
    },
    repositories: {
        [courseid:number]: { [repoType: number]: string }
    }
    theme: string,
    isLoading: boolean,
    activeCourse: number,
    activeLab: number,
    search: string,
    userSearch: Enrollment[],
    timeNow: Date,
    courseGroup: CourseGroup,
    alerts: string[],
    courseEnrollments: {
        [courseid: number]: Enrollment[]
    },
    groups: {
        [courseid: number]: Group[]
    }
}

/* Initial State */
/* To add to state, extend the State type and initialize the variable below */
export const state: State = {
    user: {avatarurl: '', email: '', id: -1, isadmin: false, name: '', remoteID: -1, studentid: -1, Token: ""},
    users: [],
    enrollments: [],
    enrollmentsByCourseId: derived((state: State) => {
        // 
        let obj: EnrollmentsByCourse = {}
        state.enrollments.forEach(enrollment => {
            obj[enrollment.getCourseid()] = enrollment
        });
        return obj
    }),
    courses: [],
    userCourses: [],
    userGroup: {},
    submissions: {},
    courseSubmissions: {},
    assignments: {},
    repositories: {},

    search: "",
    // 
    userSearch: derived((state: State) => {
        return state.users.filter(user => 
            user.getUser()?.getName().toLowerCase().includes(state.search.toLowerCase())
        ).filter(user => !state.courseGroup.enrollments.includes(user.getId()))
    }),

    courseGroup: {enrollments: [], users: [], groupName: ""},
    timeNow : new Date(),
    alerts: [],
    theme: "light",
    isLoading: true,
    activeCourse: -1,
    activeLab: -1,
    courseEnrollments: {},
    groups: {}
};

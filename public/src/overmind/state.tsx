import { derived } from "overmind";
import { Assignment, Course, Enrollment, Submission, User } from "../proto/ag_pb";


// TODO Style for members of Self should be camelCase. The JSON from /api/v1/user does not return an object with camelCase. Rewrite return on backend to comply with camelCase
export interface Self {
    remoteID: number;
    avatarurl: string;
    email: string;
    id: number;
    isadmin: boolean;
    name: string;
    studentid: number;
    AccessToken: string;
}

export interface CourseGroup {
    enrollments: number[]
    users: User[]
    groupName: string
}

export type State = {
    user: Self,
    users: Enrollment[],
    enrollments: Enrollment[],
    enrollmentsByCourseId: {
        [courseid: number]: Enrollment
    }
    courses: Course[],
    userCourses: Course[],
    submissions:{
        [courseid:number]:Submission[]
    },
    assignments: {
        [courseid:number]:Assignment[]
    },
    repositories: {
        [courseid:number]: { [repoType: number]: string }
    }
    theme: string,
    isLoading: boolean,
    activeCourse: number,
    search: string,
    userSearch: Enrollment[],
    timeNow: Date,
    cg: CourseGroup,
    alerts: string[],
    courseEnrollments: {
        [courseid: number]: Enrollment[]
    }
}

export const state: State = {
    user: {avatarurl: '', email: '', id: -1, isadmin: false, name: '', remoteID: -1, studentid: -1, AccessToken: ""},
    users: [],
    enrollments: [],
    enrollmentsByCourseId: derived((state: State) => {
        type e = {[courseid: number]: Enrollment}
        let d: e = {}
        state.enrollments.forEach(enrollment => {
            d[enrollment.getCourseid()] = enrollment
        });
        return d
    }),
    courses: [],
    userCourses: [],
    submissions: {},
    assignments: {},
    repositories: {},
    theme: "light",
    isLoading: false,
    activeCourse: -1,
    search: "",
    userSearch: derived((state: State) => {
        return state.users.filter(user => 
            user.getUser()?.getName().toLowerCase().includes(state.search.toLowerCase())
        ).filter(user => !state.cg.enrollments.includes(user.getId()))
    }),

    timeNow : new Date(),
    cg: {enrollments: [], users: [], groupName: ""},
    alerts: [],
    courseEnrollments: {}
};

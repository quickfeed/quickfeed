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

export interface Student {
    enrollments: Enrollment[]
    courses: Course[]
}

export type State = {
    user: Self,
    users: User[],
    enrollments: Enrollment[]
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
    student: Student
}

export const state: State = {
    user: {avatarurl: '', email: '', id: -1, isadmin: false, name: '', remoteID: -1, studentid: -1, AccessToken: ""},
    users: [],
    enrollments: [],
    courses: [],
    userCourses: [],
    submissions: {},
    assignments: {},
    repositories: {},
    theme: "light",
    isLoading: true,
    activeCourse: -1,
    student: derived((state: State) => { 
        return {
            courses: state.courses, 
            enrollments: state.enrollments.filter(enrollment => {
                return enrollment.getStatus() === Enrollment.UserStatus.STUDENT
            }) 
        }
    })
};
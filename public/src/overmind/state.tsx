import { derived } from "overmind";
import { Assignment, Course, Enrollment, Submission, User } from "../proto/ag_pb";


export interface IUser {
    remoteID: number;
    avatarurl: string;
    email: string;
    id: number;
    isadmin: boolean;
    name: string;
    studentid: number;
    AccessToken: string;
}

export interface IStudent {
    enrollments: Enrollment[]
    courses: Course[]
}

export type State = {
    user: IUser,
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
    student: IStudent
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

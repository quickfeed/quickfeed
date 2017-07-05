import * as Models from "../models";
import { IAssignment, ICourse, ICourseStudent, IUser } from "../models";
import { ICourseProvider } from "./CourseManager";
import { IUserProvider } from "./UserManager";

import { IMap, MapHelper, mapify } from "../map";

interface IDummyUser extends IUser {
    password: string;
}

class TempDataProvider implements IUserProvider, ICourseProvider {

    private localUsers: IMap<IDummyUser>;
    private localAssignments: IMap<IAssignment>;
    private localCourses: IMap<ICourse>;
    private localCourseStudent: ICourseStudent[];

    constructor() {
        this.addLocalAssignments();
        this.addLocalCourses();
        this.addLocalCourseStudent();
        this.addLocalUsers();
    }

    public async getAllUser(): Promise<IMap<IUser>> {
        return this.localUsers;
    }

    public async getCourses(): Promise<IMap<ICourse>> {
        return this.localCourses;
    }

    public async getCoursesStudent(): Promise<ICourseStudent[]> {
        return this.localCourseStudent;
    }

    public async getAssignments(courseId: number): Promise<IMap<IAssignment>> {
        const temp: IMap<IAssignment> = [];
        MapHelper.forEach(this.localAssignments, (a, i) => {
            if (a.courseId === courseId) {
                temp[i] = a;
            }
        });
        return temp;
    }

    public async tryLogin(username: string, password: string): Promise<IUser | null> {
        const user = MapHelper.find(this.localUsers, (u) =>
            u.email.toLocaleLowerCase() === username.toLocaleLowerCase());
        if (user && user.password === password) {
            return user;
        }
        return null;
    }

    public async tryRemoteLogin(provider: string): Promise<IUser | null> {
        let lookup = "test@testersen.no";
        if (provider === "gitlab") {
            lookup = "bob@bobsen.no";
        }
        const user = MapHelper.find(this.localUsers, (u) =>
            u.email.toLocaleLowerCase() === lookup);

        return new Promise<IUser | null>((resolve, reject) => {
            // Simulate async callback
            setTimeout(() => {
                resolve(user);
            }, 500);
        });
    }

    public async logout(user: IUser): Promise<boolean> {
        return true;
    }

    public async addUserToCourse(user: IUser, course: ICourse): Promise<boolean> {
        this.localCourseStudent.push({
            courseId: course.id,
            personId: user.id,
            state: Models.CourseStudentState.pending,
        });
        return true;
    }

    public async createNewCourse(course: any): Promise<boolean> {
        const courses = MapHelper.toArray(this.localCourses);
        course.id = courses.length;
        const courseData: ICourse = course as ICourse;
        courses.push(courseData);
        this.localCourses = mapify(courses, (ele) => ele.id);
        return true;
    }

    public async changeUserState(link: ICourseStudent, state: Models.CourseStudentState): Promise<boolean> {
        link.state = state;
        return true;
    }

    public async changeAdminRole(user: IUser): Promise<boolean> {
        user.isAdmin = !user.isAdmin;
        return true;
    }

    private addLocalUsers() {
        this.localUsers = mapify([
            {
                id: 999,
                firstName: "Test",
                lastName: "Testersen",
                email: "test@testersen.no",
                personId: 9999,
                password: "1234",
                isAdmin: true,
            },
            {
                id: 1000,
                firstName: "Admin",
                lastName: "Admin",
                email: "admin@admin",
                personId: 1000,
                password: "1234",
                isAdmin: true,
            },
            {
                id: 1,
                firstName: "Per",
                lastName: "Pettersen",
                email: "per@pettersen.no",
                personId: 1234,
                password: "1234",
                isAdmin: false,
            },
            {
                id: 2,
                firstName: "Bob",
                lastName: "Bobsen",
                email: "bob@bobsen.no",
                personId: 1234,
                password: "1234",
                isAdmin: false,
            },
            {
                id: 3,
                firstName: "Petter",
                lastName: "Pan",
                email: "petter@pan.no",
                personId: 1234,
                password: "1234",
                isAdmin: false,
            },
        ], (ele) => ele.id);
    }

    private addLocalAssignments() {
        this.localAssignments = mapify([
            {
                id: 0,
                courseId: 0,
                name: "Lab 1",
                start: new Date(2017, 5, 1),
                deadline: new Date(2017, 5, 25),
                end: new Date(2017, 5, 30),
            },
            {
                id: 1,
                courseId: 0,
                name: "Lab 2",
                start: new Date(2017, 5, 1),
                deadline: new Date(2017, 5, 25),
                end: new Date(2017, 5, 30),
            },
            {
                id: 2,
                courseId: 0,
                name: "Lab 3",
                start: new Date(2017, 5, 1),
                deadline: new Date(2017, 5, 25),
                end: new Date(2017, 5, 30),
            },
            {
                id: 3,
                courseId: 0,
                name: "Lab 4",
                start: new Date(2017, 5, 1),
                deadline: new Date(2017, 5, 25),
                end: new Date(2017, 5, 30),
            },
            {
                id: 4,
                courseId: 1,
                name: "Lab 1",
                start: new Date(2017, 5, 1),
                deadline: new Date(2017, 5, 25),
                end: new Date(2017, 5, 30),
            },
            {
                id: 5,
                courseId: 1,
                name: "Lab 2",
                start: new Date(2017, 5, 1),
                deadline: new Date(2017, 5, 25),
                end: new Date(2017, 5, 30),
            },
            {
                id: 6,
                courseId: 1,
                name: "Lab 3",
                start: new Date(2017, 5, 1),
                deadline: new Date(2017, 5, 25),
                end: new Date(2017, 5, 30),
            },
            {
                id: 7,
                courseId: 2,
                name: "Lab 1",
                start: new Date(2017, 5, 1),
                deadline: new Date(2017, 5, 25),
                end: new Date(2017, 5, 30),
            },
            {
                id: 8,
                courseId: 2,
                name: "Lab 2",
                start: new Date(2017, 5, 1),
                deadline: new Date(2017, 5, 25),
                end: new Date(2017, 5, 30),
            },
            {
                id: 9,
                courseId: 3,
                name: "Lab 1",
                start: new Date(2017, 5, 1),
                deadline: new Date(2017, 5, 25),
                end: new Date(2017, 5, 30),
            },
            {
                id: 10,
                courseId: 4,
                name: "Lab 1",
                start: new Date(2017, 5, 1),
                deadline: new Date(2017, 5, 25),
                end: new Date(2017, 5, 30),
            },
        ], (ele) => ele.id);
    }

    private addLocalCourses() {
        this.localCourses = mapify([
            {
                id: 0,
                name: "Object Oriented Programming",
                tag: "DAT100",
                year: "Spring 2017",
            },
            {
                id: 1,
                name: "Algorithms and Datastructures",
                tag: "DAT200",
                year: "Spring 2017",
            },
            {
                id: 2,
                name: "Databases",
                tag: "DAT220",
                year: "Spring 2017",
            },
            {
                id: 3,
                name: "Communication Technology",
                tag: "DAT230",
                year: "Spring 2017",
            },
            {
                id: 4,
                name: "Operating Systems",
                tag: "DAT320",
                year: "Spring 2017",
            },
        ], (ele) => ele.id);
    }

    private addLocalCourseStudent() {
        this.localCourseStudent = [
            { courseId: 0, personId: 999, state: 1 },
            { courseId: 1, personId: 999, state: 1 },
            { courseId: 0, personId: 1, state: 0 },
            { courseId: 0, personId: 2, state: 0 },
        ];
    }

}

export { TempDataProvider };

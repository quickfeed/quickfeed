import { IAssignment, ICourse, ICourseStudent, IUser } from "../models";
import { ICourseProvider } from "./CourseManager";
import { IUserProvider } from "./UserManager";

interface IDummyUser extends IUser {
    password: string;
}

class TempDataProvider implements IUserProvider, ICourseProvider {

    private localUsers: IDummyUser[];
    private localAssignments: IAssignment[];
    private localCourses: ICourse[];
    private localCourseStudent: ICourseStudent[];

    constructor() {
        this.addLocalAssignments();
        this.addLocalCourses();
        this.addLocalCourseStudent();
        this.addLocalUsers();
    }

    public getAllUser(): IUser[] {
        return this.localUsers;
    }

    public getCourses(): ICourse[] {
        return this.localCourses;
    }

    public getCourseByTag(tag: string): ICourse | null {
        for (const c of this.localCourses) {
            if (c.tag === tag) {
                return c;
            }
        }
        return null;
    }

    public getCoursesStudent(): ICourseStudent[] {
        return this.localCourseStudent;
    }

    public getAssignments(courseId: number): IAssignment[] {
        const temp: IAssignment[] = [];
        for (const a of this.localAssignments) {
            if (a.courseId === courseId) {
                temp.push(a);
            }
        }
        return temp;
    }

    public tryLogin(username: string, password: string): IUser | null {
        for (const u of this.localUsers) {
            if (u.email.toLocaleLowerCase() === username.toLocaleLowerCase()) {
                if (u.password === password) {
                    return u;
                }
                return null;
            }
        }
        return null;
    }

    public logout(user: IUser): void {
        "Do nothing";
    }

    public addUserToCourse(user: IUser, course: ICourse): void {
        this.localCourseStudent.push({ courseId: course.id, personId: user.id });
    }

    private addLocalUsers() {
        this.localUsers = [
            {
                id: 999,
                firstName: "Test",
                lastName: "Testersen",
                email: "test@testersen.no",
                personId: 9999,
                password: "1234",
            },
            {
                id: 1,
                firstName: "Per",
                lastName: "Pettersen",
                email: "per@pettersen.no",
                personId: 1234,
                password: "1234",
            },
            {
                id: 2,
                firstName: "Bob",
                lastName: "Bobsen",
                email: "bob@bobsen.no",
                personId: 1234,
                password: "1234",
            },
            {
                id: 3,
                firstName: "Petter",
                lastName: "Pan",
                email: "petter@pan.no",
                personId: 1234,
                password: "1234",
            },
        ];
    }

    private addLocalAssignments() {
        this.localAssignments = [
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
        ];
    }

    private addLocalCourses() {
        this.localCourses = [
            {
                id: 0,
                name: "Object Oriented Programming",
                tag: "DAT100",
            },
            {
                id: 1,
                name: "Algorithms and Datastructures",
                tag: "DAT200",
            },
            {
                id: 2,
                name: "Databases",
                tag: "DAT220",
            },
            {
                id: 3,
                name: "Communication Technology",
                tag: "DAT230",
            },
            {
                id: 4,
                name: "Operating Systems",
                tag: "DAT320",
            },
        ];
    }

    private addLocalCourseStudent() {
        this.localCourseStudent = [
            { courseId: 0, personId: 999 },
            { courseId: 1, personId: 999 },
        ];
    }

}

export { TempDataProvider };

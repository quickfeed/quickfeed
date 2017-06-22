import { IUser, IAssignment, ICourse, ICourseStudent } from "../models";
import { IUserProvider } from "./UserManager";
import { ICourseProvider } from "./CourseManager";

interface IDummyUser extends IUser {
    password: string;
}

class TempDataProvider implements IUserProvider, ICourseProvider{
    

    private localUsers: IDummyUser[];
    private localAssignments: IAssignment[];
    private localCourses: ICourse[];
    private localCourseStudent: ICourseStudent[];

    constructor(){
        this.addLocalAssignments();
        this.addLocalCourses();
        this.addLocalCourseStudent();
        this.addLocalUsers();
    }

    private addLocalUsers(){
        this.localUsers = [
            {
                id: 999,
                firstName: "Test",
                lastName: "Testersen",
                email: "test@testersen.no",
                personId: 9999,
                password: "1234"
            },
            {
                id: 1,
                firstName: "Per",
                lastName: "Pettersen",
                email: "per@pettersen.no",
                personId: 1234,
                password: "1234"
            },
            {
                id: 2,
                firstName: "Bob",
                lastName: "Bobsen",
                email: "bob@bobsen.no",
                personId: 1234,
                password: "1234"
            },
            {
                id: 3,
                firstName: "Petter",
                lastName: "Pan",
                email: "petter@pan.no",
                personId: 1234,
                password: "1234"
            }
        ];
    }

    private addLocalAssignments(){
        this.localAssignments = [
            {
                id: 0,
                courceId: 0,
                name: "Lab 1",
                start: new Date(2017, 5, 1),
                deadline: new Date(2017, 5, 25),
                end: new Date(2017, 5, 30)
            },
            {
                id: 1,
                courceId: 0,
                name: "Lab 2",
                start: new Date(2017, 5, 1),
                deadline: new Date(2017, 5, 25),
                end: new Date(2017, 5, 30)
            },
            {
                id: 2,
                courceId: 0,
                name: "Lab 3",
                start: new Date(2017, 5, 1),
                deadline: new Date(2017, 5, 25),
                end: new Date(2017, 5, 30)
            },
            {
                id: 3,
                courceId: 0,
                name: "Lab 4",
                start: new Date(2017, 5, 1),
                deadline: new Date(2017, 5, 25),
                end: new Date(2017, 5, 30)
            },
            {
                id: 4,
                courceId: 1,
                name: "Lab 1",
                start: new Date(2017, 5, 1),
                deadline: new Date(2017, 5, 25),
                end: new Date(2017, 5, 30)
            }
        ];
    }

    private addLocalCourses(){
        this.localCourses = [
            {
                id: 0,
                name: "Object Oriented Programming",
                tag: "DAT100"
            },
            {
                id: 1,
                name: "Algorithms and Datastructures",
                tag: "DAT200"
            }
        ];
    }

    private addLocalCourseStudent(){
        this.localCourseStudent = [
            { courseId: 0, personId: 999 }
        ];
    }

    getAllUser(): IUser[] {
        return this.localUsers;
    }

    getCourses(): ICourse[] {
        return this.localCourses;
    }

    getCoursesStudent(): ICourseStudent[] {
        return this.localCourseStudent;
    }
    
    getAssignments(courseId: number): IAssignment[] {
        let temp: IAssignment[] = [];
        for(let a of this.localAssignments){
            if (a.courceId === courseId){
                temp.push(a);
            }
        }
        return temp;
    }

    tryLogin(username: string, password: string): IUser | null {
        for(let u of this.localUsers){
            if (u.email.toLocaleLowerCase() === username.toLocaleLowerCase()){
                if (u.password === password){
                    return u;
                }
                return null;
            }
        }
        return null;
    }

}

export {TempDataProvider};
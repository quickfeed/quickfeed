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
                courceId: 1,
                name: "Lab 1",
                start: new Date(2017, 5, 1),
                deadline: new Date(2017, 5, 25),
                end: new Date(2017, 5, 30)
            }
        ];

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

    getAllUser(): IUser[] {
        return this.localUsers;
    }

    getCourses(): ICourse[] {
        return this.localCourses;
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
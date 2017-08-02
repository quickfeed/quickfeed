import * as Models from "../models";
import {
    CourseGroupStatus,
    CourseUserState,
    IAssignment,
    ICourse,
    ICourseGroup,
    ICourseUserLink,
    ICourseWithEnrollStatus,
    IError,
    INewGroup,
    IOrganization,
    ISubmission,
    IUser,
} from "../models";
import { ICourseProvider } from "./CourseManager";

import { IMap, MapHelper, mapify } from "../map";
import { IUserProvider } from "./UserManager";

import { ICourseEnrollemtnt, IUserEnrollment } from "../managers";

interface IDummyUser extends IUser {
    password: string;
}

/**
 * The TempDataProvider is a fake implemetation of the backend
 * to be able to simulate the backend for easier developtment
 */
export class TempDataProvider implements IUserProvider, ICourseProvider {

    private localUsers: IMap<IDummyUser>;
    private localAssignments: IMap<IAssignment>;
    private localCourses: IMap<ICourse>;
    private localCourseStudent: ICourseUserLink[];
    private localLabInfo: IMap<ISubmission>;
    private localCourseGroups: ICourseGroup[];

    private currentLoggedIn: IUser | null = null;

    constructor() {
        this.addLocalAssignments();
        this.addLocalCourses();
        this.addLocalCourseStudent();
        this.addLocalUsers();
        this.addLocalLabInfo();
        this.addLocalCourseGroups();
    }

    public async getDirectories(provider: string): Promise<IOrganization[]> {
        throw new Error("Not implemented");
    }

    public async getAllUser(): Promise<IUser[]> {
        return MapHelper.toArray(this.localUsers);
    }

    public async getCourses(): Promise<ICourse[]> {
        // return this.localCourses;
        return MapHelper.toArray(this.localCourses);
    }

    public async getCoursesStudent(): Promise<ICourseUserLink[]> {
        return this.localCourseStudent;
    }

    public async getAssignments(courseId: number): Promise<IMap<IAssignment>> {
        const temp: IMap<IAssignment> = [];
        MapHelper.forEach(this.localAssignments, (a, i) => {
            if (a.courseid === courseId) {
                temp[i] = a;
            }
        });
        return temp;
    }

    public async tryLogin(username: string, password: string): Promise<IUser | null> {
        const user = MapHelper.find(this.localUsers, (u) =>
            u.email.toLocaleLowerCase() === username.toLocaleLowerCase());
        if (user && user.password === password) {
            this.currentLoggedIn = user;
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
                this.currentLoggedIn = user;
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
            userid: user.id,
            state: Models.CourseUserState.pending,
        });
        return true;
    }

    /**
     * Get all userlinks for a single course
     * @param course The course userlinks should be retrived from
     * @param state Optinal. The state of the relation, all if not present
     */
    public async getUserLinksForCourse(course: ICourse, state?: CourseUserState[]): Promise<ICourseUserLink[]> {
        const users: ICourseUserLink[] = [];
        for (const c of await this.getCoursesStudent()) {
            if (course.id === c.courseId && (state === undefined || c.state === CourseUserState.student)) {
                users.push(c);
            }
        }
        return users;
    }

    public async getUsersAsMap(ids: number[]): Promise<IMap<IUser>> {
        const returnUsers: IMap<IUser> = {};
        const allUsers = await this.getAllUser();
        ids.forEach((ele) => {
            const temp = allUsers[ele];
            if (temp) {
                returnUsers[ele] = temp;
            }
        });
        return returnUsers;
    }

    public async getUsersForCourse(course: Models.ICourse, state?: CourseUserState[])
        : Promise<IUserEnrollment[]> {
        const courseStds: ICourseUserLink[] =
            await this.getUserLinksForCourse(course, state);
        const users = await this.getUsersAsMap(courseStds.map((e) => e.userid));
        return courseStds.map<IUserEnrollment>((link) => {
            const user = users[link.userid];
            if (!user) {
                // TODO: See if we should have an error here or not
                throw new Error("Link exist witout a user object");
            }
            return { courseid: link.courseId, userid: link.userid, user, status: link.state };
        });
    }

    public async createNewCourse(course: any): Promise<boolean> {
        const courses = MapHelper.toArray(this.localCourses);
        course.id = courses.length;
        const courseData: ICourse = course as ICourse;
        courses.push(courseData);
        this.localCourses = mapify(courses, (ele) => ele.id);
        return true;
    }

    public async getCourse(id: number): Promise<ICourse | null> {
        const course: ICourse | undefined = this.localCourses[id];
        if (course) {
            return course;
        }
        return null;
    }

    public async updateCourse(courseId: number, courseData: ICourse): Promise<boolean> {
        const course: ICourse | undefined = this.localCourses[courseId];
        if (course) {
            this.localCourses[courseData.id] = courseData;
            return true;
        }
        return false;
    }

    public async changeUserState(link: ICourseUserLink, state: Models.CourseUserState): Promise<boolean> {
        link.state = state;
        return true;
    }

    public async changeAdminRole(user: IUser): Promise<boolean> {
        user.isadmin = !user.isadmin;
        return true;
    }

    public async getAllLabInfos(courseId: number): Promise<IMap<ISubmission>> {
        const temp: IMap<ISubmission> = {};
        const assignments = await this.getAssignments(courseId);
        MapHelper.forEach(this.localLabInfo, (ele) => {
            if (assignments[ele.assignmentid]) {
                temp[ele.id] = ele;
            }
        });
        return temp;
    }

    public async getProviders(): Promise<string[]> {
        return ["github"];
    }

    public async getLoggedInUser(): Promise<IUser | null> {
        return this.currentLoggedIn;
    }

    public async getCoursesFor(user: IUser, state?: CourseUserState[]): Promise<ICourseEnrollemtnt[]> {
        const cLinks: ICourseUserLink[] = [];
        const temp = await this.getCoursesStudent();
        for (const c of temp) {
            if (user.id === c.userid && (state === undefined || c.state === CourseUserState.student)) {
                cLinks.push(c);
            }
        }
        const courses: ICourseEnrollemtnt[] = [];
        const tempCourses = await this.getCourses();
        for (const link of cLinks) {
            const c = tempCourses[link.courseId];
            if (c) {
                courses.push({ course: c, courseid: link.courseId, userid: link.userid, status: link.state });
            }
        }
        return courses;
    }
    public async createGroup(groupData: INewGroup, courseId: number): Promise<ICourseGroup | IError> {
        throw new Error("Method not implemented");
    }
    public async getCourseGroups(courseId: number): Promise<ICourseGroup[]> {
        return this.localCourseGroups;
    }

    public async getCourseByUserAndCourse(userid: number, courseid: number): Promise<ICourseGroup | null> {
        throw new Error("Method not implemented");
    }

    public async updateGroupStatus(groupId: number, status: CourseGroupStatus): Promise<boolean> {
        throw new Error("Method not implemented");
    }

    private addLocalUsers() {
        this.localUsers = mapify([
            {
                id: 999,
                firstname: "Test",
                lastname: "Testersen",
                email: "test@testersen.no",
                studentnr: "9999",
                password: "1234",
                isadmin: true,
            },
            {
                id: 1000,
                firstname: "Admin",
                lastname: "Admin",
                email: "admin@admin",
                studentnr: "1000",
                password: "1234",
                isadmin: true,
            },
            {
                id: 1,
                firstname: "Per",
                lastname: "Pettersen",
                email: "per@pettersen.no",
                studentnr: "1234",
                password: "1234",
                isadmin: false,
            },
            {
                id: 2,
                firstname: "Bob",
                lastname: "Bobsen",
                email: "bob@bobsen.no",
                studentnr: "1234",
                password: "1234",
                isadmin: false,
            },
            {
                id: 3,
                firstname: "Petter",
                lastname: "Pan",
                email: "petter@pan.no",
                studentnr: "1234",
                password: "1234",
                isadmin: false,
            },
        ] as IDummyUser[], (ele) => ele.id);
    }

    private addLocalAssignments() {
        this.localAssignments = mapify([
            {
                id: 0,
                courseid: 0,
                name: "Lab 1",
                // start new Date(2017, 5, 1),
                deadline: new Date(2017, 5, 25),
                // end new Date(2017, 5, 30),
            },
            {
                id: 1,
                courseid: 0,
                name: "Lab 2",
                // start new Date(2017, 5, 1),
                deadline: new Date(2017, 5, 25),
                // end new Date(2017, 5, 30),
            },
            {
                id: 2,
                courseid: 0,
                name: "Lab 3",
                // start new Date(2017, 5, 1),
                deadline: new Date(2017, 5, 25),
                // end new Date(2017, 5, 30),
            },
            {
                id: 3,
                courseid: 0,
                name: "Lab 4",
                // start new Date(2017, 5, 1),
                deadline: new Date(2017, 5, 25),
                // end new Date(2017, 5, 30),
            },
            {
                id: 4,
                courseid: 1,
                name: "Lab 1",
                // start new Date(2017, 5, 1),
                deadline: new Date(2017, 5, 25),
                // end new Date(2017, 5, 30),
            },
            {
                id: 5,
                courseid: 1,
                name: "Lab 2",
                // start new Date(2017, 5, 1),
                deadline: new Date(2017, 5, 25),
                // end new Date(2017, 5, 30),
            },
            {
                id: 6,
                courseid: 1,
                name: "Lab 3",
                // start new Date(2017, 5, 1),
                deadline: new Date(2017, 5, 25),
                // end new Date(2017, 5, 30),
            },
            {
                id: 7,
                courseid: 2,
                name: "Lab 1",
                // start new Date(2017, 5, 1),
                deadline: new Date(2017, 5, 25),
                // end new Date(2017, 5, 30),
            },
            {
                id: 8,
                courseid: 2,
                name: "Lab 2",
                // start new Date(2017, 5, 1),
                deadline: new Date(2017, 5, 25),
                // end new Date(2017, 5, 30),
            },
            {
                id: 9,
                courseid: 3,
                name: "Lab 1",
                // start new Date(2017, 5, 1),
                deadline: new Date(2017, 5, 25),
                // end new Date(2017, 5, 30),
            },
            {
                id: 10,
                courseid: 4,
                name: "Lab 1",
                // start new Date(2017, 5, 1),
                deadline: new Date(2017, 5, 25),
                // end new Date(2017, 5, 30),
            },
        ] as IAssignment[], (ele) => ele.id);
    }

    private addLocalCourses() {
        this.localCourses = mapify([
            {
                id: 0,
                name: "Object Oriented Programming",
                code: "DAT100",
                tag: "Spring",
                year: 2017,
                provider: "github",
                directoryid: 23650610,

            },
            {
                id: 1,
                name: "Algorithms and Datastructures",
                code: "DAT200",
                tag: "Spring",
                year: 2017,
                provider: "github",
                directoryid: 23650611,
            },
            {
                id: 2,
                name: "Databases",
                code: "DAT220",
                tag: "Spring",
                year: 2017,
                provider: "github",
                directoryid: 23650612,
            },
            {
                id: 3,
                name: "Communication Technology",
                code: "DAT230",
                tag: "Spring",
                year: 2017,
                provider: "github",
                directoryid: 23650613,
            },
            {
                id: 4,
                name: "Operating Systems",
                code: "DAT320",
                tag: "Spring",
                year: 2017,
                provider: "github",
                directoryid: 23650614,
            },
        ] as ICourse[], (ele) => ele.id);
    }

    private addLocalCourseStudent() {
        this.localCourseStudent = [
            { courseId: 0, userid: 999, state: 2 },
            { courseId: 1, userid: 999, state: 2 },
            { courseId: 0, userid: 1, state: 0 },
            { courseId: 0, userid: 2, state: 0 },
        ] as ICourseUserLink[];
    }

    private addLocalLabInfo() {
        this.localLabInfo = mapify<ISubmission>([
            {
                id: 1,
                assignmentid: 0,
                userid: 999,
                buildId: 1,

                buildDate: new Date(2017, 6, 4),
                buildLog: "Build log for build 1",
                executetionTime: 1,
                score: 75,

                failedTests: 2,
                passedTests: 6,
                testCases: [
                    { name: "Test 1", score: 2, points: 2, weight: 20 },
                    { name: "Test 2", score: 1, points: 3, weight: 40 },
                    { name: "Test 3", score: 3, points: 3, weight: 40 },
                ],
            },
            {
                id: 2,
                assignmentid: 1,
                userid: 999,
                buildId: 2,

                buildDate: new Date(2017, 6, 4),
                buildLog: "Build log for build 2",
                executetionTime: 1,
                score: 75,

                failedTests: 2,
                passedTests: 6,
                testCases: [
                    { name: "Test 1", score: 2, points: 2, weight: 20 },
                    { name: "Test 2", score: 1, points: 3, weight: 40 },
                    { name: "Test 3", score: 3, points: 3, weight: 40 },
                ],
            },
            {
                id: 3,
                assignmentid: 2,
                userid: 999,
                buildId: 3,

                buildDate: new Date(2017, 6, 4),
                buildLog: "Build log for build 3",
                executetionTime: 1,
                score: 75,

                failedTests: 2,
                passedTests: 6,
                testCases: [
                    { name: "Test 1", score: 2, points: 2, weight: 20 },
                    { name: "Test 2", score: 1, points: 3, weight: 40 },
                    { name: "Test 3", score: 3, points: 3, weight: 40 },
                ],
            },
            {
                id: 4,
                assignmentid: 3,
                userid: 999,
                buildId: 4,

                buildDate: new Date(2017, 6, 4),
                buildLog: "Build log for build 4",
                executetionTime: 1,
                score: 75,

                failedTests: 2,
                passedTests: 6,
                testCases: [
                    { name: "Test 1", score: 2, points: 2, weight: 20 },
                    { name: "Test 2", score: 1, points: 3, weight: 40 },
                    { name: "Test 3", score: 3, points: 3, weight: 40 },
                ],
            },
            {
                id: 5,
                assignmentid: 4,
                userid: 999,
                buildId: 5,

                buildDate: new Date(2017, 6, 4),
                buildLog: "Build log for build 5",
                executetionTime: 1,
                score: 75,

                failedTests: 2,
                passedTests: 6,
                testCases: [
                    { name: "Test 1", score: 2, points: 2, weight: 20 },
                    { name: "Test 2", score: 1, points: 3, weight: 40 },
                    { name: "Test 3", score: 3, points: 3, weight: 40 },
                ],
            },
        ] as ISubmission[], (ele: ISubmission) => {
            return ele.id;
        });
    }

    private getLocalDirectories(): IOrganization[] {
        return (
            [
                {
                    id: 23650610,
                    path: "dat520-2017",
                    avatar: "https://avatars2.githubusercontent.com/u/23650610?v=3",
                },
            ]
        );
    }

    private addLocalCourseGroups(): void {
        this.localCourseGroups = [
            {
                id: 1,
                name: "Group1",
                status: CourseGroupStatus.approved,
                courseid: 1,
                users: [
                    {
                        id: 1,
                        email: "test@example.com",
                        firstname: "Student",
                        lastname: "1",
                        isadmin: false,
                        studentnr: "12345",
                    },
                    {
                        id: 2,
                        email: "test2@example.com",
                        firstname: "Student",
                        lastname: "2",
                        isadmin: false,
                        studentnr: "12346",
                    },
                ],
            },
            {
                id: 2,
                name: "Group2",
                status: CourseGroupStatus.pending,
                courseid: 1,
                users: [
                    {
                        id: 3,
                        email: "tes3t@example.com",
                        firstname: "Student",
                        lastname: "3",
                        isadmin: false,
                        studentnr: "12347",
                    },
                    {
                        id: 4,
                        email: "test4@example.com",
                        firstname: "Student",
                        lastname: "4",
                        isadmin: false,
                        studentnr: "12348",
                    },
                ],
            },
        ];
    }

}

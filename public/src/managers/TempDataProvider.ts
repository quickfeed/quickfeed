import { Assignment, Course, Enrollment, Group, Organization, Status, User } from "../../proto/ag_pb";
import {
    INewGroup,
    ISubmission,
} from "../models";

import { ICourseEnrollment, IUserEnrollment } from "../managers";
import { ICourseProvider } from "./CourseManager";
import { IUserProvider } from "./UserManager";

import { Timestamp } from "google-protobuf/google/protobuf/timestamp_pb";
import { isNull } from "util";
import { IMap, MapHelper, mapify } from "../map";

interface IGrpcDummyUser {
    user: User;
    password: string;
}

/**
 * The TempDataProvider is a fake implemetation of the backend
 * to be able to simulate the backend for easier developtment
 */
export class TempDataProvider implements IUserProvider, ICourseProvider {

    private localUsers: IMap<IGrpcDummyUser>;
    private localAssignments: IMap<Assignment>;
    private localCourses: IMap<Course>;
    private localCourseStudent: Enrollment[];
    private localLabInfo: IMap<ISubmission>;
    private localCourseGroups: Group[];

    private currentLoggedIn: User | null = null;

    constructor() {
        this.addLocalAssignments();
        this.addLocalCourses();
        this.addLocalCourseStudent();
        this.addLocalUsers();
        this.addLocalLabInfo();
        this.addLocalCourseGroups();
    }

    public async approveSubmission(submissionid: number, courseID: number): Promise<void> {
        throw new Error("Method not implemented.");
    }

    public async getOrganizations(provider: string): Promise<Organization[]> {
        throw new Error("Not implemented");
    }

    public async getAllUser(): Promise<User[]> {
        const users: User[] = [];
        const dummyUsers = MapHelper.toArray(this.localUsers);
        dummyUsers.forEach((ele) => {
            users.push(ele.user);
        });
        return users;
    }

    public async getCourses(): Promise<Course[]> {
        // return this.localCourses;
        return MapHelper.toArray(this.localCourses);
    }

    public async getCoursesStudent(): Promise<Enrollment[]> {
        return this.localCourseStudent;
    }

    public async getAssignments(courseId: number): Promise<Assignment[]> {
        const temp: Assignment[] = [];
        MapHelper.forEach(this.localAssignments, (a, i) => {
            if (a.getCourseid() === courseId) {
                temp[i] = a;
            }
        });
        return temp;
    }

    public async tryLogin(username: string, password: string): Promise<User | null> {
        const user = MapHelper.find(this.localUsers, (u) =>
            u.user.getEmail().toLocaleLowerCase() === username.toLocaleLowerCase());
        if (user && user.password === password) {
            this.currentLoggedIn = user.user;
            return user.user;
        }
        return null;
    }

    public async tryRemoteLogin(provider: string): Promise<User | null> {
        let lookup = "test@testersen.no";
        if (provider === "gitlab") {
            lookup = "bob@bobsen.no";
        }
        const user = MapHelper.find(this.localUsers, (u) =>
            u.user.getEmail().toLocaleLowerCase() === lookup);

        return new Promise<User | null>((resolve, reject) => {
            // Simulate async callback
            setTimeout(() => {
                if (isNull(user)) {
                    this.currentLoggedIn = user;
                    resolve(user);
                } else {
                    this.currentLoggedIn = user.user;
                    resolve(user.user);
                }

            }, 500);
        });
    }

    public async isAuthorizedTeacher(): Promise<boolean> {
        return true;
    }

    public async logout(user: User): Promise<boolean> {
        return true;
    }

    public async addUserToCourse(user: User, course: Course): Promise<boolean> {
        const tempEnrollment: Enrollment = new Enrollment();
        tempEnrollment.setCourseid(course.getId());
        tempEnrollment.setUserid(user.getId());
        tempEnrollment.setStatus(Enrollment.UserStatus.PENDING);
        this.localCourseStudent.push(tempEnrollment);
        return true;
    }

    /**
     * Get all userlinks for a single course
     * @param course The course userlinks should be retrived from
     * @param state Optinal. The state of the relation, all if not present
     */
    public async getUserLinksForCourse(course: Course, state?: Enrollment.UserStatus[]): Promise<Enrollment[]> {
        const users: Enrollment[] = [];
        for (const c of await this.getCoursesStudent()) {
            if (course.getId() === c.getCourseid()
             && (state === undefined || c.getStatus() === Enrollment.UserStatus.STUDENT)) {
                users.push(c);
            }
        }
        return users;
    }

    public async getUsersAsMap(ids: number[]): Promise<IMap<User>> {
        const returnUsers: IMap<User> = {};
        const allUsers = await this.getAllUser();
        ids.forEach((ele) => {
            const temp = allUsers[ele];
            if (temp) {
                returnUsers[ele] = temp;
            }
        });
        return returnUsers;
    }

    public async getUsersForCourse(course: Course, noGroupMembers?: boolean, state?: Enrollment.UserStatus[])
        : Promise<IUserEnrollment[]> {
        const courseStds: Enrollment[] =
            await this.getUserLinksForCourse(course, state);
        const users = await this.getUsersAsMap(courseStds.map((e) => e.getUserid()));
        return courseStds.map<IUserEnrollment>((link) => {
            const user = users[link.getUserid()];
            if (!user) {
                // TODO: See if we should have an error here or not
                throw new Error("Link exist witout a user object");
            }
            return { courseid: link.getCourseid(), userid: link.getUserid(), user, status: link.getStatus() };
        });
    }

    public async createNewCourse(course: any): Promise<Course | Status> {
        throw new Error("Method not implemented");
    }

    public async getCourse(id: number): Promise<Course | null> {
        const course: Course | undefined = this.localCourses[id];
        if (course) {
            return course;
        }
        return null;
    }

    public async updateCourse(courseId: number, courseData: Course): Promise<Status> {
        throw new Error("Method not implemented");
    }

    public async changeUserState(link: Enrollment, state: Enrollment.UserStatus): Promise<boolean> {
        link.setStatus(state);
        return true;
    }

    public async changeAdminRole(user: User): Promise<boolean> {
        user.setIsadmin(!user.getIsadmin());
        return true;
    }

    public async getAllLabInfos(courseId: number): Promise<ISubmission[]> {
        const temp: ISubmission[] = [];
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

    public async grpcGetLoggedInUser(): Promise<User | null> {
        return this.currentLoggedIn;
    }

    public async getLoggedInUser(): Promise<User | null> {
        return this.currentLoggedIn;
    }

    public async getCoursesFor(user: User, state?: Enrollment.UserStatus[]): Promise<ICourseEnrollment[]> {
        const cLinks: Enrollment[] = [];
        const temp = await this.getCoursesStudent();
        for (const c of temp) {
            if (user.getId() === c.getUserid()
             && (state === undefined || c.getStatus() === Enrollment.UserStatus.STUDENT)) {
                cLinks.push(c);
            }
        }
        const courses: ICourseEnrollment[] = [];
        const tempCourses = await this.getCourses();
        for (const link of cLinks) {
            const c = tempCourses[link.getCourseid()];
            if (c) {
                courses.push({
                    course: c, courseid: link.getCourseid(), userid: link.getUserid(), status: link.getStatus() 
                });
            }
        }
        return courses;
    }
    public async createGroup(groupData: INewGroup, courseId: number): Promise<Group | Status> {
        throw new Error("Method not implemented");
    }
    public async getCourseGroups(courseId: number): Promise<Group[]> {
        return this.localCourseGroups;
    }

    public async deleteGroup(groupId: number): Promise<boolean> {
        throw new Error("Method not implemented");
    }

    public async getGroupByUserAndCourse(userid: number, courseid: number): Promise<Group | null> {
        throw new Error("Method not implemented");
    }

    public async updateGroupStatus(groupId: number, status: Group.GroupStatus): Promise<boolean> {
        throw new Error("Method not implemented");
    }
    public async getGroup(gid: number): Promise<Group | null> {
        throw new Error("Method not implemented");
    }
    public async updateGroup(groupData: Group): Promise<Status> {
        throw new Error("Method not implemented");
    }
    public async getAllGroupLabInfos(courseId: number, groupID: number): Promise<ISubmission[]> {
        throw new Error("Method not implemented.");
    }

    public async updateAssignments(courseid: number): Promise<any> {
        return new Promise((resolve, reject) => {
            setTimeout(() => {
                resolve({});
            }, 10);
        });
    }

    public async updateUser(user: User): Promise<boolean> {
        const tempUser = this.localUsers[user.getId()];
        if (tempUser) {
            tempUser.user.setName(user.getName());
            tempUser.user.setEmail(user.getEmail());
            tempUser.user.setStudentid(user.getStudentid());
            tempUser.user.setIsadmin(user.getIsadmin());
        }
        return Promise.resolve(true);
    }

    public async getRepositoryURL(cid: number, type: number): Promise<string> {
        throw new Error("Method not implemented.");
    }

    private addLocalUsers() {
        const dummyUsers: IGrpcDummyUser[] = [];
        const tempUser: User = new User();
        tempUser.setId(999);
        tempUser.setName("Test Testersen");
        tempUser.setEmail("test@testersen.no");
        tempUser.setStudentid("9999");
        tempUser.setIsadmin(true);
        const tempDummy: IGrpcDummyUser = { user: tempUser, password: "1234" };
        dummyUsers.push(tempDummy);

        tempUser.setId(1000);
        tempUser.setName("Admin Admin");
        tempUser.setEmail("admin@admin");
        tempUser.setStudentid("1000");
        tempUser.setIsadmin(true);
        const tempDummy1: IGrpcDummyUser = { user: tempUser, password: "1234" };
        dummyUsers.push(tempDummy1);

        tempUser.setId(1);
        tempUser.setName("Per Pettersen");
        tempUser.setEmail("per@pettersen.no");
        tempUser.setStudentid("1234");
        tempUser.setIsadmin(true);
        const tempDummy2: IGrpcDummyUser = { user: tempUser, password: "1234" };
        dummyUsers.push(tempDummy2);

        tempUser.setId(2);
        tempUser.setName("Bob Bobsen");
        tempUser.setEmail("bob@bobsen.no");
        tempUser.setStudentid("1234");
        tempUser.setIsadmin(true);
        const tempDummy3: IGrpcDummyUser = { user: tempUser, password: "1234" };
        dummyUsers.push(tempDummy3);

        tempUser.setId(3);
        tempUser.setName("Petter Pan");
        tempUser.setEmail("petter@pan.no");
        tempUser.setStudentid("1234");
        tempUser.setIsadmin(true);
        const tempDummy4: IGrpcDummyUser = { user: tempUser, password: "1234" };
        dummyUsers.push(tempDummy4);
        this.localUsers = mapify(dummyUsers, (ele) => ele.user.getId());
    }

    private addLocalAssignments() {
        const ts = new Timestamp();
        ts.fromDate(new Date(2017, 5, 25));
        const a0 = new Assignment();
        const a1 = new Assignment();
        const a2 = new Assignment();
        const a3 = new Assignment();
        const a4 = new Assignment();
        const a5 = new Assignment();
        const a6 = new Assignment();
        const a7 = new Assignment();
        const a8 = new Assignment();
        const a9 = new Assignment();
        const a10 = new Assignment();

        a0.setId(0);
        a0.setCourseid(0);
        a0.setName("Lab 1");
        a0.setLanguage("Go");
        a0.setDeadline(ts);

        a1.setId(1);
        a1.setCourseid(0);
        a1.setName("Lab 2");
        a1.setLanguage("Go");
        a1.setDeadline(ts);

        a2.setId(2);
        a2.setCourseid(0);
        a2.setName("Lab 3");
        a2.setLanguage("Go");
        a2.setDeadline(ts);

        a3.setId(3);
        a3.setCourseid(0);
        a3.setName("Lab 4");
        a3.setLanguage("Go");
        a3.setDeadline(ts);

        a4.setId(4);
        a4.setCourseid(1);
        a4.setName("Lab 1");
        a4.setLanguage("Go");
        a4.setDeadline(ts);

        a5.setId(5);
        a5.setCourseid(1);
        a5.setName("Lab 2");
        a5.setLanguage("Go");
        a5.setDeadline(ts);

        a6.setId(6);
        a6.setCourseid(1);
        a6.setName("Lab 3");
        a6.setLanguage("Go");
        a6.setDeadline(ts);

        a7.setId(7);
        a7.setCourseid(2);
        a7.setName("Lab 1");
        a7.setLanguage("TypeScript");
        a7.setDeadline(ts);

        a8.setId(8);
        a8.setCourseid(2);
        a8.setName("Lab 2");
        a8.setLanguage("Go");
        a8.setDeadline(ts);

        a9.setId(9);
        a9.setCourseid(3);
        a9.setName("Lab 1");
        a9.setLanguage("Go");
        a9.setDeadline(ts);

        a10.setId(10);
        a10.setCourseid(4);
        a10.setName("Lab 1");
        a10.setLanguage("TypeScript");
        a10.setDeadline(ts);

        const tempAssignments: Assignment[] = [a0, a1, a2, a3, a4, a5, a6, a7, a8, a9, a10];
        this.localAssignments = mapify(tempAssignments, (ele) => ele.getId());
    }

    private addLocalCourses() {
        const course0 = new Course();
        const course1 = new Course();
        const course2 = new Course();
        const course3 = new Course();
        const course4 = new Course();

        course0.setId(0);
        course0.setName("Object Oriented Programming");
        course0.setCode("DAT100");
        course0.setTag("Spring");
        course0.setYear(2017);
        course0.setProvider("github");
        course0.setOrganizationid(23650610);

        course1.setId(1);
        course1.setName("Algorithms and Datastructures");
        course1.setCode("DAT200");
        course1.setTag("Spring");
        course1.setYear(2017);
        course1.setProvider("github");
        course1.setOrganizationid(23650611);

        course2.setId(2);
        course2.setName("Databases");
        course2.setCode("DAT220");
        course2.setTag("Spring");
        course2.setYear(2017);
        course2.setProvider("github");
        course2.setOrganizationid(23650612);

        course3.setId(3);
        course3.setName("Communication Technology");
        course3.setCode("DAT230");
        course3.setTag("Spring");
        course3.setYear(2017);
        course3.setProvider("github");
        course3.setOrganizationid(23650613);

        course4.setId(4);
        course4.setName("Operating Systems");
        course4.setCode("DAT320");
        course4.setTag("Spring");
        course4.setYear(2017);
        course4.setProvider("github");
        course4.setOrganizationid(23650614);

        const tempCourses: Course[] = [course0, course1, course2, course3, course4];
        this.localCourses = mapify(tempCourses, (ele) => ele.getId());
    }

    private addLocalCourseStudent() {
        const localEnrols: Enrollment[] = [];
        const tempEnrol: Enrollment = new Enrollment();
        tempEnrol.setCourseid(0);
        tempEnrol.setUserid(999);
        tempEnrol.setStatus(2);
        localEnrols.push(tempEnrol);
        tempEnrol.setCourseid(1);
        localEnrols.push(tempEnrol);
        tempEnrol.setCourseid(0);
        tempEnrol.setUserid(1);
        tempEnrol.setStatus(0);
        localEnrols.push(tempEnrol);
        tempEnrol.setUserid(2);
        localEnrols.push(tempEnrol);
        this.localCourseStudent = localEnrols;
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

    private getLocalOrgs(): Organization[] {
        const localOrgs: Organization[] = [];
        const localOrg = new Organization();
        localOrg.setId(23650610);
        localOrg.setPath("dat520-2017");
        localOrg.setAvatar("https://avatars2.githubusercontent.com/u/23650610?v=3");
        localOrgs.push(localOrg);
        return localOrgs;
    }

    private addLocalCourseGroups(): void {
        const user1 = new User();
        user1.setId(1);
        user1.setEmail("test@example.com");
        user1.setName("Student 1");
        user1.setStudentid("12345");

        const user2 = new User();
        user2.setId(2);
        user2.setEmail("test2@example.com");
        user2.setName("Student 2");
        user2.setStudentid("12346");

        const group1 = new Group();
        group1.setId(1);
        group1.setName("Group 1");
        group1.setStatus(Group.GroupStatus.APPROVED);
        group1.setCourseid(1);
        group1.setUsersList([user1, user2]);

        const user3 = new User();
        user3.setId(3);
        user3.setEmail("test3@example.com");
        user3.setName("Student 3");
        user3.setStudentid("12347");

        const user4 = new User();
        user4.setId(4);
        user4.setEmail("test4@example.com");
        user4.setName("Student 4");
        user4.setStudentid("12348");

        const group2 = new Group();
        group2.setId(2);
        group2.setName("Group 2");
        group2.setStatus(Group.GroupStatus.PENDING);
        group2.setCourseid(1);
        group2.setUsersList([user3, user4]);
        this.localCourseGroups = [group1, group2];
    }
}

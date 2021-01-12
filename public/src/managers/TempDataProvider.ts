import {
    Assignment,
    Course,
    Enrollment,
    GradingBenchmark,
    GradingCriterion,
    Group,
    Organization,
    Repository,
    Status,
    User,
    Review,
    SubmissionsForCourseRequest
} from '../../proto/ag_pb';
import { IAllSubmissionsForEnrollment, ISubmission } from '../models';

import { ICourseProvider } from "./CourseManager";
import { IUserProvider } from "./UserManager";

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

    public async updateSubmission(courseID: number, submission: ISubmission): Promise<boolean> {
        throw new Error("Method not implemented.");
    }

    public async getUsers(): Promise<User[]> {
        const users: User[] = [];
        const dummyUsers = MapHelper.toArray(this.localUsers);
        dummyUsers.forEach((ele) => {
            users.push(ele.user);
        });
        return users;
    }

    public async getUser(): Promise<User> {
        if (this.currentLoggedIn) {
            return this.currentLoggedIn;
        }
        return new User();
    }
    public async getCourses(): Promise<Course[]> {
        // return this.localCourses;
        return MapHelper.toArray(this.localCourses);
    }

    public async getCoursesStudent(): Promise<Enrollment[]> {
        return this.localCourseStudent;
    }

    public async getAssignments(courseID: number): Promise<Assignment[]> {
        const temp: Assignment[] = [];
        MapHelper.forEach(this.localAssignments, (a, i) => {
            if (a.getCourseid() === courseID) {
                temp[i] = a;
            }
        });
        return temp;
    }

    public async tryLogin(username: string, password: string): Promise<User | null> {
        const user = MapHelper.find(this.localUsers, (u) =>
            u.user.getEmail().toLocaleLowerCase() === username.toLocaleLowerCase());
        if (user?.password === password) {
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
                if (!(user)) {
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

    public async addUserToCourse(course: Course, user: User): Promise<boolean> {
        const tempEnrollment: Enrollment = new Enrollment();
        tempEnrollment.setCourseid(course.getId());
        tempEnrollment.setUserid(user.getId());
        tempEnrollment.setStatus(Enrollment.UserStatus.PENDING);
        this.localCourseStudent.push(tempEnrollment);
        return true;
    }

    public async approveAll(courseID: number): Promise<boolean> {
        return true;
    }

    public async updateSubmissions(assignmentID: number, courseID: number, score: number, release: boolean, approve: boolean): Promise<boolean> {
        return true;
    }

    /**
     * Get all userlinks for a single course
     * @param course The course userlinks should be retrived from
     * @param status Optinal. The status of the relation, all if not present
     */
    public async getUserLinksForCourse(course: Course, status?: Enrollment.UserStatus[]): Promise<Enrollment[]> {
        const users: Enrollment[] = [];
        for (const c of await this.getCoursesStudent()) {
            if (course.getId() === c.getCourseid()
             && (status === undefined || c.getStatus() === Enrollment.UserStatus.STUDENT)) {
                users.push(c);
            }
        }
        return users;
    }

    public async getUsersAsMap(IDs: number[]): Promise<IMap<User>> {
        const returnUsers: IMap<User> = {};
        const allUsers = await this.getUsers();
        IDs.forEach((ele) => {
            const temp = allUsers[ele];
            if (temp) {
                returnUsers[ele] = temp;
            }
        });
        return returnUsers;
    }

    public async getUsersForCourse(course: Course, noGroupMembers?: boolean, withActivity?: boolean, status?: Enrollment.UserStatus[])
        : Promise<Enrollment[]> {
        const courseStds: Enrollment[] =
            await this.getUserLinksForCourse(course, status);
        const users = await this.getUsersAsMap(courseStds.map((e) => e.getUserid()));
        return courseStds.map<Enrollment>((link) => {
            const user = users[link.getUserid()];
            if (!user) {
                // TODO: See if we should have an error here or not
                throw new Error("Link exist witout a user object");
            }
            return link;
        });
    }

    public async createNewCourse(course: any): Promise<Course | Status> {
        throw new Error("Method not implemented");
    }

    public async getCourse(ID: number): Promise<Course | null> {
        const course: Course | undefined = this.localCourses[ID];
        if (course) {
            return course;
        }
        return null;
    }

    public async updateCourse(course: Course): Promise<Status> {
        throw new Error("Method not implemented");
    }

    public async updateCourseVisibility(enrol: Enrollment): Promise<boolean> {
        return true;
    }

    public async changeUserStatus(link: Enrollment, status: Enrollment.UserStatus): Promise<Status> {
        link.setStatus(status);
        const stat = new Status();
        stat.setCode(0);
        return stat;
    }

    public async changeAdminRole(user: User): Promise<boolean> {
        user.setIsadmin(!user.getIsadmin());
        return true;
    }

    public async getSubmissionsByUser(courseID: number): Promise<ISubmission[]> {
        const temp: ISubmission[] = [];
        const assignments = await this.getAssignments(courseID);
        MapHelper.forEach(this.localLabInfo, (ele) => {
            if (assignments[ele.assignmentid]) {
                temp[ele.id] = ele;
            }
        });
        return temp;
    }

    public async getReviewers(submissionID: number, courseID: number): Promise<User[]> {
        return [];
    }

    public async getEnrollmentsForUser(userID: number): Promise<Enrollment[]> {
        return [];
    }

    public async getSubmissionsByCourse(courseID: number, type: SubmissionsForCourseRequest.Type): Promise<IAllSubmissionsForEnrollment[]> {
        return [];
    }

    public async getProviders(): Promise<string[]> {
        return ["github"];
    }

    public async getLoggedInUser(): Promise<User | null> {
        return this.currentLoggedIn;
    }

    public async getCoursesForUser(user: User, status: Enrollment.UserStatus[]): Promise<Course[]> {
        const cLinks: Enrollment[] = [];
        const temp = await this.getCoursesStudent();
        for (const c of temp) {
            if (user.getId() === c.getUserid()
             && (status === undefined || c.getStatus() === Enrollment.UserStatus.STUDENT)) {
                cLinks.push(c);
            }
        }
        const courses: Enrollment[] = [];
        return this.getCourses();
    }
    public async createGroup(courseID: number, name: string, users: number[]): Promise<Group | Status> {
        throw new Error("Method not implemented");
    }
    public async getGroupsForCourse(courseID: number): Promise<Group[]> {
        return this.localCourseGroups;
    }

    public async deleteGroup(courseID: number, groupID: number): Promise<Status> {
        throw new Error("Method not implemented");
    }

    public async getGroupByUserAndCourse(courseID: number, userID: number): Promise<Group | null> {
        throw new Error("Method not implemented");
    }

    public async getOrganization(orgName: string): Promise<Organization | Status > {
        throw new Error("Method not implemented");
    }

    public async updateGroupStatus(groupID: number, status: Group.GroupStatus): Promise<Status> {
        throw new Error("Method not implemented");
    }
    public async getGroup(groupID: number): Promise<Group | null> {
        throw new Error("Method not implemented");
    }
    public async updateGroup(group: Group): Promise<Status> {
        throw new Error("Method not implemented");
    }
    public async getSubmissionsByGroup(courseID: number, groupID: number): Promise<ISubmission[]> {
        throw new Error("Method not implemented.");
    }

    public async isEmptyRepo(courseID: number, userID: number, groupID: number): Promise<boolean> {
        throw new Error("Method not implemented.");
    }

    public async addNewBenchmark(bm: GradingBenchmark): Promise<GradingBenchmark | null> {
        return bm;
    }

    public async addNewCriterion(c: GradingCriterion): Promise<GradingCriterion | null> {
        return c;
    }

    public async updateBenchmark(bm: GradingBenchmark): Promise<boolean> {
        return true;
    }

    public async updateCriterion(c: GradingCriterion): Promise<boolean> {
        return true;
    }

    public async deleteBenchmark(bm: GradingBenchmark): Promise<boolean> {
        return true;
    }
    public async deleteCriterion(c: GradingCriterion): Promise<boolean> {
        return true;
    }

    public async loadCriteria(assignmentID: number, courseID: number): Promise<GradingBenchmark[]> {
        return [];
    }
    public async addReview(r: Review): Promise<Review | null> {
        return r;
    }

    public async editReview(r: Review): Promise<boolean> {
        return true;
    }

    public async updateAssignments(courseID: number): Promise<any> {
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

    public async getRepositories(courseID: number, types: Repository.Type[]): Promise<Map<Repository.Type, string>> {
        throw new Error("Method not implemented");
    }

    public async rebuildSubmission(assignmentID: number, submissionID: number): Promise<ISubmission | null> {
        throw new Error("Method not implemented");
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
        const ts = new Date(2017, 5, 25);
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
        a0.setScriptfile("Go");
        a0.setDeadline(ts.toDateString());

        a1.setId(1);
        a1.setCourseid(0);
        a1.setName("Lab 2");
        a1.setScriptfile("Go");
        a1.setDeadline(ts.toDateString());

        a2.setId(2);
        a2.setCourseid(0);
        a2.setName("Lab 3");
        a2.setScriptfile("Go");
        a2.setDeadline(ts.toDateString());

        a3.setId(3);
        a3.setCourseid(0);
        a3.setName("Lab 4");
        a3.setScriptfile("Go");
        a3.setDeadline(ts.toDateString());

        a4.setId(4);
        a4.setCourseid(1);
        a4.setName("Lab 1");
        a4.setScriptfile("Go");
        a4.setDeadline(ts.toDateString());

        a5.setId(5);
        a5.setCourseid(1);
        a5.setName("Lab 2");
        a5.setScriptfile("Go");
        a5.setDeadline(ts.toDateString());

        a6.setId(6);
        a6.setCourseid(1);
        a6.setName("Lab 3");
        a6.setScriptfile("Go");
        a6.setDeadline(ts.toDateString());

        a7.setId(7);
        a7.setCourseid(2);
        a7.setName("Lab 1");
        a7.setScriptfile("TypeScript");
        a7.setDeadline(ts.toDateString());

        a8.setId(8);
        a8.setCourseid(2);
        a8.setName("Lab 2");
        a8.setScriptfile("Go");
        a8.setDeadline(ts.toDateString());

        a9.setId(9);
        a9.setCourseid(3);
        a9.setName("Lab 1");
        a9.setScriptfile("Go");
        a9.setDeadline(ts.toDateString());

        a10.setId(10);
        a10.setCourseid(4);
        a10.setName("Lab 1");
        a10.setScriptfile("TypeScript");
        a10.setDeadline(ts.toDateString());

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
                executionTime: 1,
                score: 75,

                failedTests: 2,
                passedTests: 6,
                testCases: [
                    { TestName: "Test 1", Score: 2, MaxScore: 2, Weight: 20 },
                    { TestName: "Test 2", Score: 1, MaxScore: 3, Weight: 40 },
                    { TestName: "Test 3", Score: 3, MaxScore: 3, Weight: 40 },
                ],
            },
            {
                id: 2,
                assignmentid: 1,
                userid: 999,
                buildId: 2,

                buildDate: new Date(2017, 6, 4),
                buildLog: "Build log for build 2",
                executionTime: 1,
                score: 75,

                failedTests: 2,
                passedTests: 6,
                testCases: [
                    { TestName: "Test 1", Score: 2, MaxScore: 2, Weight: 20 },
                    { TestName: "Test 2", Score: 1, MaxScore: 3, Weight: 40 },
                    { TestName: "Test 3", Score: 3, MaxScore: 3, Weight: 40 },
                ],
            },
            {
                id: 3,
                assignmentid: 2,
                userid: 999,
                buildId: 3,

                buildDate: new Date(2017, 6, 4),
                buildLog: "Build log for build 3",
                executionTime: 1,
                score: 75,

                failedTests: 2,
                passedTests: 6,
                testCases: [
                    { TestName: "Test 1", Score: 2, MaxScore: 2, Weight: 20 },
                    { TestName: "Test 2", Score: 1, MaxScore: 3, Weight: 40 },
                    { TestName: "Test 3", Score: 3, MaxScore: 3, Weight: 40 },
                ],
            },
            {
                id: 4,
                assignmentid: 3,
                userid: 999,
                buildId: 4,

                buildDate: new Date(2017, 6, 4),
                buildLog: "Build log for build 4",
                executionTime: 1,
                score: 75,

                failedTests: 2,
                passedTests: 6,
                testCases: [
                    { TestName: "Test 1", Score: 2, MaxScore: 2, Weight: 20 },
                    { TestName: "Test 2", Score: 1, MaxScore: 3, Weight: 40 },
                    { TestName: "Test 3", Score: 3, MaxScore: 3, Weight: 40 },
                ],
            },
            {
                id: 5,
                assignmentid: 4,
                userid: 999,
                buildId: 5,

                buildDate: new Date(2017, 6, 4),
                buildLog: "Build log for build 5",
                executionTime: 1,
                score: 75,

                failedTests: 2,
                passedTests: 6,
                testCases: [
                    { TestName: "Test 1", Score: 2, MaxScore: 2, Weight: 20 },
                    { TestName: "Test 2", Score: 1, MaxScore: 3, Weight: 40 },
                    { TestName: "Test 3", Score: 3, MaxScore: 3, Weight: 40 },
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

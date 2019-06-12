import {Assignment, 
        Course, 
        Enrollment, 
        User, 
        Group, 
        Submission, 
        Directory,
        Void} from "../../proto/ag_pb";
import {
    IAssignment,
    IBuildInfo,
    ICourseUserLink,
    IError,
    INewGroup,
    IStatusCode,
    ISubmission,
    ITestCases,
    IUser,
} from "../models";

import { HttpHelper, IHTTPResult } from "../HttpHelper";
import {GrpcManager} from "./GRPCManager"
import { ICourseProvider } from "./CourseManager";

import HttpStatusCode from "../HttpStatusCode";
import {
    ICourseEnrollment,
    IEnrollment,
    isCourseEnrollment,
    IUserEnrollment,
    IUserProvider,
} from "../managers";
import { IMap, mapify } from "../map";
import { ILogger } from "./LogManager";

interface IEndpoints {
    courses: string;
    users: string;
    user: string;
    group: string;
    groups: string;
    refresh: string;
    submissions: string;
    submission: string;
    assignments: string;
    directories: string;
    providers: string;
    auth: string;
    logout: string;
    api: string;
}

const URL_ENDPOINT: IEndpoints = {
    courses: "courses",
    users: "users",
    user: "user",
    group: "group",
    groups: "groups",
    refresh: "refresh",
    submissions: "submissions",
    submission: "submission",
    assignments: "assignments",
    directories: "directories",
    providers: "providers",
    auth: "auth",
    logout: "logout",
    api: "api/v1",
};

export class ServerProvider implements IUserProvider, ICourseProvider {

    private helper: HttpHelper;
    private grpcHelper: GrpcManager;
    private logger: ILogger;

    constructor(helper: HttpHelper, grpcHelper: GrpcManager, logger: ILogger) {
        this.helper = helper;
        this.grpcHelper = grpcHelper;
        this.logger = logger;
    }

    public async getCourses(): Promise<Course[]> {
        const result = await this.grpcHelper.getCourses();
        if (result.statusCode !== 0 || !result.data) {
            this.handleError(result, "getCourses");
            return [];
        }
        const data = result.data.getCoursesList().map((course) => {
            return course;
        });
        return data;
    }


    public async getCoursesFor(user: User, state?: Enrollment.UserStatus[]): Promise<ICourseEnrollment[]> {
        const result = await this.grpcHelper.getCoursesWithEnrollment(user.getId(), state);
        if (result.statusCode !== 0 || !result.data) {
            this.handleError(result, "getCoursesFor");
            return [];
        }
        const arr: ICourseEnrollment[] = [];
        result.data.getCoursesList().forEach((ele) => {
            arr.push({
                course: ele,
                status: ele.getEnrolled(),
                courseid: ele.getId(),
                userid: user.getId(),
                user,
            });
        });
        return arr;
    }


    public async getUsersForCourse(course: Course, noGroupMembers?: boolean, state?: Enrollment.UserStatus[]): Promise<IUserEnrollment[]> {
        const result = await this.grpcHelper.getEnrollmentsByCourse(course.getId(), noGroupMembers, state);
        if (result.statusCode !== 0 || !result.data) {
            this.handleError(result, "getUserForCourse");
            return [];
        }

        const arr: IUserEnrollment[] = [];
        result.data.getEnrollmentsList().forEach((ele) => {
            const enroll: IEnrollment = this.toIEnrollment(ele);
            if (isCourseEnrollment(enroll)) {
                enroll.user = this.makeUserInfo(enroll.user);
                arr.push(enroll);
            }
        });
        return arr;
    }


    public async getAssignments(courseId: number): Promise<IMap<IAssignment>> {
        const result = await this.grpcHelper.getAssignments(courseId);
        if (result.statusCode !== 0 || !result.data) {
            this.handleError(result, "getAssignments");
            throw new Error("Problem with the request");
        }
        const assignments: IAssignment[] = [];
        result.data.getAssignmentsList().forEach((ele) => {
            const asg = this.toIAssignment(ele);
            assignments.push(asg);
        });

        return mapify(assignments, (ele) => {
            ele.deadline = !ele.deadline ? new Date(2000, 1, 1) : new Date(ele.deadline);
            return ele.id;
        });
    }


    public async addUserToCourse(user: User, course: Course): Promise<boolean> {
        const result = await this.grpcHelper.createEnrollment(user.getId(), course.getId());
        if (result.statusCode === 0) {
            return true;
        } else {
            this.handleError(result, "addUserToCourse");
        }
        return false;
    }

    public async changeUserState(link: ICourseUserLink, state: Enrollment.UserStatus): Promise<boolean> {
        const result = await this.grpcHelper.updateEnrollment(link.userid, link.courseId, state);
        if (result.statusCode === 0) {
            return true;
        } else {
            this.handleError(result, "changeUserState");
        }
        return false;
    }



    public async createNewCourse(courseData: Course): Promise<Course | IError> {
        const result = await this.grpcHelper.createCourse(courseData);
        if (result.statusCode !== 0 || !result.data) {
            this.handleError(result, "createNewCourse");
            return this.parseError(result);
        }
        return result.data;
    }


    public async getCourse(id: number): Promise<Course | null> {
        const result = await this.grpcHelper.getCourse(id);
        if (result.statusCode !== 0 || !result.data) {
            this.handleError(result, "getCourse");
            return null;
        }
        return result.data;
    }


    public async updateCourse(courseId: number, courseData: Course): Promise<Void | IError> {
        const result = await this.grpcHelper.updateCourse(courseData);
        if (result.statusCode !== 0 || !result.data) {
            this.handleError(result, "updateCourse");
            return this.parseError(result);
        }
        const voidy = new Void();
        return voidy;
     }

    public async createGroup(groupData: INewGroup, courseID: number): Promise<Group | IError> {
        const result = await this.grpcHelper.createGroup(groupData, courseID);
        
        if (result.statusCode !== 0 || !result.data) {
            this.handleError(result, "createGroup");
            return this.parseError(result);
        }
        return result.data;
    }


    public async getCourseGroups(courseID: number): Promise<Group[]> {
        const result = await this.grpcHelper.getGroups(courseID);
        if (result.statusCode !== 0 || !result.data) {
            this.handleError(result, "getCourseGroups");
            throw new Error("Problem with the request");
        }
        return result.data.getGroupsList();
    }

    public async getGroupByUserAndCourse(userID: number, courseID: number): Promise<Group | null> {
        const result = await this.grpcHelper.getGroupByUserAndCourse(userID, courseID);
        if (result.statusCode !== 0 || !result.data) {
            return null;
        }
        return result.data;
    }


    public async updateGroupStatus(groupID: number, st: Group.GroupStatus): Promise<boolean> {
        const result = await this.grpcHelper.updateGroupStatus(groupID, st);
        if (result.statusCode === 0)  {
            return true;
        } else {
            this.handleError(result, "updateGroupStatus");
        }
        return false;
    }


    public async getGroup(groupID: number): Promise<Group | null> {
        const result = await this.grpcHelper.getGroup(groupID);
        if (result.statusCode !== 0 || !result.data) {
            this.handleError(result, "getGroup");
            return null;
        }
        return result.data;
    }


    public async deleteGroup(groupID: number): Promise<boolean> {
        const result = await this.grpcHelper.deleteGroup(groupID);
         if (result.statusCode === 0) {
            return true;
        } else {
            this.handleError(result, "deleteGroup");
        }
        return false;
    }


    public async updateGroup(group: Group): Promise<IStatusCode | IError> {     //(groupData: INewGroup, groupID: number, courseID: number): Promise<IStatusCode | IError> {
        const result = await this.grpcHelper.updateGroup(group);
        if (result.statusCode !== 0 || !result.data) {
            this.handleError(result, "getGroup");
            return this.parseError(result);
        }
        const code: IStatusCode = {statusCode: result.statusCode};
        return code;
    }

    public async getAllGroupLabInfos(courseID: number, groupID: number): Promise<IMap<ISubmission>> {
        const result = await this.grpcHelper.getGroupSubmissions(courseID, groupID);
        if (result.statusCode === 0 && result.data === undefined) {
            return {};
        }
        if (!result.data) {
            this.handleError(result, "getAllLabInfos");
            return {};
        }

        const isubmissions: ISubmission[] = [];
        result.data.getSubmissionsList().forEach((ele) => {
            const isbm = this.toISUbmission(ele);
            isubmissions.push(isbm);
        });
        return mapify(isubmissions, (ele) => {
            return ele.id;
        });

    }

    public async getAllLabInfos(courseID: number, userID: number): Promise<IMap<ISubmission>> {
        const result = await this.grpcHelper.getSubmissions(courseID, userID);
        if (result.statusCode === HttpStatusCode.OK && result.data === undefined) {
            return {};
        }
        if (!result.data) {
            this.handleError(result, "getAllLabInfos");
            return {};
        }
        const isubmissions: ISubmission[] = [];
        result.data.getSubmissionsList().forEach((ele) => {
            const isbm = this.toISUbmission(ele);
            isubmissions.push(isbm);
        });
        return mapify(isubmissions, (ele) => {
            return ele.id;
        });
    }


    public async tryLogin(username: string, password: string): Promise<User | null> {
        throw new Error("tryLogin This could be removed since there is no normal login.");
    }

    public async logout(user: User): Promise<boolean> {
        window.location.assign("/" + URL_ENDPOINT.logout);
        return true;
    }

    public async getAllUser(): Promise<User[]> {
        const result = await this.grpcHelper.getUsers();
        if (result.statusCode !== 0 || !result.data) {
            this.handleError(result, "getAllUser");
            return [];
        }

        const arr = result.data.getUsersList().map<User>((ele) => this.makeUserInfo(ele));
        return arr;
    }

    public async tryRemoteLogin(provider: string): Promise<User | null> {
        if (provider.length > 0) {
            const requestString = "/" + URL_ENDPOINT.auth + "/" + provider;
            window.location.assign(requestString);
        }
        return null;
    }

    public async changeAdminRole(user: User): Promise<boolean> {
        const result = await this.grpcHelper.updateUser(user, true);
        if (result.statusCode !== 0) {
            return false;
        } else {
            this.handleError(result, "changeAdminRole");
        }
        return true;
    }


    public async updateUser(user: User): Promise<boolean> {
        const result = await this.grpcHelper.updateUser(user);
        if (result.statusCode !== 0) {
            this.handleError(result, "updateUser");
        }
        return true;
    }

    public async getDirectories(provider: string): Promise<Directory[]> {
        const result = await this.grpcHelper.getDirectories(provider);
        if (result.statusCode !== 0 || !result.data) {
            this.handleError(result, "getDirectories");
        } else {
            return result.data.getDirectoriesList();
        }
        return[];
    }

    public async getProviders(): Promise<string[]> {
        const result = await this.grpcHelper.getProviders();
        if (result.statusCode !== 0 || !result.data) {
            this.handleError(result, "getProviders");
        } else {
            return result.data.getProvidersList();
        }
        return [];
    }

    public async getLoggedInUser(): Promise<User | null> {
      
        const result = await this.helper.get<IUser>(URL_ENDPOINT.user);
        if (result.statusCode !== HttpStatusCode.FOUND || !result.data) {
            this.handleError(result, "getLoggedInUser");
            return null;
        }
        const iusr = result.data;

        // We want a user with full information provided to be set as currentUser 
        // Such user is retrieved by GRPC method getUser
        const grpcResult = await this.grpcHelper.getUser(iusr.id);

        if (grpcResult.statusCode !== 0 || !grpcResult.data) {
            this.handleError(result, "getLoggedInUser");
            const usr: User = new User();
            usr.setId(iusr.id);
            usr.setStudentid(iusr.studentid);
            usr.setName(iusr.name);
            usr.setEmail(iusr.email);
            usr.setAvatarurl(iusr.avatarurl);
            usr.setIsadmin(iusr.isadmin);
            return this.makeUserInfo(usr);
        }
        return this.makeUserInfo(grpcResult.data);
    }


    public async refreshCoursesFor(courseID: number): Promise<any> {
        const result = await this.grpcHelper.refreshCourse(courseID);
        if (result.statusCode !== 0 || !result.data) {
            this.handleError(result, "refreshCoursesFor");
            return null;
        }
        return result.data
    }

    public async getRepositoryURL(courseID: number, repoType: number): Promise<string> {
        const result = await this.grpcHelper.getRepositoryURL(courseID, repoType);
        if (result.statusCode !== 0 || !result.data) {
            //this.handleError(result, "getRepositoryURL");
            return "";
        }
        return result.data.getUrl();
    }


    public async approveSubmission(submissionID: number): Promise<void> {
        const result = await this.grpcHelper.updateSubmission(submissionID);
        if (result.statusCode !== HttpStatusCode.OK) {
            this.handleError(result, "approveSubmission");
            return;
        }
        return;
    }


    private makeUserInfo(data: User): User {
        return data;
    }

    private handleError(result: IHTTPResult<any>, sender?: string): void {
        this.logger.warn("Request to server failed with status code: " + result.statusCode, true);
        if (sender) {
            this.logger.warn("Failed request from: " + sender);
        }
    }

    private parseError(result: IHTTPResult<any>): IError {
        const error: IError = {
            statusCode: result.statusCode,
        };
        if (result.data) {
            error.data = JSON.parse(JSON.stringify(result.data)) as any;
        }
        return error;
    }

    private buildURL(uri: string[]): string {
        let url = "";
        let counter = 0;
        for (const tag of uri) {
            url += tag;
            url += counter < uri.length - 1 ? "/" : "";
            ++counter;
        }
        return url;
    }

    private toISUbmission(sbm: Submission): ISubmission {      
        let buildInfoAsString = "";
        let scoreInfoAsString = "";
        if (sbm.getBuildinfo() && (sbm.getBuildinfo().trim().length > 2)) {
            buildInfoAsString = sbm.getBuildinfo();
        }
        if (sbm.getScoreobjects() && (sbm.getScoreobjects().trim().length > 2)) {
            scoreInfoAsString = sbm.getScoreobjects();
        }
        let buildInfo: IBuildInfo;
        let scoreObj: ITestCases[];
        try {
            buildInfo = JSON.parse(buildInfoAsString);
        } catch (e) {
            buildInfo = JSON.parse(
                "{\"builddate\": \"2017-07-28\", \"buildid\": 1, \"buildlog\": \"This is cool\", \"execTime\": 1}",
            );
        }
        try {
            scoreObj = JSON.parse(scoreInfoAsString);
        } catch (e) {
            scoreObj = JSON.parse(
                "[{\"name\": \"Test 1\", \"score\": 3, \"points\": 4, \"weight\": 100}]",
            );
        }

        let failed = 0;
        let passed = 0;
        scoreObj.forEach((ele) => {
            if (ele.points !== ele.score) {
                failed++;
            } else {
                passed++;
            }
        });
        const isbm: ISubmission = {
            id: sbm.getId(),
            userid: sbm.getUserid(),
            groupid: sbm.getGroupid(),
            assignmentid: sbm.getAssignmentid(),
            passedTests: passed,
            failedTests: failed,
            score: sbm.getScore(),
            buildId: buildInfo.buildid,
            buildDate: buildInfo.builddate,
            executetionTime: buildInfo.execTime,
            buildLog: buildInfo.buildlog,
            testCases: scoreObj,
            approved: sbm.getApproved()
        };

        return isbm;


    }

    // this method convert a grpc Assignment to IAssignment
    private toIAssignment(assg: Assignment): IAssignment {
        let deadline = assg.getDeadline();
        let date: Date = new Date();
        if (deadline) {
            //HACK: check the correctnes of date conversion
            let date = new Date(deadline.getSeconds());
        }
        const iassgn: IAssignment = {
            id: assg.getId(),
            name: assg.getName(),
            courseid: assg.getCourseid(),
            deadline: date,
            language: assg.getLanguage(),
            isgrouplab: assg.getIsgrouplab()
        };
        return  iassgn;
    }

    // this method convert a grpc Enrollment to IEnrollment
    private toIEnrollment(enrollment: Enrollment): IEnrollment {
        const ienroll: IEnrollment =  {
            userid: enrollment.getUserid(),
            courseid: enrollment.getCourseid(),
        };
        if (enrollment.getStatus() !== undefined) {
            ienroll.status = enrollment.getStatus().valueOf();
        }

        const user: User | undefined = enrollment.getUser();
        if (user !== undefined) {
            ienroll.user = user;
        }
        ienroll.course = enrollment.getCourse();
        

        return ienroll as IEnrollment;
    }


}

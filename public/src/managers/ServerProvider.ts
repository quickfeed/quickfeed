import {
    Assignment,
    Course,
    Enrollment,
    Group,
    Organization,
    Status,
    Submission,
    User,
    Void,
} from "../../proto/ag_pb";
import {
    IBuildInfo,
    INewGroup,
    ISubmission,
    ITestCases,
    IUser,
} from "../models";

import { HttpHelper } from "../HttpHelper";
import { ICourseProvider } from "./CourseManager";
import { GrpcManager } from "./GRPCManager";

import HttpStatusCode from "../HttpStatusCode";
import {
    IUserProvider,
} from "../managers";
import { ILogger } from "./LogManager";

interface IEndpoints {
    user: string;
    auth: string;
    logout: string;
}

const URL_ENDPOINT: IEndpoints = {
    user: "user",
    auth: "auth",
    logout: "logout",
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
        if (result.status.getCode() !== 0 || !result.data) {
            return [];
        }
        return result.data.getCoursesList();
    }

    public async getCoursesFor(user: User, state?: Enrollment.UserStatus[]): Promise<Enrollment[]> {
        const result = await this.grpcHelper.getCoursesWithEnrollment(user.getId(), state);
        if (result.status.getCode() !== 0 || !result.data) {
            return [];
        }
        const arr: Enrollment[] = [];
        result.data.getCoursesList().forEach((ele) => {
            const enr: Enrollment = new Enrollment();
            enr.setCourse(ele);
            enr.setCourseid(ele.getId());
            enr.setUser(user);
            enr.setUserid(user.getId());
            enr.setStatus(ele.getEnrolled());
            arr.push(enr);
        });
        return arr;
    }

    public async getUsersForCourse(
        course: Course, noGroupMembers?: boolean,
        state?: Enrollment.UserStatus[]): Promise<Enrollment[]> {

        const result = await this.grpcHelper.getEnrollmentsByCourse(course.getId(), noGroupMembers, state);
        if (result.status.getCode() !== 0 || !result.data) {
            return [];
        }
        return result.data.getEnrollmentsList();
    }


    public async getEnrollment(courseID: number, userID?: number, groupID?: number): Promise<Enrollment | null> {
        const result = await this.grpcHelper.getEnrollment(courseID, userID, groupID);
        if (result.status.getCode() !== 0 || !result.data) {
            return null;
        }
        return result.data;
    }

    public async getAssignments(courseId: number): Promise<Assignment[]> {
        const result = await this.grpcHelper.getAssignments(courseId);
        if (result.status.getCode() !== 0 || !result.data) {
            return [];
        }
        return result.data.getAssignmentsList();
    }

    public async addUserToCourse(user: User, course: Course): Promise<boolean> {
        const result = await this.grpcHelper.createEnrollment(user.getId(), course.getId());
        return result.status.getCode() === 0;
    }

    public async changeUserState(link: Enrollment, state: Enrollment.UserStatus): Promise<boolean> {
        const result = await this.grpcHelper.updateEnrollment(link.getUserid(), link.getCourseid(), state);
        return result.status.getCode() === 0;
    }

    public async isAuthorizedTeacher(): Promise<boolean> {
        const result = await this.grpcHelper.isAuthorizedTeacher();
        if (result.status.getCode() !== 0 || !result.data) {
            return false;
        }
        return result.data.getIsauthorized();
    }

    public async isTeacher(user: User, courseID: number): Promise<boolean> {
        const result = await this.grpcHelper.getEnrollment(courseID, user.getId());
        if (result.status.getCode() !== 0 || !result.data) {
            return false;
        }
        return result.data.getStatus() === Enrollment.UserStatus.TEACHER;
    }

    public async createNewCourse(courseData: Course): Promise<Course | Status> {
        const result = await this.grpcHelper.createCourse(courseData);
        if (result.status.getCode() !== 0 || !result.data) {
            return result.status;
        }
        return result.data;
    }

    public async getCourse(id: number): Promise<Course | null> {
        const result = await this.grpcHelper.getCourse(id);
        if (result.status.getCode() !== 0 || !result.data) {
            return null;
        }
        return result.data;
    }

    public async updateCourse(courseId: number, courseData: Course): Promise<Void | Status> {
        const result = await this.grpcHelper.updateCourse(courseData);
        if (result.status.getCode() !== 0 || !result.data) {
            return result.status;
        }
        return new Void();
    }

    public async createGroup(groupData: INewGroup, courseID: number): Promise<Group | Status> {
        const result = await this.grpcHelper.createGroup(groupData, courseID);
        if (result.status.getCode() !== 0 || !result.data) {
            return result.status;
        }
        return result.data;
    }

    public async getCourseGroups(courseID: number): Promise<Group[]> {
        const result = await this.grpcHelper.getGroups(courseID);
        if (result.status.getCode() !== 0 || !result.data) {
            return [];
        }
        return result.data.getGroupsList();
    }

    public async getGroupByUserAndCourse(userID: number, courseID: number): Promise<Group | null> {
        const result = await this.grpcHelper.getGroupByUserAndCourse(userID, courseID);
        if (result.status.getCode() !== 0 || !result.data) {
            return null;
        }
        return result.data;
    }

    public async updateGroupStatus(groupID: number, st: Group.GroupStatus): Promise<boolean> {
        const result = await this.grpcHelper.updateGroupStatus(groupID, st);
        if (result.status.getCode() !== 0) {
            return false;
        }
        return true;
    }

    public async getGroup(groupID: number): Promise<Group | null> {
        const result = await this.grpcHelper.getGroup(groupID);
        if (result.status.getCode() !== 0 || !result.data) {
            return null;
        }
        return result.data;
    }

    public async deleteGroup(groupID: number): Promise<boolean> {
        const result = await this.grpcHelper.deleteGroup(groupID);
        return result.status.getCode() === 0;
    }

    public async updateGroup(group: Group): Promise<Status> {
        const result = await this.grpcHelper.updateGroup(group);
        return result.status;
    }

    public async getAllGroupLabInfos(courseID: number, groupID: number): Promise<ISubmission[]> {
        const result = await this.grpcHelper.getGroupSubmissions(courseID, groupID);
        if (result.status.getCode() !== 0 || !result.data) {
            return [];
        }

        const isubmissions: ISubmission[] = [];
        result.data.getSubmissionsList().forEach((ele) => {
            const isbm = this.toISUbmission(ele);
            isubmissions.push(isbm);
        });
        return isubmissions;
    }

    public async getAllLabInfos(courseID: number, userID: number): Promise<ISubmission[]> {
        const result = await this.grpcHelper.getSubmissions(courseID, userID);
        if (result.status.getCode() !== 0 || !result.data) {
            return [];
        }
        const isubmissions: ISubmission[] = [];
        result.data.getSubmissionsList().forEach((ele) => {
            const isbm = this.toISUbmission(ele);
            isubmissions.push(isbm);
        });
        return isubmissions;
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
        if (result.status.getCode() !== 0 || !result.data) {
            return [];
        }
        return result.data.getUsersList();
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
        // we are not interested in user data returned in this case, only checking that there were no errors
        return result.status.getCode() === 0;
    }

    public async updateUser(user: User): Promise<boolean> {
        const result = await this.grpcHelper.updateUser(user);
        if (result.status.getCode() !== 0 || !result.data) {
            return false;
        }
        return true;
    }

    public async getOrganizations(provider: string): Promise<Organization[]> {
        const result = await this.grpcHelper.getOrganizations(provider);
        if (result.status.getCode() !== 0 || !result.data) {
            return [];
        }
        return result.data.getOrganizationsList();
    }

    public async getProviders(): Promise<string[]> {
        const result = await this.grpcHelper.getProviders();
        if (result.status.getCode() !== 0 || !result.data) {
            return [];
        }
        return result.data.getProvidersList();
    }

    public async getLoggedInUser(): Promise<User | null> {
        const result = await this.helper.get<IUser>(URL_ENDPOINT.user);
        if (result.statusCode !== HttpStatusCode.FOUND || !result.data) {
            console.log("REST request to get logged in user resulted in error: " + result.statusCode);
            return null;
        }
        const iusr = result.data;
        const usr: User = new User();
        usr.setId(iusr.id);
        usr.setStudentid(iusr.studentid);
        usr.setName(iusr.name);
        usr.setEmail(iusr.email);
        usr.setAvatarurl(iusr.avatarurl);
        usr.setIsadmin(iusr.isadmin);
        return usr;
    }

    public async updateAssignments(courseID: number): Promise<boolean> {
        const result = await this.grpcHelper.updateAssignments(courseID);
        return result.status.getCode() === 0;
    }

    public async getRepositoryURL(courseID: number, repoType: number): Promise<string> {
        const result = await this.grpcHelper.getRepositoryURL(courseID, repoType);
        if (result.status.getCode() !== 0 || !result.data) {
            return "";
        }
        return result.data.getUrl();
    }

    public async approveSubmission(submissionID: number, courseID: number): Promise<void> {
        await this.grpcHelper.approveSubmission(submissionID, courseID);
        return;
    }

    private toISUbmission(sbm: Submission): ISubmission {
        let buildInfoAsString = "";
        let scoreInfoAsString = "";
        if (sbm.getBuildinfo() && (sbm.getBuildinfo().trim().length > 2)) {
            buildInfoAsString = sbm.getBuildinfo();
        }
        console.log("toISubmission: build: " + sbm.getBuildinfo());
        console.log("toISubmission: parsed scores: " + buildInfoAsString);
        if (sbm.getScoreobjects() && (sbm.getScoreobjects().trim().length > 2)) {
            scoreInfoAsString = sbm.getScoreobjects();
        }
        console.log("toISubmission: scores: " + sbm.getScoreobjects());
        console.log("toISubmission: parsed scores: " + scoreInfoAsString);
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
        if (scoreObj) {
            scoreObj.forEach((ele) => {
                if (ele.points !== ele.score) {
                    failed++;
                } else {
                    passed++;
                }
            });
        }
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
            approved: sbm.getApproved(),
        };
        console.log("toISubmission: constructed submission " + isbm);
        return isbm;
    }
}

import {
    Assignment,
    Course,
    Enrollment,
    Group,
    Organization,
    Submission,
    User,
    Void,
} from "../../proto/ag_pb";
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

import { HttpHelper } from "../HttpHelper";
import { ICourseProvider } from "./CourseManager";
import { GrpcManager, IGrpcResponse } from "./GRPCManager";

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
        if (result.statusCode !== 0 || !result.data) {
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

    public async getUsersForCourse(
        course: Course, noGroupMembers?: boolean,
        state?: Enrollment.UserStatus[]): Promise<IUserEnrollment[]> {

        const result = await this.grpcHelper.getEnrollmentsByCourse(course.getId(), noGroupMembers, state);
        if (result.statusCode !== 0 || !result.data) {
            return [];
        }

        const arr: IUserEnrollment[] = [];
        result.data.getEnrollmentsList().forEach((ele) => {
            // TODO(meling) this conversion seems unnecessary.
            const enroll: IEnrollment = this.toIEnrollment(ele);
            // TODO(meling) this should be unnecessary to check since we get the enrollment from the backend.
            if (isCourseEnrollment(enroll)) {
                arr.push(enroll);
            }
        });
        return arr;
    }

    public async getAssignments(courseId: number): Promise<IMap<IAssignment>> {
        const result = await this.grpcHelper.getAssignments(courseId);
        if (result.statusCode !== 0 || !result.data) {
            return [];
        }
        const assignments: IAssignment[] = [];
        result.data.getAssignmentsList().forEach((ele) => {
            const asg = this.toIAssignment(ele);
            assignments.push(asg);
        });

        return mapify(assignments, (ele) => {
            return ele.id;
        });
    }

    public async addUserToCourse(user: User, course: Course): Promise<boolean> {
        const result = await this.grpcHelper.createEnrollment(user.getId(), course.getId());
        if (result.statusCode !== 0) {
            return false;
        }
        return true;
    }

    public async changeUserState(link: ICourseUserLink, state: Enrollment.UserStatus): Promise<boolean> {
        const result = await this.grpcHelper.updateEnrollment(link.userid, link.courseId, state);

        if (result.statusCode !== 0) {
            return false;
        }
        return true;
    }

    public async isAuthorizedTeacher(): Promise<boolean> {
        const result = await this.grpcHelper.isAuthorizedTeacher();
        if (result.statusCode !== 0) {
            return false;
        }
        return true;
    }

    public async createNewCourse(courseData: Course): Promise<Course | IError> {
        const result = await this.grpcHelper.createCourse(courseData);
        if (result.statusCode !== 0 || !result.data) {
            return this.parseError(result);
        }
        return result.data;
    }

    public async getCourse(id: number): Promise<Course | null> {
        const result = await this.grpcHelper.getCourse(id);
        if (result.statusCode !== 0 || !result.data) {
            return null;
        }
        return result.data;
    }

    public async updateCourse(courseId: number, courseData: Course): Promise<Void | IError> {
        const result = await this.grpcHelper.updateCourse(courseData);
        if (result.statusCode !== 0 || !result.data) {
            return this.parseError(result);
        }
        return new Void();
    }

    public async createGroup(groupData: INewGroup, courseID: number): Promise<Group | IError> {
        const result = await this.grpcHelper.createGroup(groupData, courseID);
        if (result.statusCode !== 0 || !result.data) {
            return this.parseError(result);
        }
        return result.data;
    }

    public async getCourseGroups(courseID: number): Promise<Group[]> {
        const result = await this.grpcHelper.getGroups(courseID);
        if (result.statusCode !== 0 || !result.data) {
            return [];
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
        if (result.statusCode !== 0) {
            return false;
        }
        return true;
    }

    public async getGroup(groupID: number): Promise<Group | null> {
        const result = await this.grpcHelper.getGroup(groupID);
        if (result.statusCode !== 0 || !result.data) {
            return null;
        }
        return result.data;
    }

    public async deleteGroup(groupID: number): Promise<boolean> {
        const result = await this.grpcHelper.deleteGroup(groupID);
        if (result.statusCode !== 0) {
            return false;
        }
        return true;
    }

    public async updateGroup(group: Group): Promise<IStatusCode | IError> {
        const result = await this.grpcHelper.updateGroup(group);
        if (result.statusCode !== 0 || !result.data) {
            return this.parseError(result);
        }
        const code: IStatusCode = { statusCode: result.statusCode };
        return code;
    }

    public async getAllGroupLabInfos(courseID: number, groupID: number): Promise<IMap<ISubmission>> {
        const result = await this.grpcHelper.getGroupSubmissions(courseID, groupID);
        if (result.statusCode !== 0 || !result.data) {
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
        if (result.statusCode !== 0 || !result.data) {
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
        if (result.statusCode !== 0 || !result.data) {
            return false;
        }
        return true;
    }

    public async updateUser(user: User): Promise<boolean> {
        const result = await this.grpcHelper.updateUser(user);
        if (result.statusCode !== 0 || !result.data) {
            return false;
        }
        return true;
    }

    public async getOrganizations(provider: string): Promise<Organization[]> {
        const result = await this.grpcHelper.getOrganizations(provider);
        if (result.statusCode !== 0 || !result.data) {
            return [];
        }
        return result.data.getOrganizationsList();
    }

    public async getProviders(): Promise<string[]> {
        const result = await this.grpcHelper.getProviders();
        if (result.statusCode !== 0 || !result.data) {
            return [];
        }
        return result.data.getProvidersList();
    }

    public async getLoggedInUser(): Promise<User | null> {
        const result = await this.helper.get<IUser>(URL_ENDPOINT.user);
        if (result.statusCode !== HttpStatusCode.FOUND || !result.data) {
            return null;
        }
        const iusr = result.data;

        // We want a user with full information provided to be set as currentUser
        // Such user is retrieved by GRPC method getUser
        const grpcResult = await this.grpcHelper.getUser(iusr.id);
        // If grpc method fails, construct User object from JSON
        if (grpcResult.statusCode !== 0 || !grpcResult.data) {
            const usr: User = new User();
            usr.setId(iusr.id);
            usr.setStudentid(iusr.studentid);
            usr.setName(iusr.name);
            usr.setEmail(iusr.email);
            usr.setAvatarurl(iusr.avatarurl);
            usr.setIsadmin(iusr.isadmin);
            return usr;
        }
        return grpcResult.data;
    }

    public async refreshCoursesFor(courseID: number): Promise<any> {
        const result = await this.grpcHelper.refreshCourse(courseID);
        if (result.statusCode !== 0 || !result.data) {
            return null;
        }
        return result.data.getAssignmentsList();
    }

    public async getRepositoryURL(courseID: number, repoType: number): Promise<string> {
        const result = await this.grpcHelper.getRepositoryURL(courseID, repoType);
        if (result.statusCode !== 0 || !result.data) {
            return "";
        }
        return result.data.getUrl();
    }

    public async approveSubmission(submissionID: number): Promise<void> {
        await this.grpcHelper.updateSubmission(submissionID);
        return;
    }

    private parseError(result: IGrpcResponse<any>): IError {
        const error: IError = {
            statusCode: result.statusCode,
        };
        if (result.message) {
            error.message = result.message;
        }
        return error;
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
            approved: sbm.getApproved(),
        };
        return isbm;
    }

    // this method convert a grpc Assignment to IAssignment
    private toIAssignment(assg: Assignment): IAssignment {
        const deadline = assg.getDeadline();
        let date: Date = new Date();
        if (deadline) {
            // HACK: check the correctness of date conversion
            date = new Date(deadline.getSeconds());
        }
        const iassgn: IAssignment = {
            id: assg.getId(),
            name: assg.getName(),
            courseid: assg.getCourseid(),
            deadline: date,
            language: assg.getLanguage(),
            isgrouplab: assg.getIsgrouplab(),
        };
        return iassgn;
    }

    // this method convert a grpc Enrollment to IEnrollment
    private toIEnrollment(enrollment: Enrollment): IEnrollment {
        const ienroll: IEnrollment = {
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

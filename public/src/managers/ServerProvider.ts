import {
    Assignment,
    Course,
    Enrollment,
    Group,
    Organization,
    Repository,
    Status,
    Submission,
    User,
    Void,
} from "../../proto/ag_pb";
import {
    IAssignmentLink,
    IBuildInfo,
    IStudentSubmission,
    ISubmission,
    ITestCases,
    IUser,
} from "../models";

import { HttpHelper } from "../HttpHelper";
import { ICourseProvider } from "./CourseManager";
import { GrpcManager } from "./GRPCManager";

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

    public async getAssignments(courseID: number): Promise<Assignment[]> {
        const result = await this.grpcHelper.getAssignments(courseID);
        if (result.status.getCode() !== 0 || !result.data) {
            return [];
        }
        return result.data.getAssignmentsList();
    }

    public async addUserToCourse(course: Course, user: User): Promise<boolean> {
        const result = await this.grpcHelper.createEnrollment(course.getId(), user.getId());
        return result.status.getCode() === 0;
    }

    public async changeUserState(link: Enrollment, state: Enrollment.UserStatus): Promise<boolean> {
        const result = await this.grpcHelper.updateEnrollment(link.getCourseid(), link.getUserid(), state);
        return result.status.getCode() === 0;
    }

    public async approveAll(courseID: number): Promise<boolean> {
        const result = await this.grpcHelper.updateEnrollments(courseID);
        if (result.status.getCode() !== 0 || !result.data) {
            return false;
        }
        return true;
    }

    public async isAuthorizedTeacher(): Promise<boolean> {
        const result = await this.grpcHelper.isAuthorizedTeacher();
        if (result.status.getCode() !== 0 || !result.data) {
            return false;
        }
        return result.data.getIsauthorized();
    }

    public async createNewCourse(course: Course): Promise<Course | Status> {
        const result = await this.grpcHelper.createCourse(course);
        if (result.status.getCode() !== 0 || !result.data) {
            return result.status;
        }
        return result.data;
    }

    public async getCourse(ID: number): Promise<Course | null> {
        const result = await this.grpcHelper.getCourse(ID);
        if (result.status.getCode() !== 0 || !result.data) {
            return null;
        }
        return result.data;
    }

    public async updateCourse(course: Course): Promise<Void | Status> {
        const result = await this.grpcHelper.updateCourse(course);
        if (result.status.getCode() !== 0 || !result.data) {
            return result.status;
        }
        return new Void();
    }

    public async createGroup(courseID: number, name: string, users: number[]): Promise<Group | Status> {
        const result = await this.grpcHelper.createGroup(courseID, name, users);
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

    public async getGroupByUserAndCourse(courseID: number, userID: number): Promise<Group | null> {
        const result = await this.grpcHelper.getGroupByUserAndCourse(courseID, userID);
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

    public async deleteGroup(courseID: number, groupID: number): Promise<boolean> {
        const result = await this.grpcHelper.deleteGroup(courseID, groupID);
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
            const isbm = this.toISubmission(ele);
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
            const isbm = this.toISubmission(ele);
            isubmissions.push(isbm);
        });
        return isubmissions;
    }

    public async getCourseLabs(courseID: number): Promise<IAssignmentLink[]> {
        const result = await this.grpcHelper.getCourseLabSubmissions(courseID);
        if (result.status.getCode() !== 0 || !result.data) {
            return [];
        }

        const results = result.data.getLabsList();
        const labCourse = new Course();
        labCourse.setId(courseID);
        const labs: IAssignmentLink[] = [];
        for (const studentLabs of results) {
            const subs: IStudentSubmission[] = [];
            const allSubs = studentLabs.getSubmissionsList();
            if (allSubs) {
                for (const lab of allSubs) {
                    // populate student submissions
                    const labAssignment = new Assignment();
                    labAssignment.setId(lab.getAssignmentid());
                    const ILab: IStudentSubmission = {
                        assignment:  labAssignment,
                        latest: this.toISubmission(lab),
                        authorName: studentLabs.getAuthorname(),
                    };
                    subs.push(ILab);
                }
            }
            // populate assignment links
            let enrol = studentLabs.getEnrollment();
            if (!enrol) {
                enrol = new Enrollment();
            }
            const labLink: IAssignmentLink = {
                course: labCourse,
                link: enrol,
                assignments: subs,
            };
            labs.push(labLink);
        }
        return labs;
    }

    public async tryLogin(username: string, password: string): Promise<User | null> {
        throw new Error("tryLogin This could be removed since there is no normal login.");
    }

    public async logout(user: User): Promise<boolean> {
        window.location.assign("/" + URL_ENDPOINT.logout);
        return true;
    }

    public async getUser(): Promise<User> {
        const result = await this.grpcHelper.getUser();
        if (result.status.getCode() !== 0 || !result.data) {
            return new User();
        }
        return result.data;
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

    public async getOrganization(orgName: string): Promise<Organization | Status> {
        const result = await this.grpcHelper.getOrganization(orgName);
        if (result.status.getCode() !== 0 || !result.data) {
            return result.status;
        }
        return result.data;
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
        if (result.statusCode !== 302 || !result.data) {
            console.log("failed to get logged in user; status code: " + result.statusCode);
            return null;
        }
        const iusr = result.data;
        const usr = new User();
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

    public async getRepositories(cid: number, types: Repository.Type[]): Promise<Map<Repository.Type, string>> {
        const result = await this.grpcHelper.getRepositories(cid, types);
        const tsMap = new Map<Repository.Type, string>();
        if (result.status.getCode() !== 0 || !result.data) {
            return tsMap;
        }
        // protobuf and typescript maps have class method mismatch. we need to convert one into another here
        const tmp = result.data.getUrlsMap();
        tmp.forEach((v, k) => {
            tsMap.set((Repository.Type as any)[k], v);
        });

        return tsMap;
    }

    public async approveSubmission(courseID: number, submissionID: number): Promise<boolean> {
        const result = await this.grpcHelper.approveSubmission(courseID, submissionID);
        if (result.status.getCode() !== 0) {
            return false;
        }
        return true;
    }

    public async refreshSubmission(assignmentID: number, submissionID: number): Promise<boolean> {
        const result = await this.grpcHelper.refreshSubmission(assignmentID, submissionID);
        if (result.status.getCode() !== 0) {
            return false;
        }
        return true;
    }

    public async isEmptyRepo(courseID: number, userID: number, groupID: number): Promise<boolean> {
        const result = await this.grpcHelper.isEmptyRepo(courseID, userID, groupID);
        return result.status.getCode() === 0;
    }

    private toISubmission(sbm: Submission): ISubmission {
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
                "[{\"TestName\": \"Test 1\", \"Score\": 3, \"MaxScore\": 4, \"Weight\": 100}]",
            );
        }
        let failed = 0;
        let passed = 0;
        if (scoreObj) {
            scoreObj.forEach((ele) => {
                if (ele.MaxScore !== ele.Score) {
                    failed++;
                } else {
                    passed++;
                }
            });
        }
        const bDate = new Date(buildInfo.builddate);
        const isbm: ISubmission = {
            id: sbm.getId(),
            userid: sbm.getUserid(),
            groupid: sbm.getGroupid(),
            assignmentid: sbm.getAssignmentid(),
            passedTests: passed,
            failedTests: failed,
            score: sbm.getScore(),
            buildId: buildInfo.buildid,
            buildDate: bDate,
            executetionTime: buildInfo.execTime,
            buildLog: buildInfo.buildlog,
            testCases: scoreObj,
            approved: sbm.getApproved(),
        };
        return isbm;
    }
}

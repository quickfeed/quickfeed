import {
    Assignment,
    Benchmarks,
    Course,
    CourseSubmissions,
    Enrollment,
    GradingBenchmark,
    GradingCriterion,
    Group,
    Organization,
    Repository,
    Review,
    Status,
    Submission,
    SubmissionsForCourseRequest,
    User,
} from "../../proto/ag_pb";
import {
    IAllSubmissionsForEnrollment,
    IBuildInfo,
    ISubmissionLink,
    ISubmission,
    ITestCases,
    IUser,
} from "../models";

import { HttpHelper } from "../HttpHelper";
import { ICourseProvider } from "./CourseManager";
import { GrpcManager, IGrpcResponse } from "./GRPCManager";

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
        if (!this.responseCodeSuccess(result) || !result.data) {
            return [];
        }
        return result.data.getCoursesList();
    }

    public async getCoursesForUser(user: User, statuses: Enrollment.UserStatus[]): Promise<Course[]> {
        const result = await this.grpcHelper.getCoursesByUser(user.getId(), statuses);
        if (!this.responseCodeSuccess(result) || !result.data) {
            return [];
        }
        return result.data.getCoursesList();
    }

    public async getEnrollmentsForUser(userID: number, statuses: Enrollment.UserStatus[]): Promise<Enrollment[]> {
        const result = await this.grpcHelper.getEnrollmentsByUser(userID);
        if (!this.responseCodeSuccess(result) || !result.data) {
            return [];
        }
        return result.data.getEnrollmentsList();
    }

    public async getUsersForCourse(
        course: Course,
        withoutGroupMembers?: boolean,
        withActivity?: boolean,
        status?: Enrollment.UserStatus[]): Promise<Enrollment[]> {

        const result = await this.grpcHelper.getEnrollmentsByCourse(course.getId(), withoutGroupMembers, withActivity, status);
        if (!this.responseCodeSuccess(result) || !result.data) {
            return [];
        }
        return result.data.getEnrollmentsList();
    }

    public async getAssignments(courseID: number): Promise<Assignment[]> {
        const result = await this.grpcHelper.getAssignments(courseID);
        if (!this.responseCodeSuccess(result) || !result.data) {
            return [];
        }
        return result.data.getAssignmentsList();
    }

    public async addUserToCourse(course: Course, user: User): Promise<boolean> {
        const result = await this.grpcHelper.createEnrollment(course.getId(), user.getId());
        return this.responseCodeSuccess(result);
    }

    public async changeUserStatus(enrollment: Enrollment, status: Enrollment.UserStatus): Promise<Status> {
        const originalStatus = enrollment.getStatus();
        enrollment.setStatus(status);
        const result = await this.grpcHelper.updateEnrollment(enrollment);
        if (!this.responseCodeSuccess(result)) {
            enrollment.setStatus(originalStatus);
        }
        return result.status;
    }

    public async approveAll(courseID: number): Promise<boolean> {
        const result = await this.grpcHelper.updateEnrollments(courseID);
        return result.data ? this.responseCodeSuccess(result) : false;
    }

    public async isAuthorizedTeacher(): Promise<boolean> {
        const result = await this.grpcHelper.isAuthorizedTeacher();
        if (this.responseCodeSuccess(result) && result.data) {
            return result.data.getIsauthorized();
        }
        return false;

    }

    public async createNewCourse(course: Course): Promise<Course | Status> {
        const result = await this.grpcHelper.createCourse(course);
        if (!this.responseCodeSuccess(result) || !result.data) {
            return result.status;
        }
        return result.data;
    }

    public async getCourse(courseID: number): Promise<Course | null> {
        const result = await this.grpcHelper.getCourse(courseID);
        if (!this.responseCodeSuccess(result) || !result.data) {
            return null;
        }
        return result.data;
    }

    public async updateCourse(course: Course): Promise<Status> {
        const result = await this.grpcHelper.updateCourse(course);
        return result.status;
    }

    public async updateCourseVisibility(enrol: Enrollment): Promise<boolean> {
        const result = await this.grpcHelper.updateCourseVisibility(enrol);
        return this.responseCodeSuccess(result);
    }

    public async createGroup(courseID: number, groupName: string, users: number[]): Promise<Group | Status> {
        const result = await this.grpcHelper.createGroup(courseID, groupName, users);
        if (!this.responseCodeSuccess(result) || !result.data) {
            return result.status;
        }
        return result.data;
    }

    public async getGroupsForCourse(courseID: number): Promise<Group[]> {
        const result = await this.grpcHelper.getGroupsByCourse(courseID);
        if (!this.responseCodeSuccess(result) || !result.data) {
            return [];
        }
        return result.data.getGroupsList();
    }

    public async getGroupByUserAndCourse(courseID: number, userID: number): Promise<Group | null> {
        const result = await this.grpcHelper.getGroupByUserAndCourse(courseID, userID);
        if (!this.responseCodeSuccess(result) || !result.data) {
            return null;
        }
        return result.data;
    }

    public async updateGroupStatus(groupID: number, status: Group.GroupStatus): Promise<Status> {
        const result = await this.grpcHelper.updateGroupStatus(groupID, status);
        return result.status;
    }

    public async getGroup(groupID: number): Promise<Group | null> {
        const result = await this.grpcHelper.getGroup(groupID);
        if (!this.responseCodeSuccess(result) || !result.data) {
            return null;
        }
        return result.data;
    }

    public async deleteGroup(courseID: number, groupID: number): Promise<Status> {
        const result = await this.grpcHelper.deleteGroup(courseID, groupID);
        return result.status;
    }

    public async updateGroup(group: Group): Promise<Status> {
        const result = await this.grpcHelper.updateGroup(group);
        return result.status;
    }

    public async getSubmissionsByGroup(courseID: number, groupID: number): Promise<ISubmission[]> {
        const result = await this.grpcHelper.getGroupSubmissions(courseID, groupID);
        if (!this.responseCodeSuccess(result) || !result.data) {
            return [];
        }

        const isubmissions: ISubmission[] = [];
        result.data.getSubmissionsList().forEach((ele) => {
            const isbm = this.toISubmission(ele);
            isubmissions.push(isbm);
        });
        return isubmissions;
    }

    public async getSubmissionsByUser(courseID: number, userID: number): Promise<ISubmission[]> {
        const result = await this.grpcHelper.getSubmissions(courseID, userID);
        if (!this.responseCodeSuccess(result) || !result.data) {
            return [];
        }
        const isubmissions: ISubmission[] = [];
        result.data.getSubmissionsList().forEach((ele) => {
            const isbm = this.toISubmission(ele);
            isubmissions.push(isbm);
        });
        return isubmissions;
    }

    public async getSubmissionsByCourse(courseID: number, type: SubmissionsForCourseRequest.Type): Promise<IAllSubmissionsForEnrollment[]> {
        const result = await this.grpcHelper.getSubmissionsByCourse(courseID, type);
        if (!this.responseCodeSuccess(result) || !result.data) {
            return [];
        }
        return this.toUILinks(result.data);

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
        if (!this.responseCodeSuccess(result) || !result.data) {
            return new User();
        }
        return result.data;
    }

    public async getUsers(): Promise<User[]> {
        const result = await this.grpcHelper.getUsers();
        if (!this.responseCodeSuccess(result) || !result.data) {
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

    public async changeAdminRole(user: User, promote: boolean): Promise<boolean> {
        user.setIsadmin(promote);
        const result = await this.grpcHelper.updateUser(user);
        // we are not interested in user data returned in this case, only checking that there were no errors
        return this.responseCodeSuccess(result);
    }

    public async updateUser(user: User): Promise<boolean> {
        const result = await this.grpcHelper.updateUser(user);
        return result.data ? this.responseCodeSuccess(result) : false;
    }

    public async getOrganization(orgName: string): Promise<Organization | Status> {
        const result = await this.grpcHelper.getOrganization(orgName);
        if (!this.responseCodeSuccess(result) || !result.data) {
            return result.status;
        }
        return result.data;
    }

    public async getProviders(): Promise<string[]> {
        const result = await this.grpcHelper.getProviders();
        if (!this.responseCodeSuccess(result) || !result.data) {
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
        return this.responseCodeSuccess(result);
    }

    public async getRepositories(courseID: number, types: Repository.Type[]): Promise<Map<Repository.Type, string>> {
        const result = await this.grpcHelper.getRepositories(courseID, types);
        const tsMap = new Map<Repository.Type, string>();
        if (!this.responseCodeSuccess(result) || !result.data) {
            return tsMap;
        }
        // protobuf and typescript maps have class method mismatch. we need to convert one into another here
        const tmp = result.data.getUrlsMap();
        tmp.forEach((v, k) => {
            tsMap.set((Repository.Type as any)[k], v);
        });

        return tsMap;
    }

    public async updateSubmission(courseID: number, submission: ISubmission): Promise<boolean> {
        const result = await this.grpcHelper.updateSubmission(courseID, submission);
        return this.responseCodeSuccess(result);
    }

    public async rebuildSubmission(assignmentID: number, submissionID: number): Promise<ISubmission | null> {
        const result = await this.grpcHelper.rebuildSubmission(assignmentID, submissionID);
        if (!this.responseCodeSuccess(result) || !result.data) {
            return null;
        }
        return this.toISubmission(result.data);
    }

    public async isEmptyRepo(courseID: number, userID: number, groupID: number): Promise<boolean> {
        const result = await this.grpcHelper.isEmptyRepo(courseID, userID, groupID);
        return this.responseCodeSuccess(result);
    }

    public async addNewBenchmark(bm: GradingBenchmark): Promise<GradingBenchmark | null> {
        const result = await this.grpcHelper.createBenchmark(bm);
        if (!this.responseCodeSuccess(result) || !result.data) {
            return null;
        }
        return result.data;
    }

    public async addNewCriterion(c: GradingCriterion): Promise<GradingCriterion | null> {
        const result = await this.grpcHelper.createCriterion(c);
        if (!this.responseCodeSuccess(result) || !result.data) {
            return null;
        }
        return result.data;
    }

    public async updateBenchmark(bm: GradingBenchmark): Promise<boolean> {
        const result = await this.grpcHelper.updateBenchmark(bm);
        return this.responseCodeSuccess(result);
    }

    public async updateCriterion(c: GradingCriterion): Promise<boolean> {
        const result = await this.grpcHelper.updateCriterion(c);
        return this.responseCodeSuccess(result);
    }

    public async deleteBenchmark(bm: GradingBenchmark): Promise<boolean> {
        const result = await this.grpcHelper.deleteBenchmark(bm);
        return this.responseCodeSuccess(result);
    }
    public async deleteCriterion(c: GradingCriterion): Promise<boolean> {
        const result = await this.grpcHelper.deleteCriterion(c);
        return this.responseCodeSuccess(result);
    }

    public async loadCriteria(assignmentID: number, courseID: number): Promise<GradingBenchmark[]> {
        const result = await this.grpcHelper.loadCriteria(assignmentID, courseID);
        if (!this.responseCodeSuccess(result) || !result.data) {
            return [];
        }
        return result.data.getBenchmarksList();
    }

    public async addReview(ir: Review, courseID: number): Promise<Review | null> {
        const result = await this.grpcHelper.createReview(ir, courseID);
        if (!this.responseCodeSuccess(result) || !result.data) {
            return null;
        }
        return result.data;
    }

    public async editReview(ir: Review, courseID: number): Promise<boolean> {
        const result = await this.grpcHelper.updateReview(ir, courseID);
        return this.responseCodeSuccess(result);
    }

    public async getReviewers(submissionID: number, courseID: number): Promise<User[]> {
        const result = await this.grpcHelper.getReviewers(submissionID, courseID);
        if (!this.responseCodeSuccess(result) || !result.data) {
            return [];
        }
        return result.data.getReviewersList();
    }

    public async updateSubmissions(assignmentID: number, courseID: number, score: number, release: boolean, approve: boolean): Promise<boolean> {
        const result = await this.grpcHelper.updatesubmissions(assignmentID, courseID, score, release, approve);
        return this.responseCodeSuccess(result);
    }

    private toISubmission(sbm: Submission): ISubmission {
        const buildInfoAsString = sbm.getBuildinfo();
        const scoreInfoAsString = sbm.getScoreobjects();
        let buildInfo: IBuildInfo;
        let scoreObj: ITestCases[];
        // IMPORTANT: Field names of the Score struct found in the kit/score/score.go,
        // of the ITestCases struct found in the public/src/models.ts,
        // and names in the string passed to JSON.parse() method must match.
        // If experiencing an uncaught error in the browser which results in a blank page
        // when addressing lab information for a student/group, it is likely originates from here.
        try {
            buildInfo = JSON.parse(buildInfoAsString);
        } catch (e) {
            buildInfo = JSON.parse(
                "{\"builddate\": \"2017-07-28\", \"buildid\": 1, \"buildlog\": \"No tests for this assignment\", \"execTime\": 1}",
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
            executionTime: buildInfo.execTime,
            buildLog: buildInfo.buildlog,
            testCases: scoreObj,
            reviews: sbm.getReviewsList(),
            released: sbm.getReleased(),
            status: sbm.getStatus(),
            approvedDate: sbm.getApproveddate(),
        };
        return isbm;
    }

    private responseCodeSuccess(response: IGrpcResponse<any>): boolean {
        return response.status.getCode() === 0;
    }

    // temporary fix, will be removed with manual grading update
    private toUILinks(sbLinks: CourseSubmissions): IAllSubmissionsForEnrollment[] {
        const crs = sbLinks.getCourse();
        if (!crs) {
            return [];
        }
        const uilinks: IAllSubmissionsForEnrollment[] = [];
        sbLinks.getLinksList().forEach(l => {
            const enr = l.getEnrollment();
            if (enr) {
                const allLabs: ISubmissionLink[] = [];
                l.getSubmissionsList().forEach(s => {
                    const a = s.getAssignment();
                    const sb = s.getSubmission();
                    if (a) {
                        const name = a.getIsgrouplab() ? enr.getGroup()?.getName() : enr.getUser()?.getName();
                        allLabs.push({
                            assignment: a,
                            submission: sb ? this.toISubmission(sb) : undefined,
                            authorName: name ?? "Name not found",
                        });
                    }
                });
                uilinks.push({
                    course: crs,
                    enrollment: enr,
                    labs: allLabs,
                });
             }
        });
        return uilinks;
    }
}

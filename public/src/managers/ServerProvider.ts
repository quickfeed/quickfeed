import {
    CourseGroupStatus,
    CourseUserState,
    courseUserStateToString,
    IAssignment,
    IBuildInfo,
    ICourse,
    ICourseGroup,
    ICourseUserLink,
    ICourseWithEnrollStatus,
    IError, INewCourse,
    INewGroup,
    IOrganization, IStatusCode,
    ISubmission,
    ITestCases,
    IUser,
} from "../models";

import { HttpHelper, IHTTPResult } from "../HttpHelper";
import { ICourseProvider } from "./CourseManager";

import HttpStatusCode from "../HttpStatusCode";
import {
    ICourseEnrollment,
    IEnrollment,
    isCourseEnrollment,
    isUserEnrollment,
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
    private logger: ILogger;

    constructor(helper: HttpHelper, logger: ILogger) {
        this.helper = helper;
        this.logger = logger;
    }

    public async getCourses(): Promise<ICourse[]> {
        const result = await this.helper.get<any>(URL_ENDPOINT.courses);
        if (result.statusCode !== HttpStatusCode.OK || !result.data) {
            this.handleError(result, "getCourses");
            return [];
        }
        return result.data;
        // const data = JSON.parse(JSON.stringify(result.data).toLowerCase()) as ICourse[];
        // return mapify(result.data, (ele) => ele.id);
        // throw new Error("Method not implemented.");
    }

    public async getCoursesFor(user: IUser, state?: CourseUserState[]): Promise<ICourseEnrollment[]> {
        // TODO: Fix to use correct url request
        const status = state ? "?status=" + courseUserStateToString(state) : "";
        const result = await this.helper.get<ICourseWithEnrollStatus[]>("/" + URL_ENDPOINT.users +
            "/" + user.id + "/" + URL_ENDPOINT.courses + status);

        if (result.statusCode !== HttpStatusCode.OK || !result.data) {
            this.handleError(result, "getCoursesFor");
            return [];
        }

        const arr: ICourseEnrollment[] = [];
        result.data.forEach((ele) => {
            const enroll = ele.enrolled as number >= 0 ? ele.enrolled : undefined;
            arr.push({
                course: ele as ICourse,
                status: enroll,
                courseid: ele.id,
                userid: user.id,
                user,
            });
        });
        return arr;
    }

    public async getUsersForCourse(course: ICourse, state?: CourseUserState[]): Promise<IUserEnrollment[]> {
        const status = state ? "?status=" + courseUserStateToString(state) : "";
        const result = await this.helper.get<IEnrollment[]>(URL_ENDPOINT.courses + "/" +
            course.id + "/" + URL_ENDPOINT.users + status);

        if (result.statusCode !== HttpStatusCode.OK || !result.data) {
            this.handleError(result, "getUserForCourse");
            return [];
        }

        const arr: IUserEnrollment[] = [];
        result.data.forEach((ele) => {
            if (isCourseEnrollment(ele)) {
                ele.user = this.makeUserInfo(ele.user);
                console.log("BBBB ", ele.user);
                arr.push(ele);
            }
        });
        return arr;
    }

    public async getAssignments(courseId: number): Promise<IMap<IAssignment>> {
        const result = await this.helper.get<any>(URL_ENDPOINT.courses + "/" +
            courseId.toString() + "/" + URL_ENDPOINT.assignments);

        if (result.statusCode !== HttpStatusCode.OK || !result.data) {
            console.log(result);
            this.handleError(result, "getAssignments");
            throw new Error("Problem with the request");
        }
        return mapify(result.data as IAssignment[], (ele) => {
            if (!ele.deadline) {
                ele.deadline = new Date(2000, 1, 1);
            } else {
                ele.deadline = new Date(ele.deadline);
            }
            return ele.id;
        });
    }

    public async addUserToCourse(user: IUser, course: ICourse): Promise<boolean> {
        const result = await this.helper.post<{ status: CourseUserState }, undefined>
            (URL_ENDPOINT.courses + "/" + course.id + "/" + URL_ENDPOINT.users + "/" + user.id, {
                status: CourseUserState.pending,
            });
        if (result.statusCode === HttpStatusCode.CREATED) {
            return true;
        } else {
            this.handleError(result, "addUserToCourse");
        }
        return false;
    }

    public async changeUserState(link: ICourseUserLink, state: CourseUserState): Promise<boolean> {
        const uri: string[] = [
            URL_ENDPOINT.courses, link.courseId.toString(),
            URL_ENDPOINT.users, link.userid.toString()];
        const URL = this.buildURL(uri);
        const result = await this.helper.patch<{ courseID: number, userID: number, status: CourseUserState }, undefined>
            ("/" + URL, {
                courseID: link.courseId,
                userID: link.userid,
                status: state,
            });
        if (result.statusCode <= HttpStatusCode.ACCEPTED) {
            return true;
        } else {
            this.handleError(result, "changeUserState");
        }
        return false;
    }

    public async createNewCourse(courseData: INewCourse): Promise<ICourse | IError> {
        // const uri: string = "courses";
        const result = await this.helper.post<INewCourse, ICourse>(URL_ENDPOINT.courses, courseData);
        if (result.statusCode !== HttpStatusCode.CREATED || !result.data) {
            return this.parseError(result);
        }
        return JSON.parse(JSON.stringify(result.data)) as ICourse;
    }

    public async getCourse(ID: number): Promise<ICourse | null> {
        const uri: string[] = [URL_ENDPOINT.courses, ID.toString()];
        const URL = this.buildURL(uri);
        const result = await this.helper.get<any>(URL);
        if (result.statusCode !== HttpStatusCode.OK || !result.data) {
            this.handleError(result, "getCourse");
            return null;
        }
        return JSON.parse(JSON.stringify(result.data)) as ICourse;
    }

    public async updateCourse(courseID: number, courseData: ICourse): Promise<IStatusCode | IError> {
        // const uri: string = "courses/" + courseID;
        const uri: string[] = [URL_ENDPOINT.courses, courseID.toString()];
        const URL = this.buildURL(uri);
        const result = await this.helper.put<ICourse, ICourse>(URL, courseData);
        if (result.statusCode !== HttpStatusCode.OK) {
            this.handleError(result, "updateCourse");
            return this.parseError(result);
        }
        return JSON.parse(JSON.stringify(result.statusCode)) as IStatusCode;
    }

    public async createGroup(groupData: INewGroup, courseID: number): Promise<ICourseGroup | IError> {
        // const uri: string = "courses/" + courseID + "/groups";
        const uri: string[] = [URL_ENDPOINT.courses, courseID.toString(), URL_ENDPOINT.groups];
        const URL = this.buildURL(uri);
        const result = await this.helper.post<INewGroup, ICourseGroup>(URL, groupData);
        if (result.statusCode !== HttpStatusCode.CREATED || !result.data) {
            this.handleError(result, "createGroup");
            return this.parseError(result);
        }
        return JSON.parse(JSON.stringify(result.data)) as ICourseGroup;
    }

    public async getCourseGroups(courseID: number): Promise<ICourseGroup[]> {
        // const uri: string = "courses/" + courseID + "/groups";
        const uri: string[] = [URL_ENDPOINT.courses, courseID.toString(), URL_ENDPOINT.groups];
        const URL = this.buildURL(uri);
        const result = await this.helper.get<ICourseGroup>(URL);
        if (result.statusCode !== HttpStatusCode.OK || !result.data) {
            this.handleError(result, "getCourseGroups");
            return [];
        }
        return JSON.parse(JSON.stringify(result.data)) as ICourseGroup[];
    }

    public async getGroupByUserAndCourse(userID: number, courseID: number): Promise<ICourseGroup | null> {
        // const uri: string = "users/" + userID + "/courses/" + courseID + "/group";
        const uri: string[] = [URL_ENDPOINT.users, userID.toString(),
        URL_ENDPOINT.courses, courseID.toString(), URL_ENDPOINT.group];
        const URL = this.buildURL(uri);

        const result = await this.helper.get<ICourseGroup>(URL);
        if (result.statusCode !== HttpStatusCode.FOUND || !result.data) {
            return null;
        }
        return JSON.parse(JSON.stringify(result.data)) as ICourseGroup;
    }

    public async updateGroupStatus(groupID: number, st: CourseGroupStatus): Promise<boolean> {
        // const uri: string = "groups/" + groupID;
        const data = { status: st };
        const uri: string[] = [URL_ENDPOINT.groups, groupID.toString()];
        const URL = this.buildURL(uri);

        const result = await this.helper.patch<{ status: CourseGroupStatus }, undefined>(URL, data);
        if (result.statusCode === HttpStatusCode.OK) {
            return true;
        } else {
            this.handleError(result, "updateGroupStatus");
        }
        return false;
    }

    public async getGroup(groupID: number): Promise<ICourseGroup | null> {
        // const uri: string = "groups/" + groupID;
        const uri: string[] = [URL_ENDPOINT.groups, groupID.toString()];
        const URL = this.buildURL(uri);
        const result = await this.helper.get<ICourseGroup>(URL);
        if (result.statusCode !== HttpStatusCode.OK || !result.data) {
            this.handleError(result, "getGroup");
            return null;
        }
        return JSON.parse(JSON.stringify(result.data)) as ICourseGroup;
    }

    public async deleteGroup(groupID: number): Promise<boolean> {
        // const uri: string = "groups/" + groupID;
        const uri: string = URL_ENDPOINT.groups + "/" + groupID;
        const result = await this.helper.delete(uri);
        if (result.statusCode === HttpStatusCode.OK) {
            return true;
        } else {
            this.handleError(result, "deleteGroup");
        }
        return false;
    }

    public async updateGroup(groupData: INewGroup, groupID: number, courseID: number): Promise<IStatusCode | IError> {
        // const uri: string = "courses/" + courseID + "/groups/" + groupID;
        const uri: string[] = [URL_ENDPOINT.courses, courseID.toString(), URL_ENDPOINT.groups, groupID.toString()];
        const URL = this.buildURL(uri);
        const result = await this.helper.put<INewGroup, any>(URL, groupData);
        if (result.statusCode !== HttpStatusCode.OK) {
            this.handleError(result, "updateGroup");
            return this.parseError(result);
        }
        return JSON.parse(JSON.stringify(result.statusCode)) as IStatusCode;
    }

    // getAllGroupLabInfos
    public async getAllGroupLabInfos(courseID: number, groupID: number): Promise<IMap<ISubmission>> {
        const uri: string[] = [URL_ENDPOINT.courses, courseID.toString(),
        URL_ENDPOINT.groups, groupID.toString(), URL_ENDPOINT.submissions];
        const URL = this.buildURL(uri);
        const result = await this.helper.get<ISubmission[]>(URL);
        if (result.statusCode === HttpStatusCode.OK && result.data === undefined) {
            return {};
        }
        if (!result.data) {
            this.handleError(result, "getAllLabInfos");
            return {};
        }
        return mapify(result.data, (submission) => {
            let buildInfoAsString = "";
            let scoreInfoAsString = "";
            if ((submission as any).buildinfo && ((submission as any).buildinfo as string).trim().length > 2) {
                buildInfoAsString = (submission as any).buildinfo as string;
            }
            if ((submission as any).scoreobjects && ((submission as any).scoreobjects as string).trim().length > 2) {
                scoreInfoAsString = (submission as any).scoreobjects;
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

            submission.buildDate = buildInfo.builddate;
            submission.buildId = buildInfo.buildid;
            submission.buildLog = buildInfo.buildlog;
            submission.executetionTime = buildInfo.execTime;
            submission.testCases = scoreObj;
            submission.failedTests = 0;
            submission.passedTests = 0;
            if (submission.testCases == null) submission.testCases = [];
            submission.testCases.forEach((x) => {
                if (x.points !== x.score) {
                    submission.failedTests++;
                } else {
                    submission.passedTests++;
                }
            });
            // This ID is used as index in an array created by mapify function.
            return submission.id;
        });
    }

    // TODO change to use course id instead of getting all of them
    public async getAllLabInfos(courseID: number, userID: number): Promise<IMap<ISubmission>> {
        const uri: string[] = [URL_ENDPOINT.courses, courseID.toString(),
        URL_ENDPOINT.users, userID.toString(), URL_ENDPOINT.submissions];
        const URL = this.buildURL(uri);
        console.log("Testing ", URL);
        const result = await this.helper.get<ISubmission[]>(URL);
        if (result.statusCode === HttpStatusCode.OK && result.data === undefined) {
            return {};
        }
        if (!result.data) {
            this.handleError(result, "getAllLabInfos");
            return {};
        }
        return mapify(result.data, (submission) => {
            let buildInfoAsString = "";
            let scoreInfoAsString = "";
            if ((submission as any).buildinfo && ((submission as any).buildinfo as string).trim().length > 2) {
                buildInfoAsString = (submission as any).buildinfo as string;
            }
            if ((submission as any).scoreobjects && ((submission as any).scoreobjects as string).trim().length > 2) {
                scoreInfoAsString = (submission as any).scoreobjects;
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

            submission.buildDate = buildInfo.builddate;
            submission.buildId = buildInfo.buildid;
            submission.buildLog = buildInfo.buildlog;
            submission.executetionTime = buildInfo.execTime;
            submission.testCases = scoreObj;
            submission.failedTests = 0;
            submission.passedTests = 0;
            if (submission.testCases == null) submission.testCases = [];
            submission.testCases.forEach((x) => {
                if (x.points !== x.score) {
                    submission.failedTests++;
                } else {
                    submission.passedTests++;
                }
            });
            // This ID is used as index in an array created by mapify function.
            return submission.id;
        });
    }

    public async tryLogin(username: string, password: string): Promise<IUser | null> {
        throw new Error("tryLogin This could be removed since there is no normal login.");
    }

    public async logout(user: IUser): Promise<boolean> {
        window.location.assign("/" + URL_ENDPOINT.logout);
        return true;
    }

    public async getAllUser(): Promise<IUser[]> {
        const result = await this.helper.get<IUser[]>(URL_ENDPOINT.users);
        if (result.statusCode !== HttpStatusCode.FOUND || !result.data) {
            this.handleError(result, "getAllUser");
            return [];
        }
        const newArray = result.data.map<IUser>((ele) => this.makeUserInfo(ele));
        // return mapify(newArray, (ele) => ele.id);
        return newArray;
    }

    public async tryRemoteLogin(provider: string): Promise<IUser | null> {
        // TODO this needs to be fixed to return user data from provider
        if (provider.length > 0) {
            const requestString = "/" + URL_ENDPOINT.auth + "/" + provider;
            window.location.assign(requestString);
        }
        return null;
    }

    public async changeAdminRole(user: IUser): Promise<boolean> {
        const result = await this.helper.patch<{ isadmin: boolean }, {}>("/" + URL_ENDPOINT.users +
            "/" + user.id + "", { isadmin: true });
        if (result.statusCode < HttpStatusCode.BAD_REQUEST) {
            return false;
        } else {
            this.handleError(result, "changeAdminRole");
        }
        return true;
    }

    public async updateUser(user: IUser): Promise<boolean> {
        // TODO: make actuall implementation
        const result = await this.helper.patch<{ isadmin: boolean }, {}>("/" + URL_ENDPOINT.users
            + "/" + user.id + "", user);
        if (result.statusCode < HttpStatusCode.BAD_REQUEST) {
            return false;
        } else {
            this.handleError(result, "updateUser");
        }
        return true;
    }

    // TODO: check if resp.status contain correct status
    public async getDirectories(provider: string): Promise<IOrganization[]> {
        const uri: string = URL_ENDPOINT.directories;
        const data: { provider: string } = { provider };
        const result = await this.helper.post<{ provider: string }, IOrganization[]>(uri, data);
        if (result.statusCode === HttpStatusCode.OK && result.data) {
            return result.data;
        } else {
            this.handleError(result, "getDirectories");
        }
        return [];
    }

    public async getProviders(): Promise<string[]> {
        // https://nicolasf.itest.run/api/v1/providers
        const result = await this.helper.get<string[]>(URL_ENDPOINT.providers);
        if (result.data) {
            return result.data;
        } else {
            this.handleError(result, "getProviders");
        }
        return [];
    }

    public async getLoggedInUser(): Promise<IUser | null> {
        const result = await this.helper.get<IUser>(URL_ENDPOINT.user);
        if (result.statusCode !== HttpStatusCode.FOUND || !result.data) {
            this.handleError(result, "getLoggedInUser");
            return null;
        }
        return this.makeUserInfo(result.data);
    }

    public async refreshCoursesFor(courseID: number): Promise<any> {
        const result = await this.helper.post<any, null>(URL_ENDPOINT.courses + "/" + courseID + "/" +
            URL_ENDPOINT.refresh, null);
        if (result.statusCode !== HttpStatusCode.OK || !result.data) {
            this.handleError(result, "refreshCoursesFor");
            return null;
        }
        return this.makeUserInfo(result.data);
    }

    public async approveSubmission(submissionID: number): Promise<void> {
        const result = await this.helper.patch<any, null>(URL_ENDPOINT.submissions + "/" + submissionID, null);
        if (result.statusCode !== HttpStatusCode.OK) {
            this.handleError(result, "approveSubmission");
            return;
        }
        return;
    }

    private makeUserInfo(data: IUser): IUser {
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
        for (const tag of uri) {
            url += tag;
            url += "/";
        }
        return url;
    }
}

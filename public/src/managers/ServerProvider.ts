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
    IError,
    INewGroup,
    IOrganization,
    ISubmission,
    ITestCases,
    IUser,
} from "../models";

import { HttpHelper, IHTTPResult } from "../HttpHelper";
import { ICourseProvider } from "./CourseManager";

import {
    ICourseEnrollemtnt,
    IEnrollment,
    isCourseEnrollment,
    isUserEnrollment,
    IUserEnrollment,
    IUserProvider,
} from "../managers";
import { IMap, mapify } from "../map";
import { ILogger } from "./LogManager";

import { combinePath } from "../NavigationHelper";

export class ServerProvider implements IUserProvider, ICourseProvider {
    private helper: HttpHelper;
    private logger: ILogger;

    constructor(helper: HttpHelper, logger: ILogger) {
        this.helper = helper;
        this.logger = logger;
    }

    public async getCourses(): Promise<ICourse[]> {
        const result = await this.helper.get<any>("courses");
        if (result.statusCode !== 200 || !result.data) {
            this.handleError(result);
            return [];
        }
        return result.data;
        // const data = JSON.parse(JSON.stringify(result.data).toLowerCase()) as ICourse[];
        // return mapify(result.data, (ele) => ele.id);
        // throw new Error("Method not implemented.");
    }

    public async getCoursesFor(user: IUser, state?: CourseUserState[]): Promise<ICourseEnrollemtnt[]> {
        // TODO: Fix to use correct url request
        const status = state ? "?status=" + courseUserStateToString(state) : "";
        const result = await this.helper.get<ICourseWithEnrollStatus[]>("/users/" + user.id + "/courses" + status);
        if (result.statusCode !== 200 || !result.data) {
            this.handleError(result);
            return [];
        }

        const arr: ICourseEnrollemtnt[] = [];
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
        console.log(arr);
        return arr;
    }

    public async getUsersForCourse(course: ICourse, state?: CourseUserState[]): Promise<IUserEnrollment[]> {
        const status = state ? "?status=" + courseUserStateToString(state) : "";
        const result = await this.helper.get<IEnrollment[]>("/courses/" + course.id + "/users" + status);
        if (result.statusCode !== 200 || !result.data) {
            this.handleError(result);
            return [];
        }

        const arr: IUserEnrollment[] = [];
        result.data.forEach((ele) => {
            if (isCourseEnrollment(ele)) {
                ele.user = this.makeUserInfo(ele.user);
                arr.push(ele);
            }
        });
        return arr;
    }

    public async getAssignments(courseId: number): Promise<IMap<IAssignment>> {
        const result = await this.helper.get<any>("courses/" + courseId.toString() + "/assignments");

        if (result.statusCode !== 200 || !result.data) {
            console.log(result);
            this.handleError(result);
            throw new Error("Problem with the request");
        }
        return mapify(result.data as IAssignment[], (ele) => {
            ele.deadline = new Date(2017, 7, 18);
            return ele.id;
        });
    }

    public async addUserToCourse(user: IUser, course: ICourse): Promise<boolean> {
        const result = await this.helper.put<{ courseid: number, userid: number, status: CourseUserState }, undefined>
            ("/courses/" + course.id + "/users/" + user.id, {
                courseid: course.id,
                userid: user.id,
                status: CourseUserState.pending,
            });
        if (result.statusCode <= 202) {
            return true;
        } else {
            this.handleError(result);
        }
        return false;
    }

    public async changeUserState(link: ICourseUserLink, state: CourseUserState): Promise<boolean> {
        const result = await this.helper.put<{ courseid: number, userid: number, status: CourseUserState }, undefined>
            ("/courses/" + link.courseId + "/users/" + link.userid, {
                courseid: link.courseId,
                userid: link.userid,
                status: state,
            });
        if (result.statusCode <= 202) {
            return true;
        } else {
            this.handleError(result);
        }
        return false;
    }

    public async createNewCourse(courseData: ICourse): Promise<boolean> {
        const uri: string = "courses";
        const data: ICourse = courseData;
        const result = await this.helper.post<ICourse, ICourse>(uri, data);
        console.log("res = ", result);
        return true;
    }

    public async getCourse(id: number): Promise<ICourse | null> {
        const result = await this.helper.get<any>("courses/" + id);
        if (result.statusCode !== 200 || !result.data) {
            console.log("Error =>", result);
            this.handleError(result);
            return null;
        }
        const data = JSON.parse(JSON.stringify(result.data)) as ICourse;
        return data;
    }

    public async updateCourse(courseId: number, courseData: ICourse): Promise<boolean> {
        const uri: string = "courses/" + courseId;
        const result = await this.helper.put<ICourse, ICourse>(uri, courseData);
        if (result.statusCode !== 200) {
            console.log("Error =>", result);
            this.handleError(result);
            return false;
        }
        console.log("Success => ", result);
        return true;
    }

    public async createGroup(groupData: INewGroup, courseId: number): Promise<ICourseGroup | IError> {
        const uri: string = "courses/" + courseId + "/groups";
        const result = await this.helper.post<INewGroup, ICourseGroup>(uri, groupData);
        if (result.statusCode !== 201 || !result.data) {
            this.handleError(result);
            return this.parseError(result);
        }
        const data = JSON.parse(JSON.stringify(result.data)) as ICourseGroup;
        return data;
    }

    public async getCourseGroups(courseId: number): Promise<ICourseGroup[]> {
        const uri: string = "courses/" + courseId + "/groups";
        const result = await this.helper.get<ICourseGroup>(uri);
        if (result.statusCode !== 200 || !result.data) {
            console.log("Error =>", result);
            return [];
        }
        const data = JSON.parse(JSON.stringify(result.data)) as ICourseGroup[];
        return data;
    }

    public async getCourseByUserAndCourse(userid: number, courseid: number): Promise<ICourseGroup | null> {
        const uri: string = "users/" + userid + "/courses/" + courseid + "/group";
        const result = await this.helper.get<ICourseGroup>(uri);
        if (result.statusCode !== 302 || !result.data) {
            return null;
        }
        const data = JSON.parse(JSON.stringify(result.data)) as ICourseGroup;
        return data;
    }

    public async updateGroupStatus(groupId: number, st: CourseGroupStatus): Promise<boolean> {
        const uri: string = "groups/" + groupId;
        const data = { status: st };

        const result = await this.helper.patch<{ status: CourseGroupStatus }, undefined>(uri, data);
        if (result.statusCode === 200) {
            return true;
        } else {
            this.handleError(result);
        }
        return false;
    }

    // TODO change to use course id instead of getting all of them
    public async getAllLabInfos(courseId: number): Promise<IMap<ISubmission>> {
        const result = await this.helper.get<ISubmission[]>(("courses/" + courseId.toString() + "/submissions"));
        if (!result.data) {
            this.handleError(result);
            return {};
        }
        return mapify(result.data, (e) => {
            let a = "{\"builddate\": \"2017-07-28\", \"buildid\": 1, \"buildlog\": \"This is cool\", \"execTime\": 1}";
            let b = "[{\"name\": \"Test 1\", \"score\": 3, \"points\": 4, \"weight\": 100}]";
            if ((e as any).buildinfo && ((e as any).buildinfo as string).trim().length > 2) {
                a = (e as any).buildinfo as string;
            }
            if ((e as any).scoreobjects && ((e as any).scoreobjects as string).trim().length > 2) {
                b = (e as any).scoreobjects;
            }
            console.log(a);
            let tempInfo: IBuildInfo;
            let scoreObj: ITestCases[];
            try {
                tempInfo = JSON.parse(a);
            } catch (e) {
                tempInfo = JSON.parse(
                    "{\"builddate\": \"2017-07-28\", \"buildid\": 1, \"buildlog\": \"This is cool\", \"execTime\": 1}",
                );
            }
            try {
                scoreObj = JSON.parse(b);
            } catch (e) {
                scoreObj = JSON.parse(
                    "[{\"name\": \"Test 1\", \"score\": 3, \"points\": 4, \"weight\": 100}]",
                );
            }

            e.buildDate = tempInfo.builddate;
            e.buildId = tempInfo.buildid;
            e.buildLog = tempInfo.buildlog;
            e.executetionTime = tempInfo.exectime;
            e.failedTests = 0;
            e.passedTests = 1;
            e.testCases = scoreObj;
            return e.id;
        });
    }

    public async tryLogin(username: string, password: string): Promise<IUser | null> {
        throw new Error("tryLogin This could be removed since there is no normal login.");
    }

    public async logout(user: IUser): Promise<boolean> {
        window.location.assign("/logout");
        return true;
    }

    public async getAllUser(): Promise<IUser[]> {
        const result = await this.helper.get<Array<{ id: number, isadmin: boolean }>>("users");
        if (result.statusCode !== 302 || !result.data) {
            this.handleError(result);
            return [];
        }
        const newArray = result.data.map<IUser>((ele) => this.makeUserInfo(ele));
        // return mapify(newArray, (ele) => ele.id);
        return newArray;
    }

    public async tryRemoteLogin(provider: string): Promise<IUser | null> {
        let requestString: null | string = null;
        switch (provider) {
            case "github":
                requestString = "/auth/github";
                break;
            case "gitlab":
                requestString = "/auth/gitlab";
                break;
        }
        if (requestString) {
            window.location.assign(requestString);
            return null;
        } else {
            return null;
        }
    }

    public async changeAdminRole(user: IUser): Promise<boolean> {
        const result = await this.helper.patch<{ isadmin: boolean }, {}>("/users/" + user.id + "", { isadmin: true });
        if (result.statusCode < 400) {
            return false;
        } else {
            this.handleError(result);
        }
        return true;
    }

    // TODO: check if resp.status contain correct status
    public async getDirectories(provider: string): Promise<IOrganization[]> {
        const uri: string = "directories";
        const data: { provider: string } = { provider };
        const result = await this.helper.post<{ provider: string }, IOrganization[]>(uri, data);
        if (result.data) {
            return result.data;
        } else {
            this.handleError(result);
        }
        return [];
    }

    public async getProviders(): Promise<string[]> {
        // https://nicolasf.itest.run/api/v1/providers
        const result = await this.helper.get<string[]>("providers");
        if (result.data) {
            return result.data;
        } else {
            this.handleError(result);
        }
        return [];
    }

    public async getLoggedInUser(): Promise<IUser | null> {
        const result = await this.helper.get<{ id: number, isadmin: boolean }>("user");
        if (result.statusCode !== 302 || !result.data) {
            return null;
        } else {
            this.handleError(result);
        }
        return this.makeUserInfo(result.data);
    }

    private makeUserInfo(data: { id: number, isadmin: boolean }): IUser {
        return {
            firstname: "Agent 00" + data.id,
            lastname: "NR" + data.id,
            isadmin: data.isadmin,
            id: data.id,
            personid: 1000,
            email: "00" + data.id + "@secretorganization.com",
        };
    }

    private handleError(result: IHTTPResult<any>): void {
        this.logger.warn("Request to server failed with status code: " + result.statusCode, true);
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

    /*
     {
     "ID": 1,
     "remoteidentities": [
     {
     "ID": 1,
     "Provider": "github",
     "RemoteID": 1964338,
     "UserID": 1
     }
     ]
     }
     */
}

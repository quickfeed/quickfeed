import {
    CourseUserState,
    courseUserStateToString,
    IAssignment,
    ICourse,
    ICourseUserLink,
    ICourseWithEnrollStatus,
    ILabInfo,
    IOrganization,
    IUser,
} from "../models";

import { HttpHelper } from "../HttpHelper";
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

async function request(url: string): Promise<string> {
    const req = new XMLHttpRequest();
    return new Promise<string>((resolve, reject) => {
        req.onreadystatechange = () => {
            if (req.readyState === 4) {
                if (req.status === 200) {
                    console.log(req);
                    resolve(req.responseText);
                } else {
                    reject(req);
                }
            }
        };
        req.open("GET", url, true);
        req.send();
    });
}

export class ServerProvider implements IUserProvider, ICourseProvider {
    private helper: HttpHelper;

    constructor(helper: HttpHelper) {
        this.helper = helper;
    }

    public async getCourses(): Promise<ICourse[]> {
        const result = await this.helper.get<any>("courses");
        if (result.statusCode !== 200 || !result.data) {
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
        console.log(status);
        const result = await this.helper.get<ICourseWithEnrollStatus[]>("/users/" + user.id + "/courses" + status);
        if (result.statusCode !== 200 || !result.data) {
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

    // public async getActiveCoursesFor(user: IUser): Promise<ICourseWithEnrollStatus[]> {
    //     const query = "?active=true";
    //     const result = await this.helper.get<ICourseWithEnrollStatus[]>("/users/" + user.id + "/courses" + query);
    //     if (result.statusCode !== 200 || !result.data) {
    //         return [];
    //     }
    //     return result.data;
    // }

    // public async getCoursesWithEnrollStatus(user: IUser, state?: CourseUserState)
    // : Promise<ICourseWithEnrollStatus[]> {
    //     const status = state ? "?status=" + courseUserStateToString(state) : "";
    //     const result = await this.helper.get<ICourseWithEnrollStatus[]>("/users/" + user.id + "/courses" + status);
    //     if (result.statusCode !== 200 || !result.data) {
    //         return [];
    //     }

    //     return result.data;
    // }

    public async getUsersForCourse(course: ICourse, state?: CourseUserState[]): Promise<IUserEnrollment[]> {
        const status = state ? "?status=" + courseUserStateToString(state) : "";
        const result = await this.helper.get<IEnrollment[]>("/courses/" + course.id + "/users" + status);
        if (result.statusCode !== 200 || !result.data) {
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
            throw new Error("Problem with the request");
        }
        return mapify(result.data as IAssignment[], (ele) => {
            ele.deadline = new Date(2017, 7, 18);
            return ele.id;
        });
    }

    public async addUserToCourse(user: IUser, course: ICourse): Promise<boolean> {
        const resp = await this.helper.put<{ courseid: number, userid: number, status: CourseUserState }, undefined>
            ("/courses/" + course.id + "/users/" + user.id, {
                courseid: course.id,
                userid: user.id,
                status: CourseUserState.pending,
            });
        if (resp.statusCode === 201) {
            return true;
        }
        return false;
    }

    public async changeUserState(link: ICourseUserLink, state: CourseUserState): Promise<boolean> {
        const resp = await this.helper.put<{ courseid: number, userid: number, status: CourseUserState }, undefined>
            ("/courses/" + link.courseId + "/users/" + link.userid, {
                courseid: link.courseId,
                userid: link.userid,
                status: state,
            });
        if (resp.statusCode === 201) {
            return true;
        }
        return false;
    }

    public async createNewCourse(courseData: ICourse): Promise<boolean> {
        const uri: string = "courses";
        const data: ICourse = courseData;
        const resp = await this.helper.post<ICourse, ICourse>(uri, data);
        // return resp.data;
        console.log("res = ", resp);
        return true;
    }

    public async getCourse(id: number): Promise<ICourse | null> {
        const result = await this.helper.get<any>("courses/" + id);
        if (result.statusCode !== 200 || !result.data) {
            console.log("Error =>", result);
            return null;
        }
        const data = JSON.parse(JSON.stringify(result.data)) as ICourse;
        return data;
    }

    public async updateCourse(courseId: number, courseData: ICourse): Promise<boolean> {
        const uri: string = "courses/" + courseId;
        const resp = await this.helper.put<ICourse, ICourse>(uri, courseData);
        if (resp.statusCode !== 200) {
            console.log("Error =>", resp);
            return false;
        }
        console.log("Success => ", resp);
        // const course = JSON.parse(JSON.stringify(resp.data)) as ICourse;
        return true;
    }

    public async getAllLabInfos(): Promise<IMap<ILabInfo>> {
        return {};
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
        }
        return true;
    }

    // TODO: check if resp.status contain correct status
    public async getDirectories(provider: string): Promise<IOrganization[]> {
        const uri: string = "directories";
        const data: { provider: string } = { provider };
        const resp = await this.helper.post<{ provider: string }, IOrganization[]>(uri, data);
        if (resp.data) {
            return resp.data;
        }
        return [];
    }

    public async getLoggedInUser(): Promise<IUser | null> {
        const result = await this.helper.get<{ id: number, isadmin: boolean }>("user");
        if (result.statusCode !== 302 || !result.data) {
            return null;
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

import { IUserProvider } from "../managers";
import { IMap, mapify } from "../map";
import { CourseUserState, IAssignment, ICourse, ICourseUserLink, ILabInfo, IOrganization, IUser } from "../models";
import { ICourseProvider } from "./CourseManager";

import { HttpHelper } from "../HttpHelper";

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

    public async getCourses(): Promise<IMap<ICourse>> {
        const result = await this.helper.get<any>("courses?user=0");
        if (result.statusCode !== 200 || !result.data) {
            return {};
        }
        // const data = JSON.parse(JSON.stringify(result.data).toLowerCase()) as ICourse[];
        return mapify(result.data, (ele) => ele.id);
        // throw new Error("Method not implemented.");
    }

    public async getCoursesFor(user: IUser, state?: CourseUserState): Promise<ICourse[]> {
        const result = await this.helper.get<any>("courses?user=" + user.id);
        if (result.statusCode !== 200 || !result.data) {
            return [];
        }
        // const data = JSON.parse(JSON.stringify(result.data).toLowerCase()) as ICourse[];
        return result.data;
    }

    public async getUsersForCourse(course: ICourse, state?: CourseUserState | undefined): Promise<IUser[]> {
        throw new Error("Method not implemented.");
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
        const resp = await this.helper.put<{}, undefined>
            ("/courses/" + course.id + "/users/" + user.id, {});
        if (resp.statusCode === 201) {
            return true;
        }
        return false;
    }

    public async changeUserState(link: ICourseUserLink, state: CourseUserState): Promise<boolean> {
        throw new Error("Method not implemented.");
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

    public async getAllUser(): Promise<IMap<IUser>> {
        const result = await this.helper.get<Array<{ id: number, isadmin: boolean }>>("users");
        if (result.statusCode !== 302 || !result.data) {
            return {};
        }
        const newArray = result.data.map<IUser>((ele) => this.makeUserInfo(ele));
        return mapify(newArray, (ele) => ele.id);
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
        const result = await this.helper.put<{ setadmin: boolean }, {}>("/users/" + user.id + "", { setadmin: true });
        if (result.statusCode !== 200) {
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
            firstname: "No name",
            lastname: "names",
            isadmin: data.isadmin,
            id: data.id,
            personid: 1000,
            email: "no@name.com",
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

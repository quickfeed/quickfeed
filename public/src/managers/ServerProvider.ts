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
        const result = await this.helper.get<any>("courses");
        if (result.statusCode !== 200 || !result.data) {
            return {};
        }
        const data = JSON.parse(JSON.stringify(result.data).toLowerCase()) as ICourse[];
        return mapify(data, (ele) => ele.id);
        // throw new Error("Method not implemented.");
    }

    public async getAssignments(courseId: number): Promise<IMap<IAssignment>> {
        throw new Error("Method not implemented.");
    }

    public async getCoursesStudent(): Promise<ICourseUserLink[]> {
        throw new Error("Method not implemented.");
    }

    public async addUserToCourse(user: IUser, course: ICourse): Promise<boolean> {
        throw new Error("Method not implemented.");
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

    public async updateCourse(courseData: ICourse): Promise<boolean> {
        throw new Error("Method not implemented");
    }

    public async deleteCourse(id: number): Promise<boolean> {
        throw new Error("Method not implemented");
    }

    public async getAllLabInfos(): Promise<IMap<ILabInfo>> {
        throw new Error("Method not implemented.");
    }

    public async tryLogin(username: string, password: string): Promise<IUser | null> {
        throw new Error("Method not implemented.");
    }

    public async logout(user: IUser): Promise<boolean> {
        window.location.assign("/logout");
        return true;
    }

    public async getAllUser(): Promise<IMap<IUser>> {
        const result = await this.helper.get<Array<{ ID: number }>>("users");
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
        throw new Error("Method not implemented");
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
        const result = await this.helper.get<{ ID: number }>("user");
        if (result.statusCode !== 302 || !result.data) {
            return null;
        }
        return this.makeUserInfo(result.data);
    }

    private makeUserInfo(data: { ID: number }): IUser {
        return {
            firstName: "No name",
            lastName: "names",
            isAdmin: true,
            id: data.ID,
            personId: 1000,
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

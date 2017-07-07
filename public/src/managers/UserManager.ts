import { IEventData, newEvent } from "../event";
import { IUser } from "../models";

import { ArrayHelper } from "../helper";
import { IMap, MapHelper } from "../map";

interface IUserProvider {
    tryLogin(username: string, password: string): Promise<IUser | null>;
    logout(user: IUser): Promise<boolean>;
    getAllUser(): Promise<IMap<IUser>>;
    tryRemoteLogin(provider: string): Promise<IUser | null>;
    changeAdminRole(user: IUser): Promise<boolean>;
}

interface IUserLoginEvent extends IEventData {
    user: IUser;
}

class UserManager {
    public onLogin = newEvent<IUserLoginEvent>("UserManager.onLogin");
    public onLogout = newEvent<IEventData>("UserManager.onLogout");

    private userProvider: IUserProvider;
    private currentUser: IUser | null;

    constructor(userProvider: IUserProvider) {
        this.userProvider = userProvider;
    }

    public getCurrentUser(): IUser | null {
        return this.currentUser;
    }

    public async tryLogin(username: string, password: string): Promise<IUser | null> {
        const result = await this.userProvider.tryLogin(username, password);
        if (result) {
            this.currentUser = result;
            this.onLogin({ target: this, user: this.currentUser });
        }
        return result;
    }

    public async tryRemoteLogin(provider: string): Promise<IUser | null> {
        const result = await this.userProvider.tryRemoteLogin(provider);
        if (result) {
            this.currentUser = result;
            this.onLogin({ target: this, user: this.currentUser });
        }
        return result;
    }

    public async logout() {
        if (this.currentUser) {
            await this.userProvider.logout(this.currentUser);
            this.currentUser = null;
            this.onLogout({ target: this });
        }
    }

    public isAdmin(user: IUser): boolean {
        return user.isAdmin;
    }

    public async isTeacher(user: IUser): Promise<boolean> {
        return user.id > 100;
    }

    public async getAllUser(): Promise<IUser[]> {
        return MapHelper.toArray(await this.userProvider.getAllUser());
    }

    public async getUsers(ids: number[]): Promise<IUser[]> {
        return MapHelper.toArray(await this.getUsersAsMap(ids));
    }

    public async getUsersAsMap(ids: number[]): Promise<IMap<IUser>> {
        const returnUsers: IMap<IUser> = {};
        const allUsers = await this.userProvider.getAllUser();
        ids.forEach((ele) => {
            const temp = allUsers[ele];
            if (temp) {
                returnUsers[ele] = temp;
            }
        });
        return returnUsers;
    }

    public async getUser(id: number): Promise<IUser> {
        throw new Error("Not implemented error");
    }

    public async changeAdminRole(user: IUser): Promise<boolean> {
        return this.userProvider.changeAdminRole(user);
    }
}

export { IUserProvider, UserManager };

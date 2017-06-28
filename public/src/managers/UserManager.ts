import { IEventData, newEvent } from "../event";
import { IUser } from "../models";

interface IUserProvider {
    tryLogin(username: string, password: string): IUser | null;
    logout(user: IUser): void;
    getAllUser(): IUser[];
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

    public tryLogin(username: string, password: string): IUser | null {
        const result = this.userProvider.tryLogin(username, password);
        if (result) {
            this.currentUser = result;
            this.onLogin({ target: this, user: this.currentUser });
        }
        return result;
    }

    public logout() {
        if (this.currentUser) {
            this.userProvider.logout(this.currentUser);
            this.currentUser = null;
            this.onLogout({ target: this });
        }
    }

    public isAdmin(user: IUser): boolean {
        return user.id > 100;
    }

    public isTeacher(user: IUser): boolean {
        return user.id > 100;
    }

    public getAllUser(): IUser[] {
        return this.userProvider.getAllUser();
    }

    public getUser(id: number): IUser {
        throw new Error("Not implemented error");
    }
}

export { IUserProvider, UserManager };

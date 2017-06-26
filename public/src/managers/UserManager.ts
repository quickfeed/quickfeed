import { IUser } from "../models";

interface IUserProvider {
    tryLogin(username: string, password: string): IUser | null;
    getAllUser(): IUser[];
}

class UserManager {
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
        }
        return result;
    }

    public getAllUser(): IUser[] {
        return this.userProvider.getAllUser();
    }

    public getUser(id: number): IUser {
        throw new Error("Not implemented error");
    }
}

export { IUserProvider, UserManager };

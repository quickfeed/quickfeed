import { IEventData, newEvent } from "../event";
import { ILogger } from "./LogManager";

import { Enrollment, User } from "../../proto/ag_pb";

export interface IUserProvider {
    tryLogin(username: string, password: string): Promise<User | null>;
    logout(user: User): Promise<boolean>;
    getUser(): Promise<User>;
    getUsers(): Promise<User[]>;
    tryRemoteLogin(provider: string): Promise<User | null>;
    changeAdminRole(user: User, promote: boolean): Promise<boolean>;
    getLoggedInUser(): Promise<User | null>;
    updateUser(user: User): Promise<boolean>;
    isAuthorizedTeacher(): Promise<boolean>;
}

interface IUserLoginEvent extends IEventData {
    user: User;
}

export class UserManager {
    /**
     * This event fires when the user has logged in.
     */
    public onLogin = newEvent<IUserLoginEvent>("UserManager.onLogin");
    /**
     * This event fires when the user has logged out.
     */
    public onLogout = newEvent<IEventData>("UserManager.onLogout");

    private userProvider: IUserProvider;
    private currentUser: User | null;

    /**
     * Creates a new instance of the UserManager
     * @param userProvider A user provider to get user information from
     */
    constructor(userProvider: IUserProvider, logger: ILogger) {
        this.userProvider = userProvider;
    }

    /**
     * Returns the current logged in user.
     * If there is no logged in user it returns null
     */
    public getCurrentUser(): User | null {
        return this.currentUser;
    }

    public isValidUser(user: User): boolean {
        return user.getEmail().length > 0
            && user.getName().length > 0
            && user.getStudentid().length > 0;
    }

    /**
     * Tries to login to the service with username and password
     * This is only used for testing
     */
    public async tryLogin(username: string, password: string): Promise<User | null> {
        const result = await this.userProvider.tryLogin(username, password);
        if (result) {
            this.currentUser = result;
            this.onLogin({ target: this, user: this.currentUser });
        }
        return result;
    }

    /**
     * Try to login with a remote service, like github and gitlab.
     * Normaly this function redirects before it returns.
     */
    public async tryRemoteLogin(provider: string): Promise<User | null> {
        const result = await this.userProvider.tryRemoteLogin(provider);
        if (result) {
            this.currentUser = result;
            this.onLogin({ target: this, user: this.currentUser });
        }
        return result;
    }

    /**
     * logout from the current logged in session
     */
    public async logout() {
        if (this.currentUser) {
            await this.userProvider.logout(this.currentUser);
            this.currentUser = null;
            this.onLogout({ target: this });
        }
    }

    /**
     * Checks whether current user has teacher scopes
     * @returns Returns true if user has already been authorized as teacher
     */
    public async isAuthorizedTeacher(): Promise<boolean> {
        return this.userProvider.isAuthorizedTeacher();
    }

    public async getUsers(): Promise<User[]> {
        return this.userProvider.getUsers();
    }

    public async changeAdminRole(user: User, promote: boolean): Promise<boolean> {
        return this.userProvider.changeAdminRole(user, promote);
    }

    public updateUser(user: User): Promise<boolean> {
        return this.userProvider.updateUser(user);
    }

    public getUser(): Promise<User> {
        return this.userProvider.getUser();
    }

    public async isTeacher(courseID?: number): Promise<boolean> {
        let valid = false;
        const user = await this.getUser();

        if (user) {
            user.getEnrollmentsList().forEach((ele) => {
                if (courseID) {
                    if (courseID === ele.getCourseid() && ele.getStatus() === Enrollment.UserStatus.TEACHER) {
                        valid = true;
                    }

                } else if (ele.getStatus() === Enrollment.UserStatus.TEACHER) {
                    valid = true;
                }
            });
        }
        return valid;
    }

    /**
     * Communicates with the backend to see if there is a logged in user
     */
    public async checkUserLoggedIn(): Promise<boolean> {
        const usr = await this.userProvider.getLoggedInUser();
        this.currentUser = usr;
        return usr ? true : false;
    }
}

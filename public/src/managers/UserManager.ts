import { IUser } from "../models";

interface IUserProvider {
    tryLogin(username: string, password: string): IUser | null;
    getAllUser(): IUser[];
}

class UserManager{
    private userProvider: IUserProvider;
    private currentUser: IUser | null;

    constructor(userProvider: IUserProvider){
        this.userProvider = userProvider;
    }

    getCurrentUser(): IUser | null{
        return this.currentUser;
    }

    tryLogin(username: string, password: string): IUser | null{
        let result = this.userProvider.tryLogin(username, password);
        if (result){
            this.currentUser = result;
        }
        return result;
    }

    getAllUser(): IUser[]{
        return this.userProvider.getAllUser();
    }

    getUser(id: number){
        
    }
}

export {IUserProvider, UserManager};

import { IUserProvider } from "../managers";
import { IMap } from "../map";
import { IUser } from "../models";

function request(url: string, callback: (data: string) => void) {
    const req = new XMLHttpRequest();

    req.onreadystatechange = () => {
        if (req.readyState === 4 && req.status === 200) {
            console.log(req);
            callback(req.responseText);
        }
    };
    req.open("GET", url, true);
    req.send();
}

export class ServerProvider implements IUserProvider {
    public tryLogin(username: string, password: string): IUser | null {
        throw new Error("Method not implemented.");
    }
    public logout(user: IUser): void {
        throw new Error("Method not implemented.");
    }
    public getAllUser(): IMap<IUser> {
        throw new Error("Method not implemented.");
    }
    public tryRemoteLogin(provider: string, callback: (result: IUser | null) => void): void {
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
            request(requestString, (data: string) => {
                console.log(data);
                callback(null);
            });
        } else {
            callback(null);
        }
    }
}

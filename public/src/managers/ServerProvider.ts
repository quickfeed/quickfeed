
import { IUserProvider } from "../managers";
import { IMap } from "../map";
import { IUser } from "../models";

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

export class ServerProvider implements IUserProvider {
    public async tryLogin(username: string, password: string): Promise<IUser | null> {
        throw new Error("Method not implemented.");
    }
    public async logout(user: IUser): Promise<boolean> {
        throw new Error("Method not implemented.");
    }
    public async getAllUser(): Promise<IMap<IUser>> {
        throw new Error("Method not implemented.");
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
}

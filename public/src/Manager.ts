import { IUser } from "./overmind/state";




export class Manager {
    public async tryRemoteLogin(): Promise<IUser | null> {
        
        const requestString = "/" + "auth" + "/" + "github";
        window.location.assign(requestString);
        
        return null;
    }
}
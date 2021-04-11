import { Self } from "./overmind/state";




export class Manager {
    public async tryRemoteLogin(): Promise<Self | null> {
        
        const requestString = "/" + "auth" + "/" + "github";
        window.location.assign(requestString);
        
        return null;
    }
}
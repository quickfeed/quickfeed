import * as React from "react";
import { User } from "../../proto/ag_pb";
import { NavigationManager, UserManager } from "../managers";
import { INavInfo } from "../NavigationHelper";
import { View, ViewPage } from "./ViewPage";

export class LoginPage extends ViewPage {
    private navMan: NavigationManager;
    private userMan: UserManager;

    constructor(navMan: NavigationManager, userMan: UserManager) {
        super();
        this.navMan = navMan;
        this.userMan = userMan;

        this.navHelper.defaultPage = "index";

        this.navHelper.registerFunction("index", this.index);
        this.navHelper.registerFunction("login/{provider}", this.login);
        this.navHelper.registerFunction("logout", this.logout);
    }

    // TODO(meling) remove
    public async index(info: INavInfo<{ provider: string }>): View {
        return <div>Index page</div>;
    }

    public async login(info: INavInfo<{ provider: string }>): View {
        const iUser: Promise<User | null> = this.userMan.tryRemoteLogin(info.params.provider);
        iUser.then((result: User | null) => {
            if (result) {
                this.navMan.navigateToDefault();
            } else {
                console.log("Failed");
            }
        });
        return Promise.resolve(<div>Logging in please wait</div>);
    }

    public async logout(info: INavInfo<{}>): View {
        this.userMan.logout();
        return <div>Logged out</div>;
    }
}

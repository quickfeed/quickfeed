import * as React from "react";
import { ViewPage } from "./ViewPage";

import { INavInfo } from "../NavigationHelper";

import { NavigationManager, UserManager } from "../managers";
import { IUser } from "../models";

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
    }

    public index(info: INavInfo<{ provider: string }>): JSX.Element {
        return <div>Quickly hide, you should not be here! Someone is going to get mad...</div>;
    }

    public login(info: INavInfo<{ provider: string }>): JSX.Element {
        this.userMan.tryRemoteLogin(info.params.provider, (result: IUser | null) => {
            if (result) {
                console.log("Sucessful login of: " + result);
            } else {
                console.log("Failed");
            }
        });
        return <div>Logging in please wait</div>;
    }

    public logout(info: INavInfo<{}>): JSX.Element {
        this.userMan.logout();
        return <div>Logged out</div>;
    }

    public renderContent(page: string): JSX.Element {
        const pageContent = this.navHelper.navigateTo(page);
        if (pageContent) {
            return pageContent;
        }
        return <div>404 Not found</div>;
    }
}

import * as React from "react";

import { View, ViewPage } from "./ViewPage";

import { UserProfile } from "../components/forms/UserProfile";
import { NavigationManager, UserManager } from "../managers";
import { IUser } from "../models";

export class UserPage extends ViewPage {
    private userMan: UserManager;
    private navMan: NavigationManager;
    private curUser: IUser;

    constructor(navMan: NavigationManager, userMan: UserManager) {
        super();
        this.template = "frontpage";
        this.userMan = userMan;
        this.navMan = navMan;
        this.navHelper.defaultPage = "profile";
        this.navHelper.registerFunction("profile", this.profile);
        this.curUser = this.userMan.getCurrentUser() || {
            name: "", email: "", avatarurl: "", isadmin: false, studentid: "", id: 0,
        };
    }

    public async updateUser() {
        this.curUser = this.userMan.getCurrentUser() || {
            name: "", email: "", avatarurl: "", isadmin: false, studentid: "", id: 0,
        };
        this.navMan.refresh();
    }

    public async profile(): View {
        return <UserProfile userMan={this.userMan} onEditStop={() => { this.updateUser(); }} />;
        // throw new Error("Not implemented");
    }

    public async renderMenu(index: number): Promise<JSX.Element[]> {
        if (index === 1) {
            return [<div id="0" className="jumbotron">
                <div className="centerblock container">
                    <h1>Hi, {this.curUser.name}</h1>
                </div>
            </div>];
        }
        return [];
    }
}

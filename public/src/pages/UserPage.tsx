import * as React from "react";

import { View, ViewPage } from "./ViewPage";

import { UserProfile } from "../components/forms/UserProfile";
import { NavigationManager, UserManager } from "../managers";
import { User } from "../../proto/ag_pb";

export class UserPage extends ViewPage {
    private userMan: UserManager;
    private navMan: NavigationManager;
    private curUser: User;

    constructor(navMan: NavigationManager, userMan: UserManager) {
        super();
        this.template = "frontpage";
        this.userMan = userMan;
        this.navMan = navMan;
        this.navHelper.defaultPage = "profile";
        this.navHelper.registerFunction("profile", this.profile);
        const noUser: User = new User();
        noUser.setName("");
        noUser.setEmail("");
        noUser.setAvatarurl("");
        noUser.setIsadmin(false);
        noUser.setId(0);
        noUser.setStudentid("");
        this.curUser = this.userMan.getCurrentUser() || noUser;
    }

    public async updateUser() {
        const noUser: User = new User();
        noUser.setName("");
        noUser.setEmail("");
        noUser.setAvatarurl("");
        noUser.setIsadmin(false);
        noUser.setId(0);
        noUser.setStudentid("");
        this.curUser = this.userMan.getCurrentUser() || noUser;
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
                    <h1>Hi, {this.curUser.getName()}</h1>
                </div>
            </div>];
        }
        return [];
    }
}

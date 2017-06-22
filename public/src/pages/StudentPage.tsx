import { NavigationManager, ILink } from "../managers/NavigationManager";
import { UserManager } from "../managers/UserManager";
import * as React from "react";
import { UserViewer } from "./views/UserView";
import { HelloView } from "./views/HelloView";
import { NavMenu } from "../components";
import { ViewPage } from "./ViewPage";


class StudentPage extends ViewPage {
    navMan: NavigationManager;
    constructor(users: UserManager, navMan: NavigationManager){
        super();

        this.navMan = navMan;
        this.defaultPage = "opsys/lab1";
        this.pages["opsys/lab1"] = <h1>Lab1</h1>;
        this.pages["opsys/lab2"] = <h1>Lab2</h1>;
        this.pages["opsys/lab3"] = <h1>Lab3</h1>;
        this.pages["opsys/lab4"] = <h1>Lab4</h1>;
        this.pages["user"] = <UserViewer users={users.getAllUser()}></UserViewer>;
        this.pages["hello"] = <HelloView></HelloView>;
    }

    renderMenu(key: number): JSX.Element[]{
        if (key === 0){
            let labLinks = [
                {name: "Lab 1", uri: this.pagePath + "/opsys/lab1"},
                {name: "Lab 2", uri: this.pagePath + "/opsys/lab2"}, 
                {name: "Lab 3", uri: this.pagePath + "/opsys/lab3"},
                {name: "Lab 4", uri: this.pagePath + "/opsys/lab4"},
            ];
            let settings = [
                {name: "Users", uri: this.pagePath + "/user"},
                {name: "Hello world", uri: this.pagePath + "/hello"}
            ];

            this.navMan.checkLinks(labLinks, this);
            this.navMan.checkLinks(settings, this);

            return [
                <h4 key={0}>Labs</h4>,
                <NavMenu key={1} links={labLinks} onClick={link => this.handleClick(link)}></NavMenu>,
                <h4 key={2}>Settings</h4>,
                <NavMenu key={3} links={settings} onClick={link => this.handleClick(link)}></NavMenu>
            ];
        }
        return [];
    }

    handleClick(link: ILink){
        if (link.uri){
            this.navMan.navigateTo(link.uri);
        }
    }
}

export {StudentPage};
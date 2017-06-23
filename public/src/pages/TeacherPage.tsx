import * as React from "react"
import { NavigationManager, ILink } from "../managers/NavigationManager";
import { UserManager } from "../managers/UserManager";
import { UserView } from "./views/UserView";
import { HelloView } from "./views/HelloView";
import { NavMenu } from "../components";
import { ViewPage } from "./ViewPage";

class TeacherPage extends ViewPage {

    navMan: NavigationManager;
    private pages: {[name: string]: JSX.Element} = {}
    constructor(users: UserManager, navMan: NavigationManager){
        super();

        this.navMan = navMan;
        this.defaultPage = "opsys/lab1";
        this.pages["opsys/lab1"] = <h1>Teacher Lab1</h1>;
        this.pages["opsys/lab2"] = <h1>Teacher Lab2</h1>;
        this.pages["opsys/lab3"] = <h1>Teacher Lab3</h1>;
        this.pages["opsys/lab4"] = <h1>Teacher Lab4</h1>;
        this.pages["user"] = <UserView users={users.getAllUser()}></UserView>;
        this.pages["hello"] = <HelloView></HelloView>;
    }

    handleClick(link: ILink){
        if (link.uri){
            this.navMan.navigateTo(link.uri);
        }
    }

    pageNavigation(page: string): void {
        
    }

    renderMenu(menu: number): JSX.Element[]{
        if (menu === 0){
            let labLinks = [
                {name: "Teacher Lab 1", uri: this.pagePath + "/opsys/lab1"},
                {name: "Teacher Lab 2", uri: this.pagePath + "/opsys/lab2"}, 
                {name: "Teacher Lab 3", uri: this.pagePath + "/opsys/lab3"},
                {name: "Teacher Lab 4", uri: this.pagePath + "/opsys/lab4"},
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
                <h4 key={4}>Settings</h4>,
                <NavMenu key={3} links={settings} onClick={link => this.handleClick(link)}></NavMenu>
            ];
        }
        return [];
    }

    renderContent(page: string): JSX.Element{
        if (page.length === 0){
            page = this.defaultPage;
        }
        if (this.pages[page]){
            return this.pages[page];
        }
        return <h1>404 page not found</h1>
    }
}

export {TeacherPage}
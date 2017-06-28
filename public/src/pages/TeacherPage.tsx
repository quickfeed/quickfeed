import * as React from "react";
import { NavMenu } from "../components";
import { CourseManager, ILink, ILinkCollection, NavigationManager, UserManager } from "../managers";

import { ViewPage } from "./ViewPage";
import { HelloView } from "./views/HelloView";
import { UserView } from "./views/UserView";

import { INavInfo, NavigationHelper } from "../NavigationHelper";

import { CollapsableNavMenu } from "../components/navigation/CollapsableNavMenu";

class TeacherPage extends ViewPage {

    private navMan: NavigationManager;
    private userMan: UserManager;
    private courseMan: CourseManager;

    private pages: { [name: string]: JSX.Element } = {};
    constructor(userMan: UserManager, navMan: NavigationManager, courseMan: CourseManager) {
        super();

        this.navMan = navMan;
        this.userMan = userMan;
        this.courseMan = courseMan;

        this.navHelper.defaultPage = "course/0";
        this.navHelper.registerFunction("course/{course}", this.course);
        this.navHelper.registerFunction("course/{course}/{page}", this.course);
        this.navHelper.registerFunction("user", (navInfo) => {
            return <UserView users={userMan.getAllUser()}></UserView>;
        });
        this.navHelper.registerFunction("user", (navInfo) => {
            return <HelloView></HelloView>;
        });
    }

    public course(info: INavInfo<{ course: string, page?: string }>): JSX.Element {
        if (info.params.page) {
            return <h3>You are know on page {info.params.page.toUpperCase()} in course {info.params.course}</h3>;
        }
        return <h1>Teacher Course {info.params.course}</h1>;
    }

    public generateCollectionFor(link: ILink): ILinkCollection {
        return {
            item: link,
            children: [
                { name: "Results", uri: link.uri + "/results" },
                { name: "Groups", uri: link.uri + "/groups" },
                { name: "Users", uri: link.uri + "/users" },
                { name: "Settings", uri: link.uri + "/settings" },
                { name: "Info", uri: link.uri + "/info" },
            ],
        };
    }

    public renderMenu(menu: number): JSX.Element[] {
        const curUser = this.userMan.getCurrentUser();
        if (curUser) {
            if (menu === 0) {
                const couses = this.courseMan.getCoursesFor(curUser);

                const labLinks: ILinkCollection[] = [];
                couses.forEach((e) => {
                    labLinks.push(this.generateCollectionFor({
                        name: e.tag,
                        uri: this.pagePath + "/course/" + e.id,
                    }));
                });

                const settings: ILink[] = [];

                this.navMan.checkLinkCollection(labLinks, this);
                this.navMan.checkLinks(settings, this);

                return [
                    <h4 key={0}>Courses</h4>,
                    <CollapsableNavMenu
                        key={1}
                        links={labLinks} onClick={(link) => this.handleClick(link)}>
                    </CollapsableNavMenu>,
                    <h4 key={2}>Settings</h4>,
                    <NavMenu key={3} links={settings} onClick={(link) => this.handleClick(link)}></NavMenu>,
                ];
            }
        } else {
            return [<h4>
                you are not logged in;
            </h4>];
        }
        return [];
    }

    public renderContent(page: string): JSX.Element {
        const temp = this.navHelper.navigateTo(page);
        if (temp) {
            return temp;
        }
        return <h1>404 page not found</h1>;
    }

    private handleClick(link: ILink) {
        if (link.uri) {
            this.navMan.navigateTo(link.uri);
        }
    }
}

export { TeacherPage };

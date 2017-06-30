import * as React from "react";
import { DynamicTable, NavMenu } from "../components";
import { CourseManager, ILink, ILinkCollection, NavigationManager, UserManager } from "../managers";

import { ViewPage } from "./ViewPage";
import { HelloView } from "./views/HelloView";
import { UserView } from "./views/UserView";

import { INavInfo, NavigationHelper } from "../NavigationHelper";

import { CollapsableNavMenu } from "../components/navigation/CollapsableNavMenu";
import { CourseStudentState, ICourseStudent, IUser } from "../models";

import { ArrayHelper } from "../helper";

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
        this.navHelper.registerFunction("course/{course}/members", this.courseUsers);
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

    public courseUsers(info: INavInfo<{ course: string }>): JSX.Element {
        const course = parseInt(info.params.course, 10);
        if (!isNaN(course)) {
            const tempCourse = this.courseMan.getCourse(course);
            if (tempCourse) {
                const userIds = this.courseMan.getUserIdsForCourse(tempCourse);
                const users = this.userMan.getUsers(userIds.map((e) => e.personId));

                const all = ArrayHelper.join(userIds, users, (e1, e2) => e1.personId === e2.id);
                console.log(all);
                const acceptedUsers: IUser[] = [];
                const pendingUsers: Array<{ ele1: ICourseStudent, ele2: IUser }> = [];
                all.forEach((ele, id) => {
                    switch (ele.ele1.state) {
                        case CourseStudentState.accepted:
                            acceptedUsers.push(ele.ele2);
                            break;
                        case CourseStudentState.pending:
                            pendingUsers.push(ele);
                            break;
                    }
                });
                return <div>
                    <h3>Users for {tempCourse.name} ({tempCourse.tag})</h3>
                    <UserView users={acceptedUsers}></UserView>
                    <h3>Pending users for {tempCourse.name} ({tempCourse.tag})</h3>
                    {this.createPendingTable(pendingUsers)}
                </div>;
            }
        }
        return <div>404 Page not found</div>;
    }

    public createPendingTable(pendingUsers: Array<{ ele1: ICourseStudent, ele2: IUser }>): JSX.Element {
        return <DynamicTable
            data={pendingUsers}
            header={["ID", "First name", "Last name", "Email", "StudenID", "Action"]}
            selector={(ele: { ele1: ICourseStudent, ele2: IUser }) => [
                ele.ele2.id.toString(),
                ele.ele2.firstName,
                ele.ele2.lastName,
                ele.ele2.email,
                ele.ele2.personId.toString(),
                <span>
                    <button onClick={(e) => {
                        this.courseMan.changeUserState(ele.ele1, CourseStudentState.accepted);
                        this.navMan.refresh();
                    }}
                        className="btn btn-primary">
                        Accept
                    </button>
                    <button onClick={(e) => {
                        this.courseMan.changeUserState(ele.ele1, CourseStudentState.rejected);
                        this.navMan.refresh();
                    }} className="btn btn-danger">
                        Reject
                    </button>
                </span>,
            ]}
        >
        </DynamicTable>;
    }

    public generateCollectionFor(link: ILink): ILinkCollection {
        return {
            item: link,
            children: [
                { name: "Results", uri: link.uri + "/results" },
                { name: "Groups", uri: link.uri + "/groups" },
                { name: "Members", uri: link.uri + "/members" },
                { name: "Settings", uri: link.uri + "/settings" },
                { name: "Course Info", uri: link.uri + "/courseinfo" },
            ],
        };
    }

    public renderMenu(menu: number): JSX.Element[] {
        const curUser = this.userMan.getCurrentUser();
        if (curUser) {
            if (menu === 0) {
                const courses = this.courseMan.getCoursesFor(curUser);

                const labLinks: ILinkCollection[] = [];
                courses.forEach((e) => {
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
            return [];
        }
        return [];
    }

    public renderContent(page: string): JSX.Element {
        if (!this.userMan.getCurrentUser()) {
            return <h1>You are not logged in</h1>;
        }
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

import * as React from "react";

import { DynamicTable, NavMenu } from "../components";

import { CourseManager, ILink, NavigationManager, UserManager } from "../managers";
import { INavInfo } from "../NavigationHelper";
import { ViewPage } from "./ViewPage";
import { UserView } from "./views/UserView";

import { IAssignment, ICourse } from "../models";

class AdminPage extends ViewPage {
    private navMan: NavigationManager;
    private userMan: UserManager;
    private courseMan: CourseManager;

    constructor(navMan: NavigationManager, userMan: UserManager, courseMan: CourseManager) {
        super();
        this.navMan = navMan;
        this.userMan = userMan;
        this.courseMan = courseMan;

        this.navHelper.defaultPage = "users";
        this.navHelper.registerFunction("users", this.users);
        this.navHelper.registerFunction("courses", this.courses);
        this.navHelper.registerFunction("labs", this.labs);
    }

    public users(info: INavInfo<{}>) {
        const allUsers = this.userMan.getAllUser();
        return <div>
            <h1>All Users</h1>
            <UserView users={allUsers}></UserView>
        </div>;
    }

    public courses(info: INavInfo<{}>) {
        const allCourses = this.courseMan.getCourses();
        return <div>
            <h1>All Courses</h1>
            <DynamicTable
                header={["ID", "Tag", "Name"]}
                data={allCourses}
                selector={(e: ICourse) => [e.id.toString(), e.name, e.tag]}
            >
            </DynamicTable>
        </div>;
    }

    public labs(info: INavInfo<{}>) {
        const allCourses = this.courseMan.getCourses();
        const tables = allCourses.map((e, i) => {
            const labs = this.courseMan.getAssignments(e);
            return <div key={i}>
                <h3>Labs for {e.name} ({e.tag})</h3>
                <DynamicTable
                    header={["ID", "Name", "Start", "Deadline", "End"]}
                    data={labs}
                    selector={(lab: IAssignment) => [
                        lab.id.toString(),
                        lab.name,
                        lab.start.toDateString(),
                        lab.deadline.toDateString(),
                        lab.end.toDateString(),
                    ]}>
                </DynamicTable>
            </div>;
        });
        return <div>
            {tables}
        </div>;
    }

    public renderContent(page: string): JSX.Element {
        const temp = this.navHelper.navigateTo(page);
        if (temp) {
            return temp;
        }
        return <h1>404 Page not found</h1>;
    }

    public renderMenu(index: number) {
        if (index === 0) {
            const links: ILink[] = [
                { name: "All Users", uri: this.pagePath + "/users" },
                { name: "All Courses", uri: this.pagePath + "/courses" },
                { name: "All Labs", uri: this.pagePath + "/labs" },
            ];

            this.navMan.checkLinks(links, this);

            return [
                <h4 key={0}>Admin Menu</h4>,
                <NavMenu
                    key={1}
                    links={links}
                    onClick={(e) => {
                        if (e.uri) {
                            this.navMan.navigateTo(e.uri);
                        }
                    }}
                >
                </NavMenu>,
            ];
        }
        return [];
    }
}

export { AdminPage };

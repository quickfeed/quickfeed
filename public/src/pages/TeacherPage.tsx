import * as React from "react";
import { DynamicTable, NavMenu, Results } from "../components";
import { CourseManager, ILink, ILinkCollection, NavigationManager, UserManager } from "../managers";

import { View, ViewPage } from "./ViewPage";
import { HelloView } from "./views/HelloView";
import { UserView } from "./views/UserView";

import { INavInfo } from "../NavigationHelper";

import { CollapsableNavMenu } from "../components/navigation/CollapsableNavMenu";
import { CourseStudentState, IAssignment, ICourse, ICourseStudent, IUser } from "../models";

import { ArrayHelper } from "../helper";

class TeacherPage extends ViewPage {

    private navMan: NavigationManager;
    private userMan: UserManager;
    private courseMan: CourseManager;
    private courses: ICourse[] = [];

    private pages: { [name: string]: JSX.Element } = {};

    constructor(userMan: UserManager, navMan: NavigationManager, courseMan: CourseManager) {
        super();

        this.navMan = navMan;
        this.userMan = userMan;
        this.courseMan = courseMan;
        this.navHelper.defaultPage = "course";

        this.navHelper.registerFunction("course/{course}", this.course);
        this.navHelper.registerFunction("course/{course}/members", this.courseUsers);
        this.navHelper.registerFunction("course/{course}/results", this.results);
        this.navHelper.registerFunction("course/{course}/{page}", this.course);
        this.navHelper.registerFunction("user", async (navInfo) => {
            return <UserView users={await userMan.getAllUser()}></UserView>;
        });
        this.navHelper.registerFunction("user", async (navInfo) => {
            return <HelloView></HelloView>;
        });
    }

    public async init(): Promise<void> {
        this.courses = await this.getCourses();
        this.navHelper.defaultPage = "course/" + (this.courses.length > 0 ? this.courses[0].id.toString() : "");
    }

    public async course(info: INavInfo<{ course: string, page?: string }>): View {
        this.courses = await this.getCourses();
        const courseId = parseInt(info.params.course, 10);
        const course = await this.courseMan.getCourse(courseId);
        if (course) {
            if (info.params.page) {
                return <h3>You are know on page {info.params.page.toUpperCase()} in course {info.params.course}</h3>;
            }
            return <h1>Teacher Course {info.params.course}</h1>;
        }
        return <div>404 Page not found</div>;
    }

    public async results(info: INavInfo<{ course: string }>): View {
        const courseId = parseInt(info.params.course, 10);
        const course = await this.courseMan.getCourse(courseId);
        if (course) {
            const courseStds: ICourseStudent[] =
                await this.courseMan.getUserIdsForCourse(course, CourseStudentState.accepted);
            // TODO: currently userMan.getUsers does not return correct users
            // fix: return a Map for getAllUser method in UserManager
            const students: IUser[] = await this.userMan.getUsers(courseStds.map((e) => e.personId));
            const labs: IAssignment[] = await this.courseMan.getAssignments(courseId);
            return <Results course={course} students={students} labs={labs}></Results>;
        }
        return <div>404 Page not found</div>;
    }

    public async courseUsers(info: INavInfo<{ course: string }>): View {
        this.courses = await this.getCourses();
        const courseId = parseInt(info.params.course, 10);
        const course = await this.courseMan.getCourse(courseId);
        if (course) {
            const userIds = await this.courseMan.getUserIdsForCourse(course);
            const users = await this.userMan.getUsers(userIds.map((e) => e.personId));

            const all = ArrayHelper.join(userIds, users, (e1, e2) => e1.personId === e2.id);
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
            let condPending;
            if (pendingUsers.length > 0) {
                condPending = <div><h3>Pending users</h3>{this.createPendingTable(pendingUsers)}</div>;
            }
            return <div>
                <h1>{course.name}</h1>
                <h3>Registered users</h3>
                <UserView users={acceptedUsers}></UserView>
                {condPending}
            </div>;
        }
        return <div>404 Page not found</div>;
    }

    public createPendingTable(pendingUsers: Array<{ ele1: ICourseStudent, ele2: IUser }>): JSX.Element {
        return <DynamicTable
            data={pendingUsers}
            header={["Name", "Email", "Student ID", "Action"]}
            selector={
                (ele: { ele1: ICourseStudent, ele2: IUser }) => [
                    ele.ele2.firstName + " " + ele.ele2.lastName,
                    <a href={"mailto:" + ele.ele2.email}>{ele.ele2.email}</a>,
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

    public async renderMenu(menu: number): Promise<JSX.Element[]> {
        const curUser = this.userMan.getCurrentUser();
        if (curUser && this.isTeacher(curUser)) {
            if (menu === 0) {
                const courses = await this.courseMan.getCoursesFor(curUser);

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
        }
        return [];
    }

    public async renderContent(page: string): View {
        const curUser: IUser | null = this.userMan.getCurrentUser();
        if (!curUser) {
            return <h1>You are not logged in</h1>;
        } else if (this.isTeacher(curUser)) {
            return await super.renderContent(page);
        }
        return <h1>404 page not found</h1>;
    }

    private handleClick(link: ILink) {
        if (link.uri) {
            this.navMan.navigateTo(link.uri);
        }
    }

    private async getCourses(): Promise<ICourse[]> {
        const curUsr = this.userMan.getCurrentUser();
        if (curUsr) {
            return await this.courseMan.getCoursesFor(curUsr);
        }
        return [];
    }

    private async isTeacher(curUser: IUser): Promise<boolean> {
        return this.userMan.isTeacher(curUser);
    }
}

export { TeacherPage };

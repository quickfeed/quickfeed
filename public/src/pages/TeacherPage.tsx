import * as React from "react";
import {DynamicTable, NavMenu} from "../components";
import {CourseManager, ILink, ILinkCollection, NavigationManager, UserManager} from "../managers";

import {ViewPage} from "./ViewPage";
import {HelloView} from "./views/HelloView";
import {UserView} from "./views/UserView";

import {INavInfo} from "../NavigationHelper";

import {CollapsableNavMenu} from "../components/navigation/CollapsableNavMenu";
import {CourseStudentState, ICourse, ICourseStudent, IUser} from "../models";

import {ArrayHelper} from "../helper";

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
        this.courses = this.getCourses();

        this.navHelper.defaultPage = "course/" + (this.courses.length > 0 ? this.courses[0].id.toString() : "");
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
        const courseId = parseInt(info.params.course, 10);
        const course = this.getCourseById(courseId);
        if (course) {
            if (info.params.page) {
                return <h3>You are know on page {info.params.page.toUpperCase()} in course {info.params.course}</h3>;
            }
            return <h1>Teacher Course {info.params.course}</h1>;
        }
        return <div>404 Page not found</div>;
    }

    public courseUsers(info: INavInfo<{ course: string }>): JSX.Element {
        const courseId = parseInt(info.params.course, 10);
        const course = this.getCourseById(courseId);
        if (course) {
            const userIds = this.courseMan.getUserIdsForCourse(course);
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
                <h3>Users for {course.name} ({course.tag})</h3>
                <UserView users={acceptedUsers}></UserView>
                <h3>Pending users for {course.name} ({course.tag})</h3>
                {this.createPendingTable(pendingUsers)}
            </div>;
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
                {name: "Results", uri: link.uri + "/results"},
                {name: "Groups", uri: link.uri + "/groups"},
                {name: "Members", uri: link.uri + "/members"},
                {name: "Settings", uri: link.uri + "/settings"},
                {name: "Course Info", uri: link.uri + "/courseinfo"},
            ],
        };
    }

    public renderMenu(menu: number): JSX.Element[] {
        const curUser = this.userMan.getCurrentUser();
        if (curUser && this.isTeacher(curUser)) {
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
        }
        return [];
    }

    public renderContent(page: string): JSX.Element {
        const curUser: IUser | null = this.userMan.getCurrentUser();
        if (!curUser) {
            return <h1>You are not logged in</h1>;
        } else if (this.isTeacher(curUser)) {
            const temp = this.navHelper.navigateTo(page);
            if (temp) {
                return temp;
            }
        }
        return <h1>404 page not found</h1>;
    }

    private handleClick(link: ILink) {
        if (link.uri) {
            this.navMan.navigateTo(link.uri);
        }
    }

    private getCourses(): ICourse[] {
        const curUsr = this.userMan.getCurrentUser();
        if (curUsr) {
            return this.courseMan.getCoursesFor(curUsr, 1);
        }
        return [];
    }

    private getCourseById(courseId: number): ICourse | null {
        if (!isNaN(courseId)) {
            const tempCourse = ArrayHelper.find(this.courses, (e, i) => {
                if (e.id === courseId) {
                    return true;
                }
                return false;
            });
            return tempCourse;
        }
        return null;
    }

    private isTeacher(curUser: IUser): boolean {
        return this.userMan.isTeacher(curUser);
    }
}

export {TeacherPage};

import * as React from "react";

import {Button, CourseForm, DynamicTable, NavMenu} from "../components";

import {CourseManager, ILink, NavigationManager, UserManager} from "../managers";
import {INavInfo} from "../NavigationHelper";
import {ViewPage} from "./ViewPage";
import {UserView} from "./views/UserView";

import {IAssignment, ICourse} from "../models";

class AdminPage extends ViewPage {
    private navMan: NavigationManager;
    private userMan: UserManager;
    private courseMan: CourseManager;
    private flashMessages: string[] | null;

    constructor(navMan: NavigationManager, userMan: UserManager, courseMan: CourseManager) {
        super();
        this.navMan = navMan;
        this.userMan = userMan;
        this.courseMan = courseMan;
        this.flashMessages = null;

        this.navHelper.defaultPage = "users";
        this.navHelper.registerFunction("users", this.users);
        this.navHelper.registerFunction("courses", this.courses);
        this.navHelper.registerFunction("labs", this.labs);
        this.navHelper.registerFunction("courses/new", this.newCourse);
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
            <Button className="btn btn-primary pull-right" text="+Create New"
                    onClick={() => this.handleNewCourse()}
            />
            <h1>All Courses</h1>
            <DynamicTable
                header={["ID", "Name", "Tag", "Year/Semester"]}
                data={allCourses}
                selector={(e: ICourse) => [e.id.toString(), e.name, e.tag, e.year]}
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

    public newCourse(info: INavInfo<{}>): JSX.Element {

        let flashHolder: JSX.Element = <div></div>;
        if (this.flashMessages) {
            const errors: JSX.Element[] = [];
            for (const fm of this.flashMessages) {
                errors.push(<li>{fm}</li>);
            }

            flashHolder = <div className="alert alert-danger">
                <h4>{errors.length} errors prohibited Course from being saved: </h4>
                <ul>
                    {errors}
                </ul>
            </div>;
        }

        return (
            <div>
                <h1>Create New Course</h1>
                {flashHolder}
                <CourseForm className="form-horizontal"
                            onSubmit={(formData, errors) => this.createNewCourse(formData, errors)}
                />
            </div>
        );
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
                {name: "All Users", uri: this.pagePath + "/users"},
                {name: "All Courses", uri: this.pagePath + "/courses"},
                {name: "All Labs", uri: this.pagePath + "/labs"},
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

    private handleNewCourse(flashMessage?: string[]): void {
        if (flashMessage) {
            this.flashMessages = flashMessage;
        }
        this.navMan.navigateTo(this.pagePath + "/courses/new");
    }

    private createNewCourse(fd: any, errors: string[]): void {
        if (errors.length === 0) {
            this.courseMan.createNewCourse(fd);
            this.flashMessages = null;
            this.navMan.navigateTo(this.pagePath + "/courses");
        } else {
            this.handleNewCourse(errors);
        }
    }

}

export {AdminPage};

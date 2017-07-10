import * as React from "react";

import { Button, CourseForm, DynamicTable, NavMenu } from "../components";

import { CourseManager, ILink, NavigationManager, UserManager } from "../managers";
import { INavInfo } from "../NavigationHelper";
import { View, ViewPage } from "./ViewPage";

import { CourseView } from "./views/CourseView";
import { UserView } from "./views/UserView";

import { IAssignment } from "../models";

export class AdminPage extends ViewPage {
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
        this.navHelper.registerFunction("courses/{id}/edit", this.editCourse);
    }

    public async users(info: INavInfo<{}>): View {
        const allUsers = await this.userMan.getAllUser();
        return <div>
            <h1>All Users</h1>
            <UserView users={allUsers} userMan={this.userMan} navMan={this.navMan} addSearchOption={true} />
        </div>;
    }

    public async courses(info: INavInfo<{}>): View {
        const allCourses = await this.courseMan.getCourses();
        return <div>
            <Button className="btn btn-primary pull-right" text="+Create New"
                onClick={() => this.handleNewCourse()}
            />
            <h1>All Courses</h1>
            <CourseView courses={allCourses}
                onEditClick={(id: number) => this.handleEditCourseClick(id)}
                onDeleteClick={(id: number) => this.handleDeleteCourseClick(id)}
            />
        </div>;
    }

    public async labs(info: INavInfo<{}>): View {
        const allCourses = await this.courseMan.getCourses();
        const tables: JSX.Element[] = [];
        for (let i = 0; i < allCourses.length; i++) {
            const e = allCourses[i];
            const labs = await this.courseMan.getAssignments(e);
            tables.push(<div key={i}>
                <h3>Labs for {e.name} ({e.code})</h3>
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
            </div>);
        }
        return <div>
            {tables}
        </div>;
    }

    public async newCourse(info: INavInfo<{}>): View {

        let flashHolder: JSX.Element = <div></div>;
        if (this.flashMessages) {
            const errors: JSX.Element[] = [];
            for (let i: number = 0; i < this.flashMessages.length; i++) {
                errors.push(<li key={i}>{this.flashMessages[i]}</li>);
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
                {flashHolder}
                <CourseForm className="form-horizontal"
                    courseMan={this.courseMan}
                    onSubmit={(formData, errors) => this.createNewCourse(formData, errors)}
                />
            </div>
        );
    }

    public async editCourse(info: INavInfo<{ id: string }>): View {
        const courseId = parseInt(info.params.id, 10);
        const course = await this.courseMan.getCourse(courseId);
        if (course) {
            let flashHolder: JSX.Element = <div></div>;
            if (this.flashMessages) {
                const errors: JSX.Element[] = [];
                for (let i: number = 0; i < this.flashMessages.length; i++) {
                    errors.push(<li key={i}>{this.flashMessages[i]}</li>);
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
                    {flashHolder}
                    <CourseForm className="form-horizontal"
                        courseMan={this.courseMan}
                        onSubmit={(formData, errors) => this.updateCourse(formData, errors)}
                        courseData={course}
                    />
                </div>
            );
        }
        return <h1>Page not found</h1>;
    }

    public async renderMenu(index: number): Promise<JSX.Element[]> {
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

    private handleNewCourse(flashMessage?: string[]): void {
        if (flashMessage) {
            this.flashMessages = flashMessage;
        }
        this.navMan.navigateTo(this.pagePath + "/courses/new");
    }

    private async createNewCourse(fd: any, errors: string[]): Promise<void> {
        if (errors.length === 0) {
            // TODO: check returned value from backend and take appropriate action
            const a = await this.courseMan.createNewCourse(fd);
            if (!a) {
                console.log("Failed to create course");
            }
            this.flashMessages = null;
            this.navMan.navigateTo(this.pagePath + "/courses");
        } else {
            this.handleNewCourse(errors);
        }
    }

    private handleEditCourseClick(id: number): void {
        this.navMan.navigateTo(this.pagePath + "/courses/" + id + "/edit");
    }

    private handleDeleteCourseClick(id: number): void {
        if (confirm("Are you sure?")) {
            this.courseMan.deleteCourse(id);
        }
        this.navMan.navigateTo(this.pagePath + "/courses");
    }

    private updateCourse(fd: any, errors: string[]): void {
        if (errors.length === 0) {
            // TODO: check returned value from backend and take appropriate action
            this.courseMan.updateCourse(fd);
            this.flashMessages = null;
            this.navMan.navigateTo(this.pagePath + "/courses");
        } else {
            this.flashMessages = errors;
            this.navMan.navigateTo(this.pagePath + "/courses/" + fd.id + "/edit");
        }
    }

}

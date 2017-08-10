import * as React from "react";

import { BootstrapButton, CourseForm, DynamicTable, NavMenu } from "../components";

import { CourseManager, ILink, NavigationManager, UserManager } from "../managers";
import { INavInfo } from "../NavigationHelper";
import { View, ViewPage } from "./ViewPage";

import { CourseView } from "./views/CourseView";
import { ActionType, UserView } from "./views/UserView";

import { IAssignment, IUserRelation } from "../models";

export class AdminPage extends ViewPage {
    private navMan: NavigationManager;
    private userMan: UserManager;
    private courseMan: CourseManager;

    constructor(navMan: NavigationManager, userMan: UserManager, courseMan: CourseManager) {
        super();
        this.navMan = navMan;
        this.userMan = userMan;
        this.courseMan = courseMan;
        if (!localStorage.getItem("admin")) {
            this.template = "frontpage";
        }
        this.navHelper.defaultPage = "users";
        this.navHelper.registerFunction("users", this.users);
        this.navHelper.registerFunction("courses", this.courses);
        this.navHelper.registerFunction("labs", this.labs);
        this.navHelper.registerFunction("courses/new", this.newCourse);
        this.navHelper.registerFunction("courses/{id}/edit", this.editCourse);
    }

    public async users(info: INavInfo<{}>): View {
        const allUsers = (await this.userMan.getAllUser()).map((user) => {
            return {
                user,
                link: {
                    userid: user.id,
                    courseId: 0,
                    state: 0,
                },
            };
        });

        return <div>
            <h1>All Users</h1>
            <UserView
                users={allUsers}
                optionalActions={(user: IUserRelation) => {
                    if (this.userMan.isAdmin(user.user)) {
                        return [{ uri: "demote", name: "Demote", extra: "danger" }];
                    }
                    return [{ uri: "promote", name: "Promote", extra: "primary" }];

                }}
                linkType={ActionType.InRow}
                actionClick={
                    (user, link) => {
                        this.handleAdminRoleClick(user);
                    }
                }
                userMan={this.userMan}
                navMan={this.navMan}
                searchable={true} />
        </div>;
    }

    public async handleAdminRoleClick(user: IUserRelation): Promise<boolean> {
        if (this.userMan && this.navMan) {
            const res = await this.userMan.changeAdminRole(user.user);
            this.navMan.refresh();
            return res;
        }
        return false;
    }

    public async courses(info: INavInfo<{}>): View {
        const allCourses = await this.courseMan.getCourses();
        return <div>
            <BootstrapButton classType="primary"
                className="pull-right"
                onClick={() => this.handleNewCourse()}>
                New Course
            </BootstrapButton>
            <h1>All Courses</h1>
            <CourseView courses={allCourses}
                onEditClick={(id: number) => this.handleEditCourseClick(id)}
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
                    header={["ID", "Name", /*"Start",*/ "Deadline"/*, "End"*/]}
                    data={labs}
                    selector={(lab: IAssignment) => [
                        lab.id.toString(),
                        lab.name,
                        // lab.start.toDateString(),
                        lab.deadline.toDateString(),
                        // lab.end.toDateString(),
                    ]}>
                </DynamicTable>
            </div>);
        }
        return <div>
            {tables}
        </div>;
    }

    public async newCourse(info: INavInfo<{}>): View {
        const providers = await this.courseMan.getProviders();

        return (
            <div>
                <CourseForm className="form-horizontal"
                    courseMan={this.courseMan}
                    navMan={this.navMan}
                    pagePath={this.pagePath}
                    providers={providers}
                />
            </div>
        );
    }

    public async editCourse(info: INavInfo<{ id: string }>): View {
        const courseId = parseInt(info.params.id, 10);
        const course = await this.courseMan.getCourse(courseId);
        if (course) {
            const providers = await this.courseMan.getProviders();
            return (
                <CourseForm className="form-horizontal"
                    providers={providers}
                    courseMan={this.courseMan}
                    navMan={this.navMan}
                    pagePath={this.pagePath}
                    courseData={course}
                />
            );
        }
        return <h1>Page not found</h1>;
    }

    public async renderMenu(index: number): Promise<JSX.Element[]> {
        if (index === 0) {
            const links: ILink[] = [
                { name: "Users", uri: this.pagePath + "/users" },
                { name: "Courses", uri: this.pagePath + "/courses" },
                { name: "Labs", uri: this.pagePath + "/labs" },
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

    private handleNewCourse(): void {
        this.navMan.navigateTo(this.pagePath + "/courses/new");
    }

    private handleEditCourseClick(id: number): void {
        this.navMan.navigateTo(this.pagePath + "/courses/" + id + "/edit");
    }

}

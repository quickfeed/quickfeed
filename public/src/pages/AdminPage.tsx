import * as React from "react";

import { BootstrapButton, CourseForm, DynamicTable, NavMenu } from "../components";

import { CourseManager, ILink, NavigationManager, UserManager } from "../managers";
import { INavInfo } from "../NavigationHelper";
import { View, ViewPage } from "./ViewPage";

import { CourseView } from "./views/CourseView";
import { ActionType, UserView } from "./views/UserView";

import { Assignment, Enrollment } from "../../proto/ag_pb";
import { formatDate } from "../helper"
import { IUserRelation } from "../models";

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
            const enrol = new Enrollment();
            enrol.setUserid(user.getId());
            enrol.setCourseid(0);
            enrol.setStatus(0);
            return {
                user,
                link: enrol,
            };
        });

        return <div>
            <h1>All Users</h1>
            <UserView
                users={allUsers}
                isCourseList={false}
                optionalActions={(user: IUserRelation) => {
                    if (user.user.getIsadmin()) {
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
            // promote non-admin user to admin, demote otherwise
            console.log("Changing admin role: current role: " + user.user.getIsadmin());
            const res = await this.userMan.changeAdminRole(user.user, !user.user.getIsadmin());
            console.log("Changing to IsAdmin = " + !user.user.getIsadmin());
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
            const labs = await this.courseMan.getAssignments(e.getId());
            tables.push(<div key={i}>
                <h3>Labs for {e.getName()} ({e.getCode()})</h3>
                <DynamicTable
                    header={["ID", "Name", "Lab Type", "Deadline"]}
                    data={labs}
                    selector={(lab: Assignment) => [
                        lab.getId().toString(),
                        lab.getName(),
                        lab.getIsgrouplab() ? "Group lab" : "Individual",
                        formatDate(lab.getDeadline()),
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
                    courseMan={this.courseMan}
                    navMan={this.navMan}
                    pagePath={this.pagePath}
                    providers={providers}
                    courseData={course}
                />
            );
        }
        return <h1>Page not found</h1>;
    }

    public async renderMenu(index: number): Promise<JSX.Element[]> {
        // if user has no teacher scopes, redirect to authorization page
        const authorized = await this.userMan.isAuthorizedTeacher();
        if (!authorized) {
            window.location.href = "https://" + window.location.hostname + "/auth/github-teacher";
        }
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

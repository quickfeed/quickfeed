import * as React from "react";

import { BootstrapButton, CourseGroup, DynamicTable, GroupForm, NavMenu, Results } from "../components";
import { CourseManager, ILink, ILinkCollection, NavigationManager, UserManager } from "../managers";

import { View, ViewPage } from "./ViewPage";
import { UserView } from "./views/UserView";

import { INavInfo } from "../NavigationHelper";

import { CollapsableNavMenu } from "../components/navigation/CollapsableNavMenu";
import {
    CourseGroupStatus,
    CourseUserState,
    IAssignment,
    ICourse,
    ICourseGroup,
    ICourseUserLink,
    IUser,
    IUserCourseWithUser,
    IUserRelation,
    IGroupCourseWithGroup,
} from "../models";

import { MemberView } from "./views/MemberView";
import { GroupResults } from "../components/teacher/GroupResults";

export class TeacherPage extends ViewPage {

    private navMan: NavigationManager;
    private userMan: UserManager;
    private courseMan: CourseManager;
    private courses: ICourse[] = [];

    private pages: { [name: string]: JSX.Element } = {};
    private curUser: IUser | null;
    private refreshState = 0;

    constructor(userMan: UserManager, navMan: NavigationManager, courseMan: CourseManager) {
        super();

        this.navMan = navMan;
        this.userMan = userMan;
        this.courseMan = courseMan;
        this.navHelper.defaultPage = "course";
        this.navHelper.checkAuthentication = () => this.checkAuthentication();

        this.navHelper.registerFunction("courses/{course}", this.course);
        this.navHelper.registerFunction("courses/{course}/members", this.courseUsers);
        this.navHelper.registerFunction("courses/{course}/results", this.results);
        this.navHelper.registerFunction("courses/{course}/groupresults", this.groupresults);
        this.navHelper.registerFunction("courses/{course}/groups", this.groups);
        this.navHelper.registerFunction("courses/{cid}/groups/{gid}/edit", this.editGroup);
    }

    public checkAuthentication(): boolean {
        this.curUser = this.userMan.getCurrentUser();
        if (this.curUser && this.userMan.isTeacher(this.curUser)) {
            return true;
        }
        this.curUser = null;
        return false;
    }

    public async init(): Promise<void> {
        this.courses = await this.getCourses();
        this.navHelper.defaultPage = "courses/" + (this.courses.length > 0 ? this.courses[0].id.toString() : "");
    }

    public async course(info: INavInfo<{ course: string, page?: string }>): View {
        const courseId = parseInt(info.params.course, 10);
        const course = await this.courseMan.getCourse(courseId);
        if (course) {
            if (info.params.page) {
                return <h3>You are know on page {info.params.page.toUpperCase()} in course {info.params.course}</h3>;
            }
            // return <h1>Teacher Course {info.params.course}</h1>;
            let button;
            switch (this.refreshState) {
                case 0:
                    button = <BootstrapButton
                        classType="primary"
                        onClick={(e) => {
                            this.refreshState = 1;
                            this.courseMan.refreshCoursesFor(course.id)
                                .then((value) => {
                                    this.refreshState = 2;
                                    this.navMan.refresh();
                                });
                            this.navMan.refresh();
                        }}
                    >
                        Refresh course info
                </BootstrapButton>;
                    break;
                case 1:
                    button = <BootstrapButton
                        classType="default"
                        disabled={true}>
                        Refreshing Course information
                </BootstrapButton>;
                    break;
                case 2:
                    button = <BootstrapButton
                        classType="success"
                        disabled={false}
                        onClick={(e) => {
                            this.refreshState = 1;
                            this.courseMan.refreshCoursesFor(course.id)
                                .then((value) => {
                                    this.refreshState = 2;
                                    this.navMan.refresh();
                                });
                            this.navMan.refresh();
                        }}
                    >
                        Info refreshed
                </BootstrapButton>;
                    break;
            }
            return <div>
                <h1>Overview for {course.name}</h1>
                {button}
            </div>;
        }
        return <div>404 Page not found</div>;
    }

    public async results(info: INavInfo<{ course: string }>): View {
        const courseId = parseInt(info.params.course, 10);
        const course = await this.courseMan.getCourse(courseId);
        if (course) {
            const students = await this.courseMan.getUsersForCourse(course, this.userMan,
                [
                    CourseUserState.student,
                    CourseUserState.teacher,
                ]);
            const linkedStudents: IUserCourseWithUser[] = [];
            for (const student of students) {
                const temp = await this.courseMan.getStudentCourse(student.user, course);
                if (temp) {
                    linkedStudents.push({ course: temp, user: student.user });
                }
            }
            const labs: IAssignment[] = await this.courseMan.getAssignments(courseId);
            return <Results
                course={course}
                labs={labs}
                students={linkedStudents}
                onApproveClick={async (submissionID: number) => {
                    await this.courseMan.approveSubmission(submissionID);
                    this.navMan.refresh();
                }}
            >
            </Results>;
        }
        return <div>404 Page not found</div>;
    }

    public async groupresults(info: INavInfo<{ course: string }>): View {
        const courseId = parseInt(info.params.course, 10);
        const course = await this.courseMan.getCourse(courseId);
        if (course) {
            var linkedGroups: IGroupCourseWithGroup[] = [];
            const groupCourses = await this.courseMan.getCourseGroups(course.id);
            for (const grpCourse of groupCourses) {
                const grp = await this.courseMan.getGroupCourse(grpCourse, course);
                if (grpCourse && grp) {
                    linkedGroups.push({
                        course: grp,
                        group: grpCourse,

                    });
                }
            }
            const labs: IAssignment[] = await this.courseMan.getAssignments(courseId);
            return <GroupResults
                course={course}
                labs={labs}
                groups={linkedGroups}
                onApproveClick={async (submissionID: number) => {
                    await this.courseMan.approveSubmission(submissionID);
                    this.navMan.refresh();
                }}
            >
            </GroupResults>;
        }
        return <div>404 Page not found</div>;
    }

    public async groups(info: INavInfo<{ course: string }>): View {
        return this.courseFunc(info.params.course, async (course) => {
            const groups = await this.courseMan.getCourseGroups(course.id);
            const approvedGroups: ICourseGroup[] = [];
            const pendingGroups: ICourseGroup[] = [];
            const rejectedGroups: ICourseGroup[] = [];
            for (const grp of groups) {
                switch (grp.status) {
                    case CourseGroupStatus.approved:
                        approvedGroups.push(grp);
                        break;
                    case CourseGroupStatus.pending:
                        pendingGroups.push(grp);
                        break;
                    case CourseGroupStatus.rejected:
                        rejectedGroups.push(grp);
                        break;
                }
            }
            return <CourseGroup
                approvedGroups={approvedGroups}
                pendingGroups={pendingGroups}
                rejectedGroups={rejectedGroups}
                course={course}
                navMan={this.navMan}
                courseMan={this.courseMan}
                pagePath={this.pagePath}
            />;
        })
    }

    public async editGroup(info: INavInfo<{ cid: string, gid: string }>): View {
        const courseId = parseInt(info.params.cid, 10);
        const groupId = parseInt(info.params.gid, 10);
        const course = await this.courseMan.getCourse(courseId);
        const curUser = this.userMan.getCurrentUser();
        const group: ICourseGroup | null = await this.courseMan.getGroup(groupId);
        if (course && curUser && group) {
            const students = await this.courseMan
                .getUsersForCourse(course, this.userMan, [CourseUserState.student, CourseUserState.teacher]);
            return <GroupForm
                className="form-horizontal"
                students={students}
                course={course}
                curUser={curUser}
                courseMan={this.courseMan}
                navMan={this.navMan}
                pagePath={this.pagePath}
                groupData={group}
            />;
        }
        return <div>404 Page not found</div>;
    }

    public async courseUsers(info: INavInfo<{ course: string }>): View {
        const courseId = parseInt(info.params.course, 10);
        const course = await this.courseMan.getCourse(courseId);
        if (course) {
            const all = await this.courseMan.getUsersForCourse(
                course,
                this.userMan);
            const acceptedUsers: IUserRelation[] = [];
            const pendingUsers: IUserRelation[] = [];
            // Sorts all the users to the correct tables, and ignores the rejected once
            // TODO: Maybe move this to the Members view
            all.forEach((user, id) => {
                switch (user.link.state) {
                    case CourseUserState.teacher:
                    case CourseUserState.student:
                        acceptedUsers.push(user);
                        break;
                    case CourseUserState.pending:
                        pendingUsers.push(user);
                        break;
                }
            });
            return <MemberView
                acceptedUsers={acceptedUsers}
                course={course}
                navMan={this.navMan}
                pendingUsers={pendingUsers}
                courseMan={this.courseMan}
            >
            </MemberView>;
        }
        return <div>404 Page not found</div>;
    }

    public generateCollectionFor(link: ILink): ILinkCollection {
        return {
            item: link,
            children: [
                { name: "Results", uri: link.uri + "/results" },
                { name: "Group Results", uri: link.uri + "/groupresults" },
                { name: "Groups", uri: link.uri + "/groups" },
                { name: "Members", uri: link.uri + "/members" },
                // {name: "Settings", uri: link.uri + "/settings" },
                // {name: "Course Info", uri: link.uri + "/courseinfo" },
            ],
        };
    }

    public async renderMenu(menu: number): Promise<JSX.Element[]> {
        const curUser = this.userMan.getCurrentUser();
        if (curUser) {
            if (menu === 0) {
                const states = [CourseUserState.teacher];
                if (this.userMan.isAdmin(curUser)) {
                    states.push(CourseUserState.pending);
                    states.push(CourseUserState.student);
                }
                const courses = await this.courseMan.getCoursesFor(curUser, states);
                // const courses = await this.courseMan.getActiveCoursesFor(curUser);

                const labLinks: ILinkCollection[] = [];
                courses.forEach((e) => {
                    labLinks.push(this.generateCollectionFor({
                        name: e.code,
                        uri: this.pagePath + "/courses/" + e.id,
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

    private async courseFunc(courseParam: string, fn: (course: ICourse) => View): View {
        const courseId = parseInt(courseParam, 10);
        const course = await this.courseMan.getCourse(courseId);
        if (course) {
            return fn(course)
        }
        return <div>404 Page not found</div>;
    }

}

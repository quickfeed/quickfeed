import * as React from "react";
import { CoursesOverview, GroupForm, GroupInfo, NavMenu, SingleCourseOverview, StudentLab } from "../components";

import { CourseManager } from "../managers/CourseManager";
import { ILink, NavigationManager } from "../managers/NavigationManager";
import { UserManager } from "../managers/UserManager";

import {
    CourseUserState, ICourse, ICourseGroup,
    ICourseLinkAssignment, IGroupCourse, IStudentSubmission, 
    IUserCourse, isIUserCourse, IGroupCourse

} from "../models";

import { View, ViewPage } from "./ViewPage";

import { INavInfo } from "../NavigationHelper";

import { CollapsableNavMenu } from "../components/navigation/CollapsableNavMenu";
import { ILinkCollection } from "../managers";
import { EnrollmentView } from "./views/EnrollmentView";

export class StudentPage extends ViewPage {
    private navMan: NavigationManager;
    private userMan: UserManager;
    private courseMan: CourseManager;

    // Single user
    private userCourses: ICourseLinkAssignment[] = [];
    private activeUserCourses: ICourseLinkAssignment[] = [];
    private selectedUserCourse: ICourseLinkAssignment | undefined;

    // Group user
    private GroupUserCourses: ICourseLinkAssignment[] = [];
    private selectedUserGroupCourse: ICourseLinkAssignment | undefined;

    private selectedAssignment: IStudentSubmission | undefined;

    constructor(users: UserManager, navMan: NavigationManager, courseMan: CourseManager) {
        super();

        this.navMan = navMan;
        this.userMan = users;
        this.courseMan = courseMan;

        this.navHelper.defaultPage = "index";

        this.navHelper.checkAuthentication = () => this.checkAuthentication();

        this.navHelper.registerFunction<any>("index", this.index);
        this.navHelper.registerFunction<any>("courses/{courseid:number}", this.course);
        this.navHelper.registerFunction<any>("courses/{courseid:number}/lab/{labid:number}", this.courseWithLab);
        this.navHelper.registerFunction<any>("courses/{courseid:number}/grouplab/{labid:number}", this.courseWithGroupLab);
        this.navHelper.registerFunction<any>("courses/{courseid:number}/members", this.members);
        this.navHelper.registerFunction<any>("courses/{courseid:number}/info", this.courseInformation);
        this.navHelper.registerFunction<any>("courses/{courseid:number}/{page}", this.courseMissing);
        this.navHelper.registerFunction<any>("enroll", this.enroll);
    }

    public checkAuthentication(): boolean {
        const curUser = this.userMan.getCurrentUser();
        if (curUser) {
            return true;
        }
        return false;

    }

    public async index(navInfo: INavInfo<any>): View {
        await this.setupData();
        if (this.activeUserCourses) {
            return (<CoursesOverview
                courseOverview={this.activeUserCourses as IUserCourse[]}
                groupCourseOverview={this.GroupUserCourses as IGroupCourse[]}
                navMan={this.navMan}
            />);
        }
        return <h1>404</h1>;
    }

    public async enroll(navInfo: INavInfo<any>): View {
        await this.setupData();
        const curUser = this.userMan.getCurrentUser();
        if (!curUser) {
            return <h1>404</h1>;
        }
        return <div>
            <h1>Enrollment page</h1>
            <EnrollmentView
                courses={await this.courseMan.getCoursesWithState(curUser)}
                onEnrollmentClick={(course: ICourse) => {
                    this.courseMan.addUserToCourse(curUser, course);
                    this.navMan.refresh();
                }}>
            </EnrollmentView>
        </div >;
    }

    public async course(navInfo: INavInfo<{ courseid: number }>): View {
        await this.setupData();
        this.selectCourse(navInfo.params.courseid);
        this.selectGroupCourse(navInfo.params.courseid);
        if (this.selectedUserCourse) {
            return (<SingleCourseOverview
                courseAndLabs={this.selectedUserCourse as IUserCourse}
                groupAndLabs={this.selectedUserGroupCourse as IGroupCourse}
                onLabClick={(courseId: number, labId: number) => this.handleLabClick(courseId, labId)}
                onGroupLabClick={(courseId: number, labId: number) => this.handleGroupLabClick(courseId, labId)} />);
        }
        return <h1>404 not found</h1>;
    }

    public async courseWithLab(navInfo: INavInfo<{ courseid: number, labid: number }>): View {
        await this.setupData();
        this.selectCourse(navInfo.params.courseid);
        if (this.selectedUserCourse) {
            await this.selectAssignment(navInfo.params.labid);
            if (this.selectedAssignment) {
                return <StudentLab
                    course={this.selectedUserCourse.course}
                    assignment={this.selectedAssignment}
                    showApprove={false}
                    onRebuildClick={() => { }}
                    onApproveClick={() => { }}>
                </StudentLab>;
            }
        }
        return <div>404 not found</div>;
    }

    // TODO - Instead of requesting to server for each time
    // preload grouplab the same way as normal labs are loaded.
    public async courseWithGroupLab(navInfo: INavInfo<{ courseid: number, labid: number }>): View {
        await this.setupData();
        this.selectGroupCourse(navInfo.params.courseid);
        if (this.selectedUserCourse) {
            await this.selectGroupAssignment(navInfo.params.labid);
            if (this.selectedAssignment) {
                return <StudentLab
                    course={this.selectedUserCourse.course}
                    assignment={this.selectedAssignment}
                    showApprove={false}
                    onRebuildClick={() => { }}
                    onApproveClick={() => { }}>
                </StudentLab>;
            }
        }
        // Need to show something if person is not part of group yet.
        return this.courseWithLab(navInfo);
    }

    public async members(navInfo: INavInfo<{ courseid: number }>): View {
        await this.setupData();
        const courseId = navInfo.params.courseid;
        const course = await this.courseMan.getCourse(courseId);
        const curUser = this.userMan.getCurrentUser();
        if (course && curUser) {
            const grp: ICourseGroup | null = await this.courseMan.getGroupByUserAndCourse(curUser.id, course.id);
            if (grp) {
                return <GroupInfo group={grp} course={course} />;
            } else {
                const students = await this.courseMan
                    .getUsersForCourse(course, this.userMan, [CourseUserState.student, CourseUserState.teacher]);
                return <GroupForm className="form-horizontal"
                    students={students}
                    course={course}
                    curUser={curUser}
                    courseMan={this.courseMan}
                    navMan={this.navMan}
                    pagePath={this.pagePath} />;
            }

        }
        return <div>404 not found</div>;
    }

    public async courseInformation(navInfo: INavInfo<{ courseid: number }>): View {
        const informationURL = await this.courseMan.getCourseInformationURL(navInfo.params.courseid);
        if (informationURL === "") {
            return <div> 404 not found</div>;
        }

        // Open new window for course information.
        const win = window.open(informationURL, "_blank");
        // Focus on the new window.
        const test = win ? win.focus() : null;
        // We have to deliver a view back to user, so we deliver a link to the user
        // incase a popup blocker is present.

        // TODO replace the <a href> with something smarter,
        // since it crashes the program if we do it like that and user tries to go back in history
        return <div> Course information found <a href={informationURL}> here </a> </div>;
    }

    public async courseMissing(navInfo: INavInfo<{ courseid: number, page: string }>): View {
        return <div>The page {navInfo.params.page} is not yet implemented</div >;
    }

    public async renderMenu(key: number): Promise<JSX.Element[]> {
        if (key === 0) {
            const coursesLinks: ILinkCollection[] = this.activeUserCourses.map(
                (course, i) => {
                    const allLinks: ILink[] = [];
                    allLinks.push({ name: "Labs" });
                    const labs = course.assignments;
                    const gLabs: ILink[] = [];
                    labs.forEach((lab) => {
                        if (lab.assignment.isgrouplab) {
                            gLabs.push({
                                name: lab.assignment.name,
                                uri: this.pagePath + "/courses/" + course.course.id + "/grouplab/" + lab.assignment.id,
                            });
                        } else {
                            allLinks.push({
                                name: lab.assignment.name,
                                uri: this.pagePath + "/courses/" + course.course.id + "/lab/" + lab.assignment.id,
                            });
                        }
                    });
                    allLinks.push({ name: "Group Labs" });
                    allLinks.push(...gLabs);
                    allLinks.push({ name: "Settings" });
                    allLinks.push({
                        name: "Members", uri: this.pagePath + "/courses/" + course.course.id + "/members",
                    });
                    allLinks.push({
                        name: "Course Info", uri: this.pagePath + "/courses/" + course.course.id + "/info",
                    });
                    return {
                        item: { name: course.course.code, uri: this.pagePath + "/courses/" + course.course.id },
                        children: allLinks,
                    };
                });

            const settings = [
                { name: "Join course", uri: this.pagePath + "/enroll" },
            ];

            this.navMan.checkLinkCollection(coursesLinks, this);
            this.navMan.checkLinks(settings, this);

            return [
                <h4 key={0}>Courses</h4>,
                <CollapsableNavMenu key={1} links={coursesLinks} onClick={(link) => this.handleClick(link)}>
                </CollapsableNavMenu>,
                <h4 key={2}>Settings</h4>,
                <NavMenu key={3} links={settings} onClick={(link) => this.handleClick(link)}></NavMenu>,
            ];
        }
        return [];
    }

    private onlyActiveCourses(studentCourse: IUserCourse[]): IUserCourse[] {
        const userCourses: IUserCourse[] = [];
        studentCourse.forEach((a) => {
            if (a.link && (a.link.state === CourseUserState.student || a.link.state === CourseUserState.teacher)) {
                userCourses.push(a);
            }
        });
        return userCourses;
    }

    // Loads and cache information when user enters a page.
    private async setupData() {
        const curUser = this.userMan.getCurrentUser();
        if (curUser) {
            this.userCourses = await this.courseMan.getStudentCourses(curUser,
                [
                    CourseUserState.student,
                    CourseUserState.teacher,
                ]);
            this.activeUserCourses = this.onlyActiveCourses(this.userCourses as IUserCourse[]);

            // preloading groupdata.
            this.GroupUserCourses = [];

            for (const course of this.activeUserCourses) {
                const group = await this.courseMan.getGroupByUserAndCourse(curUser.id, course.course.id);
                if (group != null) {
                    const groupCourse = await this.courseMan.getGroupCourse(group, course.course);
                    if (groupCourse) {
                        this.GroupUserCourses.push(groupCourse);
                    }
                }
            }
        }
    }

    private selectCourse(course: number) {
        this.selectedUserCourse = undefined;
        this.selectedUserCourse = this.activeUserCourses.find(
            (e) => e.course.id === course);
    }

    private selectGroupCourse(course: number) {
        this.selectedUserGroupCourse = undefined;
        this.selectedUserGroupCourse = this.GroupUserCourses.find(
            (e) => e.course.id === course);
    }

    private selectAssignment(labId: number) {
        if (this.selectedUserCourse) {
            // TODO: Be carefull not to return anything that sould not be able to be returned
            this.selectedAssignment = this.selectedUserCourse.assignments.find(
                (e) => e.assignment.id === labId,
            );
        }
    }

    private selectGroupAssignment(labId: number) {
        if (this.selectedUserGroupCourse) {
            // TODO: Be carefull not to return anything that sould not be able to be returned
            this.selectedAssignment = this.selectedUserGroupCourse.assignments.find(
                (e) => e.assignment.id === labId,
            );
        }
    }

    private handleClick(link: ILink) {
        if (link.uri) {
            this.navMan.navigateTo(link.uri);
        }
    }

    private handleLabClick(courseId: number, labId: number): void {
        this.navMan.navigateTo(this.pagePath + "/courses/" + courseId + "/lab/" + labId);
    }

    private handleGroupLabClick(courseId: number, labId: number): void {
        this.navMan.navigateTo(this.pagePath + "/courses/" + courseId + "/grouplab/" + labId);
    }
}

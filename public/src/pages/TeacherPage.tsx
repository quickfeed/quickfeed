import * as React from "react";

import { BootstrapButton, CourseGroup, GroupForm, NavMenu, Results } from "../components";
import { CourseManager, ILink, ILinkCollection, NavigationManager, UserManager } from "../managers";

import { View, ViewPage } from "./ViewPage";

import { INavInfo } from "../NavigationHelper";

import { CollapsableNavMenu } from "../components/navigation/CollapsableNavMenu";
import {
    IAssignment,
    IGroupCourseWithGroup,
    IUserCourseWithUser,
    IUserRelation,
} from "../models";
import {Course, Group, Enrollment, Repository} from "../../proto/ag_pb";


import { GroupResults } from "../components/teacher/GroupResults";
import { MemberView } from "./views/MemberView";

export class TeacherPage extends ViewPage {

    private navMan: NavigationManager;
    private userMan: UserManager;
    private courseMan: CourseManager;
    private courses: Course[] = [];

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
        this.navHelper.registerFunction("courses/{cid}/info", this.courseInformation);
        this.navHelper.registerFunction("courses/{cid}/assignmentinfo", this.assignmentInformation);
        this.navHelper.registerFunction("courses/{cid}/testinfo", this.testInformation);
        this.navHelper.registerFunction("courses/{cid}/solutioninfo", this.solutionInformation);
    }

    public checkAuthentication(): boolean {
        const curUser = this.userMan.getCurrentUser();
        if (curUser && this.userMan.isTeacher(curUser)) {
            if (!this.userMan.isAuthorizedTeacher()) {
                window.location.assign("https://" + window.location.hostname + "/auth/github-teacher");
                //window.location.href="https://" + window.location.hostname + "/auth/github-teacher";                
            }
            return true;
        }
        return false;
    }

    public async init(): Promise<void> {
        this.courses = await this.getCourses();
        this.navHelper.defaultPage = "courses/" + (this.courses.length > 0 ? this.courses[0].getId().toString() : "");
    }

    public async course(info: INavInfo<{ course: string, page?: string }>): View {
        return this.courseFunc(info.params.course, async (course) => {
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
                            this.courseMan.refreshCoursesFor(course.getId())
                                .then((value) => {
                                    this.refreshState = 2;
                                    this.navMan.refresh();
                                });
                            this.navMan.refresh();
                        }}>
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
                            this.courseMan.refreshCoursesFor(course.getId())
                                .then((value) => {
                                    this.refreshState = 2;
                                    this.navMan.refresh();
                                });
                            this.navMan.refresh();
                        }}>
                        Info refreshed
                    </BootstrapButton>;
                    break;
            }
            return <div>
                <h1>Overview for {course.getName()}</h1>
                {button}
            </div>;
        });
    }

    public async results(info: INavInfo<{ course: string }>): View {
        return this.courseFunc(info.params.course, async (course) => {
            const labs: IAssignment[] = await this.courseMan.getAssignments(course.getId());

            const students = await this.courseMan.getUsersForCourse(course, this.userMan, false,
                [
                    Enrollment.UserStatus.STUDENT,
                    Enrollment.UserStatus.TEACHER,
                ]);
            const linkedStudents: IUserCourseWithUser[] = [];
            for (const student of students) {
                const userCourses = await this.courseMan.getStudentCourseForTeacher(student, course, labs);
                if (userCourses) {
                    linkedStudents.push({ course: userCourses, user: student.user });
                }
            }
            return <Results
                course={course}
                labs={labs}
                students={linkedStudents}
                onApproveClick={async (submissionID: number) => {
                    await this.courseMan.approveSubmission(submissionID);
                    // this.navMan.refresh();
                }}
            >
            </Results>;
        });
    }

    public async groupresults(info: INavInfo<{ course: string }>): View {
        return this.courseFunc(info.params.course, async (course) => {
            const linkedGroups: IGroupCourseWithGroup[] = [];
            const groupCourses = await this.courseMan.getCourseGroups(course.getId());
            const labs: IAssignment[] = await this.courseMan.getAssignments(course.getId());

            for (const grpCourse of groupCourses) {
                const grp = await this.courseMan.getGroupCourseForTeacher(grpCourse, course, labs);
                if (grpCourse && grp) {
                    linkedGroups.push({
                        course: grp,
                        group: grpCourse,

                    });
                }
            }

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
        });
    }

    public async groups(info: INavInfo<{ course: string }>): View {
        return this.courseFunc(info.params.course, async (course) => {
            const groups = await this.courseMan.getCourseGroups(course.getId());
            const approvedGroups: Group[] = [];
            const pendingGroups: Group[] = [];
            const rejectedGroups: Group[] = [];
            for (const grp of groups) {
                switch (grp.getStatus()) {
                    case Group.GroupStatus.APPROVED:
                        approvedGroups.push(grp);
                        break;
                    case Group.GroupStatus.PENDING:
                        pendingGroups.push(grp);
                        break;
                    case Group.GroupStatus.REJECTED:
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
        });
    }

    public async editGroup(info: INavInfo<{ cid: string, gid: string }>): View {
        const courseId = parseInt(info.params.cid, 10);
        const groupId = parseInt(info.params.gid, 10);  

        const course = await this.courseMan.getCourse(courseId);
        const curUser = await this.userMan.getCurrentUser();
        const group: Group | null = await this.courseMan.getGroup(groupId);
        if (course && curUser && group) {
            // get full list of students and teachers
            const students = await this.courseMan
                .getUsersForCourse(course, this.userMan, false, [Enrollment.UserStatus.STUDENT, Enrollment.UserStatus.TEACHER]);
            // get list of users who are not in group
            const freeStudents = await this.courseMan
                .getUsersForCourse(course, this.userMan, true, [Enrollment.UserStatus.STUDENT, Enrollment.UserStatus.TEACHER]);
            return <GroupForm
                className="form-horizontal"
                students={students}
                freeStudents={freeStudents}
                course={course}
                curUser={curUser}
                courseMan={this.courseMan}
                userMan={this.userMan}
                navMan={this.navMan}
                pagePath={this.pagePath}
                groupData={group}
            />;
        }
        return <div>404 Page not found</div>;
    }

    public async courseUsers(info: INavInfo<{ course: string }>): View {
        return this.courseFunc(info.params.course, async (course) => {
            const all = await this.courseMan.getUsersForCourse(course, this.userMan);
            const acceptedUsers: IUserRelation[] = [];
            const pendingUsers: IUserRelation[] = [];
            const rejectedUsers: IUserRelation[] = [];
            // Sorts all the users to the correct tables, and ignores the rejected once
            // TODO: Maybe move this to the Members view
            all.forEach((user, id) => {
                switch (user.link.state) {
                    case Enrollment.UserStatus.TEACHER:
                    case Enrollment.UserStatus.STUDENT:
                        acceptedUsers.push(user);
                        break;
                    case Enrollment.UserStatus.PENDING:
                        pendingUsers.push(user);
                        break;
                    case Enrollment.UserStatus.REJECTED:
                        rejectedUsers.push(user);
                        break;
                }
            });
            return <MemberView
                acceptedUsers={acceptedUsers}
                course={course}
                navMan={this.navMan}
                pendingUsers={pendingUsers}
                rejectedUsers={rejectedUsers}
                courseMan={this.courseMan}
            >
            </MemberView>;
        });
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
                { name: "Repositories" },
                { name: "Course Info", uri: link.uri + "/info" },
                { name: "Assignments", uri: link.uri + "/assignmentinfo" },
                { name: "Tests", uri: link.uri + "/testinfo" },
                { name: "Solutions", uri: link.uri + "/solutioninfo" },
            ],
        };
    }

    public async courseInformation(navInfo: INavInfo<{ cid: string }>): View {
        const courseId = parseInt(navInfo.params.cid, 10);
        const informationURL = await this.courseMan.getRepositoryURL(courseId, Repository.Type.COURSEINFO);
        if (informationURL === "") {
            return <div> 404 not found</div>;
        }

        // Open new window for course information.
        const popup = window.open(informationURL, "_blank");

        if (!popup) {
            return <div> Course information found <a href={informationURL}> here </a> </div>;
        } else {
            this.navMan.navigateTo(this.pagePath + "/" + this.currentPage);
        }

        // If for some reason navigateTo did not succeed, show this error message.
        return <div> Popup blocker prevented the page to load. </div>;
    }

    public async assignmentInformation(navInfo: INavInfo<{ cid: string }>): View {
        const courseId = parseInt(navInfo.params.cid, 10);
        const assignmentURL = await this.courseMan.getRepositoryURL(courseId,
            Repository.Type.ASSIGNMENTS);
        if (assignmentURL === "") {
            return <div> 404 not found</div>;
        }

        // Open new window for course information.
        const popup = window.open(assignmentURL, "_blank");

        if (!popup) {
            return <div> Assignments found <a href={assignmentURL}> here </a> </div>;
        } else {
            this.navMan.navigateTo(this.pagePath + "/" + this.currentPage);
        }

        // If for some reason navigateTo did not succeed, show this error message.
        return <div> Popup blocker prevented the page to load. </div>;
    }

    public async testInformation(navInfo: INavInfo<{ cid: string }>): View {
        const courseId = parseInt(navInfo.params.cid, 10);
        const testInformationURL = await this.courseMan.getRepositoryURL(courseId, Repository.Type.TESTS);
        if (testInformationURL === "") {
            return <div> 404 not found</div>;
        }

        // Open new window for course information.
        const popup = window.open(testInformationURL, "_blank");

        if (!popup) {
            return <div> Test repository found <a href={testInformationURL}> here </a> </div>;
        } else {
            this.navMan.navigateTo(this.pagePath + "/" + this.currentPage);
        }

        // If for some reason navigateTo did not succeed, show this error message.
        return <div> Popup blocker prevented the page to load. </div>;
    }

    public async solutionInformation(navInfo: INavInfo<{ cid: string }>): View {
        const courseId = parseInt(navInfo.params.cid, 10);
        const solutionURL = await this.courseMan.getRepositoryURL(courseId,
            Repository.Type.SOLUTIONS);
        if (solutionURL === "") {
            return <div> 404 not found</div>;
        }

        // Open new window for course information.
        const popup = window.open(solutionURL, "_blank");

        if (!popup) {
            return <div> solution repository found <a href={solutionURL}> here </a> </div>;
        } else {
            this.navMan.navigateTo(this.pagePath + "/" + this.currentPage);
        }

        // If for some reason navigateTo did not succeed, show this error message.
        return <div> Popup blocker prevented the page to load. </div>;
    }

    public async renderMenu(menu: number): Promise<JSX.Element[]> {
        const curUser = this.userMan.getCurrentUser();
        if (curUser) {
            if (menu === 0) {
                const states = [Enrollment.UserStatus.TEACHER];
                if (this.userMan.isAdmin(curUser)) {
                    states.push(Enrollment.UserStatus.PENDING);
                    states.push(Enrollment.UserStatus.STUDENT);
                }
                const courses = await this.courseMan.getCoursesFor(curUser, states);
                // const courses = await this.courseMan.getActiveCoursesFor(curUser);

                const labLinks: ILinkCollection[] = [];
                courses.forEach((e) => {
                    labLinks.push(this.generateCollectionFor({
                        name: e.getCode(),
                        uri: this.pagePath + "/courses/" + e.getId(),
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

    private async getCourses(): Promise<Course[]> {
        const curUsr = this.userMan.getCurrentUser();
        if (curUsr) {
            return await this.courseMan.getCoursesFor(curUsr);
        }
        return [];
    }

    private async courseFunc(courseParam: string, fn: (course: Course) => View): View {
        const courseId = parseInt(courseParam, 10);
        const course = await this.courseMan.getCourse(courseId);
        if (course) {
            return fn(course);
        }
        return <div>404 Page not found</div>;
    }

}

import * as React from "react";

import { BootstrapButton, CourseGroup, GroupForm, Results } from "../components";
import { CourseManager, ILink, ILinkCollection, NavigationManager, UserManager } from "../managers";

import { View, ViewPage } from "./ViewPage";

import { INavInfo } from "../NavigationHelper";

import { Assignment, Course, Enrollment, Group, Repository } from "../../proto/ag_pb";
import { CollapsableNavMenu } from "../components/navigation/CollapsableNavMenu";
import { IUserRelation } from "../models";
import { GroupResults } from "../components/teacher/GroupResults";
import { MemberView } from "./views/MemberView";
import { showLoader } from '../loader';

export class TeacherPage extends ViewPage {

    private navMan: NavigationManager;
    private userMan: UserManager;
    private courseMan: CourseManager;
    private courses: Course[] = [];
    private repositories: Map<Repository.Type, string>;

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
        this.navHelper.registerFunction("courses/{cid}/new_group", this.newGroup);
        this.navHelper.registerFunction("courses/{cid}/groups/{gid}/edit", this.editGroup);
    }

    public checkAuthentication(): boolean {
        const curUser = this.userMan.getCurrentUser();
        if (curUser && (curUser.getIsadmin() || this.userMan.isTeacher())) {
            this.userMan.isAuthorizedTeacher().then((answer) => {
                if (!answer) {
                    window.location.href = "https://" + window.location.hostname + "/auth/github-teacher";
                }
            });
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
                return <h3>You are now on page {info.params.page.toUpperCase()} in course {info.params.course}</h3>;
            }
            let button;
            switch (this.refreshState) {
                case 0:
                    button = <BootstrapButton
                        classType="primary"
                        onClick={(e) => {
                            this.refreshState = 1;
                            this.courseMan.updateAssignments(course.getId())
                                .then(() => {
                                    this.refreshState = 2;
                                    this.navMan.refresh();
                                });
                            this.navMan.refresh();
                        }}>
                        Update Course Assignments
                    </BootstrapButton>;
                    break;
                case 1:
                    button = <BootstrapButton
                        classType="default"
                        disabled={true}>
                        Updating Course Assignments
                    </BootstrapButton>;
                    break;
                case 2:
                    button = <BootstrapButton
                        classType="success"
                        disabled={false}
                        onClick={(e) => {
                            this.refreshState = 1;
                            this.courseMan.updateAssignments(course.getId())
                                .then(() => {
                                    this.refreshState = 2;
                                    this.navMan.refresh();
                                });
                            this.navMan.refresh();
                        }}>
                        Course Assignments Updated
                    </BootstrapButton>;
                    break;
            }
            return <div key="head">
                <h1>Overview for {course.getName()}</h1>
                {button}
            </div>;
        });
    }

    // TODO(meling) consolidate these two result functions?
    public async results(info: INavInfo<{ course: string }>): View {
        return this.courseFunc(info.params.course, async (course) => {
            const labs: Assignment[] = await this.courseMan.getAssignments(course.getId());
            const results = await this.courseMan.getLabsForCourse(course.getId(), false);
            const labResults = await this.courseMan.fillLabLinks(course, results, labs);
            return <Results
                course={course}
                courseURL={await this.getCourseURL(course.getId())}
                labs={labs}
                students={labResults}
                onRebuildClick={async (assignmentID: number, submissionID: number) => {
                    const ans = await this.courseMan.rebuildSubmission(assignmentID, submissionID);
                    // update refreshed submission in the labResults
                    // make a separate method for this that could be used for group and non-group labs
                    this.navMan.refresh();
                    return ans;
                }}
                    onApproveClick={async (submissionID: number, approve: boolean): Promise<boolean> => {
                    return this.approveFunc(submissionID, course.getId(), approve);
                }}
            >
            </Results>;
        });
    }

    public async groupresults(info: INavInfo<{ course: string }>): View {
        return this.courseFunc(info.params.course, async (course) => {
            const results = await this.courseMan.getLabsForCourse(course.getId(), true);
            const labs = await this.courseMan.getAssignments(course.getId());
            const labResults = await this.courseMan.fillLabLinks(course, results, labs);

            return <GroupResults
                course={course}
                courseURL={await this.getCourseURL(course.getId())}
                labs={labs}
                groups={labResults}
                onRebuildClick={async (assignmentID: number, submissionID: number) => {
                    const ans = await this.courseMan.rebuildSubmission(assignmentID, submissionID);
                    this.navMan.refresh();
                    return ans;
                }}
                onApproveClick={async (submissionID: number, approve: boolean): Promise<boolean> => {
                    return this.approveFunc(submissionID, course.getId(), approve);
                }}
            >
            </GroupResults>;
        });
    }

    public async groups(info: INavInfo<{ course: string }>): View {
        return this.courseFunc(info.params.course, async (course) => {
            const groups = await this.courseMan.getGroupsForCourse(course.getId());
            const approvedGroups: Group[] = [];
            const pendingGroups: Group[] = [];
            for (const grp of groups) {
                switch (grp.getStatus()) {
                    case Group.GroupStatus.APPROVED:
                        approvedGroups.push(grp);
                        break;
                    case Group.GroupStatus.PENDING:
                        pendingGroups.push(grp);
                        break;
                }
            }
            return <CourseGroup
                approvedGroups={approvedGroups}
                pendingGroups={pendingGroups}
                course={course}
                courseURL={await this.getCourseURL(course.getId())}
                navMan={this.navMan}
                courseMan={this.courseMan}
                pagePath={this.pagePath}
            />;
        });
    }

    public async newGroup(info: INavInfo<{ cid: number }>): View {
        const courseId = info.params.cid;
        const course = await this.courseMan.getCourse(courseId);
        const curUser = this.userMan.getCurrentUser();

        if (course && curUser) {
            // get full list of students and teachers
            const students = await this.courseMan.getUsersForCourse(
                course, false, [Enrollment.UserStatus.STUDENT, Enrollment.UserStatus.TEACHER]);
            // get list of users who are not in group
            const freeStudents = await this.courseMan.getUsersForCourse(
                course, true, [Enrollment.UserStatus.STUDENT, Enrollment.UserStatus.TEACHER]);
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
            />;
        }
        return showLoader();
    }

    public async editGroup(info: INavInfo<{ cid: string, gid: string }>): View {
        const courseId = parseInt(info.params.cid, 10);
        const groupId = parseInt(info.params.gid, 10);

        const course = await this.courseMan.getCourse(courseId);
        const curUser = this.userMan.getCurrentUser();
        const group: Group | null = await this.courseMan.getGroup(groupId);
        if (course && curUser && group) {
            // get full list of students and teachers
            const students = await this.courseMan.getUsersForCourse(
                course, false, [Enrollment.UserStatus.STUDENT, Enrollment.UserStatus.TEACHER]);
            // get list of users who are not in group
            const freeStudents = await this.courseMan.getUsersForCourse(
                course, true, [Enrollment.UserStatus.STUDENT, Enrollment.UserStatus.TEACHER]);
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
        return showLoader();
    }

    public async courseUsers(info: INavInfo<{ course: string }>): View {
        return this.courseFunc(info.params.course, async (course) => {
            const all = await this.courseMan.getUsersForCourse(course);
            const acceptedUsers: IUserRelation[] = [];
            const pendingUsers: IUserRelation[] = [];
            // TODO: Maybe move this to the Members view
            all.forEach((user) => {
                switch (user.enrollment.getStatus()) {
                    case Enrollment.UserStatus.TEACHER:
                    case Enrollment.UserStatus.STUDENT:
                        acceptedUsers.push(user);
                        break;
                    case Enrollment.UserStatus.PENDING:
                        pendingUsers.push(user);
                        break;
                }
            });

            // sorting accepted user so that teachers show first
            acceptedUsers.sort((x, y) => (x.enrollment.getStatus() < y.enrollment.getStatus()) ? 1 : -1);

            return <MemberView
                course={course}
                courseURL={await this.getCourseURL(course.getId())}
                navMan={this.navMan}
                pendingUsers={pendingUsers}
                acceptedUsers={acceptedUsers}
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
                { name: "New Group", uri: link.uri + "/new_group"},
                { name: "Repositories" },
                { name: "Course Info", uri: this.repositories.get(Repository.Type.COURSEINFO), absolute: true },
                { name: "Assignments", uri: this.repositories.get(Repository.Type.ASSIGNMENTS), absolute: true },
                { name: "Tests", uri: this.repositories.get(Repository.Type.TESTS), absolute: true },
                { name: "Solutions", uri: this.repositories.get(Repository.Type.SOLUTIONS), absolute: true },
            ],
        };
    }

    public async approveFunc(submissionID: number, courseID: number, approve: boolean): Promise<boolean> {
        if (confirm(
            `Do you want to ${this.setConfirmString(approve)} this lab?`,
        )) {
            const ans = await this.courseMan.updateSubmission(courseID, submissionID, approve);
            this.navMan.refresh();
            return ans;
        }
        return false;
    }

    public async renderMenu(menu: number): Promise<JSX.Element[]> {
        const curUser = this.userMan.getCurrentUser();
        const confirmedTeacher = await this.userMan.isTeacher();
        if (curUser) {
            if (menu === 0) {
                const status = [Enrollment.UserStatus.TEACHER];
                if (curUser.getIsadmin() || confirmedTeacher) {
                    status.push(Enrollment.UserStatus.PENDING);
                    status.push(Enrollment.UserStatus.STUDENT);
                }
                const courses = await this.courseMan.getCoursesForUser(curUser, status);

                const labLinks: ILinkCollection[] = [];
                courses.forEach((e) => {
                    labLinks.push(this.generateCollectionFor({
                        name: e.getCode(),
                        uri: this.pagePath + "/courses/" + e.getId(),
                    }));
                });

                this.navMan.checkLinkCollection(labLinks, this);

                return [
                    <h4 key={0}>Courses</h4>,
                    <CollapsableNavMenu
                        key={1}
                        links={labLinks} onClick={(link) => this.handleClick(link)}>
                    </CollapsableNavMenu>,
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
            return this.courseMan.getCoursesForUser(curUsr, []);
        }
        return [];
    }

    private async getCourseURL(courseID: number): Promise<string> {
        const repoMap = await this.courseMan.getRepositories(
            courseID,
            [Repository.Type.COURSEINFO],
            );
        const fullRepoName = repoMap.get(Repository.Type.COURSEINFO);
        return fullRepoName ? fullRepoName.split("course-info")[0] : "";
    }

    private async courseFunc(courseParam: string, fn: (course: Course) => View): View {
        const courseId = parseInt(courseParam, 10);
        const course = await this.courseMan.getCourse(courseId);
        if (course) {
            this.repositories = await this.courseMan.getRepositories(courseId,
                [Repository.Type.COURSEINFO,
                Repository.Type.ASSIGNMENTS,
                Repository.Type.TESTS,
                Repository.Type.SOLUTIONS]);
            return fn(course);
        }
        return showLoader();
    }

    private setConfirmString(approve: boolean): string {
        return approve ? "approve" : "undo approval for";
    }

}

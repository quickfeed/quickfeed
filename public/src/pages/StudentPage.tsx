import * as React from "react";
import { Course, Enrollment, Group, Repository } from "../../proto/ag/ag_pb";
import { CoursesOverview, GroupForm, GroupInfo, SingleCourseOverview, StudentLab } from "../components";
import { CollapsableNavMenu } from "../components/navigation/CollapsableNavMenu";
import { ILinkCollection } from "../managers";
import { CourseManager } from "../managers/CourseManager";
import { ILink, NavigationManager } from "../managers/NavigationManager";
import { UserManager } from "../managers/UserManager";
import { IAllSubmissionsForEnrollment, ISubmissionLink } from "../models";
import { INavInfo } from "../NavigationHelper";
import { View, ViewPage } from "./ViewPage";
import { EnrollmentView } from "./views/EnrollmentView";
import { showLoader } from "../loader";
import { CourseListView } from "./views/CourseListView";
import { sortEnrollmentsByVisibility } from "../componentHelper";

export class StudentPage extends ViewPage {
    private navMan: NavigationManager;
    private userMan: UserManager;
    private courseMan: CourseManager;

    // Single user
    private userCourses: IAllSubmissionsForEnrollment[] = [];
    private activeUserCourses: IAllSubmissionsForEnrollment[] = [];
    private selectedUserCourse: IAllSubmissionsForEnrollment | undefined;

    // Group user
    private GroupUserCourses: IAllSubmissionsForEnrollment[] = [];
    private selectedUserGroupCourse: IAllSubmissionsForEnrollment | undefined;

    private selectedAssignment: ISubmissionLink | undefined;

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
        this.navHelper.registerFunction<any>("courses/{courseid:number}/{page}", this.courseMissing);
        this.navHelper.registerFunction<any>("enroll", this.enroll);
        this.navHelper.registerFunction<any>("courses/list", this.courseList);
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
                courseOverview={this.activeUserCourses as IAllSubmissionsForEnrollment[]}
                groupCourseOverview={this.GroupUserCourses as IAllSubmissionsForEnrollment[]}
                navMan={this.navMan}
            />);
        }
        return showLoader();
    }

    public async enroll(navInfo: INavInfo<any>): View {
        await this.setupData();
        const curUser = this.userMan.getCurrentUser();
        if (!curUser) {
            return showLoader();
        }
        return <div>
            <h1>Enrollment page</h1>
            <EnrollmentView
                courses={await this.courseMan.getAllCoursesForEnrollmentPage(curUser)}
                onEnrollmentClick={(course: Course) => {
                    this.courseMan.addUserToCourse(course, curUser);
                    this.navMan.refresh();
                }}>
            </EnrollmentView>
        </div >;
    }

    public async courseList(navInfo: INavInfo<any>): View {
        await this.setupData();
        const curUser = this.userMan.getCurrentUser();
        if (!curUser) {
            return showLoader();
        }
        return <div>
            <h1>Course list</h1>
            <CourseListView
                enrollments={await this.courseMan.getEnrollmentsForUser(curUser.getId())}
                onChangeClick={(enrol: Enrollment) => {
                    return this.courseMan.updateCourseVisibility(enrol);
                }}
            ></CourseListView>
        </div>
    }

    public async course(navInfo: INavInfo<{ courseid: number }>): View {
        await this.setupData();
        this.selectCourse(navInfo.params.courseid);
        this.selectGroupCourse(navInfo.params.courseid);
        if (this.selectedUserCourse) {
            return (<SingleCourseOverview
                courseAndLabs={this.selectedUserCourse as IAllSubmissionsForEnrollment}
                groupAndLabs={this.selectedUserGroupCourse as IAllSubmissionsForEnrollment}
                onLabClick={(courseId: number, labId: number) => this.handleLabClick(courseId, labId)}
                onGroupLabClick={(courseId: number, labId: number) => this.handleGroupLabClick(courseId, labId)} />);
        }
        return showLoader();
    }

    public async courseWithLab(navInfo: INavInfo<{ courseid: number, labid: number }>): View {
        await this.setupData();
        this.selectCourse(navInfo.params.courseid);
        const curUser = this.userMan.getCurrentUser();
        if (!curUser) {
            return showLoader();
        }
        if (this.selectedUserCourse) {
            this.selectAssignment(navInfo.params.labid, false);
            if (this.selectedAssignment) {
                return <StudentLab
                    studentSubmission={this.selectedAssignment}
                    teacherPageView={false}
                    slipdays={this.selectedUserCourse.enrollment.getSlipdaysremaining()}
                    student={curUser}
                    courseURL={""}
                    onSubmissionRebuild={async (assignmentID: number, submissionID: number) => {
                        const ans = await this.courseMan.rebuildSubmission(assignmentID, submissionID);
                        this.navMan.refresh();
                        return ans ? true : false;
                    }}
                    onSubmissionStatusUpdate={async () => {
                        return false;
                    }}
                    >
                </StudentLab>;
            }
        }
        return showLoader();
    }

    public async courseWithGroupLab(navInfo: INavInfo<{ courseid: number, labid: number }>): View {
        await this.setupData();
        const curUser = this.userMan.getCurrentUser();
        if (!curUser) {
            return showLoader();
        }
        this.selectGroupCourse(navInfo.params.courseid);
        if (this.selectedUserGroupCourse) {
            this.selectAssignment(navInfo.params.labid, true);
            if (this.selectedAssignment) {
                return <StudentLab
                    studentSubmission={this.selectedAssignment}
                    teacherPageView={false}
                    slipdays={this.selectedUserGroupCourse.enrollment.getSlipdaysremaining()}
                    courseURL={""}
                    student={curUser}
                    onSubmissionRebuild={async (assignmentID: number, submissionID: number) => {
                        const ans = await this.courseMan.rebuildSubmission(assignmentID, submissionID);
                        this.navMan.refresh();
                        return ans ? true : false;
                    }}
                    onSubmissionStatusUpdate={async () => {
                        return false;
                    }}
                  >
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
            const grp: Group | null = await this.courseMan.getGroupByUserAndCourse(course.getId(), curUser.getId());
            if (grp) {
                return <GroupInfo group={grp} course={course} />;
            } else {
                const students = await this.courseMan.getUsersForCourse(course, false, false,
                    [Enrollment.UserStatus.STUDENT]);
                const freeStudents = await this.courseMan.getUsersForCourse(course, true, false,
                    [Enrollment.UserStatus.STUDENT]);
                return <GroupForm className="form-horizontal"
                    students={students}
                    freeStudents={freeStudents}
                    course={course}
                    curUser={curUser}
                    courseMan={this.courseMan}
                    userMan={this.userMan}
                    navMan={this.navMan}
                    pagePath={this.pagePath} />;
            }

        }
        return showLoader();
    }

    public async courseMissing(navInfo: INavInfo<{ courseid: number, page: string }>): View {
        return <div>The page {navInfo.params.page} is not yet implemented</div >;
    }

    public async renderMenu(key: number): Promise<JSX.Element[]> {
        if (key === 0) {
            const coursesLinks: ILinkCollection[] = [];
            for (const course of this.activeUserCourses) {
                if (course.enrollment.getState() !== Enrollment.DisplayState.HIDDEN) {
                    const courseID = course.course.getId();
                    const studentLinks: ILink[] = [];
                    const labLinks: ILink[] = [];
                    studentLinks.push({ name: "Labs" });
                    const labs = course.labs;
                    labs.forEach((lab) => {
                        if (lab.assignment.getIsgrouplab()) {
                            labLinks.push({
                                name: lab.assignment.getName(),
                                uri: this.pagePath + "/courses/" + courseID + "/grouplab/" + lab.assignment.getId(),
                            });
                        } else {
                            studentLinks.push({
                                name: lab.assignment.getName(),
                                uri: this.pagePath + "/courses/" + courseID + "/lab/" + lab.assignment.getId(),
                            });
                        }
                    });
                    studentLinks.push({ name: "Group Labs" }, ...labLinks, ...await this.generateRepoLinks(courseID))
                    coursesLinks.push({
                        item: { name: course.course.getCode(), uri: this.pagePath + "/courses/" + courseID },
                        children: studentLinks,
                    });
                }
            }

            this.navMan.checkLinkCollection(coursesLinks, this);

            return [
                <h4 key={0}>Courses</h4>,
                <CollapsableNavMenu key={1} links={coursesLinks} onClick={(link) => this.handleClick(link)}>
                </CollapsableNavMenu>,
            ];
        }
        return [];
    }

    private async generateRepoLinks(courseID: number): Promise<ILink[]> {
        const links: ILink[] = [];
        links.push({ name: "Repositories" });

        const repos =  await this.courseMan.getRepositories(
            courseID,
            [Repository.Type.USER,
            Repository.Type.GROUP,
            Repository.Type.COURSEINFO,
            Repository.Type.ASSIGNMENTS],
            );

        links.push({
            name: "User Repository", uri: repos.get(Repository.Type.USER), absolute: true,
        }, {
            name: "Group Repository", uri: repos.get(Repository.Type.GROUP), absolute: true,
        }, {
            name: "Course Info", uri: repos.get(Repository.Type.COURSEINFO), absolute: true,
        }, {
            name: "Assignments", uri: repos.get(Repository.Type.ASSIGNMENTS), absolute: true,
        }, {
            name: "New Group", uri: this.pagePath + "/courses/" + courseID + "/members",
        });
        return links;
    }


    // Loads and cache information when user enters a page.
    private async setupData() {
        const curUser = this.userMan.getCurrentUser();
        if (curUser) {
            const userEnrolls = sortEnrollmentsByVisibility(await this.courseMan.getEnrollmentsForUser(curUser.getId()), false);
            this.userCourses = [];
            this.activeUserCourses = [];
            this.GroupUserCourses = [];

            for (const enrol of userEnrolls) {
                const crs = enrol.getCourse()
                if (crs) {
                    const newCourseLink: IAllSubmissionsForEnrollment = {
                        course: crs,
                        enrollment: enrol,
                        labs: []
                    }
                    await this.courseMan.fillLinks(newCourseLink, curUser, undefined, crs.getAssignmentsList())
                    this.userCourses.push(newCourseLink);
                    if (enrol.getStatus() === Enrollment.UserStatus.STUDENT || enrol.getStatus() === Enrollment.UserStatus.TEACHER) {
                        this.activeUserCourses.push(newCourseLink);
                    }
                    const grp = enrol.getGroup()
                    if (grp) {
                        await this.courseMan.fillLinks(newCourseLink, undefined, grp, crs.getAssignmentsList())
                        this.GroupUserCourses.push(newCourseLink);
                    }
                }
            }
        }
    }

    private selectCourse(course: number) {
        this.selectedUserCourse = undefined;
        this.selectedUserCourse = this.activeUserCourses.find(
            (e) => e.course.getId() === course);
    }

    private selectGroupCourse(course: number) {
        this.selectedUserGroupCourse = undefined;
        this.selectedUserGroupCourse = this.GroupUserCourses.find(
            (e) => e.course.getId() === course);
    }

    private selectAssignment(labId: number, groupLab: boolean) {
        if (this.selectedUserCourse && !groupLab) {
            this.selectedAssignment = this.selectedUserCourse.labs.find(
                (e) => e.assignment.getId() === labId,
            );
        }
        if (this.selectedUserGroupCourse && groupLab) {
            this.selectedAssignment = this.selectedUserGroupCourse.labs.find(
                (e) => e.assignment.getId() === labId,
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

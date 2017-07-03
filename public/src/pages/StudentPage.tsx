import * as React from "react";
import { CoursesOverview, NavMenu, SingleCourseOverview, StudentLab } from "../components";

import { CourseManager } from "../managers/CourseManager";
import { ILink, NavigationManager } from "../managers/NavigationManager";
import { UserManager } from "../managers/UserManager";

import { IAssignment, ICourse, ICourseStudent, ICoursesWithAssignments, IUser } from "../models";

import { View, ViewPage } from "./ViewPage";
import { HelloView } from "./views/HelloView";
import { UserView } from "./views/UserView";

import { ArrayHelper } from "../helper";
import { INavInfo, INavInfoEvent } from "../NavigationHelper";

import { CollapsableNavMenu } from "../components/navigation/CollapsableNavMenu";
import { ILinkCollection } from "../managers";
import { EnrollmentView } from "./views/EnrollmentView";

class StudentPage extends ViewPage {
    private navMan: NavigationManager;
    private userMan: UserManager;
    private courseMan: CourseManager;

    private selectedCourse: ICourse | null = null;
    private selectedAssignment: IAssignment | null = null;

    private courses: ICourse[] = [];

    private foundId: number = -1;

    constructor(users: UserManager, navMan: NavigationManager, courseMan: CourseManager) {
        super();

        this.navMan = navMan;
        this.userMan = users;
        this.courseMan = courseMan;

        this.navHelper.defaultPage = "index";
        this.navHelper.onPreNavigation.addEventListener((e) => this.setupNavData(e));

        this.navHelper.registerFunction<any>("index", this.index);
        this.navHelper.registerFunction<any>("course/{courseid}", this.course);
        this.navHelper.registerFunction<any>("course/{courseid}/lab/{labid}", this.courseWithLab);
        this.navHelper.registerFunction<any>("course/{coruseid}/{page}", this.courseMissing);
        this.navHelper.registerFunction<any>("enroll", this.enroll);

        // Only for testing purposes
        this.navHelper.registerFunction<any>("user", this.getUsers);
        this.navHelper.registerFunction<any>("hello", (navInfo) => Promise.resolve(<HelloView></HelloView>));
    }

    public async getUsers(navInfo: INavInfo<any>): View {
        await this.setupData();
        return <UserView users={await this.userMan.getAllUser()}>
        </UserView>;
    }

    public async index(navInfo: INavInfo<any>): View {
        await this.setupData();
        const courseOverview: ICoursesWithAssignments[] = await this.getCoursesWithAssignments();
        return (<CoursesOverview course_overview={courseOverview} navMan={this.navMan} />);
    }

    public async enroll(navInfo: INavInfo<any>): View {
        await this.setupData();
        return <div>
            <h1>Enrollment page</h1>
            <EnrollmentView
                courses={await this.courseMan.getCourses()}
                studentCourses={await this.getRelations()}
                curUser={this.userMan.getCurrentUser()}
                onEnrollmentClick={(user: IUser, course: ICourse) => {
                    this.courseMan.addUserToCourse(user, course);
                    this.navMan.refresh();
                }}>
            </EnrollmentView>
        </div >;
    }

    public async course(navInfo: INavInfo<{ courseid: string }>): View {
        await this.setupData();
        this.selectCourse(navInfo.params.courseid);
        if (this.selectedCourse) {
            const courseAndLabs: ICoursesWithAssignments | null = await this.getLabs();
            if (courseAndLabs) {
                return (<SingleCourseOverview courseAndLabs={courseAndLabs} />);
            }
        }
        return <h1>404 not found</h1>;
    }

    public async courseWithLab(navInfo: INavInfo<{ courseid: string, labid: string }>): View {
        await this.setupData();
        this.selectCourse(navInfo.params.courseid);
        if (this.selectedCourse) {
            await this.selectAssignment(navInfo.params.labid);
            if (this.selectedAssignment) {
                return <StudentLab course={this.selectedCourse} assignment={this.selectedAssignment}></StudentLab>;
            }
        }
        return <div>404 not found</div>;
    }

    public async courseMissing(navInfo: INavInfo<{ courseid: string, page: string }>): View {
        return <div>The page {navInfo.params.page} is not yet implemented</div>;
    }

    public async renderMenu(key: number): Promise<JSX.Element[]> {
        if (key === 0) {
            const coursesLinks: ILinkCollection[] = await ArrayHelper.mapAsync(this.courses, async (course, i) => {
                const allLinks: ILink[] = [];
                allLinks.push({ name: "Labs" });
                const labs = await this.getLabsfor(course);
                allLinks.push(...labs.map((lab, ind) => {
                    return { name: lab.name, uri: this.pagePath + "/course/" + course.id + "/lab/" + lab.id };
                }));
                allLinks.push({ name: "Group Labs" });
                allLinks.push({ name: "Settings" });
                allLinks.push({ name: "Members", uri: this.pagePath + "/course/" + course.id + "/members" });
                allLinks.push({ name: "Coruse Info", uri: this.pagePath + "/course/" + course.id + "/info" });
                return {
                    item: { name: course.tag, uri: this.pagePath + "/course/" + course.id },
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

    private async setupNavData(data: INavInfoEvent) {
        await this.setupData();
    }

    private async setupData() {
        this.courses = await this.getCourses();
    }

    private selectCourse(courseId: string) {
        this.selectedCourse = null;
        const course = parseInt(courseId, 10);
        if (!isNaN(course)) {
            this.selectedCourse = ArrayHelper.find(this.courses, (e, i) => {
                if (e.id === course) {
                    this.foundId = i;
                    return true;
                }
                return false;
            });
        }
    }

    private async selectAssignment(labIdString: string) {
        this.selectedAssignment = null;
        const labId = parseInt(labIdString, 10);
        if (this.selectedCourse && !isNaN(labId)) {
            // TODO: Be carefull not to return anything that sould not be able to be returned
            const lab = await this.courseMan.getAssignment(this.selectedCourse, labId);
            if (lab) {
                this.selectedAssignment = lab;
            }
        }
    }

    private handleClick(link: ILink) {
        if (link.uri) {
            this.navMan.navigateTo(link.uri);
        }
    }

    private async  getRelations(): Promise<ICourseStudent[]> {
        const curUsr = this.userMan.getCurrentUser();
        if (curUsr) {
            return this.courseMan.getRelationsFor(curUsr);
        }
        return [];
    }

    private async getCourses(): Promise<ICourse[]> {
        const curUsr = this.userMan.getCurrentUser();
        if (curUsr) {
            return this.courseMan.getCoursesFor(curUsr, 1);
        }
        return [];
    }

    private async getLabsfor(course: ICourse): Promise<IAssignment[]> {
        return this.courseMan.getAssignments(course);

    }

    private async getLabs(): Promise<{ course: ICourse, labs: IAssignment[] } | null> {
        const curUsr = this.userMan.getCurrentUser();
        if (curUsr && !this.selectedCourse) {
            this.selectedCourse = (await this.courseMan.getCoursesFor(curUsr))[0];
        }

        if (this.selectedCourse) {

            const labs = await this.courseMan.getAssignments(this.selectedCourse);
            return { course: this.selectedCourse, labs };
        }
        return null;
    }

    private async getCoursesWithAssignments(): Promise<ICoursesWithAssignments[]> {
        const courseLabs: ICoursesWithAssignments[] = [];
        if (this.courses.length === 0) {
            this.courses = await this.getCourses();
        }

        if (this.courses.length > 0) {
            for (const crs of this.courses) {
                const labs = await this.courseMan.getAssignments(crs);
                const cl = { course: crs, labs };
                courseLabs.push(cl);
            }
            return courseLabs;
        }
        return [];
    }
}

export { StudentPage };

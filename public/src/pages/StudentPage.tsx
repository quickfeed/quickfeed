import * as React from "react";
import { CoursesOverview, DynamicTable, NavMenu, SingleCourseOverview, StudentLab } from "../components";

import { CourseManager } from "../managers/CourseManager";
import { ILink, NavigationManager } from "../managers/NavigationManager";
import { UserManager } from "../managers/UserManager";

import { IAssignment, ICourse, ICourseStudent, ICoursesWithAssignments, IUser } from "../models";

import { ViewPage } from "./ViewPage";
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

    private currentPage: string = "";

    private courses: ICourse[] = [];

    private foundId: number = -1;

    constructor(users: UserManager, navMan: NavigationManager, courseMan: CourseManager) {
        super();

        this.navMan = navMan;
        this.userMan = users;
        this.courseMan = courseMan;

        this.navHelper.defaultPage = "index";
        this.navHelper.onPreNavigation.addEventListener((e) => this.setupData(e));

        this.navHelper.registerFunction<any>("index", this.index);
        this.navHelper.registerFunction<any>("course/{courseid}", this.course);
        this.navHelper.registerFunction<any>("course/{courseid}/lab/{labid}", this.courseWithLab);
        this.navHelper.registerFunction<any>("enroll", this.enroll);

        // Only for testing purposes
        this.navHelper.registerFunction<any>("user", (navInfo) => <UserView users={users.getAllUser()}></UserView>);
        this.navHelper.registerFunction<any>("hello", (INavInfo) => <HelloView></HelloView>);
    }

    public index(navInfo: INavInfo<any>): JSX.Element {
        const courseOverview: ICoursesWithAssignments[] = this.getCoursesWithAssignments();
        return (<CoursesOverview course_overview={courseOverview} navMan={this.navMan} />);
    }

    public enroll(navInfo: INavInfo<any>): JSX.Element {
        return <div>
            <h1>Enrollment page</h1>
            <EnrollmentView
                courses={this.courseMan.getCourses()}
                studentCourses={this.getRelations()}
                curUser={this.userMan.getCurrentUser()}
                onEnrollmentClick={(user: IUser, course: ICourse) => {
                    this.courseMan.addUserToCourse(user, course);
                    this.navMan.refresh();
                }}>
            </EnrollmentView>
        </div >;
    }

    public course(navInfo: INavInfo<{ courseid: string }>): JSX.Element {
        this.selectCourse(navInfo.params.courseid);
        const courseAndLabs: ICoursesWithAssignments | null = this.getLabs();
        if (this.selectedCourse && courseAndLabs) {
            return (<SingleCourseOverview courseAndLabs={courseAndLabs} />);
        }
        return <div>404 not found</div>;
    }

    public courseWithLab(navInfo: INavInfo<{ courseid: string, labid: string }>): JSX.Element {
        this.selectCourse(navInfo.params.courseid);
        if (this.selectedCourse) {
            this.selectAssignment(navInfo.params.labid);
            if (this.selectedAssignment) {
                return <StudentLab course={this.selectedCourse} assignment={this.selectedAssignment}></StudentLab>;
            }
        }
        return <div>404 not found</div>;
    }

    public renderMenu(key: number): JSX.Element[] {
        if (key === 0) {
            const coursesLinks: ILinkCollection[] = this.courses.map((course, i) => {
                return {
                    item: { name: course.tag, uri: this.pagePath + "/course/" + course.id },
                    children: this.getLabsfor(course).map((lab, ind) => {
                        return { name: lab.name, uri: this.pagePath + "/course/" + course.id + "/lab/" + lab.id };
                    }),
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

    public renderContent(page: string): JSX.Element {
        const pageContent = this.navHelper.navigateTo(page);
        this.currentPage = page;
        if (pageContent) {
            return pageContent;
        }
        return <div>404 Not found</div>;
    }

    private setupData(data: INavInfoEvent) {
        this.courses = this.getCourses();
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

    private selectAssignment(labIdString: string) {
        this.selectedAssignment = null;
        const labId = parseInt(labIdString, 10);
        if (this.selectedCourse && !isNaN(labId)) {
            // TODO: Be carefull not to return anything that sould not be able to be returned
            const lab = this.courseMan.getAssignment(this.selectedCourse, labId);
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

    private getRelations(): ICourseStudent[] {
        const curUsr = this.userMan.getCurrentUser();
        if (curUsr) {
            return this.courseMan.getRelationsFor(curUsr);
        }
        return [];
    }

    private getCourses(): ICourse[] {
        const curUsr = this.userMan.getCurrentUser();
        if (curUsr) {
            return this.courseMan.getCoursesFor(curUsr, 1);
        }
        return [];
    }

    private getLabsfor(course: ICourse): IAssignment[] {
        return this.courseMan.getAssignments(course);

    }

    private getLabs(): { course: ICourse, labs: IAssignment[] } | null {
        const curUsr = this.userMan.getCurrentUser();
        if (curUsr && !this.selectedCourse) {
            this.selectedCourse = this.courseMan.getCoursesFor(curUsr)[0];
        }

        if (this.selectedCourse) {

            const labs = this.courseMan.getAssignments(this.selectedCourse);
            return { course: this.selectedCourse, labs };
        }
        return null;
    }

    private getCoursesWithAssignments(): ICoursesWithAssignments[] {
        const courseLabs: ICoursesWithAssignments[] = [];
        if (this.courses.length === 0) {
            this.courses = this.getCourses();
        }

        if (this.courses.length > 0) {
            for (const crs of this.courses) {
                const labs = this.courseMan.getAssignments(crs);
                const cl = { course: crs, labs };
                courseLabs.push(cl);
            }
            return courseLabs;
        }
        return [];
    }
}

export { StudentPage };

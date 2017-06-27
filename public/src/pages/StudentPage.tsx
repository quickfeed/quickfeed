import * as React from "react";
import { NavDropdown, NavMenu, StudentLab, CourseOverview } from "../components";

import { CourseManager } from "../managers/CourseManager";
import { ILink, NavigationManager } from "../managers/NavigationManager";
import { UserManager } from "../managers/UserManager";

import { IAssignment, ICourse, ICoursesWithAssignments } from "../models";

import { ViewPage } from "./ViewPage";
import { HelloView } from "./views/HelloView";
import { UserView } from "./views/UserView";

import { ArrayHelper } from "../helper";
import { INavInfo, INavInfoEvent, NavigationHelper } from "../NavigationHelper";

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

        // Only for testing purposes
        this.navHelper.registerFunction<any>("user", (navInfo) => <UserView users={users.getAllUser()}></UserView>);
        this.navHelper.registerFunction<any>("hello", (INavInfo) => <HelloView></HelloView>);
    }

    public index(navInfo: INavInfo<any>): JSX.Element {
        const course_overview: ICoursesWithAssignments[] = this.getCoursesWithAssignments();
        return(<CourseOverview course_overview={course_overview} navMan={this.navMan}/>);
    }

    public course(navInfo: INavInfo<{ courseid: string }>): JSX.Element {
        this.selectCourse(navInfo.params.courseid);
        if (this.selectedCourse) {
            return <div>This is the CourseView for {this.selectedCourse.name}</div>;
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
            const coursesLinks: ILink[] = this.courses.map((e, i) => {
                return { name: e.tag, uri: this.pagePath + "/course/" + e.id };
            });
            const labs = this.getLabs();
            let labLinks: ILink[] = [];
            if (labs) {
                labLinks = labs.labs.map((l, i) => {
                    return { name: l.name, uri: this.pagePath + "/course/" + labs.course.id + "/lab/" + l.id };
                });
            }

            const settings = [
                { name: "Users", uri: this.pagePath + "/user" },
                { name: "Hello world", uri: this.pagePath + "/hello" },
            ];

            this.navMan.checkLinks(labLinks, this);
            this.navMan.checkLinks(settings, this);

            return [
                <h4>Course</h4>,
                <NavDropdown
                    key={1}
                    selectedIndex={this.foundId}
                    items={coursesLinks}
                    itemClick={(link) => { this.handleClick(link); }}>
                </NavDropdown>,
                <h4 key={2}>Labs</h4>,
                <NavMenu key={3} links={labLinks} onClick={(link) => this.handleClick(link)}></NavMenu>,
                <h4 key={4}>Settings</h4>,
                <NavMenu key={5} links={settings} onClick={(link) => this.handleClick(link)}></NavMenu>,
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

    private getCourses(): ICourse[] {
        const curUsr = this.userMan.getCurrentUser();
        if (curUsr) {
            return this.courseMan.getCoursesFor(curUsr);
        }
        return [];
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
        const course_labs:ICoursesWithAssignments[] = [];
        if (this.courses.length === 0){
            this.courses = this.getCourses();
        }

        if (this.courses.length > 0 ) {
            for (const course of this.courses){
                const labs = this.courseMan.getAssignments(course);
                const cl = { course: course, labs };
                course_labs.push(cl)
            }
            return course_labs;
        }
        return [];
    }
}

export { StudentPage };

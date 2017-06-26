import * as React from "react";
import { NavDropdown, NavMenu, StudentLab } from "../components";
import { CourseManager } from "../managers/CourseManager";
import { ILink, NavigationManager } from "../managers/NavigationManager";
import { UserManager } from "../managers/UserManager";
import { IAssignment, ICourse} from "../models";

import { ViewPage } from "./ViewPage";
import { HelloView } from "./views/HelloView";
import { UserView } from "./views/UserView";

class StudentPage extends ViewPage {
    private navMan: NavigationManager;
    private userMan: UserManager;
    private courseMan: CourseManager;

    private pages: { [key: string]: JSX.Element } = {};

    private selectedCourse: ICourse | null = null;
    private selectedAssignment: IAssignment | null = null;

    private currentPage: string = "";

    constructor(users: UserManager, navMan: NavigationManager, courseMan: CourseManager) {
        super();

        this.navMan = navMan;
        this.userMan = users;
        this.courseMan = courseMan;
        this.defaultPage = "opsys/lab1";

        this.pages["opsys/lab1"] = <h1>Lab1</h1>;
        this.pages["opsys/lab2"] = <h1>Lab2</h1>;
        this.pages["opsys/lab3"] = <h1>Lab3</h1>;
        this.pages["opsys/lab4"] = <h1>Lab4</h1>;
        this.pages.user = <UserView users={users.getAllUser()}></UserView>;
        this.pages.hello = <HelloView></HelloView>;
    }

    public pageNavigation(page: string): void {
        this.currentPage = page;
        const parts = this.navMan.getParts(page);
        if (parts.length > 1) {
            if (parts[0] === "course") {
                const course = parseInt(parts[1], 10);
                if (!isNaN(course) && (!this.selectedCourse || this.selectedCourse.id !== course)) {
                    this.selectedCourse = this.courseMan.getCourse(course);
                }

                if (parts.length > 3 && this.selectedCourse) {
                    const labId = parseInt(parts[3], 10);
                    if (!isNaN(labId)) {
                        // TODO: Be carefull not to return anything that sould not be able to be returned
                        const lab = this.courseMan.getAssignment({ id: 0, name: "", tag: "" }, labId);
                        if (lab) {
                            this.selectedAssignment = lab;
                        }
                    }
                }
            }
        }
    }

    public renderMenu(key: number): JSX.Element[] {
        if (key === 0) {
            const courses = this.getCourses();
            const coursesLinks: ILink[] = [];
            for (const a of courses) {
                coursesLinks.push({ name: a.tag, uri: this.pagePath + "/course/" + a.id });
            }
            const labs = this.getLabs();
            const labLinks: ILink[] = [];
            if (labs) {
                for (const l of labs.labs) {
                    labLinks.push({ name: l.name, uri: this.pagePath + "/course/" + labs.course.id + "/lab/" + l.id });
                }
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
                    selectedIndex={0}
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
        if (page.length === 0) {
            page = this.defaultPage;
        }
        if (this.pages[page]) {
            return this.pages[page];
        }

        if (this.selectedAssignment && this.selectedCourse) {
            return <StudentLab course={this.selectedCourse} assignment={this.selectedAssignment}></StudentLab>;
        }
        return <div>404 Not found</div>;
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
}

export { StudentPage };

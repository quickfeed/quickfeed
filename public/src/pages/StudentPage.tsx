import { NavigationManager, ILink } from "../managers/NavigationManager";
import { UserManager } from "../managers/UserManager";
import * as React from "react";
import { UserView } from "./views/UserView";
import { HelloView } from "./views/HelloView";
import { NavMenu, StudentLab } from "../components";
import { ViewPage } from "./ViewPage";
import { CourseManager } from "../managers/CourseManager";
import { IAssignment, ICourse } from "../models";



class StudentPage extends ViewPage {
    private navMan: NavigationManager;
    private userMan: UserManager;
    private courseMan: CourseManager;

    private pages: { [key: string]: JSX.Element} = {};

    constructor(users: UserManager, navMan: NavigationManager, courseMan: CourseManager){
        super();

        this.navMan = navMan;
        this.userMan = users;
        this.courseMan = courseMan;
        this.defaultPage = "opsys/lab1";
        this.pages["opsys/lab1"] = <h1>Lab1</h1>;
        this.pages["opsys/lab2"] = <h1>Lab2</h1>;
        this.pages["opsys/lab3"] = <h1>Lab3</h1>;
        this.pages["opsys/lab4"] = <h1>Lab4</h1>;
        this.pages["user"] = <UserView users={users.getAllUser()}></UserView>;
        this.pages["hello"] = <HelloView></HelloView>;
    }

    getLabs(): {course: ICourse, labs: IAssignment[]} | null {
        let curUsr = this.userMan.getCurrentUser();
        if (curUsr){
            let courses = this.courseMan.getCoursesFor(curUsr);
            let labs = this.courseMan.getAssignments(courses[0]);
            return { course: courses[0], labs: labs };
        }
        return null;
    }

    private selectedCourse: ICourse | null = null;
    private selectedAssignment: IAssignment | null = null;

    pageNavigation(page: string): void{
        let parts = this.navMan.getParts(page);
        if (parts.length > 1){
            if (parts[0] === "course"){
                let course = parseInt(parts[1]);
                if (!isNaN(course) && (!this.selectedCourse || this.selectedCourse.id !== course)){
                    this.selectedCourse = this.courseMan.getCourse(course);
                }

                if (parts.length > 3 && this.selectedCourse){
                    let labId = parseInt(parts[3]);
                    if (!isNaN(labId)){
                        // TODO: Be carefull not to return anything that sould not be able to be returned
                        let lab = this.courseMan.getAssignment({id:0, name: "", tag: ""}, labId);
                        if (lab){
                            this.selectedAssignment = lab;
                        }
                    }
                }
            }
        }
    }

    renderMenu(key: number): JSX.Element[]{
        if (key === 0){
            let labs = this.getLabs();
            let labLinks: ILink[] = []
            if (labs){
                for(let l of labs.labs){
                    labLinks.push({name: l.name, uri: this.pagePath + "/course/" + labs.course.id + "/lab/" + l.id});
                }
            }

            let settings = [
                {name: "Users", uri: this.pagePath + "/user"},
                {name: "Hello world", uri: this.pagePath + "/hello"}
            ];

            this.navMan.checkLinks(labLinks, this);
            this.navMan.checkLinks(settings, this);

            return [
                <h4 key={0}>Labs</h4>,
                <NavMenu key={1} links={labLinks} onClick={link => this.handleClick(link)}></NavMenu>,
                <h4 key={2}>Settings</h4>,
                <NavMenu key={3} links={settings} onClick={link => this.handleClick(link)}></NavMenu>
            ];
        }
        return [];
    }

    renderContent(page: string): JSX.Element{
        if (page.length === 0){
            page = this.defaultPage;
        }
        if (this.pages[page]){
            return this.pages[page];
        }
        if (this.selectedAssignment && this.selectedCourse){
            return <StudentLab course={this.selectedCourse} assignment={this.selectedAssignment}></StudentLab>
        }
        return <div>404 Not found</div>
    }

    handleClick(link: ILink){
        if (link.uri){
            this.navMan.navigateTo(link.uri);
        }
    }
}

export {StudentPage};
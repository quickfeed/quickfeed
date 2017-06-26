import { NavigationManager, ILink } from "../managers/NavigationManager";
import { UserManager } from "../managers/UserManager";
import * as React from "react";
import { UserView } from "./views/UserView";
import { HelloView } from "./views/HelloView";
import { NavMenu } from "../components";
import { ViewPage } from "./ViewPage";
import { CourseManager } from "../managers/CourseManager";
import { IAssignment, ICourse, ITestCases, ILabInfo } from "../models";
import {LabResultView} from "./views/LabResultView";


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

    renderMenu(key: number): JSX.Element[]{
        if (key === 0){
            let labs = this.getLabs();
            let labLinks: ILink[] = []
            if (labs){
                for(let l of labs.labs){
                    labLinks.push({name: l.name, uri: this.pagePath + "/course/" + labs.course.tag + "/lab/" + l.id});
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
        let parts = this.navMan.getParts(page);
        if (parts.length > 1){
            if (parts[0] === "course"){
                let course_tag = parts[1];
                let course = this.courseMan.getCourseByTag(course_tag);
                if (parts.length > 3){
                    let labId = parseInt(parts[3]);
                    if (course !== null && labId !== undefined){
                        // TODO: Be carefull not to return anything that sould not be able to be returned
                        let lab = this.courseMan.getAssignment({id:0, name: "", tag: ""}, labId);
                        console.log(lab);
                        if (lab){
                            // TODO: fetch real data from backend database for corresponding course assignment
                            let testCases: ITestCases[] = [
                                {name: "Test Case 1", score: 60, points: 100, weight: 1},
                                {name: "Test Case 2", score: 50, points: 100, weight: 1},
                                {name: "Test Case 3", score: 40, points: 100, weight: 1},
                                {name: "Test Case 4", score: 30, points: 100, weight: 1},
                                {name: "Test Case 5", score: 20, points: 100, weight: 1}
                            ];

                            let labInfo: ILabInfo = {
                                lab: lab.name,
                                course: course.name,
                                score: 50,
                                weight: 100,
                                test_cases: testCases,
                                pass_tests: 10,
                                fail_tests: 20,
                                exec_time: 0.33,
                                build_time: new Date(2017, 5, 25),
                                build_id: 10
                            };
                            return <LabResultView labInfo={labInfo}></LabResultView>
                        }
                        return <h1>Could not find that lab</h1>
                    }
                }
            }
        }

        return <div></div>
    }

    handleClick(link: ILink){
        if (link.uri){
            this.navMan.navigateTo(link.uri);
        }
    }
}

export {StudentPage};
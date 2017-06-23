import * as React from "react"
import * as ReactDOM from "react-dom"

import { NavBar, Row } from "./components"
import { ILink, NavigationManager, INavEvent } from "./managers/NavigationManager";
import { ViewPage } from "./pages/ViewPage";
import { UserManager } from "./managers/UserManager";
import { StudentPage } from "./pages/StudentPage";
import { TempDataProvider } from "./managers/TempDataProvider";
import { CourseManager } from "./managers/CourseManager";
import { HomePage } from "./pages/HomePage";
import { ErrorPage } from "./pages/ErrorPage";
import { TeacherPage } from "./pages/TeacherPage";
import {HelpPage} from "./pages/HelpPage";

let topLinks: ILink[] = [
    { name: "Teacher", uri: "app/teacher/", active: false },
    { name: "Student", uri: "app/student/", active: false },
    { name: "Admin",   uri: "app/admin",    active: false },
    { name: "Help",   uri: "app/help",    active: false }
]

interface AutoGraderState{
    activePage?: ViewPage;
    topLink: ILink[];
}

interface AutoGraderProps{
    userManager: UserManager;
    navigationManager: NavigationManager;
}

class AutoGrader extends React.Component<AutoGraderProps, AutoGraderState>{
    private userManager: UserManager;
    private navMan: NavigationManager;
    private subPage: string;
    
    constructor(props: any){
        super();
        
        this.userManager = props.userManager;
        this.navMan = props.navigationManager;

        this.state = {
            activePage: undefined,
            topLink: topLinks
        }
        
        this.navMan.onNavigate.addEventListener((e: INavEvent) => {
            this.subPage = e.subPage;
            let old = this.state.activePage;
            let tempLink = this.state.topLink.slice();
            this.checkLinks(tempLink);
            this.setState({activePage: e.page, topLink: tempLink});
        });
    }

    componentDidMount() {
        let curUrl = location.pathname;
        if (curUrl === "/"){
            this.navMan.navigateToDefault();
        }
        else{
            this.navMan.navigateTo(curUrl);
        }
    }

    private handleClick(link: ILink){
        if (link.uri){
            this.navMan.navigateTo(link.uri);
        }
        else{
            console.warn("Warning! Empty link detected", link);
        }
    }

    private renderActiveMenu(menu: number): JSX.Element[] | string {
        if (this.state.activePage){
            return this.state.activePage.renderMenu(menu);
        }
        return "";
    }

    private renderActivePage(page: string):JSX.Element{
        let curPage = this.state.activePage;
        if (curPage) {
            return curPage.renderContent(page);
            /*if(!curPage.pages[curPage.defaultPage]){
                console.warn("Warning! Missing default page for " + (curPage as any).constructor.name, curPage);
            }
            if (curPage.pages[page]){
                return curPage.pages[page];
            }
            else if (page.length === 0 && curPage.pages[curPage.defaultPage]){
                return curPage.pages[curPage.defaultPage];
            }*/
        }
        return <h1>404 Page not found</h1>
    }

    private checkLinks(links: ILink[]): void{
        this.navMan.checkLinks(links);
    }

    private renderTemplate(name: string | null){
        let body: JSX.Element;
        //console.log("rendering template: " + name);
        switch(name){
            case "frontpage":
                body = (
                <Row className="container-fluid">
                    <div className="col-xs-12">
                        { this.renderActivePage(this.subPage) }
                    </div>
                </Row>
                );
            default:
                body = (
                <Row className="container-fluid">
                    <div className="col-md-2 col-sm-3 col-xs-12">
                        { this.renderActiveMenu(0) }
                    </div>
                    <div className="col-md-10 col-sm-9 col-xs-12">
                        { this.renderActivePage(this.subPage) }
                    </div>
                </Row>
                );
        }
        return (
        <div>
            <NavBar id="top-bar" isFluid={false} isInverse={true} links={topLinks} onClick={(link) => this.handleClick(link)} user={this.userManager.getCurrentUser()} brandName="Auto Grader"></NavBar>
            {body}
        </div>);
    }

    render(){
        if (this.state.activePage){
            return this.renderTemplate(this.state.activePage.template)
        }
        else{
            return <h1>404 not found</h1>;
        }
    }
}




/**
 * @description The main entry point for the application. No other code should be executet outside this function
 */
function main(){
    
    let tempData = new TempDataProvider();

    let userMan = new UserManager(tempData);
    let courseMan = new CourseManager(tempData);
    let navMan = new NavigationManager(history);

    (window as any).debugData = {tempData: tempData, userMan: userMan, courseMan: courseMan, navMan: navMan};

    let user = userMan.tryLogin("test@testersen.no","1234");

    navMan.setDefaultPath("app/home");
    navMan.registerPage("app/home", new HomePage());
    navMan.registerPage("app/student", new StudentPage(userMan, navMan, courseMan));
    navMan.registerPage("app/teacher", new TeacherPage(userMan, navMan));
    navMan.registerPage("app/help", new HelpPage(navMan));

    navMan.registerErrorPage(404, new ErrorPage());
    navMan.onNavigate.addEventListener((e ) => {console.log(e)});

    ReactDOM.render(
        <AutoGrader userManager={userMan} navigationManager={navMan}>

        </AutoGrader>,
        document.getElementById("root")
    );
}
main();

// <NavMenu links={ this.state.pages[this.state.currentPage] }  onClick={ (link) => { this.handleClick(link) } }></NavMenu><NavMenuFormatable items={ this.userManager.getAllUser() } formater={ (item: IUser) => { return item.firstName + " " + item.lastName }} onClick={(item: IUser) => { console.log("You clicked on: "); console.log(item);}}></NavMenuFormatable>
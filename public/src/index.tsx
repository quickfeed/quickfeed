import * as React from "react";
import * as ReactDOM from "react-dom";

import { NavBar, Row } from "./components";
import { CourseManager, ILink, INavEvent, NavigationManager, TempDataProvider, UserManager } from "./managers";

import { ErrorPage } from "./pages/ErrorPage";
import { HelpPage } from "./pages/HelpPage";
import { HomePage } from "./pages/HomePage";
import { StudentPage } from "./pages/StudentPage";
import { TeacherPage } from "./pages/TeacherPage";
import { ViewPage } from "./pages/ViewPage";

import { IUser } from "./models";
import { AdminPage } from "./pages/AdminPage";

import { NavBarLogin } from "./components/navigation/NavBarLogin";
import { NavBarMenu } from "./components/navigation/NavBarMenu";
import { LoginPage } from "./pages/LoginPage";

import { ServerProvider } from "./managers/ServerProvider";

interface IAutoGraderState {
    activePage?: ViewPage;
    topLinks: ILink[];
    curUser: IUser | null;
}

interface IAutoGraderProps {
    userManager: UserManager;
    navigationManager: NavigationManager;
}

class AutoGrader extends React.Component<IAutoGraderProps, IAutoGraderState> {
    private userMan: UserManager;
    private navMan: NavigationManager;
    private subPage: string;

    constructor(props: any) {
        super();

        this.userMan = props.userManager;
        this.navMan = props.navigationManager;

        const curUser = this.userMan.getCurrentUser();

        this.state = {
            activePage: undefined,
            topLinks: this.generateTopLinksFor(curUser),
            curUser,
        };

        this.navMan.onNavigate.addEventListener((e: INavEvent) => {
            this.subPage = e.subPage;
            const old = this.state.activePage;
            const tempLink = this.state.topLinks.slice();
            this.checkLinks(tempLink);
            this.setState({ activePage: e.page, topLinks: tempLink });
        });

        this.userMan.onLogin.addEventListener((e) => {
            this.setState({
                curUser: e.user,
                topLinks: this.generateTopLinksFor(e.user),
            });
        });

        this.userMan.onLogout.addEventListener((e) => {
            this.setState({
                curUser: null,
                topLinks: this.generateTopLinksFor(null),
            });
        });
    }

    public generateTopLinksFor(user: IUser | null): ILink[] {
        if (user) {
            const basis: ILink[] = [];
            if (this.userMan.isTeacher(user)) {
                basis.push({ name: "Teacher", uri: "app/teacher/", active: false });
            }
            basis.push({ name: "Student", uri: "app/student/", active: false });
            if (this.userMan.isAdmin(user)) {
                basis.push({ name: "Admin", uri: "app/admin", active: false });
            }
            basis.push({ name: "Help", uri: "app/help", active: false });
            return basis;
        } else {
            return [{ name: "Help", uri: "app/help", active: false }];
        }
    }

    public componentDidMount() {
        const curUrl = location.pathname;
        if (curUrl === "/") {
            this.navMan.navigateToDefault();
        } else {
            this.navMan.navigateTo(curUrl);
        }
    }

    public render() {
        if (this.state.activePage) {
            return this.renderTemplate(this.state.activePage.template);
        } else {
            return <h1>404 not found</h1>;
        }
    }

    private handleClick(link: ILink) {
        if (link.uri) {
            this.navMan.navigateTo(link.uri);
        } else {
            console.warn("Warning! Empty link detected", link);
        }
    }

    private renderActiveMenu(menu: number): JSX.Element[] | string {
        if (this.state.activePage) {
            return this.state.activePage.renderMenu(menu);
        }
        return "";
    }

    private renderActivePage(page: string): JSX.Element {
        const curPage = this.state.activePage;
        if (curPage) {
            return curPage.renderContent(page);
        }
        return <h1>404 Page not found</h1>;
    }

    private checkLinks(links: ILink[]): void {
        this.navMan.checkLinks(links);
    }

    private renderTemplate(name: string | null) {
        let body: JSX.Element;
        const content = this.renderActivePage(this.subPage);
        const loginLink: ILink[] = [
            { name: "Github", uri: "app/login/login/github" },
            { name: "Gitlab", uri: "app/login/login/gitlab" },
        ];
        switch (name) {
            case "frontpage":
                body = (
                    <Row className="container-fluid">
                        <div className="col-xs-12">
                            {content}
                        </div>
                    </Row>
                );
            default:
                body = (
                    <Row className="container-fluid">
                        <div className="col-md-2 col-sm-3 col-xs-12">
                            {this.renderActiveMenu(0)}
                        </div>
                        <div className="col-md-10 col-sm-9 col-xs-12">
                            {content}
                        </div>
                    </Row>
                );
        }
        return (
            <div>
                <NavBar id="top-bar"
                    isFluid={false}
                    isInverse={true}
                    onClick={(link) => this.handleClick(link)}
                    brandName="Auto Grader">
                    <NavBarMenu links={this.state.topLinks}
                        onClick={(link) => this.handleClick(link)}>
                    </NavBarMenu>
                    <NavBarLogin
                        user={this.state.curUser}
                        links={loginLink}
                        onClick={(link) => this.handleClick(link)}>
                    </NavBarLogin>
                </NavBar>
                {body}
            </div>);
    }
}

/**
 * @description The main entry point for the application. No other code should be executet outside this function
 */
function main() {
    const DEBUG_BROWSER = "DEBUG_BROWSER";
    const DEBUG_SERVER = "DEBUG_SERVER";

    let curRunning: string;
    curRunning = DEBUG_BROWSER;

    const tempData = new TempDataProvider();

    let userMan: UserManager;
    let courseMan: CourseManager;
    let navMan: NavigationManager;

    if (curRunning === DEBUG_SERVER) {
        const serverData = new ServerProvider();

        userMan = new UserManager(serverData);
        courseMan = new CourseManager(tempData);
        navMan = new NavigationManager(history);
    } else {
        userMan = new UserManager(tempData);
        courseMan = new CourseManager(tempData);
        navMan = new NavigationManager(history);
    }

    (window as any).debugData = { tempData, userMan, courseMan, navMan };

    // const user = userMan.tryLogin("test@testersen.no", "1234");

    navMan.setDefaultPath("app/home");
    navMan.registerPage("app/home", new HomePage());
    navMan.registerPage("app/student", new StudentPage(userMan, navMan, courseMan));
    navMan.registerPage("app/teacher", new TeacherPage(userMan, navMan, courseMan));
    navMan.registerPage("app/admin", new AdminPage(navMan, userMan, courseMan));
    navMan.registerPage("app/help", new HelpPage(navMan));
    navMan.registerPage("app/login", new LoginPage(navMan, userMan));

    navMan.registerErrorPage(404, new ErrorPage());
    navMan.onNavigate.addEventListener((e) => {
        console.log(e);
    });

    ReactDOM.render(
        <AutoGrader userManager={userMan} navigationManager={navMan}>

        </AutoGrader>,
        document.getElementById("root"),
    );
}

main();

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

import { HttpHelper } from "./HttpHelper";
import { ILogEntry, LogManager } from "./managers/LogManager";

import { PageInfo } from "./components/information/PageInfo";

interface IAutoGraderState {
    activePage?: ViewPage;
    currentContent: JSX.Element;
    topLinks: ILink[];
    curUser: IUser | null;
    curMessage?: ILogEntry;
}

interface IAutoGraderProps {
    userManager: UserManager;
    navigationManager: NavigationManager;
    logManager: LogManager;
}

class AutoGrader extends React.Component<IAutoGraderProps, IAutoGraderState> {
    private userMan: UserManager;
    private navMan: NavigationManager;
    private logMan: LogManager;
    private subPage: string;

    constructor(props: IAutoGraderProps) {
        super();

        this.userMan = props.userManager;
        this.navMan = props.navigationManager;
        this.logMan = props.logManager;
        this.logMan.onshowuser.addEventListener(async (e) => {
            console.log("OnShowUser Event: ", e);
            this.setState({ curMessage: e.entry });
            this.setState({ currentContent: await this.refreshActivePage() });
            console.log("State: ", this.state);
        });

        const curUser = this.userMan.getCurrentUser();

        this.state = {
            activePage: undefined,
            topLinks: [],
            curUser,
            currentContent: <div>No Content Available</div>,
        };

        (async () => {
            this.setState({ topLinks: await this.generateTopLinksFor(curUser) });
        })();

        this.navMan.onNavigate.addEventListener((e: INavEvent) => this.handleNavigation(e));

        this.userMan.onLogin.addEventListener(async (e) => {
            console.log("Sign in");
            this.setState({
                curUser: e.user,
                topLinks: await this.generateTopLinksFor(e.user),
            });
        });

        this.userMan.onLogout.addEventListener(async (e) => {
            console.log("Sign out");
            this.setState({
                curUser: null,
                topLinks: await this.generateTopLinksFor(null),
            });
        });
    }

    public async handleNavigation(e: INavEvent) {
        this.subPage = e.subPage;
        const newContent = await this.renderTemplate(e.page, e.page.template);

        const tempLink = this.state.topLinks.slice();
        this.checkLinks(tempLink);

        this.setState({ activePage: e.page, topLinks: tempLink, currentContent: newContent });
    }

    public async generateTopLinksFor(user: IUser | null): Promise<ILink[]> {
        if (user) {
            const basis: ILink[] = [];
            if (await this.userMan.isTeacher(user)) {
                basis.push({ name: "Teacher", uri: "app/teacher/", active: false });
            }
            basis.push({ name: "Student", uri: "app/student/", active: false });
            if (this.userMan.isAdmin(user)) {
                basis.push({ name: "Admin", uri: "app/admin", active: false });
            }
            basis.push({ name: "Help", uri: "app/help", active: false });
            return basis;
        } else {
            return [
                { name: "Help", uri: "app/help", active: false },
                { name: "New Course", uri: "app/admin/courses/new", active: false },
            ];
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
        console.log("Log from index.tsx");

        if (this.state.activePage) {
            return this.state.currentContent;
        } else {
            return <h1>404 not found</h1>;
        }
    }

    private async refreshActivePage(): Promise<JSX.Element> {
        if (this.state.activePage) {
            return await this.renderTemplate(this.state.activePage, this.state.activePage.template);
        }
        return <div>404 Error</div>;
    }

    private handleClick(link: ILink) {
        if (link.uri) {
            this.navMan.navigateTo(link.uri);
        } else {
            console.warn("Warning! Empty link detected", link);
        }
    }

    private async renderActiveMenu(page: ViewPage, menu: number): Promise<JSX.Element[] | string> {
        if (page) {
            return await page.renderMenu(menu);
        }
        return "";
    }

    private async renderActivePage(page: ViewPage, subPage: string): Promise<JSX.Element> {
        if (page) {
            return await page.renderContent(subPage);
        }
        return <h1>404 Page not found</h1>;
    }

    private checkLinks(links: ILink[]): void {
        this.navMan.checkLinks(links);
    }

    private async renderTemplate(page: ViewPage, name: string | null): Promise<JSX.Element> {
        console.log("render");
        let body: JSX.Element;
        const content = await this.renderActivePage(page, this.subPage);
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
                            {await this.renderActiveMenu(page, 0)}
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
                <PageInfo entry={this.state.curMessage} onclose={async () => {
                    this.setState({ curMessage: undefined });
                    this.setState({ currentContent: await this.refreshActivePage() });
                }} />
                {body}
            </div>);
    }
}

/**
 * @description The main entry point for the application. No other code should be executet outside this function
 */
async function main(): Promise<void> {
    const DEBUG_BROWSER = "DEBUG_BROWSER";
    const DEBUG_SERVER = "DEBUG_SERVER";

    let curRunning: string;
    curRunning = DEBUG_SERVER;

    if (window.location.host.match("localhost")
        || localStorage.getItem("debug")) {
        curRunning = DEBUG_BROWSER;
    }

    const tempData = new TempDataProvider();

    let userMan: UserManager;
    let courseMan: CourseManager;
    const logMan = new LogManager();
    const navMan: NavigationManager = new NavigationManager(history, logMan.createLogger("NavigationManager"));

    if (curRunning === DEBUG_SERVER) {
        const httpHelper = new HttpHelper("/api/v1");
        const serverData = new ServerProvider(httpHelper, logMan.createLogger("ServerProvider"));

        userMan = new UserManager(serverData, logMan.createLogger("UserManager"));
        courseMan = new CourseManager(serverData, logMan.createLogger("CourseManager"));
    } else {
        userMan = new UserManager(tempData, logMan.createLogger("UserManager"));
        courseMan = new CourseManager(tempData, logMan.createLogger("CourseManager"));

        const user = await userMan.tryLogin("test@testersen.no", "1234");
    }

    await userMan.checkUserLoggedIn();

    (window as any).debugData = { tempData, userMan, courseMan, navMan, logMan };

    navMan.setDefaultPath("app/home");
    const all: Array<Promise<void>> = [];
    all.push(navMan.registerPage("app/home", new HomePage()));
    all.push(navMan.registerPage("app/student", new StudentPage(userMan, navMan, courseMan)));
    all.push(navMan.registerPage("app/teacher", new TeacherPage(userMan, navMan, courseMan)));
    all.push(navMan.registerPage("app/admin", new AdminPage(navMan, userMan, courseMan)));
    all.push(navMan.registerPage("app/help", new HelpPage(navMan)));
    all.push(navMan.registerPage("app/login", new LoginPage(navMan, userMan)));

    Promise.all(all);

    navMan.registerErrorPage(404, new ErrorPage());
    navMan.onNavigate.addEventListener((e) => {
        console.log(e);
    });

    ReactDOM.render(
        <AutoGrader userManager={userMan} navigationManager={navMan} logManager={logMan}>

        </AutoGrader>,
        document.getElementById("root"),
    );
}

main();

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

import { User } from "../proto/ag/ag_pb";
import { AdminPage } from "./pages/AdminPage";

import { NavBarLogin } from "./components/navigation/NavBarLogin";
import { NavBarMenu } from "./components/navigation/NavBarMenu";
import { LoginPage } from "./pages/LoginPage";

import { ServerProvider } from "./managers/ServerProvider";

import { HttpHelper } from "./HttpHelper";
import { ILogEntry, LogManager } from "./managers/LogManager";

import { PageInfo } from "./components/information/PageInfo";
import { UserPage } from "./pages/UserPage";

import { AddMenu } from "./components/navigation/AddMenu";
import { GrpcManager } from "./managers/GRPCManager";
import { showLoader } from "./loader";

interface IAutoGraderState {
    currentContent: JSX.Element;
    topLinks: ILink[];
    curMessage?: ILogEntry;
}

interface IAutoGraderProps {
    userManager: UserManager;
    navigationManager: NavigationManager;
    logManager: LogManager;
}

class AutoGrader extends React.Component<IAutoGraderProps, IAutoGraderState> {
    private activePage?: ViewPage;
    private userMan: UserManager;
    private navMan: NavigationManager;
    private logMan: LogManager;
    private subPage: string;
    private currentBodyContent?: JSX.Element;
    private currentMenuContent: JSX.Element[][] = [];
    private curUser: User | null;

    constructor(props: IAutoGraderProps) {
        super(props);

        this.userMan = props.userManager;
        this.navMan = props.navigationManager;
        this.logMan = props.logManager;
        this.logMan.onshowuser.addEventListener(async (e) => {
            this.setState({ curMessage: e.entry });
            this.setState({ currentContent: await this.refreshActivePage() });
        });

        this.curUser = this.userMan.getCurrentUser();
        this.state = {
            topLinks: [],
            currentContent: <div>No Content Available</div>,
        };

        (async () => {
            this.setState({ topLinks: await this.generateTopLinksFor(this.curUser) });
        })();

        this.navMan.onNavigate.addEventListener((e: INavEvent) => this.handleNavigation(e));

        this.userMan.onLogin.addEventListener(async (e) => {
            const user = e.user;
            this.curUser = user;
            this.setState({
                topLinks: await this.generateTopLinksFor(user),
            });
        });

        this.userMan.onLogout.addEventListener(async (e) => {
            this.setState({
                topLinks: [],
            });
        });
    }

    public async handleNavigation(e: INavEvent) {
        if (!this.checkloggedInUser() && !e.uri.match(/app\/user/)) {
            this.navMan.navigateTo("app/user");
            return;
        }
        const topLinks = await this.generateTopLinksFor(this.curUser);
        this.checkLinks(topLinks);
        this.setState({ topLinks });

        this.currentBodyContent = undefined;
        this.currentMenuContent = [];

        this.subPage = e.subPage;

        const newContent = await this.renderTemplate(e.page, e.page.template);

        this.activePage = e.page
        this.setState({currentContent: newContent });
    }

    public async generateTopLinksFor(user: User | null): Promise<ILink[]> {
        if (user) {
            const basis: ILink[] = [];
            const confirmedTeacher = await this.userMan.isTeacher();
            if (user.getIsadmin() || confirmedTeacher) {
                basis.push({ name: "Teacher", uri: "app/teacher/", active: false });
            }
            basis.push({ name: "Courses", uri: "app/student/", active: false });
            if (user.getIsadmin() && localStorage.getItem("admin")) {
                basis.push({ name: "Admin", uri: "app/admin", active: false });
            }
            return basis;
        }
        return [
            { name: "Help", uri: "app/help", active: false },
        ];
    }

    public componentDidMount() {
        const curUrl = location.pathname;
        if (curUrl === "/") {
            this.navMan.navigateToDefault();
        } else {
            this.navMan.navigateTo(curUrl);
        }
    }

    public checkloggedInUser(): boolean {
        const cur = this.userMan.getCurrentUser();
        if (cur) {
            return this.userMan.isValidUser(cur);
        }
        return true;
    }

    public render() {
        if (this.activePage) {
            return this.state.currentContent;
        } else {
            return showLoader();
        }
    }

    private async refreshActivePage(): Promise<JSX.Element> {
        if (this.activePage) {
            return this.renderTemplate(this.activePage, this.activePage.template);
        }
        return showLoader();
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
            this.currentMenuContent[menu] = await page.renderMenu(menu);
            return this.currentMenuContent[menu];
        }
        return "";
    }

    private async renderActivePage(page: ViewPage, subPage: string): Promise<JSX.Element> {
        if (this.currentBodyContent) {
            return this.currentBodyContent;
        } else if (page) {
            this.currentBodyContent = await page.renderContent(subPage);
            return this.currentBodyContent;
        }
        return showLoader();
    }

    private checkLinks(links: ILink[]): void {
        this.navMan.checkLinks(links);
    }

    private async renderTemplate(page: ViewPage, name: string | null): Promise<JSX.Element> {
        let body: JSX.Element;
        let content: JSX.Element;
        let leftMenu: JSX.Element[] | null | string = null;
        let topArea: JSX.Element[] | null | string = null;
        content = await this.renderActivePage(page, this.subPage);
        leftMenu = await this.renderActiveMenu(page, 0);
        topArea = await this.renderActiveMenu(page, 1);

        const loginLink: ILink[] = [
            { name: "Github", uri: "app/login/login/github" },
            { name: "Gitlab", uri: "app/login/login/gitlab" },
        ];
        const dropDownMenuLinks: ILink[] = [
            { name: "Join Course", uri: "app/student/enroll" },
            { name: "Course List", uri: "app/student/courses/list"},
        ];
        if (this.curUser && this.curUser.getIsadmin()) {
            dropDownMenuLinks.push({ name: "New Course", uri: "app/admin/courses/new" });
        }
        const userLinks: ILink[] = [
            { name: "Signed in as: " + (this.curUser ? this.curUser.getName() : "") },
            { name: "#separator" },
            { name: "Your profile", uri: "/app/user" },
            { name: "Help", uri: "/app/help" },
            { name: "#separator" },
            { name: "Sign out", uri: "app/login/logout" },
        ];
        const adminLinks: ILink[] = [
            { name: "Signed in as: " + (this.curUser ? this.curUser.getName() : "") },
            { name: "#separator" },
            { name: "Your profile", uri: "/app/user" },
            { name: "Help", uri: "/app/help" },
            { name: "#separator" },
            { name: "Manage courses", uri: "app/admin/courses" },
            { name: "Manage users", uri: "app/admin/users" },
            { name: "#separator" },
            { name: "Sign out", uri: "app/login/logout" },
        ];
        switch (name) {
            case "frontpage":
                body = (
                    <Row>
                        <div className="col-xs-12">
                            {content}
                        </div>
                    </Row>
                );
                break;
            default:
                body = (
                    <Row>
                        <div className="col-md-2 col-sm-3 col-xs-12">
                            {leftMenu}
                        </div>
                        <div className="col-md-10 col-sm-9 col-xs-12">
                            {content}
                        </div>
                    </Row>
                );
                break;
        }
        let currentLinks = userLinks;
        if (this.curUser && this.curUser.getIsadmin()) {
            currentLinks = adminLinks;
        }
        return (
            <div>
                <NavBar id="top-bar"
                    isFluid={false}
                    isInverse={true}
                    onClick={(link) => this.handleClick(link)}
                    brandName="Autograder">
                    <NavBarMenu links={this.state.topLinks}
                        onClick={(link) => this.handleClick(link)}>
                    </NavBarMenu>
                    <NavBarLogin
                        user={this.curUser}
                        loginLinks={loginLink}
                        userLinks={currentLinks}
                        onClick={(link) => this.handleClick(link)}>
                    </NavBarLogin>
                    <AddMenu
                        user={this.curUser}
                        links={dropDownMenuLinks}
                        onClick={(link) => this.handleClick(link)}>
                    </AddMenu>
                </NavBar>
                <PageInfo entry={this.state.curMessage} onclose={async () => {
                    this.setState({ curMessage: undefined });
                    this.setState({ currentContent: await this.refreshActivePage() });
                }} />
                {topArea}
                <div className="container-fluid spacefix">
                    {body}
                </div>
            </div>);
    }
}

/**
 * @description The main entry point for the application.
 * No other code should be executed outside this function.
 */
async function main(): Promise<void> {
    const DEBUG_BROWSER = "DEBUG_BROWSER";
    const DEBUG_SERVER = "DEBUG_SERVER";

    let curRunning = DEBUG_SERVER;

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
        // TODO: use process.env.QUICKFEED_PORT
        // https://webpack.js.org/plugins/environment-plugin/
        const grpcHelper = new GrpcManager("https://" + window.location.hostname + ":8080");
        const serverData = new ServerProvider(httpHelper, grpcHelper, logMan.createLogger("ServerProvider"));
        userMan = new UserManager(serverData, logMan.createLogger("UserManager"));
        grpcHelper.setUserMan(userMan);
        courseMan = new CourseManager(serverData, logMan.createLogger("CourseManager"));
    } else {
        userMan = new UserManager(tempData, logMan.createLogger("UserManager"));
        courseMan = new CourseManager(tempData, logMan.createLogger("CourseManager"));
        const user = await userMan.tryLogin("test@testersen.no", "1234");
    }

    await userMan.checkUserLoggedIn();

    (window as any).debugData = { tempData, userMan, courseMan, navMan, logMan };

    navMan.setDefaultPath("app/home");
    const all: Promise<void>[] = [];
    all.push(navMan.registerPage("app/home", new HomePage()));
    all.push(navMan.registerPage("app/student", new StudentPage(userMan, navMan, courseMan)));
    all.push(navMan.registerPage("app/teacher", new TeacherPage(userMan, navMan, courseMan)));
    all.push(navMan.registerPage("app/admin", new AdminPage(navMan, userMan, courseMan)));
    all.push(navMan.registerPage("app/help", new HelpPage(navMan)));
    all.push(navMan.registerPage("app/login", new LoginPage(navMan, userMan)));
    all.push(navMan.registerPage("app/user", new UserPage(navMan, userMan)));

    Promise.all(all);

    navMan.registerErrorPage(404, new ErrorPage());

    ReactDOM.render(
        <AutoGrader userManager={userMan} navigationManager={navMan} logManager={logMan}>

        </AutoGrader>,
        document.getElementById("root"),
    );
}

main();

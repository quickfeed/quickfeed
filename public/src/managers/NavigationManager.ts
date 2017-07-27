import { IEventData, newEvent } from "../event";
import { NavigationHelper } from "../NavigationHelper";
import { isViewPage, ViewPage } from "../pages/ViewPage";
import { ILogger, ILoggerServer } from "./LogManager";

export interface IPageContainer {
    [name: string]: IPageContainer | ViewPage;
}

export interface INavEvent extends IEventData {
    uri: string;
    page: ViewPage;
    subPage: string;
}

export interface ILink {
    name: string;
    description?: string;
    uri?: string;
    active?: boolean;
}

export interface ILinkCollection {
    item: ILink;
    children?: ILink[];
}

export function isILinkCollection(item: any): item is ILinkCollection {
    if (item.item) {
        return true;
    }
    return false;
}

export class NavigationManager {
    public onNavigate = newEvent<INavEvent>("NavigationManager.onNavigate");

    private pages: IPageContainer = {};
    private errorPages: ViewPage[] = [];
    private defaultPath: string = "";
    private currentPath: string = "";
    private browserHistory: History;
    private logger: ILogger;

    constructor(history: History, logger: ILogger) {
        this.browserHistory = history;
        this.logger = logger;
        window.addEventListener("popstate", (e: PopStateEvent) => {
            this.navigateTo(location.pathname, true);
        });

    }

    public setDefaultPath(path: string) {
        this.defaultPath = path;
    }

    public navigateTo(path: string, preventPush?: boolean) {
        if (path === "/") {
            this.navigateToDefault();
            return;
        }
        const parts = NavigationHelper.getParts(path);
        let curPage: IPageContainer | ViewPage = this.pages;
        this.currentPath = parts.join("/");
        if (!preventPush) {
            this.browserHistory.pushState({}, "Autograder", "/" + this.currentPath);
        }
        for (let i = 0; i < parts.length; i++) {
            const a = parts[i];
            if (isViewPage(curPage)) {
                this.onNavigate({
                    page: curPage,
                    subPage: parts.slice(i, parts.length).join("/"),
                    target: this,
                    uri: path,
                });
                return;
            } else {
                const cur: IPageContainer | ViewPage = curPage[a];
                if (!cur) {
                    this.onNavigate({ target: this, page: this.getErrorPage(404), subPage: "", uri: path });
                    return;
                    // throw Error("404 Page not found");
                }
                curPage = cur;
            }
        }
        if (isViewPage(curPage)) {
            this.onNavigate({ target: this, page: curPage, uri: path, subPage: "" });
            return;
        } else {
            this.onNavigate({ target: this, page: this.getErrorPage(404), subPage: "", uri: path });
            // throw Error("404 Page not found");
        }
    }

    public navigateToDefault(): void {
        this.navigateTo(this.defaultPath);
    }

    public navigateToError(statusCode: number): void {
        this.onNavigate({ target: this, page: this.getErrorPage(statusCode), subPage: "", uri: statusCode.toString() });
    }

    public async registerPage(path: string, page: ViewPage): Promise<void> {
        const parts = NavigationHelper.getParts(path);
        if (parts.length === 0) {
            throw Error("Can't add page to index element");
        }
        page.setPath(parts.join("/"));
        let curObj = this.pages;

        for (let i = 0; i < parts.length - 1; i++) {
            const a = parts[i];
            if (a.length === 0) {
                continue;
            }
            let temp: IPageContainer | ViewPage = curObj[a];
            if (!temp) {
                temp = {};
                curObj[a] = temp;
            } else if (!isViewPage(temp)) {
                temp = curObj[a];
            }

            if (isViewPage(temp)) {
                throw Error("Can't assign a IPageContainer to a ViewPage");
            }
            curObj = temp;
        }
        curObj[parts[parts.length - 1]] = page;
        await page.init();
    }

    public registerErrorPage(statusCode: number, page: ViewPage) {
        this.errorPages[statusCode] = page;
    }

    /**
     * Checks to see if the link is part of the current path,
     * or the default page to the given ViewPage. Also mark them as active if they are.
     * @param links The links to check
     * @param viewPage ViewPage to get defaultPage information from
     */
    public checkLinks(links: ILink[], viewPage?: ViewPage): void {
        let checkUrl = this.currentPath;
        if (viewPage && viewPage.pagePath === checkUrl) {
            checkUrl += "/" + viewPage.navHelper.defaultPage;
        }
        const long = NavigationHelper.getParts(checkUrl);
        for (const l of links) {
            if (!l.uri) {
                continue;
            }
            const short = NavigationHelper.getParts(l.uri);
            l.active = this.checkPartEqual(long, short);
        }
    }

    public checkLinkCollection(links: ILinkCollection[], viewPage?: ViewPage): void {
        let checkUrl = this.currentPath;
        if (viewPage && viewPage.pagePath === checkUrl) {
            checkUrl += "/" + viewPage.navHelper.defaultPage;
        }
        const long = NavigationHelper.getParts(checkUrl);
        for (const l of links) {
            if (!l.item.uri) {
                continue;
            }
            const short = NavigationHelper.getParts(l.item.uri);
            l.item.active = this.checkPartEqual(long, short);
            if (l.children) {
                this.checkLinks(l.children, viewPage);
            }
        }
    }

    public refresh() {
        this.navigateTo(this.currentPath);
    }

    private checkPartEqual(long: string[], short: string[]): boolean {
        if (short.length > long.length) {
            return false;
        }
        for (let i = 0; i < short.length; i++) {
            if (short[i] !== long[i]) {
                return false;
            }
        }
        return true;
    }

    private getErrorPage(statusCode: number): ViewPage {
        if (this.errorPages[statusCode]) {
            return this.errorPages[statusCode];
        }
        throw Error("Status page: " + statusCode + " is not defined");
    }
}

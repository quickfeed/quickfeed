import * as React from "react";
import { ILink, NavigationManager } from "../managers/NavigationManager";
import { ViewPage } from "./ViewPage";
import { HelpView } from "./views/HelpView";

class HelpPage extends ViewPage {

    private navMan: NavigationManager;
    private pages: { [name: string]: JSX.Element } = {};

    constructor(navMan: NavigationManager) {
        super();
        this.navMan = navMan;
        this.defaultPage = "help";
        this.pages.help = <HelpView></HelpView>;
    }

    public pageNavigation(page: string): void {
        "Not used";
    }

    public renderContent(page: string): JSX.Element {
        if (page.length === 0) {
            page = this.defaultPage;
        }
        if (this.pages[page]) {
            return this.pages[page];
        }
        return <h1>404 page not found</h1>;
    }
}
export { HelpPage };

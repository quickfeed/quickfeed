import * as React from "react";
import { ILink, NavigationManager } from "../managers";
import { ViewPage } from "./ViewPage";
import { HelpView } from "./views/HelpView";

import { INavInfo, NavigationHelper } from "../NavigationHelper";

class HelpPage extends ViewPage {

    private navMan: NavigationManager;
    private pages: { [name: string]: JSX.Element } = {};

    constructor(navMan: NavigationManager) {
        super();
        this.navMan = navMan;
        this.navHelper.defaultPage = "help";
        this.navHelper.registerFunction("help", this.help);
    }

    public help(info: INavInfo<any>): JSX.Element {
        return <HelpView></HelpView>;
    }

    public renderContent(page: string): JSX.Element {
        const temp = this.navHelper.navigateTo(page);
        if (temp) {
            return temp;
        }
        return <h1>404 page not found</h1>;
    }
}
export { HelpPage };

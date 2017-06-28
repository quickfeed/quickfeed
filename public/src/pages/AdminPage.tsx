import * as React from "react";

import { NavigationManager } from "../managers";
import { ViewPage } from "./ViewPage";

class AdminPage extends ViewPage {
    private navMan: NavigationManager;

    constructor(navMan: NavigationManager) {
        super();
        this.navMan = navMan;
    }

    public renderContent(page: string) {
        return <div>Not yet implemented</div>;
    }
}

export { AdminPage };

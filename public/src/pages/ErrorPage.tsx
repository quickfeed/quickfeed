import * as React from "react";
import { ViewPage } from "./ViewPage";

import { NavigationHelper } from "../NavigationHelper";

class ErrorPage extends ViewPage {
    private pages: { [key: string]: JSX.Element } = {};

    constructor() {
        super();
        this.navHelper.defaultPage = "404";
        this.navHelper.registerFunction("404", (navInfo) => {
            return <div>
                <h1>404 Page not found</h1>
                <p>The page you where looking for does not exist</p>
            </div>;
        });
    }

    public renderContent(page: string): JSX.Element {
        let content = this.navHelper.navigateTo(page);
        if (!content) {
            content = this.navHelper.navigateTo("404");
        }
        if (!content) {
            throw new Error("There is a problem with the navigation");
        }
        return content;
    }
}

export { ErrorPage };

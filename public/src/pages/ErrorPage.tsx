import * as React from "react";
import { View, ViewPage } from "./ViewPage";

export class ErrorPage extends ViewPage {
    private pages: { [key: string]: JSX.Element } = {};

    constructor() {
        super();
        this.navHelper.defaultPage = "404";
        this.navHelper.registerFunction("404", async (navInfo) => {
            return <div>
                <h1>404 Page not found</h1>
                <p>The page you where looking for does not exist</p>
            </div>;
        });
    }

    public async renderContent(page: string): View {
        let content = await this.navHelper.navigateTo(page);
        if (!content) {
            content = await this.navHelper.navigateTo("404");
        }
        if (!content) {
            throw new Error("There is a problem with the navigation");
        }
        return content;
    }
}

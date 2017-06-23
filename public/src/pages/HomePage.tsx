import * as React from "react";
import { ViewPage } from "./ViewPage";

class HomePage extends ViewPage {

    constructor() {
        super();
        this.defaultPage = "index";
    }

    public pageNavigation(page: string): void {
        "Not used";
    }

    public renderContent(page: string): JSX.Element {
        return <h1>Welcome to autograder</h1>;
    }
}

export { HomePage };

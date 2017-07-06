import * as React from "react";
import { ViewPage } from "./ViewPage";

export class HomePage extends ViewPage {
    constructor() {
        super();
    }

    public async renderContent(page: string): Promise<JSX.Element> {
        return <h1>Welcome to autograder</h1>;
    }
}

import * as React from "react";
import {ViewPage} from "./ViewPage"
import { NavigationManager, ILink } from "../managers/NavigationManager";
import {HelpView} from "./views/HelpView";

class HelpPage extends ViewPage {

    navMan: NavigationManager;
    private pages: {[name: string]: JSX.Element} = {};

    constructor(navMan: NavigationManager){
        super();
        this.navMan = navMan;
        this.defaultPage = "help";
        this.pages["help"] = <HelpView></HelpView>;
    }

    pageNavigation(page: string): void {
        
    }

    renderContent(page: string): JSX.Element{
        if (page.length === 0){
            page = this.defaultPage;
        }
        if (this.pages[page]){
            return this.pages[page];
        }
        return <h1>404 page not found</h1>
    }
}
export {HelpPage}
import * as React from "react";

import { NavigationHelper } from "../NavigationHelper";

export type View = Promise<JSX.Element>;

export function isViewPage(item: any): item is ViewPage {
    if (item instanceof ViewPage) {
        return true;
    }
    return false;
}

export abstract class ViewPage {
    public template: string | null = null;
    public pagePath: string;
    public navHelper: NavigationHelper = new NavigationHelper(this);
    public currentPage: string = "";

    public async init(): Promise<void> {
        return;
    }

    public setPath(path: string) {
        this.pagePath = path;
    }

    public async renderMenu(menu: number): Promise<JSX.Element[]> {
        return [];
    }

    public async renderContent(page: string): View {
        const pageContent = await this.navHelper.navigateTo(page);
        this.currentPage = page;
        return pageContent ?? <div className="info"><p>Choose a course or create a new one.</p></div>;
    }
}

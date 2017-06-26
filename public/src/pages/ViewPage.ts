function isViewPage(item: any): item is ViewPage {
    if (item instanceof ViewPage) {
        return true;
    }
    return false;
}

abstract class ViewPage {
    public template: string | null = null;
    public defaultPage: string = "";
    public pagePath: string;

    public setPath(path: string) {
        this.pagePath = path;
    }

    public renderMenu(menu: number): JSX.Element[] {
        return [];
    }

    public abstract renderContent(page: string): JSX.Element;
}

export { isViewPage, ViewPage };

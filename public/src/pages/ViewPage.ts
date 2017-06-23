function isViewPage(item: any): item is ViewPage {
    if (item instanceof ViewPage){
        return true;
    }
    return false;
}

abstract class ViewPage{
    template: string | null = null;
    defaultPage: string = "";
    pagePath: string;

    setPath(path: string){
        this.pagePath = path;
    }

    renderMenu(menu:number): JSX.Element[] {
        return [];
    }


    abstract pageNavigation(page: string): void;
    abstract renderContent(page: string): JSX.Element;
}

export {isViewPage, ViewPage}
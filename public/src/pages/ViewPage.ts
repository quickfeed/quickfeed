function isViewPage(item: any): item is ViewPage {
    if (item instanceof ViewPage){
        return true;
    }
    return false;
}

abstract class ViewPage{
    pages: any = {};
    template: string | null = null;
    defaultPage: string = "";
    pagePath: string;

    setPath(path: string){
        this.pagePath = path;
    }

    renderMenu(menu:number): JSX.Element[] {
        return [];
    }
}

export {isViewPage, ViewPage}
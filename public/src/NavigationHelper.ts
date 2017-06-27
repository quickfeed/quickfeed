
import { IEventData, newEvent } from "./event";

interface INavPath {
    [path: string]: INavPath | INavObject;
}

interface INavInfo<T> {
    matchPath: string[];
    realPath: string[];
    params: T;
}

interface INavInfoEvent extends IEventData {
    navInfo: INavInfo<any>;
}

interface INavObject {
    path: string[];
    func: (navInfo: INavInfo<any>) => JSX.Element;
}

class NavigationHelper {
    public static getParts(path: string): string[] {
        return this.removeEmptyEntries(path.split("/"));
    }

    public static removeEmptyEntries(array: string[]): string[] {
        const newArray: string[] = [];
        array.map((v) => {
            if (v.length > 0) {
                newArray.push(v);
            }
        });
        return newArray;
    }

    public static getOptionalField(field: string): string | null {
        const tField = field.trim();
        if (tField.length > 2 && tField.charAt(0) === "{" && tField.charAt(tField.length - 1) === "}") {
            return tField.substr(1, tField.length - 2);
        }
        return null;
    }

    public static isINavObject(obj: any): obj is INavObject {
        return obj && obj.path;
    }

    public static handleClick(e: React.MouseEvent<HTMLAnchorElement>, callback: () => void) {
        if (e.shiftKey || e.ctrlKey || e.button === 1) {
            return;
        } else {
            e.preventDefault();
            callback();
        }
    }

    public onPreNavigation = newEvent<INavInfoEvent>("NavigationHelper.onPreNavigation");

    private DEFAULT_VALUE: string = "";
    private navObj = "__navObj";
    private path: INavPath = {};
    private thisObject: any;

    get defaultPage(): string {
        return this.DEFAULT_VALUE;
    }

    set defaultPage(value: string) {
        this.DEFAULT_VALUE = value;
    }

    constructor(thisObject: any) {
        this.thisObject = thisObject;
    }

    public registerFunction<T>(path: string, callback: (navInfo: INavInfo<T>) => JSX.Element): void {
        const pathParts = NavigationHelper.getParts(path);
        if (pathParts.length === 0) {
            throw new Error("Can't register function on empty path");
        }
        const curObj = this.createNavPath(pathParts);

        const temp: INavObject = {
            path: pathParts,
            func: callback,
        };
        curObj[this.navObj] = temp;
    }

    public navigateTo(path: string): JSX.Element | null {
        if (path.length === 0) {
            path = this.DEFAULT_VALUE;
        }
        const pathParts = NavigationHelper.getParts(path);
        if (pathParts.length === 0) {
            throw new Error("Can't navigate to an empty path");
        }

        const curObj = this.getNavPath(pathParts);
        if (!curObj || !curObj[this.navObj]) {
            return null;
        }
        const navObj = curObj[this.navObj] as INavObject;
        const navInfo: INavInfo<any> = {
            matchPath: navObj.path,
            realPath: pathParts,
            params: this.createParamsObj(navObj.path, pathParts),
        };
        this.onPreNavigation({ target: this, navInfo });
        return navObj.func.call(this.thisObject, navInfo);
    }

    private createParamsObj(matchPath: string[], realPath: string[]): any {
        if (matchPath.length !== realPath.length) {
            throw new Error("trying to match different paths");
        }
        const returnObj: any = {};
        for (let i = 0; i < matchPath.length; i++) {
            const param = NavigationHelper.getOptionalField(matchPath[i]);
            if (param) {
                returnObj[param] = realPath[i];
            }
        }
        return returnObj;
    }

    private getNavPath(pathParts: string[]): INavPath | null {
        let curObj = this.path;
        for (const part of pathParts) {
            let curIndex = part;

            if (!curObj[curIndex]) {
                curIndex = "*";
            }

            const curWrap = curObj[curIndex];

            if (NavigationHelper.isINavObject(curWrap) || curIndex === this.navObj) {
                throw new Error("Can't navigate to: " + curIndex);
            }
            if (!curWrap) {
                return null;
            }
            curObj = curWrap;
        }
        return curObj;
    }

    private createNavPath(pathParts: string[]): INavPath {
        let curObj = this.path;
        for (const part of pathParts) {
            let curIndex = part;
            const optional = NavigationHelper.getOptionalField(curIndex);
            if (optional) {
                curIndex = "*";
            }
            let curWrap = curObj[curIndex];
            if (NavigationHelper.isINavObject(curWrap) || curIndex === this.navObj) {
                throw new Error("Can't assign path to: " + curIndex);
            }
            if (!curWrap) {
                curWrap = {};
                curObj[curIndex] = curWrap;
            }
            curObj = curWrap;
        }
        return curObj;
    }
}

export { INavInfo, INavInfoEvent, INavObject, INavPath, NavigationHelper };

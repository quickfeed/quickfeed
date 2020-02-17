
import { IEventData, newEvent } from "./event";

export function trimChars(str: string, char: string): string {
    if (str.length === 0) {
        return "";
    }
    let start = 0;
    let end = str.length - 1;
    while (str[start] === char) {
        start++;
    }
    while (str[end] === char) {
        end--;
    }
    // No more of the string left
    if (start >= end) {
        return "";
    }
    return str.substring(start, end + 1);
}

export function combinePath(...parts: string[]): string {
    if (parts.length === 0) {
        return "";
    }
    let newPath = "";
    for (let i = 0; i < parts.length; i++) {
        if (i !== 0 || (i === 0 && parts[0].length > 0 && parts[0][0] === "/")) {
            newPath += "/";
        }
        newPath += trimChars(parts[i], "/");
    }
    return newPath;
}

export interface INavPath {
    [path: string]: INavPath | INavObject;
}

export interface INavInfo<T> {
    matchPath: string[];
    realPath: string[];
    params: T;
}

export interface INavInfoEvent extends IEventData {
    navInfo: INavInfo<any>;
}

export interface INavObject {
    path: string[];
    func: (navInfo: INavInfo<any>) => Promise<JSX.Element>;
}

export class NavigationHelper {
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

    public static getOptionalField(field: string): { name: string, type?: string } | null {
        const tField = field.trim();
        if (tField.length > 2 && tField.charAt(0) === "{" && tField.charAt(tField.length - 1) === "}") {
            const parts = tField.substr(1, tField.length - 2).split(":");
            return { name: parts[0], type: (parts.length > 1 ? parts[1] : undefined) };
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
    public checkAuthentication: ((navInfo: INavInfo<any>) => boolean) | undefined;

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

    /**
     * Registers a function that is called when the assigned path is requested.
     * The path supports wildcards and those are passed to the INavInfo<T>#params object.
     * The wild cards have format "path/{wild1}/test/{wild2}" params object would look like:
     * { wild1: "value 1", wild2: "value 2" }.
     * Types is also match and @see ITypeMap for supported types
     * The format for types is "path/{wild1:boolean}/test/{wild2:number}"
     * @param path The path to register the callback on
     * @param callback the callback to call when navigation to that path accurre
     */
    public registerFunction<T>(path: string, callback: (navInfo: INavInfo<T>) => Promise<JSX.Element>): void {
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

    public async navigateTo(path: string): Promise<JSX.Element | null> {
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
        if (!navInfo.params) {
            // TODO: Proper 404 handling here. Wrong type of one or more parameters
            console.error("One or more parameteres has wrong value", navInfo.matchPath, navInfo.realPath);
            return null;
        }
        this.onPreNavigation({ target: this, navInfo });
        if (this.checkAuthentication) {
            if (!this.checkAuthentication(navInfo)) {
                // TODO: Same as above, should be proper error handling
                return null;
            }
        }
        return navObj.func.call(this.thisObject, navInfo);
    }

    private parseValue(value: string, type: string): string | number | boolean | undefined {
        switch (type) {
            case "string":
                return value;
            case "number":
                const num = parseFloat(value);
                if (isNaN(num)) {
                   return undefined;
                }
                return num;
            case "boolean":
                if (value.toLowerCase() === "true") {
                    return true;
                } else if (value.toLowerCase() === "false") {
                    return false;
                }
                return undefined;
        }
        return undefined;
    }

    private createParamsObj(matchPath: string[], realPath: string[]): any | undefined {
        if (matchPath.length !== realPath.length) {
            throw new Error("trying to match different paths");
        }
        const returnObj: any = {};
        for (let i = 0; i < matchPath.length; i++) {
            const param = NavigationHelper.getOptionalField(matchPath[i]);
            if (param) {
                if (param.type) {
                    if (param.type === "string" || param.type === "boolean" || param.type === "number") {
                        const value = this.parseValue(realPath[i], param.type);
                        if (value !== undefined) {
                            returnObj[param.name] = value;
                        } else {
                            return undefined;
                        }
                    } else {
                        console.error("Type not supported in navigation path: ", param.type, matchPath, realPath);
                    }
                } else {
                    returnObj[param.name] = realPath[i];
                }
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

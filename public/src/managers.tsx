interface IUser{
    id: number;
    firstName: string;
    lastName: string
    email: string;
    personId: number;

}

interface IDummyUser extends IUser{
    password: string;
}

interface IUserProvider{
    tryLogin(username: string, password: string): IUser | null;
    getAllUser(): IUser[];
}

class TempDataProvider implements IUserProvider, ICourseProvider{

    private localUsers: IDummyUser[];
    private localAssignments: IAssignment[];
    private localCourses: ICourse[];

    constructor(){
        this.localCourses = [
            {
                id: 0,
                name: "Object Oriented Programming",
                tag: "DAT100"
            },
            {
                id: 1,
                name: "Algorithms and Datastructures",
                tag: "DAT200"
            }
        ];

        this.localAssignments = [
            {
                id: 0,
                courceId: 0,
                name: "Lab 1",
                start: new Date(2017, 5, 1),
                deadline: new Date(2017, 5, 25),
                end: new Date(2017, 5, 30)
            },
            {
                id: 1,
                courceId: 1,
                name: "Lab 1",
                start: new Date(2017, 5, 1),
                deadline: new Date(2017, 5, 25),
                end: new Date(2017, 5, 30)
            }
        ];

        this.localUsers = [
            {
                id: 999,
                firstName: "Test",
                lastName: "Testersen",
                email: "test@testersen.no",
                personId: 9999,
                password: "1234"
            },
            {
                id: 1,
                firstName: "Per",
                lastName: "Pettersen",
                email: "per@pettersen.no",
                personId: 1234,
                password: "1234"
            },
            {
                id: 2,
                firstName: "Bob",
                lastName: "Bobsen",
                email: "bob@bobsen.no",
                personId: 1234,
                password: "1234"
            },
            {
                id: 3,
                firstName: "Petter",
                lastName: "Pan",
                email: "petter@pan.no",
                personId: 1234,
                password: "1234"
            }
        ];
    }

    getAllUser(): IUser[] {
        return this.localUsers;
    }

    getCourses(): ICourse[] {
        return this.localCourses;
    }
    
    getAssignments(courseId: number): IAssignment[] {
        let temp: IAssignment[] = [];
        for(let a of this.localAssignments){
            if (a.courceId === courseId){
                temp.push(a);
            }
        }
        return temp;
    }

    tryLogin(username: string, password: string): IUser | null {
        for(let u of this.localUsers){
            if (u.email.toLocaleLowerCase() === username.toLocaleLowerCase()){
                if (u.password === password){
                    return u;
                }
                return null;
            }
        }
        return null;
    }

}

function isCourse(value: any): value is ICourse{
    console.log(value);
    return value && typeof value.id === "number" && value.name && value.tag;
}

interface ICourse{
    id: number;
    name: string;
    tag: string;
}

interface IAssignment{
    id: number;
    courceId: number;
    name: string;
    start: Date;
    deadline: Date;
    end: Date;
}

interface ICourseStudent{
    personId: number;
    courseId: number;
}

interface ICourseProvider{
    getCourses(): ICourse[];
    getAssignments(courseId: number): IAssignment[];
}

class CourseManager {
    courseProvider: ICourseProvider;
    constructor(courseProvider: ICourseProvider){
        this.courseProvider = courseProvider;
    }

    getCourses():ICourse[]{
        return this.courseProvider.getCourses();
    }

    getAssignments(courseId: number): IAssignment[];
    getAssignments(course: ICourse): IAssignment[];
    getAssignments(courseId: number | ICourse): IAssignment[] {
        if (isCourse(courseId)){
            courseId = courseId.id;
            console.log(courseId);
        }
        return this.courseProvider.getAssignments(courseId);
    }
    
}

class UserManager{
    private userProvider: IUserProvider;
    private currentUser: IUser | null;

    constructor(userProvider: IUserProvider){
        this.userProvider = userProvider;
    }

    getCurrentUser(): IUser | null{
        return this.currentUser;
    }

    tryLogin(username: string, password: string): IUser | null{
        let result = this.userProvider.tryLogin(username, password);
        if (result){
            this.currentUser = result;
        }
        return result;
    }

    getAllUser(): IUser[]{
        return this.userProvider.getAllUser();
    }

    getUser(id: number){
        
    }
}

interface IPageContainer{
    [name: string]: IPageContainer | ViewPage;
}

interface INavEvent extends IEventData{
    uri: string;
    page: ViewPage;
    subPage: string;
}

class NavigationManager{
    private pages: IPageContainer = { };
    private errorPages: ViewPage[] = [];
    onNavigate = newEvent<INavEvent>("NavigationManager.onNavigate");
    private defaultPath: string = "";
    private currentPath: string = "";

    // TODO: Move out to utility
    private getParts(path: string): string[]{
        return this.removeEmptyEntries(path.split("/"));
    }

    // TODO: Move out to utility
    private removeEmptyEntries(array: string[]): string[]{
        let newArray: string[] = [];
        array.map((v) => {
            if (v.length > 0){
                newArray.push(v);
            }
        });
        return newArray;
    }

    private getErrorPage(statusCode: number): ViewPage{
        if (this.errorPages[statusCode]){
            return this.errorPages[statusCode];
        }
        throw Error("Status page: " + statusCode + " is not defined");
    }

    public setDefaultPath(path: string){
        this.defaultPath = path;
    }

    public navigateTo(path: string){
        let parts = this.getParts(path);
        let curPage: IPageContainer | ViewPage = this.pages;
        this.currentPath = parts.join("/");
        for(let i = 0; i < parts.length; i++){
            let a = parts[i];
            if (isViewPage(curPage)){
                this.onNavigate({target: this, page: curPage, uri: path, subPage: parts.slice(i, parts.length).join("/")});
                return;
            }
            else{
                let cur: IPageContainer | ViewPage = curPage[a];
                if (!cur){
                    this.onNavigate({target: this, page: this.getErrorPage(404), subPage: "", uri: path });
                    return;
                    //throw Error("404 Page not found");
                }
                curPage = cur;
            }
        }
        if (isViewPage(curPage)){
            this.onNavigate({target: this, page: curPage, uri: path, subPage: ""});
            return;
        }
        else{
            this.onNavigate({target: this, page: this.getErrorPage(404), subPage: "", uri: path });
            //throw Error("404 Page not found");
        }
    }

    public navigateToDefault(): void{
        this.navigateTo(this.defaultPath);
    }

    public navigateToError(statusCode: number): void{
        this.onNavigate({target: this, page: this.getErrorPage(statusCode), subPage: "", uri: statusCode.toString() });
    }

    public registerPage(path: string, page: ViewPage){
        let parts = this.getParts(path);
        if (parts.length === 0){
            throw Error("Can't add page to index element");
        }
        page.setPath(parts.join("/"));
        let curObj = this.pages;

        
        for(let i = 0; i < parts.length - 1; i++){
            let a = parts[i];
            if (a.length === 0){
                continue;
            }
            let temp: IPageContainer | ViewPage = curObj[a];
            if (!temp){
                temp = {};
                curObj[a] = temp;
            }
            else if (!isViewPage(temp)){
                temp = curObj[a];
            }
            if (isViewPage(temp)){
                throw Error("Can't assign a IPageContainer to a ViewPage");
            }
            curObj = temp;
        }
        curObj[parts[parts.length - 1]] = page;
        
    }

    public registerErrorPage(statusCode: number, page: ViewPage){
        this.errorPages[statusCode] = page;
    }

    /**
     * Checks to see if the link is part of the current path, and also mark them as active if they are.
     * @param links The links to check
     */
    public checkLinks(links: ILink[]): void
    /**
     * Checks to see if the link is part of the current path, or the default page to the given ViewPage. Also mark them as active if they are.
     * @param links The links to check
     * @param viewPage ViewPage to get defaultPage information from
     */
    public checkLinks(links: ILink[], viewPage: ViewPage): void
    public checkLinks(links: ILink[], viewPage?: ViewPage): void {
        let checkUrl = this.currentPath;
        if (viewPage && viewPage.pagePath === checkUrl){
            checkUrl += "/" + viewPage.defaultPage;
        }
        for(let l of links){
            if (!l.uri){
                continue;
            }
            let a = this.getParts(l.uri).join("/");
            l.active = a === checkUrl.substr(0, a.length)
        }
    }

    public refresh(){
        this.navigateTo(this.currentPath);
    }
}
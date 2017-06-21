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

class DummyUserProvider implements IUserProvider{
    private localData: IDummyUser[];

    constructor(){
        this.localData = [
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
        return this.localData;
    }

    tryLogin(username: string, password: string): IUser | null {
        for(let u of this.localData){
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

interface IAssignementProvider{

}

class AssignmentManager {

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
    pages: IPageContainer = { };
    onNavigate = newEvent<INavEvent>("NavigationManager.onNavigate");
    currentPath: string = "";

    getParts(path: string): string[]{
        return this.removeEmptyEntries(path.split("/"));
    }

    removeEmptyEntries(array: string[]): string[]{
        let newArray: string[] = [];
        array.map((v) => {
            if (v.length > 0){
                newArray.push(v);
            }
        });
        return newArray;
    }

    setDefaultPath(path: string){
        this.currentPath = path;
    }

    navigateToDefault(): void{
        this.navigateTo(this.currentPath);
    }

    registerPage(path: string, page: ViewPage){
        let parts = this.getParts(path);
        if (parts.length === 0){
            throw Error("Can't add page to index element");
        }
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

    navigateTo(path: string){
        let parts = this.getParts(path);
        let curPage: IPageContainer | ViewPage = this.pages;
        for(let i = 0; i < parts.length; i++){
            let a = parts[i];
            if (isViewPage(curPage)){
                this.onNavigate({target: this, page: curPage, uri: path, subPage: parts.slice(i, parts.length).join("/")});
                return;
            }
            else{
                let cur: IPageContainer | ViewPage = curPage[a];
                if (!cur){
                    throw Error("404 Page not found");
                }
                curPage = cur;
            }
        }
        if (isViewPage(curPage)){
            this.onNavigate({target: this, page: curPage, uri: path, subPage: ""});
            return;
        }
        else{
            throw Error("404 Page not found");
        }
    }
}
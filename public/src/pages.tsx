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

class UserViewer extends React.Component<any, undefined> {
    render(){
        return <DynamicTable 
            header={["ID","First name", "Last name", "Email", "StudentID"]} 
            data={this.props.users} 
            selector={(item: IUser) => [item.id.toString(), item.firstName, item.lastName, item.email, item.personId.toString()]} 
            >
        </DynamicTable>
    }
}

class HelloView extends React.Component<any, undefined>{
    render(){
        return <h1>Hello world</h1>
    }
}

class StudentPage extends ViewPage {
    navMan: NavigationManager;
    constructor(users: UserManager, navMan: NavigationManager){
        super();

        this.navMan = navMan;
        this.defaultPage = "opsys/lab1";
        this.pages["opsys/lab1"] = <h1>Lab1</h1>;
        this.pages["opsys/lab2"] = <h1>Lab2</h1>;
        this.pages["opsys/lab3"] = <h1>Lab3</h1>;
        this.pages["opsys/lab4"] = <h1>Lab4</h1>;
        this.pages["user"] = <UserViewer users={users.getAllUser()}></UserViewer>;
        this.pages["hello"] = <HelloView></HelloView>;
    }

    renderMenu(key: number): JSX.Element[]{
        if (key === 0){
            let labLinks = [
                {name: "Lab 1", uri: this.pagePath + "/opsys/lab1"},
                {name: "Lab 2", uri: this.pagePath + "/opsys/lab2"}, 
                {name: "Lab 3", uri: this.pagePath + "/opsys/lab3"},
                {name: "Lab 4", uri: this.pagePath + "/opsys/lab4"},
            ];
            let settings = [
                {name: "Users", uri: this.pagePath + "/user"},
                {name: "Hello world", uri: this.pagePath + "/hello"}
            ];

            this.navMan.checkLinks(labLinks, this);
            this.navMan.checkLinks(settings, this);

            return [
                <h4 key={0}>Labs</h4>,
                <NavMenu key={1} links={labLinks} onClick={link => this.handleClick(link)}></NavMenu>,
                <h4 key={2}>Settings</h4>,
                <NavMenu key={3} links={settings} onClick={link => this.handleClick(link)}></NavMenu>
            ];
        }
        return [];
    }

    handleClick(link: ILink){
        if (link.uri){
            this.navMan.navigateTo(link.uri);
        }
    }
}

class TeacherPage extends ViewPage {
    navMan: NavigationManager;
    constructor(users: UserManager, navMan: NavigationManager){
        super();

        this.navMan = navMan;
        this.defaultPage = "opsys/lab1";
        this.pages["opsys/lab1"] = <h1>Teacher Lab1</h1>;
        this.pages["opsys/lab2"] = <h1>Teacher Lab2</h1>;
        this.pages["opsys/lab3"] = <h1>Teacher Lab3</h1>;
        this.pages["opsys/lab4"] = <h1>Teacher Lab4</h1>;
        this.pages["user"] = <UserViewer users={users.getAllUser()}></UserViewer>;
        this.pages["hello"] = <HelloView></HelloView>;
    }

    handleClick(link: ILink){
        if (link.uri){
            this.navMan.navigateTo(link.uri);
        }
    }

    renderMenu(menu: number): JSX.Element[]{
        if (menu === 0){
            let labLinks = [
                {name: "Teacher Lab 1", uri: this.pagePath + "/opsys/lab1"},
                {name: "Teacher Lab 2", uri: this.pagePath + "/opsys/lab2"}, 
                {name: "Teacher Lab 3", uri: this.pagePath + "/opsys/lab3"},
                {name: "Teacher Lab 4", uri: this.pagePath + "/opsys/lab4"},
            ];

            let settings = [
                {name: "Users", uri: this.pagePath + "/user"},
                {name: "Hello world", uri: this.pagePath + "/hello"}
            ];

            this.navMan.checkLinks(labLinks, this);
            this.navMan.checkLinks(settings, this);

            return [
                <h4 key={0}>Labs</h4>,
                <NavMenu key={1} links={labLinks} onClick={link => this.handleClick(link)}></NavMenu>,
                <h4 key={4}>Settings</h4>,
                <NavMenu key={3} links={settings} onClick={link => this.handleClick(link)}></NavMenu>
            ];
        }
        return [];
    }
}

class HomePage extends ViewPage{
    constructor(){
        super();
        this.defaultPage = "index";
        this.pages["index"] = <h1>Welcome to autograder</h1>;
    }
}

class ErrorPage extends ViewPage{
    constructor(){
        super();
        this.defaultPage = "404";
        this.pages["404"] = <div><h1>404 Page not found</h1><p>The page you where looking for does not exist</p></div>
    }
}
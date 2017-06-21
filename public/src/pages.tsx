function isViewPage(item: any): item is ViewPage {
    if (item.getMenu){
        return true;
    }
    return false;
}

class ViewPage{
    menus: (JSX.Element[])[] = [];
    pages: any = {};
    template: string | null = null;
    defaultPage: string = "";

    getMenu(menu: number): JSX.Element[] | null{
        if (this.menus.length > menu){
            return this.menus[menu];
        }
        return null;
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
        this.menus[0] = [
            <h4 key={0}>Labs</h4>,
            <NavMenu key={1} links={[
                    {name: "Lab 1", uri: "opsys/lab1"},
                    {name: "Lab 2", uri: "opsys/lab2"}, 
                    {name: "Lab 3", uri: "opsys/lab3"},
                    {name: "Lab 4", uri: "opsys/lab4"},
                    
                ]} 
                onClick={(link) => {this.handleClick(link)}}></NavMenu>,
            <h4 key={4}>Settings</h4>,
            <NavMenu key={3} links={[
                    {name: "Users", uri: "user"},
                    {name: "Hello world", uri: "hello"}
                ]}
                onClick={(link) => {this.handleClick(link)}}></NavMenu>
        ];
        this.defaultPage = "opsys/lab1";
        this.pages["opsys/lab1"] = <h1>Lab1</h1>;
        this.pages["opsys/lab2"] = <h1>Lab2</h1>;
        this.pages["opsys/lab3"] = <h1>Lab3</h1>;
        this.pages["opsys/lab4"] = <h1>Lab4</h1>;
        this.pages["user"] = <UserViewer users={users.getAllUser()}></UserViewer>;
        this.pages["hello"] = <HelloView></HelloView>;
    }

    handleClick(link: ILink){
        this.navMan.navigateTo("app/student/" + link.uri);
    }
}

class TeacherPage extends ViewPage {
    navMan: NavigationManager;
    constructor(users: UserManager, navMan: NavigationManager){
        super();

        this.navMan = navMan;
        this.menus[0] = [
            <h4 key={0}>Labs</h4>,
            <NavMenu key={1} links={[
                    {name: "Teacher Lab 1", uri: "opsys/lab1"},
                    {name: "Teacher Lab 2", uri: "opsys/lab2"}, 
                    {name: "Teacher Lab 3", uri: "opsys/lab3"},
                    {name: "Teacher Lab 4", uri: "opsys/lab4"},
                    
                ]} 
                onClick={(link) => {this.handleClick(link)}}></NavMenu>,
            <h4 key={4}>Settings</h4>,
            <NavMenu key={3} links={[
                    {name: "Users", uri: "user"},
                    {name: "Hello world", uri: "hello"}
                ]}
                onClick={(link) => {this.handleClick(link)}}></NavMenu>
        ];
        this.defaultPage = "opsys/lab1";
        this.pages["opsys/lab1"] = <h1>Teacher Lab1</h1>;
        this.pages["opsys/lab2"] = <h1>Teacher Lab2</h1>;
        this.pages["opsys/lab3"] = <h1>Teacher Lab3</h1>;
        this.pages["opsys/lab4"] = <h1>Teacher Lab4</h1>;
        this.pages["user"] = <UserViewer users={users.getAllUser()}></UserViewer>;
        this.pages["hello"] = <HelloView></HelloView>;
    }

    handleClick(link: ILink){
        this.navMan.navigateTo("app/teacher/" + link.uri);
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
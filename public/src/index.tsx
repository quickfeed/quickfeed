let topLinks: ILink[] = [
    { name: "Teacher", uri: "app/teacher/" },
    { name: "Student", uri: "app/student/" },
    { name: "Admin", uri: "app/admin" }
]

interface AutoGraderState{
    pages:(ILink[])[];
    activePage?: ViewPage;
    currentPage: number;
}

interface AutoGraderProps{
    userManager: UserManager;
    navigationManager: NavigationManager;
}

class AutoGrader extends React.Component<AutoGraderProps, AutoGraderState>{
    private userManager: UserManager;
    private navMan: NavigationManager;
    private studentPage: StudentPage;
    private subPage: string;
    
    constructor(props: any){
        super();
        
        this.userManager = props.userManager;
        this.navMan = props.navigationManager;

        this.state = {
            activePage: undefined,
            pages: [ ],
            currentPage: 0
        }
        
        this.navMan.onNavigate.addEventListener((e: INavEvent) => {
            this.subPage = e.subPage;
            let old = this.state.activePage;
            this.setState({activePage: e.page});            
        });
    }

    componentDidMount() {
        this.navMan.navigateToDefault();
    }

    private handleClick(link: ILink){
        if (link.uri){
            this.navMan.navigateTo(link.uri);
        }
        else{
            console.warn("Warning! Empty link detected", link);
        }
    }

    private renderActiveMenu(menu: number): JSX.Element[] | string {
        if (this.state.activePage){
            let temp = this.state.activePage.getMenu(menu);
            if (temp){
                return temp;
            }
        }
        return "";
    }

    private renderActivePage(page: string):JSX.Element{
        if (this.state.activePage){
            if(!this.state.activePage.pages[this.state.activePage.defaultPage]){
                console.warn("Warning! Missing default page for " + (this.state.activePage as any).constructor.name, this.state.activePage);
            }
            if (this.state.activePage.pages[page]){
                return this.state.activePage.pages[page];
            }
            else if (this.state.activePage.pages[this.state.activePage.defaultPage]){
                return this.state.activePage.pages[this.state.activePage.defaultPage];
            }
        }
        return <h1>404 Page not found</h1>
    }

    private renderTemplate(name: string | null){
        let body: JSX.Element;
        console.log("rendering template: " + name);
        switch(name){
            case "frontpage":
                body = (
                <Row className="container-fluid">
                    <div className="col-xs-12">
                        { this.renderActivePage(this.subPage) }
                    </div>
                </Row>
                );
            default:
                body = (
                <Row className="container-fluid">
                    <div className="col-md-2 col-sm-3 col-xs-12">
                        { this.renderActiveMenu(0) }
                    </div>
                    <div className="col-md-10 col-sm-9 col-xs-12">
                        { this.renderActivePage(this.subPage) }
                    </div>
                </Row>
                );
        }
        return (
        <div>
            <NavBar id="top-bar" isFluid={false} isInverse={true} links={topLinks} onClick={(link) => this.handleClick(link)} brandName="Auto Grader"></NavBar>
            {body}
        </div>);
    }

    render(){
        if (this.state.activePage){
            return this.renderTemplate(this.state.activePage.template)
        }
        else{
            return <h1>404 not found</h1>;
        }
    }
}

// Just to make them globaly available for easier debugging
let userMan = new UserManager(new DummyUserProvider());
let navMan = new NavigationManager();

/**
 * @description The main entry point for the application. No other code should be executet outside this function
 */
function main(){
    
    let user = userMan.tryLogin("test@testersen.no","1234");

    navMan.setDefaultPath("app/home");
    navMan.registerPage("app/home", new HomePage());
    navMan.registerPage("app/student", new StudentPage(userMan, navMan));
    navMan.registerPage("app/teacher", new TeacherPage(userMan, navMan));

    navMan.registerErrorPage(404, new ErrorPage());
    navMan.onNavigate.addEventListener((e) => {console.log(e)});

    ReactDOM.render(
        <AutoGrader userManager={userMan} navigationManager={navMan}>

        </AutoGrader>,
        document.getElementById("root")
    );
}
main();

// <NavMenu links={ this.state.pages[this.state.currentPage] }  onClick={ (link) => { this.handleClick(link) } }></NavMenu><NavMenuFormatable items={ this.userManager.getAllUser() } formater={ (item: IUser) => { return item.firstName + " " + item.lastName }} onClick={(item: IUser) => { console.log("You clicked on: "); console.log(item);}}></NavMenuFormatable>
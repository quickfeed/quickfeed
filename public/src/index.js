var __extends = (this && this.__extends) || (function () {
    var extendStatics = Object.setPrototypeOf ||
        ({ __proto__: [] } instanceof Array && function (d, b) { d.__proto__ = b; }) ||
        function (d, b) { for (var p in b) if (b.hasOwnProperty(p)) d[p] = b[p]; };
    return function (d, b) {
        extendStatics(d, b);
        function __() { this.constructor = d; }
        d.prototype = b === null ? Object.create(b) : (__.prototype = b.prototype, new __());
    };
})();
var topLinks = [
    { name: "Teacher", uri: "app/teacher/" },
    { name: "Student", uri: "app/student/" },
    { name: "Admin", uri: "app/admin" }
];
var AutoGrader = (function (_super) {
    __extends(AutoGrader, _super);
    function AutoGrader(props) {
        var _this = _super.call(this) || this;
        _this.userManager = props.userManager;
        _this.navMan = props.navigationManager;
        _this.state = {
            activePage: undefined,
            pages: [],
            currentPage: 0
        };
        _this.navMan.onNavigate.addEventListener(function (e) {
            _this.subPage = e.subPage;
            var old = _this.state.activePage;
            _this.setState({ activePage: e.page });
        });
        return _this;
    }
    AutoGrader.prototype.componentDidMount = function () {
        this.navMan.navigateToDefault();
    };
    AutoGrader.prototype.handleClick = function (link) {
        if (link.uri) {
            this.navMan.navigateTo(link.uri);
        }
        else {
            console.warn("Warning! Empty link detected", link);
        }
    };
    AutoGrader.prototype.renderActiveMenu = function (menu) {
        if (this.state.activePage) {
            var temp = this.state.activePage.getMenu(menu);
            if (temp) {
                return temp;
            }
        }
        return "";
    };
    AutoGrader.prototype.renderActivePage = function (page) {
        if (this.state.activePage) {
            if (!this.state.activePage.pages[this.state.activePage.defaultPage]) {
                console.warn("Warning! Missing default page for " + this.state.activePage.constructor.name, this.state.activePage);
            }
            if (this.state.activePage.pages[page]) {
                return this.state.activePage.pages[page];
            }
            else if (this.state.activePage.pages[this.state.activePage.defaultPage]) {
                return this.state.activePage.pages[this.state.activePage.defaultPage];
            }
        }
        return React.createElement("h1", null, "404 Page not found");
    };
    AutoGrader.prototype.renderTemplate = function (name) {
        var _this = this;
        var body;
        console.log("rendering template: " + name);
        switch (name) {
            case "frontpage":
                body = (React.createElement(Row, { className: "container-fluid" },
                    React.createElement("div", { className: "col-xs-12" }, this.renderActivePage(this.subPage))));
            default:
                body = (React.createElement(Row, { className: "container-fluid" },
                    React.createElement("div", { className: "col-md-2 col-sm-3 col-xs-12" }, this.renderActiveMenu(0)),
                    React.createElement("div", { className: "col-md-10 col-sm-9 col-xs-12" }, this.renderActivePage(this.subPage))));
        }
        return (React.createElement("div", null,
            React.createElement(NavBar, { id: "top-bar", isFluid: false, isInverse: true, links: topLinks, onClick: function (link) { return _this.handleClick(link); }, brandName: "Auto Grader" }),
            body));
    };
    AutoGrader.prototype.render = function () {
        if (this.state.activePage) {
            return this.renderTemplate(this.state.activePage.template);
        }
        else {
            return React.createElement("h1", null, "404 not found");
        }
    };
    return AutoGrader;
}(React.Component));
var tempData = new TempDataProvider();
var userMan = new UserManager(tempData);
var courseMan = new CourseManager(tempData);
var navMan = new NavigationManager();
function main() {
    var user = userMan.tryLogin("test@testersen.no", "1234");
    navMan.setDefaultPath("app/home");
    navMan.registerPage("app/home", new HomePage());
    navMan.registerPage("app/student", new StudentPage(userMan, navMan));
    navMan.registerPage("app/teacher", new TeacherPage(userMan, navMan));
    navMan.registerErrorPage(404, new ErrorPage());
    navMan.onNavigate.addEventListener(function (e) { console.log(e); });
    ReactDOM.render(React.createElement(AutoGrader, { userManager: userMan, navigationManager: navMan }), document.getElementById("root"));
}
main();

/******/ (function(modules) { // webpackBootstrap
/******/ 	// The module cache
/******/ 	var installedModules = {};
/******/
/******/ 	// The require function
/******/ 	function __webpack_require__(moduleId) {
/******/
/******/ 		// Check if module is in cache
/******/ 		if(installedModules[moduleId]) {
/******/ 			return installedModules[moduleId].exports;
/******/ 		}
/******/ 		// Create a new module (and put it into the cache)
/******/ 		var module = installedModules[moduleId] = {
/******/ 			i: moduleId,
/******/ 			l: false,
/******/ 			exports: {}
/******/ 		};
/******/
/******/ 		// Execute the module function
/******/ 		modules[moduleId].call(module.exports, module, module.exports, __webpack_require__);
/******/
/******/ 		// Flag the module as loaded
/******/ 		module.l = true;
/******/
/******/ 		// Return the exports of the module
/******/ 		return module.exports;
/******/ 	}
/******/
/******/
/******/ 	// expose the modules object (__webpack_modules__)
/******/ 	__webpack_require__.m = modules;
/******/
/******/ 	// expose the module cache
/******/ 	__webpack_require__.c = installedModules;
/******/
/******/ 	// define getter function for harmony exports
/******/ 	__webpack_require__.d = function(exports, name, getter) {
/******/ 		if(!__webpack_require__.o(exports, name)) {
/******/ 			Object.defineProperty(exports, name, {
/******/ 				configurable: false,
/******/ 				enumerable: true,
/******/ 				get: getter
/******/ 			});
/******/ 		}
/******/ 	};
/******/
/******/ 	// getDefaultExport function for compatibility with non-harmony modules
/******/ 	__webpack_require__.n = function(module) {
/******/ 		var getter = module && module.__esModule ?
/******/ 			function getDefault() { return module['default']; } :
/******/ 			function getModuleExports() { return module; };
/******/ 		__webpack_require__.d(getter, 'a', getter);
/******/ 		return getter;
/******/ 	};
/******/
/******/ 	// Object.prototype.hasOwnProperty.call
/******/ 	__webpack_require__.o = function(object, property) { return Object.prototype.hasOwnProperty.call(object, property); };
/******/
/******/ 	// __webpack_public_path__
/******/ 	__webpack_require__.p = "";
/******/
/******/ 	// Load entry module and return exports
/******/ 	return __webpack_require__(__webpack_require__.s = 6);
/******/ })
/************************************************************************/
/******/ ([
/* 0 */
/***/ (function(module, exports) {

module.exports = React;

/***/ }),
/* 1 */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
function isViewPage(item) {
    if (item instanceof ViewPage) {
        return true;
    }
    return false;
}
exports.isViewPage = isViewPage;
var ViewPage = (function () {
    function ViewPage() {
        this.pages = {};
        this.template = null;
        this.defaultPage = "";
    }
    ViewPage.prototype.setPath = function (path) {
        this.pagePath = path;
    };
    ViewPage.prototype.renderMenu = function (menu) {
        return [];
    };
    return ViewPage;
}());
exports.ViewPage = ViewPage;


/***/ }),
/* 2 */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

function __export(m) {
    for (var p in m) if (!exports.hasOwnProperty(p)) exports[p] = m[p];
}
Object.defineProperty(exports, "__esModule", { value: true });
__export(__webpack_require__(8));
__export(__webpack_require__(3));
__export(__webpack_require__(9));
__export(__webpack_require__(10));
__export(__webpack_require__(11));
__export(__webpack_require__(12));


/***/ }),
/* 3 */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

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
Object.defineProperty(exports, "__esModule", { value: true });
var React = __webpack_require__(0);
var NavHeaderBar = (function (_super) {
    __extends(NavHeaderBar, _super);
    function NavHeaderBar() {
        return _super !== null && _super.apply(this, arguments) || this;
    }
    NavHeaderBar.prototype.componentDidMount = function () {
        console.log(this.refs.button);
        var temp = this.refs.button;
        temp.setAttribute("data-toggle", "collapse");
        temp.setAttribute("data-target", "#" + this.props.id);
        temp.setAttribute("aria-expanded", "false");
    };
    NavHeaderBar.prototype.render = function () {
        return React.createElement("div", { className: "navbar-header" },
            React.createElement("button", { ref: "button", type: "button", className: "navbar-toggle collapsed" },
                React.createElement("span", { className: "sr-only" }, "Toggle navigation"),
                React.createElement("span", { className: "icon-bar" }),
                React.createElement("span", { className: "icon-bar" }),
                React.createElement("span", { className: "icon-bar" })),
            React.createElement("a", { className: "navbar-brand", onClick: function (e) { e.preventDefault(); }, href: "#" }, this.props.brandName));
    };
    return NavHeaderBar;
}(React.Component));
exports.NavHeaderBar = NavHeaderBar;


/***/ }),
/* 4 */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

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
Object.defineProperty(exports, "__esModule", { value: true });
var React = __webpack_require__(0);
var components_1 = __webpack_require__(2);
var UserViewer = (function (_super) {
    __extends(UserViewer, _super);
    function UserViewer() {
        return _super !== null && _super.apply(this, arguments) || this;
    }
    UserViewer.prototype.render = function () {
        return React.createElement(components_1.DynamicTable, { header: ["ID", "First name", "Last name", "Email", "StudentID"], data: this.props.users, selector: function (item) { return [item.id.toString(), item.firstName, item.lastName, item.email, item.personId.toString()]; } });
    };
    return UserViewer;
}(React.Component));
exports.UserViewer = UserViewer;


/***/ }),
/* 5 */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

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
Object.defineProperty(exports, "__esModule", { value: true });
var React = __webpack_require__(0);
var HelloView = (function (_super) {
    __extends(HelloView, _super);
    function HelloView() {
        return _super !== null && _super.apply(this, arguments) || this;
    }
    HelloView.prototype.render = function () {
        return React.createElement("h1", null, "Hello world");
    };
    return HelloView;
}(React.Component));
exports.HelloView = HelloView;


/***/ }),
/* 6 */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

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
Object.defineProperty(exports, "__esModule", { value: true });
var React = __webpack_require__(0);
var ReactDOM = __webpack_require__(7);
var components_1 = __webpack_require__(2);
var NavigationManager_1 = __webpack_require__(13);
var UserManager_1 = __webpack_require__(15);
var StudentPage_1 = __webpack_require__(16);
var TempDataProvider_1 = __webpack_require__(17);
var CourseManager_1 = __webpack_require__(18);
var HomePage_1 = __webpack_require__(20);
var ErrorPage_1 = __webpack_require__(21);
var TeacherPage_1 = __webpack_require__(22);
var topLinks = [
    { name: "Teacher", uri: "app/teacher/", active: false },
    { name: "Student", uri: "app/student/", active: false },
    { name: "Admin", uri: "app/admin", active: false }
];
var AutoGrader = (function (_super) {
    __extends(AutoGrader, _super);
    function AutoGrader(props) {
        var _this = _super.call(this) || this;
        _this.userManager = props.userManager;
        _this.navMan = props.navigationManager;
        _this.state = {
            activePage: undefined,
            topLink: topLinks
        };
        _this.navMan.onNavigate.addEventListener(function (e) {
            _this.subPage = e.subPage;
            var old = _this.state.activePage;
            var tempLink = _this.state.topLink.slice();
            _this.checkLinks(tempLink);
            _this.setState({ activePage: e.page, topLink: tempLink });
        });
        return _this;
    }
    AutoGrader.prototype.componentDidMount = function () {
        var curUrl = location.pathname;
        if (curUrl === "/") {
            this.navMan.navigateToDefault();
        }
        else {
            this.navMan.navigateTo(curUrl);
        }
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
            return this.state.activePage.renderMenu(menu);
        }
        return "";
    };
    AutoGrader.prototype.renderActivePage = function (page) {
        var curPage = this.state.activePage;
        if (curPage) {
            if (!curPage.pages[curPage.defaultPage]) {
                console.warn("Warning! Missing default page for " + curPage.constructor.name, curPage);
            }
            if (curPage.pages[page]) {
                return curPage.pages[page];
            }
            else if (curPage.pages[curPage.defaultPage]) {
                return curPage.pages[curPage.defaultPage];
            }
        }
        return React.createElement("h1", null, "404 Page not found");
    };
    AutoGrader.prototype.checkLinks = function (links) {
        this.navMan.checkLinks(links);
    };
    AutoGrader.prototype.renderTemplate = function (name) {
        var _this = this;
        var body;
        console.log("rendering template: " + name);
        switch (name) {
            case "frontpage":
                body = (React.createElement(components_1.Row, { className: "container-fluid" },
                    React.createElement("div", { className: "col-xs-12" }, this.renderActivePage(this.subPage))));
            default:
                body = (React.createElement(components_1.Row, { className: "container-fluid" },
                    React.createElement("div", { className: "col-md-2 col-sm-3 col-xs-12" }, this.renderActiveMenu(0)),
                    React.createElement("div", { className: "col-md-10 col-sm-9 col-xs-12" }, this.renderActivePage(this.subPage))));
        }
        return (React.createElement("div", null,
            React.createElement(components_1.NavBar, { id: "top-bar", isFluid: false, isInverse: true, links: topLinks, onClick: function (link) { return _this.handleClick(link); }, brandName: "Auto Grader" }),
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
var tempData = new TempDataProvider_1.TempDataProvider();
var userMan = new UserManager_1.UserManager(tempData);
var courseMan = new CourseManager_1.CourseManager(tempData);
var navMan = new NavigationManager_1.NavigationManager(history);
function main() {
    var user = userMan.tryLogin("test@testersen.no", "1234");
    navMan.setDefaultPath("app/home");
    navMan.registerPage("app/home", new HomePage_1.HomePage());
    navMan.registerPage("app/student", new StudentPage_1.StudentPage(userMan, navMan));
    navMan.registerPage("app/teacher", new TeacherPage_1.TeacherPage(userMan, navMan));
    navMan.registerErrorPage(404, new ErrorPage_1.ErrorPage());
    navMan.onNavigate.addEventListener(function (e) { console.log(e); });
    ReactDOM.render(React.createElement(AutoGrader, { userManager: userMan, navigationManager: navMan }), document.getElementById("root"));
}
main();


/***/ }),
/* 7 */
/***/ (function(module, exports) {

module.exports = ReactDOM;

/***/ }),
/* 8 */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

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
Object.defineProperty(exports, "__esModule", { value: true });
var React = __webpack_require__(0);
var NavHeaderBar_1 = __webpack_require__(3);
var NavBar = (function (_super) {
    __extends(NavBar, _super);
    function NavBar() {
        return _super !== null && _super.apply(this, arguments) || this;
    }
    NavBar.prototype.renderIsFluid = function () {
        var name = "container";
        if (this.props.isFluid) {
            name += "-fluid";
        }
        return name;
    };
    NavBar.prototype.renderNavBarClass = function () {
        var name = "navbar navbar-absolute-top";
        if (this.props.isInverse) {
            name += " navbar-inverse";
        }
        else {
            name += " navbar-default";
        }
        return name;
    };
    NavBar.prototype.handleClick = function (link) {
        if (this.props.onClick) {
            this.props.onClick(link);
        }
    };
    NavBar.prototype.render = function () {
        var _this = this;
        var items = this.props.links.map(function (v, i) {
            var active = "";
            if (v.active) {
                active = "active";
            }
            return React.createElement("li", { className: active, key: i },
                React.createElement("a", { href: "/" + v.uri, onClick: function (e) { e.preventDefault(); _this.handleClick(v); } }, v.name));
        });
        return React.createElement("nav", { className: this.renderNavBarClass() },
            React.createElement("div", { className: this.renderIsFluid() },
                React.createElement(NavHeaderBar_1.NavHeaderBar, { id: this.props.id, brandName: this.props.brandName }),
                React.createElement("div", { className: "collapse navbar-collapse", id: this.props.id },
                    React.createElement("ul", { className: "nav navbar-nav" }, items))));
    };
    return NavBar;
}(React.Component));
exports.NavBar = NavBar;


/***/ }),
/* 9 */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

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
Object.defineProperty(exports, "__esModule", { value: true });
var React = __webpack_require__(0);
var NavMenu = (function (_super) {
    __extends(NavMenu, _super);
    function NavMenu() {
        return _super !== null && _super.apply(this, arguments) || this;
    }
    NavMenu.prototype.render = function () {
        var _this = this;
        var items = this.props.links.map(function (v, i) {
            var active = "";
            if (v.active) {
                active = "active";
            }
            return React.createElement("li", { className: active, key: i },
                React.createElement("a", { href: "/" + v.uri, onClick: function (e) { e.preventDefault(); if (_this.props.onClick)
                        _this.props.onClick(v); } }, v.name));
        });
        return React.createElement("ul", { className: "nav nav-pills nav-stacked" }, items);
    };
    return NavMenu;
}(React.Component));
exports.NavMenu = NavMenu;


/***/ }),
/* 10 */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

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
Object.defineProperty(exports, "__esModule", { value: true });
var React = __webpack_require__(0);
var NavMenuFormatable = (function (_super) {
    __extends(NavMenuFormatable, _super);
    function NavMenuFormatable() {
        return _super !== null && _super.apply(this, arguments) || this;
    }
    NavMenuFormatable.prototype.renderObj = function (item) {
        if (this.props.formater) {
            return this.props.formater(item);
        }
        return item.toString();
    };
    NavMenuFormatable.prototype.handleItemClick = function (item) {
        if (this.props.onClick) {
            this.props.onClick(item);
        }
    };
    NavMenuFormatable.prototype.render = function () {
        var _this = this;
        var items = this.props.items.map(function (v, i) {
            return React.createElement("li", { key: i },
                React.createElement("a", { href: "#", onClick: function () { _this.handleItemClick(v); } }, _this.renderObj(v)));
        });
        return React.createElement("ul", { className: "nav nav-pills nav-stacked" }, items);
    };
    return NavMenuFormatable;
}(React.Component));
exports.NavMenuFormatable = NavMenuFormatable;


/***/ }),
/* 11 */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

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
Object.defineProperty(exports, "__esModule", { value: true });
var React = __webpack_require__(0);
var DynamicTable = (function (_super) {
    __extends(DynamicTable, _super);
    function DynamicTable() {
        return _super !== null && _super.apply(this, arguments) || this;
    }
    DynamicTable.prototype.renderCells = function (values) {
        return values.map(function (v, i) {
            return React.createElement("td", { key: i }, v);
        });
    };
    DynamicTable.prototype.renderRow = function (item, i) {
        return React.createElement("tr", { key: i }, this.renderCells(this.props.selector(item)));
    };
    DynamicTable.prototype.render = function () {
        var _this = this;
        var rows = this.props.data.map(function (v, i) {
            return _this.renderRow(v, i);
        });
        return React.createElement("table", { className: "table" },
            React.createElement("thead", null,
                React.createElement("tr", null, this.renderCells(this.props.header))),
            React.createElement("tbody", null, rows));
    };
    return DynamicTable;
}(React.Component));
exports.DynamicTable = DynamicTable;


/***/ }),
/* 12 */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
var React = __webpack_require__(0);
function Row(props) {
    return React.createElement("div", { className:  true ? props.className : "" }, props.children);
}
exports.Row = Row;


/***/ }),
/* 13 */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
var event_1 = __webpack_require__(14);
var ViewPage_1 = __webpack_require__(1);
var NavigationManager = (function () {
    function NavigationManager(history) {
        var _this = this;
        this.pages = {};
        this.errorPages = [];
        this.onNavigate = event_1.newEvent("NavigationManager.onNavigate");
        this.defaultPath = "";
        this.currentPath = "";
        this.browserHistory = history;
        window.addEventListener("popstate", function (e) {
            _this.navigateTo(location.pathname, true);
        });
    }
    NavigationManager.prototype.getParts = function (path) {
        return this.removeEmptyEntries(path.split("/"));
    };
    NavigationManager.prototype.removeEmptyEntries = function (array) {
        var newArray = [];
        array.map(function (v) {
            if (v.length > 0) {
                newArray.push(v);
            }
        });
        return newArray;
    };
    NavigationManager.prototype.getErrorPage = function (statusCode) {
        if (this.errorPages[statusCode]) {
            return this.errorPages[statusCode];
        }
        throw Error("Status page: " + statusCode + " is not defined");
    };
    NavigationManager.prototype.setDefaultPath = function (path) {
        this.defaultPath = path;
    };
    NavigationManager.prototype.navigateTo = function (path, preventPush) {
        var parts = this.getParts(path);
        var curPage = this.pages;
        this.currentPath = parts.join("/");
        if (!preventPush) {
            this.browserHistory.pushState({}, "Autograder", "/" + this.currentPath);
        }
        for (var i = 0; i < parts.length; i++) {
            var a = parts[i];
            if (ViewPage_1.isViewPage(curPage)) {
                this.onNavigate({ target: this, page: curPage, uri: path, subPage: parts.slice(i, parts.length).join("/") });
                return;
            }
            else {
                var cur = curPage[a];
                if (!cur) {
                    this.onNavigate({ target: this, page: this.getErrorPage(404), subPage: "", uri: path });
                    return;
                }
                curPage = cur;
            }
        }
        if (ViewPage_1.isViewPage(curPage)) {
            this.onNavigate({ target: this, page: curPage, uri: path, subPage: "" });
            return;
        }
        else {
            this.onNavigate({ target: this, page: this.getErrorPage(404), subPage: "", uri: path });
        }
    };
    NavigationManager.prototype.navigateToDefault = function () {
        this.navigateTo(this.defaultPath);
    };
    NavigationManager.prototype.navigateToError = function (statusCode) {
        this.onNavigate({ target: this, page: this.getErrorPage(statusCode), subPage: "", uri: statusCode.toString() });
    };
    NavigationManager.prototype.registerPage = function (path, page) {
        var parts = this.getParts(path);
        if (parts.length === 0) {
            throw Error("Can't add page to index element");
        }
        page.setPath(parts.join("/"));
        var curObj = this.pages;
        for (var i = 0; i < parts.length - 1; i++) {
            var a = parts[i];
            if (a.length === 0) {
                continue;
            }
            var temp = curObj[a];
            if (!temp) {
                temp = {};
                curObj[a] = temp;
            }
            else if (!ViewPage_1.isViewPage(temp)) {
                temp = curObj[a];
            }
            if (ViewPage_1.isViewPage(temp)) {
                throw Error("Can't assign a IPageContainer to a ViewPage");
            }
            curObj = temp;
        }
        curObj[parts[parts.length - 1]] = page;
    };
    NavigationManager.prototype.registerErrorPage = function (statusCode, page) {
        this.errorPages[statusCode] = page;
    };
    NavigationManager.prototype.checkLinks = function (links, viewPage) {
        var checkUrl = this.currentPath;
        if (viewPage && viewPage.pagePath === checkUrl) {
            checkUrl += "/" + viewPage.defaultPage;
        }
        for (var _i = 0, links_1 = links; _i < links_1.length; _i++) {
            var l = links_1[_i];
            if (!l.uri) {
                continue;
            }
            var a = this.getParts(l.uri).join("/");
            l.active = a === checkUrl.substr(0, a.length);
        }
    };
    NavigationManager.prototype.refresh = function () {
        this.navigateTo(this.currentPath);
    };
    return NavigationManager;
}());
exports.NavigationManager = NavigationManager;


/***/ }),
/* 14 */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
function newEvent(info) {
    var callbacks = [];
    var handler = function EventHandler(event) {
        callbacks.map(function (v) { return v(event); });
    };
    handler.info = info;
    handler.addEventListener = function (callback) {
        callbacks.push(callback);
    };
    handler.removeEventListener = function (callback) {
        var index = callbacks.indexOf(callback);
        if (index < 0) {
            console.log(callback);
            throw Error("Event does noe exist");
        }
        callbacks.splice(index, 1);
    };
    return handler;
}
exports.newEvent = newEvent;


/***/ }),
/* 15 */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
var UserManager = (function () {
    function UserManager(userProvider) {
        this.userProvider = userProvider;
    }
    UserManager.prototype.getCurrentUser = function () {
        return this.currentUser;
    };
    UserManager.prototype.tryLogin = function (username, password) {
        var result = this.userProvider.tryLogin(username, password);
        if (result) {
            this.currentUser = result;
        }
        return result;
    };
    UserManager.prototype.getAllUser = function () {
        return this.userProvider.getAllUser();
    };
    UserManager.prototype.getUser = function (id) {
    };
    return UserManager;
}());
exports.UserManager = UserManager;


/***/ }),
/* 16 */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

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
Object.defineProperty(exports, "__esModule", { value: true });
var React = __webpack_require__(0);
var UserView_1 = __webpack_require__(4);
var HelloView_1 = __webpack_require__(5);
var components_1 = __webpack_require__(2);
var ViewPage_1 = __webpack_require__(1);
var StudentPage = (function (_super) {
    __extends(StudentPage, _super);
    function StudentPage(users, navMan) {
        var _this = _super.call(this) || this;
        _this.navMan = navMan;
        _this.defaultPage = "opsys/lab1";
        _this.pages["opsys/lab1"] = React.createElement("h1", null, "Lab1");
        _this.pages["opsys/lab2"] = React.createElement("h1", null, "Lab2");
        _this.pages["opsys/lab3"] = React.createElement("h1", null, "Lab3");
        _this.pages["opsys/lab4"] = React.createElement("h1", null, "Lab4");
        _this.pages["user"] = React.createElement(UserView_1.UserViewer, { users: users.getAllUser() });
        _this.pages["hello"] = React.createElement(HelloView_1.HelloView, null);
        return _this;
    }
    StudentPage.prototype.renderMenu = function (key) {
        var _this = this;
        if (key === 0) {
            var labLinks = [
                { name: "Lab 1", uri: this.pagePath + "/opsys/lab1" },
                { name: "Lab 2", uri: this.pagePath + "/opsys/lab2" },
                { name: "Lab 3", uri: this.pagePath + "/opsys/lab3" },
                { name: "Lab 4", uri: this.pagePath + "/opsys/lab4" },
            ];
            var settings = [
                { name: "Users", uri: this.pagePath + "/user" },
                { name: "Hello world", uri: this.pagePath + "/hello" }
            ];
            this.navMan.checkLinks(labLinks, this);
            this.navMan.checkLinks(settings, this);
            return [
                React.createElement("h4", { key: 0 }, "Labs"),
                React.createElement(components_1.NavMenu, { key: 1, links: labLinks, onClick: function (link) { return _this.handleClick(link); } }),
                React.createElement("h4", { key: 2 }, "Settings"),
                React.createElement(components_1.NavMenu, { key: 3, links: settings, onClick: function (link) { return _this.handleClick(link); } })
            ];
        }
        return [];
    };
    StudentPage.prototype.handleClick = function (link) {
        if (link.uri) {
            this.navMan.navigateTo(link.uri);
        }
    };
    return StudentPage;
}(ViewPage_1.ViewPage));
exports.StudentPage = StudentPage;


/***/ }),
/* 17 */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
var TempDataProvider = (function () {
    function TempDataProvider() {
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
    TempDataProvider.prototype.getAllUser = function () {
        return this.localUsers;
    };
    TempDataProvider.prototype.getCourses = function () {
        return this.localCourses;
    };
    TempDataProvider.prototype.getAssignments = function (courseId) {
        var temp = [];
        for (var _i = 0, _a = this.localAssignments; _i < _a.length; _i++) {
            var a = _a[_i];
            if (a.courceId === courseId) {
                temp.push(a);
            }
        }
        return temp;
    };
    TempDataProvider.prototype.tryLogin = function (username, password) {
        for (var _i = 0, _a = this.localUsers; _i < _a.length; _i++) {
            var u = _a[_i];
            if (u.email.toLocaleLowerCase() === username.toLocaleLowerCase()) {
                if (u.password === password) {
                    return u;
                }
                return null;
            }
        }
        return null;
    };
    return TempDataProvider;
}());
exports.TempDataProvider = TempDataProvider;


/***/ }),
/* 18 */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
var models_1 = __webpack_require__(19);
var CourseManager = (function () {
    function CourseManager(courseProvider) {
        this.courseProvider = courseProvider;
    }
    CourseManager.prototype.getCourses = function () {
        return this.courseProvider.getCourses();
    };
    CourseManager.prototype.getAssignments = function (courseId) {
        if (models_1.isCourse(courseId)) {
            courseId = courseId.id;
            console.log(courseId);
        }
        return this.courseProvider.getAssignments(courseId);
    };
    return CourseManager;
}());
exports.CourseManager = CourseManager;


/***/ }),
/* 19 */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
function isCourse(value) {
    console.log(value);
    return value && typeof value.id === "number" && value.name && value.tag;
}
exports.isCourse = isCourse;


/***/ }),
/* 20 */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

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
Object.defineProperty(exports, "__esModule", { value: true });
var React = __webpack_require__(0);
var ViewPage_1 = __webpack_require__(1);
var HomePage = (function (_super) {
    __extends(HomePage, _super);
    function HomePage() {
        var _this = _super.call(this) || this;
        _this.defaultPage = "index";
        _this.pages["index"] = React.createElement("h1", null, "Welcome to autograder");
        return _this;
    }
    return HomePage;
}(ViewPage_1.ViewPage));
exports.HomePage = HomePage;


/***/ }),
/* 21 */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

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
Object.defineProperty(exports, "__esModule", { value: true });
var React = __webpack_require__(0);
var ViewPage_1 = __webpack_require__(1);
var ErrorPage = (function (_super) {
    __extends(ErrorPage, _super);
    function ErrorPage() {
        var _this = _super.call(this) || this;
        _this.defaultPage = "404";
        _this.pages["404"] = React.createElement("div", null,
            React.createElement("h1", null, "404 Page not found"),
            React.createElement("p", null, "The page you where looking for does not exist"));
        return _this;
    }
    return ErrorPage;
}(ViewPage_1.ViewPage));
exports.ErrorPage = ErrorPage;


/***/ }),
/* 22 */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

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
Object.defineProperty(exports, "__esModule", { value: true });
var React = __webpack_require__(0);
var UserView_1 = __webpack_require__(4);
var HelloView_1 = __webpack_require__(5);
var components_1 = __webpack_require__(2);
var ViewPage_1 = __webpack_require__(1);
var TeacherPage = (function (_super) {
    __extends(TeacherPage, _super);
    function TeacherPage(users, navMan) {
        var _this = _super.call(this) || this;
        _this.navMan = navMan;
        _this.defaultPage = "opsys/lab1";
        _this.pages["opsys/lab1"] = React.createElement("h1", null, "Teacher Lab1");
        _this.pages["opsys/lab2"] = React.createElement("h1", null, "Teacher Lab2");
        _this.pages["opsys/lab3"] = React.createElement("h1", null, "Teacher Lab3");
        _this.pages["opsys/lab4"] = React.createElement("h1", null, "Teacher Lab4");
        _this.pages["user"] = React.createElement(UserView_1.UserViewer, { users: users.getAllUser() });
        _this.pages["hello"] = React.createElement(HelloView_1.HelloView, null);
        return _this;
    }
    TeacherPage.prototype.handleClick = function (link) {
        if (link.uri) {
            this.navMan.navigateTo(link.uri);
        }
    };
    TeacherPage.prototype.renderMenu = function (menu) {
        var _this = this;
        if (menu === 0) {
            var labLinks = [
                { name: "Teacher Lab 1", uri: this.pagePath + "/opsys/lab1" },
                { name: "Teacher Lab 2", uri: this.pagePath + "/opsys/lab2" },
                { name: "Teacher Lab 3", uri: this.pagePath + "/opsys/lab3" },
                { name: "Teacher Lab 4", uri: this.pagePath + "/opsys/lab4" },
            ];
            var settings = [
                { name: "Users", uri: this.pagePath + "/user" },
                { name: "Hello world", uri: this.pagePath + "/hello" }
            ];
            this.navMan.checkLinks(labLinks, this);
            this.navMan.checkLinks(settings, this);
            return [
                React.createElement("h4", { key: 0 }, "Labs"),
                React.createElement(components_1.NavMenu, { key: 1, links: labLinks, onClick: function (link) { return _this.handleClick(link); } }),
                React.createElement("h4", { key: 4 }, "Settings"),
                React.createElement(components_1.NavMenu, { key: 3, links: settings, onClick: function (link) { return _this.handleClick(link); } })
            ];
        }
        return [];
    };
    return TeacherPage;
}(ViewPage_1.ViewPage));
exports.TeacherPage = TeacherPage;


/***/ })
/******/ ]);
//# sourceMappingURL=bundle.js.map
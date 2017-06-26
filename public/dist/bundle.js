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
__export(__webpack_require__(13));
__export(__webpack_require__(14));
__export(__webpack_require__(15));
__export(__webpack_require__(16));


/***/ }),
/* 2 */
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
        var _this = this;
        return React.createElement("div", { className: "navbar-header" },
            React.createElement("button", { ref: "button", type: "button", className: "navbar-toggle collapsed" },
                React.createElement("span", { className: "sr-only" }, "Toggle navigation"),
                React.createElement("span", { className: "icon-bar" }),
                React.createElement("span", { className: "icon-bar" }),
                React.createElement("span", { className: "icon-bar" })),
            React.createElement("a", { className: "navbar-brand", onClick: function (e) { e.preventDefault(); _this.props.brandClick(); }, href: "/" }, this.props.brandName));
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
var components_1 = __webpack_require__(1);
var UserView = (function (_super) {
    __extends(UserView, _super);
    function UserView() {
        return _super !== null && _super.apply(this, arguments) || this;
    }
    UserView.prototype.render = function () {
        return React.createElement(components_1.DynamicTable, { header: ["ID", "First name", "Last name", "Email", "StudentID"], data: this.props.users, selector: function (item) { return [item.id.toString(), item.firstName, item.lastName, item.email, item.personId.toString()]; } });
    };
    return UserView;
}(React.Component));
exports.UserView = UserView;


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
var components_1 = __webpack_require__(1);
var NavigationManager_1 = __webpack_require__(17);
var UserManager_1 = __webpack_require__(19);
var StudentPage_1 = __webpack_require__(20);
var TempDataProvider_1 = __webpack_require__(22);
var CourseManager_1 = __webpack_require__(23);
var HomePage_1 = __webpack_require__(25);
var ErrorPage_1 = __webpack_require__(26);
var TeacherPage_1 = __webpack_require__(27);
var HelpPage_1 = __webpack_require__(28);
var topLinks = [
    { name: "Teacher", uri: "app/teacher/", active: false },
    { name: "Student", uri: "app/student/", active: false },
    { name: "Admin", uri: "app/admin", active: false },
    { name: "Help", uri: "app/help", active: false }
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
            return curPage.renderContent(page);
        }
        return React.createElement("h1", null, "404 Page not found");
    };
    AutoGrader.prototype.checkLinks = function (links) {
        this.navMan.checkLinks(links);
    };
    AutoGrader.prototype.renderTemplate = function (name) {
        var _this = this;
        var body;
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
            React.createElement(components_1.NavBar, { id: "top-bar", isFluid: false, isInverse: true, links: topLinks, onClick: function (link) { return _this.handleClick(link); }, user: this.userManager.getCurrentUser(), brandName: "Auto Grader" }),
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
function main() {
    var tempData = new TempDataProvider_1.TempDataProvider();
    var userMan = new UserManager_1.UserManager(tempData);
    var courseMan = new CourseManager_1.CourseManager(tempData);
    var navMan = new NavigationManager_1.NavigationManager(history);
    window.debugData = { tempData: tempData, userMan: userMan, courseMan: courseMan, navMan: navMan };
    var user = userMan.tryLogin("test@testersen.no", "1234");
    navMan.setDefaultPath("app/home");
    navMan.registerPage("app/home", new HomePage_1.HomePage());
    navMan.registerPage("app/student", new StudentPage_1.StudentPage(userMan, navMan, courseMan));
    navMan.registerPage("app/teacher", new TeacherPage_1.TeacherPage(userMan, navMan));
    navMan.registerPage("app/help", new HelpPage_1.HelpPage(navMan));
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
    NavBar.prototype.renderUser = function (user) {
        if (user) {
            return "Hello " + user.firstName;
        }
        return "Not logged in";
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
                React.createElement(NavHeaderBar_1.NavHeaderBar, { id: this.props.id, brandName: this.props.brandName, brandClick: function () { return _this.handleClick({ name: "Home", uri: "/" }); } }),
                React.createElement("div", { className: "collapse navbar-collapse", id: this.props.id },
                    React.createElement("ul", { className: "nav navbar-nav" }, items),
                    React.createElement("p", { className: "navbar-text navbar-right" }, this.renderUser(this.props.user)))));
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
        if (this.props.footer) {
            return (React.createElement("table", { className: "table" },
                React.createElement("thead", null,
                    React.createElement("tr", null, this.renderCells(this.props.header))),
                React.createElement("tbody", null, rows),
                React.createElement("tfoot", null,
                    React.createElement("tr", null, this.renderCells(this.props.footer)))));
        }
        return (React.createElement("table", { className: "table" },
            React.createElement("thead", null,
                React.createElement("tr", null, this.renderCells(this.props.header))),
            React.createElement("tbody", null, rows)));
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
var ProgressBar = (function (_super) {
    __extends(ProgressBar, _super);
    function ProgressBar() {
        return _super !== null && _super.apply(this, arguments) || this;
    }
    ProgressBar.prototype.render = function () {
        var progressBarStyle = {
            width: this.props.progress + "%"
        };
        return (React.createElement("div", { className: "progress" },
            React.createElement("div", { className: "progress-bar", role: "progressbar", "aria-valuenow": this.props.progress, "aria-valuemin": "0", "aria-valuemax": "100", style: progressBarStyle },
                this.props.progress,
                "%")));
    };
    return ProgressBar;
}(React.Component));
exports.ProgressBar = ProgressBar;


/***/ }),
/* 14 */
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
var components_1 = __webpack_require__(1);
var LabResult = (function (_super) {
    __extends(LabResult, _super);
    function LabResult() {
        return _super !== null && _super.apply(this, arguments) || this;
    }
    LabResult.prototype.render = function () {
        return (React.createElement(components_1.Row, null,
            React.createElement("div", { className: "col-lg-12" },
                React.createElement("h1", null, this.props.course_name),
                React.createElement("p", { className: "lead" },
                    "Your progress on ",
                    React.createElement("strong", null,
                        React.createElement("span", { id: "lab-headline" }, this.props.lab))),
                React.createElement(components_1.ProgressBar, { progress: this.props.progress })),
            React.createElement("div", { className: "col-lg-6" },
                React.createElement("p", null,
                    React.createElement("strong", { id: "status" }, "Status: Nothing built yet."))),
            React.createElement("div", { className: "col-lg-6" },
                React.createElement("p", null,
                    React.createElement("strong", { id: "pushtime" }, "Code delievered: - ")))));
    };
    return LabResult;
}(React.Component));
exports.LabResult = LabResult;


/***/ }),
/* 15 */
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
var components_1 = __webpack_require__(1);
var LastBuild = (function (_super) {
    __extends(LastBuild, _super);
    function LastBuild() {
        return _super !== null && _super.apply(this, arguments) || this;
    }
    LastBuild.prototype.render = function () {
        return (React.createElement(components_1.Row, null,
            React.createElement("div", { className: "col-lg-12" },
                React.createElement(components_1.DynamicTable, { header: ["Test name", "Score", "Weight"], data: this.props.test_cases, selector: function (item) { return [item.name, item.score.toString() + "/" + item.points.toString() + " pts", item.weight.toString() + " pts"]; }, footer: ["Total score", this.props.score.toString() + "%", this.props.weight.toString() + "%"] }))));
    };
    return LastBuild;
}(React.Component));
exports.LastBuild = LastBuild;


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
var components_1 = __webpack_require__(1);
var LastBuildInfo = (function (_super) {
    __extends(LastBuildInfo, _super);
    function LastBuildInfo() {
        return _super !== null && _super.apply(this, arguments) || this;
    }
    LastBuildInfo.prototype.handleClick = function () {
        console.log("Rebuilding...");
    };
    LastBuildInfo.prototype.render = function () {
        var _this = this;
        return (React.createElement(components_1.Row, null,
            React.createElement("div", { className: "col-lg-8" },
                React.createElement("h2", null, "Latest build"),
                React.createElement("p", { id: "passes" },
                    "Number of passed tests:  ",
                    this.props.pass_tests),
                React.createElement("p", { id: "fails" },
                    "Number of failed tests:  ",
                    this.props.fail_tests),
                React.createElement("p", { id: "buildtime" },
                    "Execution time:  ",
                    this.props.exec_time),
                React.createElement("p", { id: "timedate" },
                    "Build date:  ",
                    this.props.build_time.toString()),
                React.createElement("p", { id: "buildid" },
                    "Build ID: ",
                    this.props.build_id)),
            React.createElement("div", { className: "col-lg-4 hidden-print" },
                React.createElement("h2", null, "Actions"),
                React.createElement(components_1.Row, null,
                    React.createElement("div", { className: "col-lg-12" },
                        React.createElement("p", null,
                            React.createElement("button", { type: "button", id: "rebuild", className: "btn btn-primary", onClick: function () { return _this.handleClick(); } }, "Rebuild")))))));
    };
    return LastBuildInfo;
}(React.Component));
exports.LastBuildInfo = LastBuildInfo;


/***/ }),
/* 17 */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
var event_1 = __webpack_require__(18);
var ViewPage_1 = __webpack_require__(2);
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
        if (path === "/") {
            this.navigateToDefault();
            return;
        }
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
/* 18 */
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
/* 19 */
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
var UserView_1 = __webpack_require__(4);
var HelloView_1 = __webpack_require__(5);
var components_1 = __webpack_require__(1);
var ViewPage_1 = __webpack_require__(2);
var LabResultView_1 = __webpack_require__(21);
var StudentPage = (function (_super) {
    __extends(StudentPage, _super);
    function StudentPage(users, navMan, courseMan) {
        var _this = _super.call(this) || this;
        _this.pages = {};
        _this.navMan = navMan;
        _this.userMan = users;
        _this.courseMan = courseMan;
        _this.defaultPage = "opsys/lab1";
        _this.pages["opsys/lab1"] = React.createElement("h1", null, "Lab1");
        _this.pages["opsys/lab2"] = React.createElement("h1", null, "Lab2");
        _this.pages["opsys/lab3"] = React.createElement("h1", null, "Lab3");
        _this.pages["opsys/lab4"] = React.createElement("h1", null, "Lab4");
        _this.pages["user"] = React.createElement(UserView_1.UserView, { users: users.getAllUser() });
        _this.pages["hello"] = React.createElement(HelloView_1.HelloView, null);
        return _this;
    }
    StudentPage.prototype.getLabs = function () {
        var curUsr = this.userMan.getCurrentUser();
        if (curUsr) {
            var courses = this.courseMan.getCoursesFor(curUsr);
            var labs = this.courseMan.getAssignments(courses[0]);
            return { course: courses[0], labs: labs };
        }
        return null;
    };
    StudentPage.prototype.renderMenu = function (key) {
        var _this = this;
        if (key === 0) {
            var labs = this.getLabs();
            var labLinks = [];
            if (labs) {
                for (var _i = 0, _a = labs.labs; _i < _a.length; _i++) {
                    var l = _a[_i];
                    labLinks.push({ name: l.name, uri: this.pagePath + "/course/" + labs.course.tag + "/lab/" + l.id });
                }
            }
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
    StudentPage.prototype.renderContent = function (page) {
        if (page.length === 0) {
            page = this.defaultPage;
        }
        if (this.pages[page]) {
            return this.pages[page];
        }
        var parts = this.navMan.getParts(page);
        if (parts.length > 1) {
            if (parts[0] === "course") {
                var course_tag = parts[1];
                var course = this.courseMan.getCourseByTag(course_tag);
                if (parts.length > 3) {
                    var labId = parseInt(parts[3]);
                    if (course !== null && labId !== undefined) {
                        var lab = this.courseMan.getAssignment({ id: 0, name: "", tag: "" }, labId);
                        console.log(lab);
                        if (lab) {
                            var testCases = [
                                { name: "Test Case 1", score: 60, points: 100, weight: 1 },
                                { name: "Test Case 2", score: 50, points: 100, weight: 1 },
                                { name: "Test Case 3", score: 40, points: 100, weight: 1 },
                                { name: "Test Case 4", score: 30, points: 100, weight: 1 },
                                { name: "Test Case 5", score: 20, points: 100, weight: 1 }
                            ];
                            var labInfo = {
                                lab: lab.name,
                                course: course.name,
                                score: 50,
                                weight: 100,
                                test_cases: testCases,
                                pass_tests: 10,
                                fail_tests: 20,
                                exec_time: 0.33,
                                build_time: new Date(2017, 5, 25),
                                build_id: 10
                            };
                            return React.createElement(LabResultView_1.LabResultView, { labInfo: labInfo });
                        }
                        return React.createElement("h1", null, "Could not find that lab");
                    }
                }
            }
        }
        return React.createElement("div", null);
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
var components_1 = __webpack_require__(1);
var LabResultView = (function (_super) {
    __extends(LabResultView, _super);
    function LabResultView() {
        return _super !== null && _super.apply(this, arguments) || this;
    }
    LabResultView.prototype.render = function () {
        return (React.createElement("div", { className: "col-md-9 col-sm-9 col-xs-12" },
            React.createElement("div", { className: "result-content", id: "resultview" },
                React.createElement("section", { id: "result" },
                    React.createElement(components_1.LabResult, { course_name: this.props.labInfo.course, lab: this.props.labInfo.lab, progress: this.props.labInfo.score }),
                    React.createElement(components_1.LastBuild, { test_cases: this.props.labInfo.test_cases, score: this.props.labInfo.score, weight: this.props.labInfo.weight }),
                    React.createElement(components_1.LastBuildInfo, { pass_tests: this.props.labInfo.pass_tests, fail_tests: this.props.labInfo.fail_tests, exec_time: this.props.labInfo.exec_time, build_time: this.props.labInfo.build_time, build_id: this.props.labInfo.build_id }),
                    React.createElement(components_1.Row, null,
                        React.createElement("div", { className: "col-lg-12" },
                            React.createElement("div", { className: "well" },
                                React.createElement("code", { id: "logs" }, "# There is no build for this lab yet."))))))));
    };
    return LabResultView;
}(React.Component));
exports.LabResultView = LabResultView;


/***/ }),
/* 22 */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
var TempDataProvider = (function () {
    function TempDataProvider() {
        this.addLocalAssignments();
        this.addLocalCourses();
        this.addLocalCourseStudent();
        this.addLocalUsers();
    }
    TempDataProvider.prototype.addLocalUsers = function () {
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
    };
    TempDataProvider.prototype.addLocalAssignments = function () {
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
                courceId: 0,
                name: "Lab 2",
                start: new Date(2017, 5, 1),
                deadline: new Date(2017, 5, 25),
                end: new Date(2017, 5, 30)
            },
            {
                id: 2,
                courceId: 0,
                name: "Lab 3",
                start: new Date(2017, 5, 1),
                deadline: new Date(2017, 5, 25),
                end: new Date(2017, 5, 30)
            },
            {
                id: 3,
                courceId: 0,
                name: "Lab 4",
                start: new Date(2017, 5, 1),
                deadline: new Date(2017, 5, 25),
                end: new Date(2017, 5, 30)
            },
            {
                id: 4,
                courceId: 1,
                name: "Lab 1",
                start: new Date(2017, 5, 1),
                deadline: new Date(2017, 5, 25),
                end: new Date(2017, 5, 30)
            }
        ];
    };
    TempDataProvider.prototype.addLocalCourses = function () {
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
    };
    TempDataProvider.prototype.addLocalCourseStudent = function () {
        this.localCourseStudent = [
            { courseId: 0, personId: 999 }
        ];
    };
    TempDataProvider.prototype.getAllUser = function () {
        return this.localUsers;
    };
    TempDataProvider.prototype.getCourses = function () {
        return this.localCourses;
    };
    TempDataProvider.prototype.getCoursesStudent = function () {
        return this.localCourseStudent;
    };
    TempDataProvider.prototype.getCourseByTag = function (tag) {
        for (var _i = 0, _a = this.localCourses; _i < _a.length; _i++) {
            var c = _a[_i];
            if (c.tag === tag) {
                return c;
            }
        }
        return null;
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
/* 23 */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
var models_1 = __webpack_require__(24);
var CourseManager = (function () {
    function CourseManager(courseProvider) {
        this.courseProvider = courseProvider;
    }
    CourseManager.prototype.getCourses = function () {
        return this.courseProvider.getCourses();
    };
    CourseManager.prototype.getCourseByTag = function (tag) {
        return this.courseProvider.getCourseByTag(tag);
    };
    CourseManager.prototype.getCoursesFor = function (user) {
        var cLinks = [];
        for (var _i = 0, _a = this.courseProvider.getCoursesStudent(); _i < _a.length; _i++) {
            var c = _a[_i];
            if (user.id === c.personId) {
                cLinks.push(c);
            }
        }
        var courses = [];
        for (var _b = 0, _c = this.getCourses(); _b < _c.length; _b++) {
            var c = _c[_b];
            for (var _d = 0, cLinks_1 = cLinks; _d < cLinks_1.length; _d++) {
                var link = cLinks_1[_d];
                if (c.id === link.courseId) {
                    courses.push(c);
                    break;
                }
            }
        }
        return courses;
    };
    CourseManager.prototype.getAssignment = function (course, assignmentId) {
        var temp = this.getAssignments(course);
        console.log(temp);
        for (var _i = 0, temp_1 = temp; _i < temp_1.length; _i++) {
            var a = temp_1[_i];
            if (a.id === assignmentId) {
                return a;
            }
        }
        return null;
    };
    CourseManager.prototype.getAssignments = function (courseId) {
        if (models_1.isCourse(courseId)) {
            courseId = courseId.id;
        }
        return this.courseProvider.getAssignments(courseId);
    };
    return CourseManager;
}());
exports.CourseManager = CourseManager;


/***/ }),
/* 24 */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
function isCourse(value) {
    return value
        && typeof value.id === "number"
        && typeof value.name === "string"
        && typeof value.tag === "string";
}
exports.isCourse = isCourse;


/***/ }),
/* 25 */
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
var ViewPage_1 = __webpack_require__(2);
var HomePage = (function (_super) {
    __extends(HomePage, _super);
    function HomePage() {
        var _this = _super.call(this) || this;
        _this.defaultPage = "index";
        return _this;
    }
    HomePage.prototype.renderContent = function (page) {
        return React.createElement("h1", null, "Welcome to autograder");
    };
    return HomePage;
}(ViewPage_1.ViewPage));
exports.HomePage = HomePage;


/***/ }),
/* 26 */
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
var ViewPage_1 = __webpack_require__(2);
var ErrorPage = (function (_super) {
    __extends(ErrorPage, _super);
    function ErrorPage() {
        var _this = _super.call(this) || this;
        _this.defaultPage = "404";
        return _this;
    }
    ErrorPage.prototype.renderContent = function (page) {
        if (page.length === 0) {
            page = this.defaultPage;
        }
        return React.createElement("div", null,
            React.createElement("h1", null, "404 Page not found"),
            React.createElement("p", null, "The page you where looking for does not exist"));
    };
    return ErrorPage;
}(ViewPage_1.ViewPage));
exports.ErrorPage = ErrorPage;


/***/ }),
/* 27 */
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
var components_1 = __webpack_require__(1);
var ViewPage_1 = __webpack_require__(2);
var TeacherPage = (function (_super) {
    __extends(TeacherPage, _super);
    function TeacherPage(users, navMan) {
        var _this = _super.call(this) || this;
        _this.pages = {};
        _this.navMan = navMan;
        _this.defaultPage = "opsys/lab1";
        _this.pages["opsys/lab1"] = React.createElement("h1", null, "Teacher Lab1");
        _this.pages["opsys/lab2"] = React.createElement("h1", null, "Teacher Lab2");
        _this.pages["opsys/lab3"] = React.createElement("h1", null, "Teacher Lab3");
        _this.pages["opsys/lab4"] = React.createElement("h1", null, "Teacher Lab4");
        _this.pages["user"] = React.createElement(UserView_1.UserView, { users: users.getAllUser() });
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
    TeacherPage.prototype.renderContent = function (page) {
        if (page.length === 0) {
            page = this.defaultPage;
        }
        if (this.pages[page]) {
            return this.pages[page];
        }
        return React.createElement("h1", null, "404 page not found");
    };
    return TeacherPage;
}(ViewPage_1.ViewPage));
exports.TeacherPage = TeacherPage;


/***/ }),
/* 28 */
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
var ViewPage_1 = __webpack_require__(2);
var HelpView_1 = __webpack_require__(29);
var HelpPage = (function (_super) {
    __extends(HelpPage, _super);
    function HelpPage(navMan) {
        var _this = _super.call(this) || this;
        _this.pages = {};
        _this.navMan = navMan;
        _this.defaultPage = "help";
        _this.pages["help"] = React.createElement(HelpView_1.HelpView, null);
        return _this;
    }
    HelpPage.prototype.renderContent = function (page) {
        if (page.length === 0) {
            page = this.defaultPage;
        }
        if (this.pages[page]) {
            return this.pages[page];
        }
        return React.createElement("h1", null, "404 page not found");
    };
    return HelpPage;
}(ViewPage_1.ViewPage));
exports.HelpPage = HelpPage;


/***/ }),
/* 29 */
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
var components_1 = __webpack_require__(1);
var HelpView = (function (_super) {
    __extends(HelpView, _super);
    function HelpView() {
        return _super !== null && _super.apply(this, arguments) || this;
    }
    HelpView.prototype.render = function () {
        return (React.createElement(components_1.Row, { className: "container-fluid" },
            React.createElement("div", { className: "col-md-2 col-sm-3 col-xs-12" },
                React.createElement("div", { className: "list-group" },
                    React.createElement("a", { href: "#", className: "list-group-item disabled" }, "Help"),
                    React.createElement("a", { href: "#autograder", className: "list-group-item" }, "Autograder"),
                    React.createElement("a", { href: "#reg", className: "list-group-item" }, "Registration"),
                    React.createElement("a", { href: "#signup", className: "list-group-item" }, "Sign up for a course"))),
            React.createElement("div", { className: "col-md-8 col-sm-9 col-xs-12" },
                React.createElement("article", null,
                    React.createElement("h1", { id: "autograder" }, "Autograder"),
                    React.createElement("p", null, "Autograder is a new tool for students and teaching staff for submitting and validating lab assignments and is developed at the University of Stavanger. All lab submissions from students are handled using Git, a source code management system, and GitHub, a web-based hosting service for Git source repositories."),
                    React.createElement("p", null, "Students push their updated lab submissions to GitHub. Every lab submission is then processed by a custom continuous integration tool. This tool will run several test cases on the submitted code. Autograder generates feedback that let the students verify if their submission implements the required functionality. This feedback is available through a web interface. The feedback from the Autograder system can be used by students to improve their submissions."),
                    React.createElement("p", null, "Below is a step-by-step explanation of how to register and sign up for the lab project in Autograder."),
                    React.createElement("h1", { id: "reg" }, "Registration"),
                    React.createElement("ol", null,
                        React.createElement("li", null,
                            React.createElement("p", null,
                                "Go to ",
                                React.createElement("a", { href: "http://github.com" }, "GitHub"),
                                " and register. A GitHub account is required to sign in to Autograder. You can skip this step if you already have an account.")),
                        React.createElement("li", null,
                            React.createElement("p", null, "Click the \"Sign in with GitHub\" button to register. You will then be taken to GitHub's website.")),
                        React.createElement("li", null,
                            React.createElement("p", null, "Approve that our Autograder application may have permission to access to the requested parts of your account. It is possible to make a separate GitHub account for system if you do not want Autograder to access your personal one with the requested permissions."))),
                    React.createElement("h1", { id: "signup" }, "Signing up for a course"),
                    React.createElement("ol", null,
                        React.createElement("li", null,
                            React.createElement("p", null, "Click the course menu item.")),
                        React.createElement("li", null,
                            React.createElement("p", null, "In the course menu click on \u201CNew Course\u201D. Available courses will be listed.")),
                        React.createElement("li", null,
                            React.createElement("p", null, "Find the course you are signing up for and click sign up.")),
                        React.createElement("li", null,
                            React.createElement("p", null, "Read through and accept the terms. You will then be invited to the course organization on GitHub.")),
                        React.createElement("li", null,
                            React.createElement("p", null, "An invitation will be sent to your email address registered with GitHub account. Accept the invitation using the received email.")),
                        React.createElement("li", null,
                            React.createElement("p", null, "Wait for the teaching staff to verify your Autograder-registration.")),
                        React.createElement("li", null,
                            React.createElement("p", null, "You will get your own repository in the organization \"uis-dat520\" on GitHub after your registration is verified. You will also have access to the feedback pages for this course on Autograder.")))))));
    };
    return HelpView;
}(React.Component));
exports.HelpView = HelpView;


/***/ })
/******/ ]);
//# sourceMappingURL=bundle.js.map
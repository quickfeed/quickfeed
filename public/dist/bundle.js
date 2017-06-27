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
/******/ 	return __webpack_require__(__webpack_require__.s = 9);
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
__export(__webpack_require__(11));
__export(__webpack_require__(3));
__export(__webpack_require__(12));
__export(__webpack_require__(13));
__export(__webpack_require__(14));
__export(__webpack_require__(15));
__export(__webpack_require__(16));
__export(__webpack_require__(18));
__export(__webpack_require__(19));
__export(__webpack_require__(20));
__export(__webpack_require__(21));
__export(__webpack_require__(22));
__export(__webpack_require__(35));
__export(__webpack_require__(37));


/***/ }),
/* 2 */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
var NavigationHelper_1 = __webpack_require__(6);
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
        this.navHelper = new NavigationHelper_1.NavigationHelper(this);
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
            React.createElement("a", { className: "navbar-brand", onClick: function (e) { e.preventDefault(); _this.props.brandClick(); }, href: ";/" }, this.props.brandName));
    };
    return NavHeaderBar;
}(React.Component));
exports.NavHeaderBar = NavHeaderBar;


/***/ }),
/* 4 */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
var ArrayHelper = (function () {
    function ArrayHelper() {
    }
    ArrayHelper.find = function (array, predicate) {
        for (var i = 0; i < array.length; i++) {
            var cur = array[i];
            if (predicate.call(array, cur, i, array)) {
                return cur;
            }
        }
        return null;
    };
    return ArrayHelper;
}());
exports.ArrayHelper = ArrayHelper;


/***/ }),
/* 5 */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
function newEvent(info) {
    var callbacks = [];
    var handler = function EventHandler(event) {
        callbacks.map((function (v) { return v(event); }));
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
/* 6 */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
var event_1 = __webpack_require__(5);
var NavigationHelper = (function () {
    function NavigationHelper(thisObject) {
        this.onPreNavigation = event_1.newEvent("NavigationHelper.onPreNavigation");
        this.DEFAULT_VALUE = "";
        this.navObj = "__navObj";
        this.path = {};
        this.thisObject = thisObject;
    }
    NavigationHelper.getParts = function (path) {
        return this.removeEmptyEntries(path.split("/"));
    };
    NavigationHelper.removeEmptyEntries = function (array) {
        var newArray = [];
        array.map(function (v) {
            if (v.length > 0) {
                newArray.push(v);
            }
        });
        return newArray;
    };
    NavigationHelper.getOptionalField = function (field) {
        var tField = field.trim();
        if (tField.length > 2 && tField.charAt(0) === "{" && tField.charAt(tField.length - 1) === "}") {
            return tField.substr(1, tField.length - 2);
        }
        return null;
    };
    NavigationHelper.isINavObject = function (obj) {
        return obj && obj.path;
    };
    Object.defineProperty(NavigationHelper.prototype, "defaultPage", {
        get: function () {
            return this.DEFAULT_VALUE;
        },
        set: function (value) {
            this.DEFAULT_VALUE = value;
        },
        enumerable: true,
        configurable: true
    });
    NavigationHelper.prototype.registerFunction = function (path, callback) {
        var pathParts = NavigationHelper.getParts(path);
        if (pathParts.length === 0) {
            throw new Error("Can't register function on empty path");
        }
        var curObj = this.createNavPath(pathParts);
        var temp = {
            path: pathParts,
            func: callback,
        };
        curObj[this.navObj] = temp;
    };
    NavigationHelper.prototype.navigateTo = function (path) {
        if (path.length === 0) {
            path = this.DEFAULT_VALUE;
        }
        var pathParts = NavigationHelper.getParts(path);
        if (pathParts.length === 0) {
            throw new Error("Can't navigate to an empty path");
        }
        var curObj = this.getNavPath(pathParts);
        if (!curObj || !curObj[this.navObj]) {
            return null;
        }
        var navObj = curObj[this.navObj];
        var navInfo = {
            matchPath: navObj.path,
            realPath: pathParts,
            params: this.createParamsObj(navObj.path, pathParts),
        };
        this.onPreNavigation({ target: this, navInfo: navInfo });
        return navObj.func.call(this.thisObject, navInfo);
    };
    NavigationHelper.prototype.createParamsObj = function (matchPath, realPath) {
        if (matchPath.length !== realPath.length) {
            throw new Error("trying to match different paths");
        }
        var returnObj = {};
        for (var i = 0; i < matchPath.length; i++) {
            var param = NavigationHelper.getOptionalField(matchPath[i]);
            if (param) {
                returnObj[param] = realPath[i];
            }
        }
        return returnObj;
    };
    NavigationHelper.prototype.getNavPath = function (pathParts) {
        var curObj = this.path;
        for (var _i = 0, pathParts_1 = pathParts; _i < pathParts_1.length; _i++) {
            var part = pathParts_1[_i];
            var curIndex = part;
            if (!curObj[curIndex]) {
                curIndex = "*";
            }
            var curWrap = curObj[curIndex];
            if (NavigationHelper.isINavObject(curWrap) || curIndex === this.navObj) {
                throw new Error("Can't navigate to: " + curIndex);
            }
            if (!curWrap) {
                return null;
            }
            curObj = curWrap;
        }
        return curObj;
    };
    NavigationHelper.prototype.createNavPath = function (pathParts) {
        var curObj = this.path;
        for (var _i = 0, pathParts_2 = pathParts; _i < pathParts_2.length; _i++) {
            var part = pathParts_2[_i];
            var curIndex = part;
            var optional = NavigationHelper.getOptionalField(curIndex);
            if (optional) {
                curIndex = "*";
            }
            var curWrap = curObj[curIndex];
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
    };
    return NavigationHelper;
}());
exports.NavigationHelper = NavigationHelper;


/***/ }),
/* 7 */
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
var components_1 = __webpack_require__(1);
var UserView = (function (_super) {
    __extends(UserView, _super);
    function UserView() {
        return _super !== null && _super.apply(this, arguments) || this;
    }
    UserView.prototype.render = function () {
        return React.createElement(components_1.DynamicTable, { header: ["ID", "F irst name", "Last name", "Email", "StudentID"], data: this.props.users, selector: function (item) { return [
                item.id.toString(),
                item.firstName,
                item.lastName,
                item.email,
                item.personId.toString(),
            ]; } });
    };
    return UserView;
}(React.Component));
exports.UserView = UserView;


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
var ReactDOM = __webpack_require__(10);
var components_1 = __webpack_require__(1);
var managers_1 = __webpack_require__(23);
var ErrorPage_1 = __webpack_require__(29);
var HelpPage_1 = __webpack_require__(30);
var HomePage_1 = __webpack_require__(32);
var StudentPage_1 = __webpack_require__(33);
var TeacherPage_1 = __webpack_require__(34);
var topLinks = [
    { name: "Teacher", uri: "app/teacher/", active: false },
    { name: "Student", uri: "app/student/", active: false },
    { name: "Admin", uri: "app/admin", active: false },
    { name: "Help", uri: "app/help", active: false },
];
var AutoGrader = (function (_super) {
    __extends(AutoGrader, _super);
    function AutoGrader(props) {
        var _this = _super.call(this) || this;
        _this.userManager = props.userManager;
        _this.navMan = props.navigationManager;
        _this.state = {
            activePage: undefined,
            topLink: topLinks,
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
    AutoGrader.prototype.render = function () {
        if (this.state.activePage) {
            return this.renderTemplate(this.state.activePage.template);
        }
        else {
            return React.createElement("h1", null, "404 not found");
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
        var content = this.renderActivePage(this.subPage);
        switch (name) {
            case "frontpage":
                body = (React.createElement(components_1.Row, { className: "container-fluid" },
                    React.createElement("div", { className: "col-xs-12" }, content)));
            default:
                body = (React.createElement(components_1.Row, { className: "container-fluid" },
                    React.createElement("div", { className: "col-md-2 col-sm-3 col-xs-12" }, this.renderActiveMenu(0)),
                    React.createElement("div", { className: "col-md-10 col-sm-9 col-xs-12" }, content)));
        }
        return (React.createElement("div", null,
            React.createElement(components_1.NavBar, { id: "top-bar", isFluid: false, isInverse: true, links: topLinks, onClick: function (link) { return _this.handleClick(link); }, user: this.userManager.getCurrentUser(), brandName: "Auto Grader" }),
            body));
    };
    return AutoGrader;
}(React.Component));
function main() {
    var tempData = new managers_1.TempDataProvider();
    var userMan = new managers_1.UserManager(tempData);
    var courseMan = new managers_1.CourseManager(tempData);
    var navMan = new managers_1.NavigationManager(history);
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
/* 10 */
/***/ (function(module, exports) {

module.exports = ReactDOM;

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
var NavHeaderBar_1 = __webpack_require__(3);
var NavBar = (function (_super) {
    __extends(NavBar, _super);
    function NavBar() {
        return _super !== null && _super.apply(this, arguments) || this;
    }
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
    return NavBar;
}(React.Component));
exports.NavBar = NavBar;


/***/ }),
/* 12 */
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
                React.createElement("a", { href: "/" + v.uri, onClick: function (e) { return _this.handleClick(e, v); } }, v.name));
        });
        return React.createElement("ul", { className: "nav nav-pills nav-stacked" }, items);
    };
    NavMenu.prototype.handleClick = function (e, v) {
        e.preventDefault();
        if (this.props.onClick) {
            this.props.onClick(v);
        }
    };
    return NavMenu;
}(React.Component));
exports.NavMenu = NavMenu;


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
var NavMenuFormatable = (function (_super) {
    __extends(NavMenuFormatable, _super);
    function NavMenuFormatable() {
        return _super !== null && _super.apply(this, arguments) || this;
    }
    NavMenuFormatable.prototype.render = function () {
        var _this = this;
        var items = this.props.items.map(function (v, i) {
            return React.createElement("li", { key: i },
                React.createElement("a", { href: "#", onClick: function () { _this.handleItemClick(v); } }, _this.renderObj(v)));
        });
        return React.createElement("ul", { className: "nav nav-pills nav-stacked" }, items);
    };
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
    return NavMenuFormatable;
}(React.Component));
exports.NavMenuFormatable = NavMenuFormatable;


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
var DynamicTable = (function (_super) {
    __extends(DynamicTable, _super);
    function DynamicTable() {
        return _super !== null && _super.apply(this, arguments) || this;
    }
    DynamicTable.prototype.render = function () {
        var _this = this;
        var rows = this.props.data.map(function (v, i) {
            return _this.renderRow(v, i);
        });
        if (this.props.footer) {
            return this.tableWithFooter(rows, this.props.footer);
        }
        return this.tableWithNoFooter(rows);
    };
    DynamicTable.prototype.renderCells = function (values, th) {
        if (th === void 0) { th = false; }
        return values.map(function (v, i) {
            if (th) {
                return React.createElement("th", { key: i }, v);
            }
            return React.createElement("td", { key: i }, v);
        });
    };
    DynamicTable.prototype.renderRow = function (item, i) {
        var _this = this;
        return (React.createElement("tr", { key: i, onClick: function (e) { return _this.handleRowClick(e, item); } }, this.renderCells(this.props.selector(item))));
    };
    DynamicTable.prototype.tableWithFooter = function (rows, footer) {
        return (React.createElement("table", { className: this.props.onRowClick ? "table table-hover" : "table" },
            React.createElement("thead", null,
                React.createElement("tr", null, this.renderCells(this.props.header, true))),
            React.createElement("tbody", null, rows),
            React.createElement("tfoot", null,
                React.createElement("tr", null, this.renderCells(footer)))));
    };
    DynamicTable.prototype.tableWithNoFooter = function (rows) {
        return (React.createElement("table", { className: this.props.onRowClick ? "table table-hover" : "table" },
            React.createElement("thead", null,
                React.createElement("tr", null, this.renderCells(this.props.header, true))),
            React.createElement("tbody", null, rows)));
    };
    DynamicTable.prototype.handleRowClick = function (e, item) {
        e.preventDefault();
        if (this.props.onRowClick && this.props.row_links && this.props.link_key_identifier) {
            var identifier = this.props.link_key_identifier;
            this.props.onRowClick(this.props.row_links[item[identifier]]);
        }
    };
    return DynamicTable;
}(React.Component));
exports.DynamicTable = DynamicTable;


/***/ }),
/* 15 */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
var React = __webpack_require__(0);
function Row(props) {
    return React.createElement("div", { className: props.className ? "row " + props.className : "row" }, props.children);
}
exports.Row = Row;


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
var LabResultView_1 = __webpack_require__(17);
var StudentLab = (function (_super) {
    __extends(StudentLab, _super);
    function StudentLab() {
        return _super !== null && _super.apply(this, arguments) || this;
    }
    StudentLab.prototype.render = function () {
        var testCases = [
            { name: "Test Case 1", score: 60, points: 100, weight: 1 },
            { name: "Test Case 2", score: 50, points: 100, weight: 1 },
            { name: "Test Case 3", score: 40, points: 100, weight: 1 },
            { name: "Test Case 4", score: 30, points: 100, weight: 1 },
            { name: "Test Case 5", score: 20, points: 100, weight: 1 }
        ];
        var labInfo = {
            lab: this.props.assignment.name,
            course: this.props.course.name,
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
    };
    return StudentLab;
}(React.Component));
exports.StudentLab = StudentLab;


/***/ }),
/* 17 */
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
/* 18 */
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
var NavDropdown = (function (_super) {
    __extends(NavDropdown, _super);
    function NavDropdown() {
        var _this = _super.call(this) || this;
        _this.state = {
            isOpen: false,
        };
        return _this;
    }
    NavDropdown.prototype.render = function () {
        var _this = this;
        var children = this.props.items.map(function (item, index) {
            return React.createElement("li", { key: index },
                React.createElement("a", { href: "/" + item.uri, onClick: function (e) {
                        e.preventDefault();
                        _this.toggleOpen();
                        _this.props.itemClick(item, index);
                    } }, item.name));
        });
        return React.createElement("div", { className: this.getButtonClass() },
            React.createElement("button", { className: "btn btn-default dropdown-toggle", type: "button", onClick: function () { return _this.toggleOpen(); } },
                this.renderActive(),
                React.createElement("span", { className: "caret" })),
            React.createElement("ul", { className: "dropdown-menu" }, children));
    };
    NavDropdown.prototype.getButtonClass = function () {
        if (this.state.isOpen) {
            return "button open";
        }
        else {
            return "button";
        }
    };
    NavDropdown.prototype.toggleOpen = function () {
        var newState = !this.state.isOpen;
        this.setState({ isOpen: newState });
    };
    NavDropdown.prototype.renderActive = function () {
        if (this.props.items.length === 0) {
            return "";
        }
        var curIndex = this.props.selectedIndex;
        if (curIndex >= this.props.items.length || curIndex < 0) {
            curIndex = 0;
        }
        return this.props.items[curIndex].name;
    };
    return NavDropdown;
}(React.Component));
exports.NavDropdown = NavDropdown;


/***/ }),
/* 19 */
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
/* 23 */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

function __export(m) {
    for (var p in m) if (!exports.hasOwnProperty(p)) exports[p] = m[p];
}
Object.defineProperty(exports, "__esModule", { value: true });
__export(__webpack_require__(24));
__export(__webpack_require__(26));
__export(__webpack_require__(27));
__export(__webpack_require__(28));


/***/ }),
/* 24 */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
var helper_1 = __webpack_require__(4);
var models_1 = __webpack_require__(25);
var CourseManager = (function () {
    function CourseManager(courseProvider) {
        this.courseProvider = courseProvider;
    }
    CourseManager.prototype.getCourse = function (id) {
        return helper_1.ArrayHelper.find(this.getCourses(), function (a) { return a.id === id; });
    };
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
/* 25 */
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
/* 26 */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
var event_1 = __webpack_require__(5);
var NavigationHelper_1 = __webpack_require__(6);
var ViewPage_1 = __webpack_require__(2);
var NavigationManager = (function () {
    function NavigationManager(history) {
        var _this = this;
        this.onNavigate = event_1.newEvent("NavigationManager.onNavigate");
        this.pages = {};
        this.errorPages = [];
        this.defaultPath = "";
        this.currentPath = "";
        this.browserHistory = history;
        window.addEventListener("popstate", function (e) {
            _this.navigateTo(location.pathname, true);
        });
    }
    NavigationManager.prototype.setDefaultPath = function (path) {
        this.defaultPath = path;
    };
    NavigationManager.prototype.navigateTo = function (path, preventPush) {
        if (path === "/") {
            this.navigateToDefault();
            return;
        }
        var parts = NavigationHelper_1.NavigationHelper.getParts(path);
        var curPage = this.pages;
        this.currentPath = parts.join("/");
        if (!preventPush) {
            this.browserHistory.pushState({}, "Autograder", "/" + this.currentPath);
        }
        for (var i = 0; i < parts.length; i++) {
            var a = parts[i];
            if (ViewPage_1.isViewPage(curPage)) {
                this.onNavigate({
                    page: curPage,
                    subPage: parts.slice(i, parts.length).join("/"),
                    target: this,
                    uri: path,
                });
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
        var parts = NavigationHelper_1.NavigationHelper.getParts(path);
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
            checkUrl += "/" + viewPage.navHelper.defaultPage;
        }
        for (var _i = 0, links_1 = links; _i < links_1.length; _i++) {
            var l = links_1[_i];
            if (!l.uri) {
                continue;
            }
            var a = NavigationHelper_1.NavigationHelper.getParts(l.uri).join("/");
            l.active = a === checkUrl.substr(0, a.length);
        }
    };
    NavigationManager.prototype.refresh = function () {
        this.navigateTo(this.currentPath);
    };
    NavigationManager.prototype.getErrorPage = function (statusCode) {
        if (this.errorPages[statusCode]) {
            return this.errorPages[statusCode];
        }
        throw Error("Status page: " + statusCode + " is not defined");
    };
    return NavigationManager;
}());
exports.NavigationManager = NavigationManager;


/***/ }),
/* 27 */
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
    TempDataProvider.prototype.getAllUser = function () {
        return this.localUsers;
    };
    TempDataProvider.prototype.getCourses = function () {
        return this.localCourses;
    };
    TempDataProvider.prototype.getCoursesStudent = function () {
        return this.localCourseStudent;
    };
    TempDataProvider.prototype.getAssignments = function (courseId) {
        var temp = [];
        for (var _i = 0, _a = this.localAssignments; _i < _a.length; _i++) {
            var a = _a[_i];
            if (a.courseId === courseId) {
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
    TempDataProvider.prototype.addLocalUsers = function () {
        this.localUsers = [
            {
                id: 999,
                firstName: "Test",
                lastName: "Testersen",
                email: "test@testersen.no",
                personId: 9999,
                password: "1234",
            },
            {
                id: 1,
                firstName: "Per",
                lastName: "Pettersen",
                email: "per@pettersen.no",
                personId: 1234,
                password: "1234",
            },
            {
                id: 2,
                firstName: "Bob",
                lastName: "Bobsen",
                email: "bob@bobsen.no",
                personId: 1234,
                password: "1234",
            },
            {
                id: 3,
                firstName: "Petter",
                lastName: "Pan",
                email: "petter@pan.no",
                personId: 1234,
                password: "1234",
            },
        ];
    };
    TempDataProvider.prototype.addLocalAssignments = function () {
        this.localAssignments = [
            {
                id: 0,
                courseId: 0,
                name: "Lab 1",
                start: new Date(2017, 5, 1),
                deadline: new Date(2017, 5, 25),
                end: new Date(2017, 5, 30),
            },
            {
                id: 1,
                courseId: 0,
                name: "Lab 2",
                start: new Date(2017, 5, 1),
                deadline: new Date(2017, 5, 25),
                end: new Date(2017, 5, 30),
            },
            {
                id: 2,
                courseId: 0,
                name: "Lab 3",
                start: new Date(2017, 5, 1),
                deadline: new Date(2017, 5, 25),
                end: new Date(2017, 5, 30),
            },
            {
                id: 3,
                courseId: 0,
                name: "Lab 4",
                start: new Date(2017, 5, 1),
                deadline: new Date(2017, 5, 25),
                end: new Date(2017, 5, 30),
            },
            {
                id: 4,
                courseId: 1,
                name: "Lab 1",
                start: new Date(2017, 5, 1),
                deadline: new Date(2017, 5, 25),
                end: new Date(2017, 5, 30),
            },
        ];
    };
    TempDataProvider.prototype.addLocalCourses = function () {
        this.localCourses = [
            {
                id: 0,
                name: "Object Oriented Programming",
                tag: "DAT100",
            },
            {
                id: 1,
                name: "Algorithms and Datastructures",
                tag: "DAT200",
            },
        ];
    };
    TempDataProvider.prototype.addLocalCourseStudent = function () {
        this.localCourseStudent = [
            { courseId: 0, personId: 999 },
            { courseId: 1, personId: 999 },
        ];
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
    return TempDataProvider;
}());
exports.TempDataProvider = TempDataProvider;


/***/ }),
/* 28 */
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
        throw new Error("Not implemented error");
    };
    return UserManager;
}());
exports.UserManager = UserManager;


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
var ViewPage_1 = __webpack_require__(2);
var ErrorPage = (function (_super) {
    __extends(ErrorPage, _super);
    function ErrorPage() {
        var _this = _super.call(this) || this;
        _this.pages = {};
        _this.navHelper.defaultPage = "404";
        _this.navHelper.registerFunction("404", function (navInfo) {
            return React.createElement("div", null,
                React.createElement("h1", null, "404 Page not found"),
                React.createElement("p", null, "The page you where looking for does not exist"));
        });
        return _this;
    }
    ErrorPage.prototype.renderContent = function (page) {
        var content = this.navHelper.navigateTo(page);
        if (!content) {
            content = this.navHelper.navigateTo("404");
        }
        if (!content) {
            throw new Error("There is a problem with the navigation");
        }
        return content;
    };
    return ErrorPage;
}(ViewPage_1.ViewPage));
exports.ErrorPage = ErrorPage;


/***/ }),
/* 30 */
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
var HelpView_1 = __webpack_require__(31);
var HelpPage = (function (_super) {
    __extends(HelpPage, _super);
    function HelpPage(navMan) {
        var _this = _super.call(this) || this;
        _this.pages = {};
        _this.navMan = navMan;
        _this.navHelper.defaultPage = "help";
        _this.navHelper.registerFunction("help", _this.help);
        return _this;
    }
    HelpPage.prototype.help = function (info) {
        return React.createElement(HelpView_1.HelpView, null);
    };
    HelpPage.prototype.renderContent = function (page) {
        var temp = this.navHelper.navigateTo(page);
        if (temp) {
            return temp;
        }
        return React.createElement("h1", null, "404 page not found");
    };
    return HelpPage;
}(ViewPage_1.ViewPage));
exports.HelpPage = HelpPage;


/***/ }),
/* 31 */
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


/***/ }),
/* 32 */
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
        return _super.call(this) || this;
    }
    HomePage.prototype.renderContent = function (page) {
        return React.createElement("h1", null, "Welcome to autograder");
    };
    return HomePage;
}(ViewPage_1.ViewPage));
exports.HomePage = HomePage;


/***/ }),
/* 33 */
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
var ViewPage_1 = __webpack_require__(2);
var HelloView_1 = __webpack_require__(7);
var UserView_1 = __webpack_require__(8);
var helper_1 = __webpack_require__(4);
var StudentPage = (function (_super) {
    __extends(StudentPage, _super);
    function StudentPage(users, navMan, courseMan) {
        var _this = _super.call(this) || this;
        _this.selectedCourse = null;
        _this.selectedAssignment = null;
        _this.currentPage = "";
        _this.courses = [];
        _this.foundId = -1;
        _this.navMan = navMan;
        _this.userMan = users;
        _this.courseMan = courseMan;
        _this.navHelper.defaultPage = "index";
        _this.navHelper.onPreNavigation.addEventListener(function (e) { return _this.setupData(e); });
        _this.navHelper.registerFunction("index", _this.index);
        _this.navHelper.registerFunction("course/{courseid}", _this.course);
        _this.navHelper.registerFunction("course/{courseid}/lab/{labid}", _this.courseWithLab);
        _this.navHelper.registerFunction("user", function (navInfo) { return React.createElement(UserView_1.UserView, { users: users.getAllUser() }); });
        _this.navHelper.registerFunction("hello", function (INavInfo) { return React.createElement(HelloView_1.HelloView, null); });
        return _this;
    }
    StudentPage.prototype.index = function (navInfo) {
        var course_overview = this.getCoursesWithAssignments();
        return (React.createElement(components_1.CourseOverview, { course_overview: course_overview, navMan: this.navMan }));
    };
    StudentPage.prototype.course = function (navInfo) {
        this.selectCourse(navInfo.params.courseid);
        if (this.selectedCourse) {
            return React.createElement("div", null,
                "This is the CourseView for ",
                this.selectedCourse.name);
        }
        return React.createElement("div", null, "404 not found");
    };
    StudentPage.prototype.courseWithLab = function (navInfo) {
        this.selectCourse(navInfo.params.courseid);
        if (this.selectedCourse) {
            this.selectAssignment(navInfo.params.labid);
            if (this.selectedAssignment) {
                return React.createElement(components_1.StudentLab, { course: this.selectedCourse, assignment: this.selectedAssignment });
            }
        }
        return React.createElement("div", null, "404 not found");
    };
    StudentPage.prototype.renderMenu = function (key) {
        var _this = this;
        if (key === 0) {
            var coursesLinks = this.courses.map(function (e, i) {
                return { name: e.tag, uri: _this.pagePath + "/course/" + e.id };
            });
            var labs_1 = this.getLabs();
            var labLinks = [];
            if (labs_1) {
                labLinks = labs_1.labs.map(function (l, i) {
                    return { name: l.name, uri: _this.pagePath + "/course/" + labs_1.course.id + "/lab/" + l.id };
                });
            }
            var settings = [
                { name: "Users", uri: this.pagePath + "/user" },
                { name: "Hello world", uri: this.pagePath + "/hello" },
            ];
            this.navMan.checkLinks(labLinks, this);
            this.navMan.checkLinks(settings, this);
            return [
                React.createElement("h4", null, "Course"),
                React.createElement(components_1.NavDropdown, { key: 1, selectedIndex: this.foundId, items: coursesLinks, itemClick: function (link) { _this.handleClick(link); } }),
                React.createElement("h4", { key: 2 }, "Labs"),
                React.createElement(components_1.NavMenu, { key: 3, links: labLinks, onClick: function (link) { return _this.handleClick(link); } }),
                React.createElement("h4", { key: 4 }, "Settings"),
                React.createElement(components_1.NavMenu, { key: 5, links: settings, onClick: function (link) { return _this.handleClick(link); } }),
            ];
        }
        return [];
    };
    StudentPage.prototype.renderContent = function (page) {
        var pageContent = this.navHelper.navigateTo(page);
        this.currentPage = page;
        if (pageContent) {
            return pageContent;
        }
        return React.createElement("div", null, "404 Not found");
    };
    StudentPage.prototype.setupData = function (data) {
        this.courses = this.getCourses();
    };
    StudentPage.prototype.selectCourse = function (courseId) {
        var _this = this;
        this.selectedCourse = null;
        var course = parseInt(courseId, 10);
        if (!isNaN(course)) {
            this.selectedCourse = helper_1.ArrayHelper.find(this.courses, function (e, i) {
                if (e.id === course) {
                    _this.foundId = i;
                    return true;
                }
                return false;
            });
        }
    };
    StudentPage.prototype.selectAssignment = function (labIdString) {
        this.selectedAssignment = null;
        var labId = parseInt(labIdString, 10);
        if (this.selectedCourse && !isNaN(labId)) {
            var lab = this.courseMan.getAssignment(this.selectedCourse, labId);
            if (lab) {
                this.selectedAssignment = lab;
            }
        }
    };
    StudentPage.prototype.handleClick = function (link) {
        if (link.uri) {
            this.navMan.navigateTo(link.uri);
        }
    };
    StudentPage.prototype.getCourses = function () {
        var curUsr = this.userMan.getCurrentUser();
        if (curUsr) {
            return this.courseMan.getCoursesFor(curUsr);
        }
        return [];
    };
    StudentPage.prototype.getLabs = function () {
        var curUsr = this.userMan.getCurrentUser();
        if (curUsr && !this.selectedCourse) {
            this.selectedCourse = this.courseMan.getCoursesFor(curUsr)[0];
        }
        if (this.selectedCourse) {
            var labs = this.courseMan.getAssignments(this.selectedCourse);
            return { course: this.selectedCourse, labs: labs };
        }
        return null;
    };
    StudentPage.prototype.getCoursesWithAssignments = function () {
        var course_labs = [];
        if (this.courses.length === 0) {
            this.courses = this.getCourses();
        }
        if (this.courses.length > 0) {
            for (var _i = 0, _a = this.courses; _i < _a.length; _i++) {
                var course = _a[_i];
                var labs = this.courseMan.getAssignments(course);
                var cl = { course: course, labs: labs };
                course_labs.push(cl);
            }
            return course_labs;
        }
        return [];
    };
    return StudentPage;
}(ViewPage_1.ViewPage));
exports.StudentPage = StudentPage;


/***/ }),
/* 34 */
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
var ViewPage_1 = __webpack_require__(2);
var HelloView_1 = __webpack_require__(7);
var UserView_1 = __webpack_require__(8);
var TeacherPage = (function (_super) {
    __extends(TeacherPage, _super);
    function TeacherPage(users, navMan) {
        var _this = _super.call(this) || this;
        _this.pages = {};
        _this.navMan = navMan;
        _this.navHelper.defaultPage = "opsys/lab1";
        _this.navHelper.registerFunction("opsys/{lab}", _this.course);
        _this.navHelper.registerFunction("user", function (navInfo) {
            return React.createElement(UserView_1.UserView, { users: users.getAllUser() });
        });
        _this.navHelper.registerFunction("user", function (navInfo) {
            return React.createElement(HelloView_1.HelloView, null);
        });
        return _this;
    }
    TeacherPage.prototype.course = function (info) {
        return React.createElement("h1", null,
            "Teacher ",
            info.params.lab);
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
                { name: "Hello world", uri: this.pagePath + "/hello" },
            ];
            this.navMan.checkLinks(labLinks, this);
            this.navMan.checkLinks(settings, this);
            return [
                React.createElement("h4", { key: 0 }, "Labs"),
                React.createElement(components_1.NavMenu, { key: 1, links: labLinks, onClick: function (link) { return _this.handleClick(link); } }),
                React.createElement("h4", { key: 4 }, "Settings"),
                React.createElement(components_1.NavMenu, { key: 3, links: settings, onClick: function (link) { return _this.handleClick(link); } }),
            ];
        }
        return [];
    };
    TeacherPage.prototype.renderContent = function (page) {
        var temp = this.navHelper.navigateTo(page);
        if (temp) {
            return temp;
        }
        return React.createElement("h1", null, "404 page not found");
    };
    TeacherPage.prototype.handleClick = function (link) {
        if (link.uri) {
            this.navMan.navigateTo(link.uri);
        }
    };
    return TeacherPage;
}(ViewPage_1.ViewPage));
exports.TeacherPage = TeacherPage;


/***/ }),
/* 35 */
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
var CourseOverview = (function (_super) {
    __extends(CourseOverview, _super);
    function CourseOverview() {
        return _super !== null && _super.apply(this, arguments) || this;
    }
    CourseOverview.prototype.render = function () {
        var _this = this;
        var courses = this.props.course_overview.map(function (val, key) {
            return React.createElement(components_1.CoursePanel, { course: val.course, labs: val.labs, navMan: _this.props.navMan });
        });
        var index = 3;
        var l = courses.length;
        for (index; index < l; index += 3) {
            console.log("index", index);
            courses.splice(index, 0, React.createElement("div", { className: "visible-lg-block visible-md-block clearfix" }));
            l += 1;
            index += 1;
        }
        return (React.createElement("div", null,
            React.createElement("h1", null, "Your Courses"),
            React.createElement(components_1.Row, null, courses)));
    };
    return CourseOverview;
}(React.Component));
exports.CourseOverview = CourseOverview;


/***/ }),
/* 36 */,
/* 37 */
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
var CoursePanel = (function (_super) {
    __extends(CoursePanel, _super);
    function CoursePanel() {
        return _super !== null && _super.apply(this, arguments) || this;
    }
    CoursePanel.prototype.render = function () {
        var _this = this;
        var pathPrefix = "app/student/course/" + this.props.course.id + "/lab/";
        var rowLinks = {};
        for (var _i = 0, _a = this.props.labs; _i < _a.length; _i++) {
            var lab = _a[_i];
            rowLinks[lab.id] = pathPrefix + lab.id;
        }
        return (React.createElement("div", { className: "col-lg-3 col-sm-6" },
            React.createElement("div", { className: "panel panel-primary" },
                React.createElement("div", { className: "panel-heading clickable", onClick: function () { return _this.handleCourseClick(); } }, this.props.course.name),
                React.createElement("div", { className: "panel-body" },
                    React.createElement(components_1.DynamicTable, { header: ["Labs", "Score", "Weight"], data: this.props.labs, selector: function (item) { return [item.name, "50%", "100%"]; }, onRowClick: function (row) { return _this.handleRowClick(row); }, row_links: rowLinks, link_key_identifier: "id" })))));
    };
    CoursePanel.prototype.handleRowClick = function (path) {
        if (path) {
            this.props.navMan.navigateTo(path);
        }
    };
    CoursePanel.prototype.handleCourseClick = function () {
        var uri = "app/student/course/" + this.props.course.id;
        this.props.navMan.navigateTo(uri);
    };
    return CoursePanel;
}(React.Component));
exports.CoursePanel = CoursePanel;


/***/ })
/******/ ]);
//# sourceMappingURL=bundle.js.map
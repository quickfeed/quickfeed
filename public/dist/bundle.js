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
/******/ 	return __webpack_require__(__webpack_require__.s = 14);
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
__export(__webpack_require__(16));
__export(__webpack_require__(9));
__export(__webpack_require__(10));
__export(__webpack_require__(17));
__export(__webpack_require__(18));
__export(__webpack_require__(19));
__export(__webpack_require__(20));
__export(__webpack_require__(22));
__export(__webpack_require__(11));
__export(__webpack_require__(23));
__export(__webpack_require__(24));
__export(__webpack_require__(25));
__export(__webpack_require__(26));
__export(__webpack_require__(27));
__export(__webpack_require__(28));
__export(__webpack_require__(29));
__export(__webpack_require__(30));
__export(__webpack_require__(31));
__export(__webpack_require__(32));


/***/ }),
/* 2 */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __awaiter = (this && this.__awaiter) || function (thisArg, _arguments, P, generator) {
    return new (P || (P = Promise))(function (resolve, reject) {
        function fulfilled(value) { try { step(generator.next(value)); } catch (e) { reject(e); } }
        function rejected(value) { try { step(generator["throw"](value)); } catch (e) { reject(e); } }
        function step(result) { result.done ? resolve(result.value) : new P(function (resolve) { resolve(result.value); }).then(fulfilled, rejected); }
        step((generator = generator.apply(thisArg, _arguments || [])).next());
    });
};
Object.defineProperty(exports, "__esModule", { value: true });
const React = __webpack_require__(0);
const NavigationHelper_1 = __webpack_require__(3);
function isViewPage(item) {
    if (item instanceof ViewPage) {
        return true;
    }
    return false;
}
exports.isViewPage = isViewPage;
class ViewPage {
    constructor() {
        this.template = null;
        this.navHelper = new NavigationHelper_1.NavigationHelper(this);
        this.currentPage = "";
    }
    init() {
        return __awaiter(this, void 0, void 0, function* () {
            return;
        });
    }
    setPath(path) {
        this.pagePath = path;
    }
    renderMenu(menu) {
        return __awaiter(this, void 0, void 0, function* () {
            return [];
        });
    }
    renderContent(page) {
        return __awaiter(this, void 0, void 0, function* () {
            const pageContent = yield this.navHelper.navigateTo(page);
            this.currentPage = page;
            if (pageContent) {
                return pageContent;
            }
            return React.createElement("div", null, "404 Not found");
        });
    }
}
exports.ViewPage = ViewPage;


/***/ }),
/* 3 */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __awaiter = (this && this.__awaiter) || function (thisArg, _arguments, P, generator) {
    return new (P || (P = Promise))(function (resolve, reject) {
        function fulfilled(value) { try { step(generator.next(value)); } catch (e) { reject(e); } }
        function rejected(value) { try { step(generator["throw"](value)); } catch (e) { reject(e); } }
        function step(result) { result.done ? resolve(result.value) : new P(function (resolve) { resolve(result.value); }).then(fulfilled, rejected); }
        step((generator = generator.apply(thisArg, _arguments || [])).next());
    });
};
Object.defineProperty(exports, "__esModule", { value: true });
const event_1 = __webpack_require__(5);
class NavigationHelper {
    constructor(thisObject) {
        this.onPreNavigation = event_1.newEvent("NavigationHelper.onPreNavigation");
        this.DEFAULT_VALUE = "";
        this.navObj = "__navObj";
        this.path = {};
        this.thisObject = thisObject;
    }
    static getParts(path) {
        return this.removeEmptyEntries(path.split("/"));
    }
    static removeEmptyEntries(array) {
        const newArray = [];
        array.map((v) => {
            if (v.length > 0) {
                newArray.push(v);
            }
        });
        return newArray;
    }
    static getOptionalField(field) {
        const tField = field.trim();
        if (tField.length > 2 && tField.charAt(0) === "{" && tField.charAt(tField.length - 1) === "}") {
            return tField.substr(1, tField.length - 2);
        }
        return null;
    }
    static isINavObject(obj) {
        return obj && obj.path;
    }
    static handleClick(e, callback) {
        if (e.shiftKey || e.ctrlKey || e.button === 1) {
            return;
        }
        else {
            e.preventDefault();
            callback();
        }
    }
    get defaultPage() {
        return this.DEFAULT_VALUE;
    }
    set defaultPage(value) {
        this.DEFAULT_VALUE = value;
    }
    registerFunction(path, callback) {
        const pathParts = NavigationHelper.getParts(path);
        if (pathParts.length === 0) {
            throw new Error("Can't register function on empty path");
        }
        const curObj = this.createNavPath(pathParts);
        const temp = {
            path: pathParts,
            func: callback,
        };
        curObj[this.navObj] = temp;
    }
    navigateTo(path) {
        return __awaiter(this, void 0, void 0, function* () {
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
            const navObj = curObj[this.navObj];
            const navInfo = {
                matchPath: navObj.path,
                realPath: pathParts,
                params: this.createParamsObj(navObj.path, pathParts),
            };
            this.onPreNavigation({ target: this, navInfo });
            return navObj.func.call(this.thisObject, navInfo);
        });
    }
    createParamsObj(matchPath, realPath) {
        if (matchPath.length !== realPath.length) {
            throw new Error("trying to match different paths");
        }
        const returnObj = {};
        for (let i = 0; i < matchPath.length; i++) {
            const param = NavigationHelper.getOptionalField(matchPath[i]);
            if (param) {
                returnObj[param] = realPath[i];
            }
        }
        return returnObj;
    }
    getNavPath(pathParts) {
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
    createNavPath(pathParts) {
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
exports.NavigationHelper = NavigationHelper;


/***/ }),
/* 4 */
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
var CourseStudentState;
(function (CourseStudentState) {
    CourseStudentState[CourseStudentState["pending"] = 0] = "pending";
    CourseStudentState[CourseStudentState["accepted"] = 1] = "accepted";
    CourseStudentState[CourseStudentState["rejected"] = 2] = "rejected";
})(CourseStudentState = exports.CourseStudentState || (exports.CourseStudentState = {}));


/***/ }),
/* 5 */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
function newEvent(info) {
    const callbacks = [];
    const handler = function EventHandler(event) {
        callbacks.map(((v) => v(event)));
    };
    handler.info = info;
    handler.addEventListener = (callback) => {
        callbacks.push(callback);
    };
    handler.removeEventListener = (callback) => {
        const index = callbacks.indexOf(callback);
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
class MapHelper {
    static mapTo(map, callback) {
        const returnArray = [];
        const keys = Object.keys(map);
        for (const a of keys) {
            const index = parseInt(a, 10);
            returnArray.push(callback(map[index], index, map));
        }
        return returnArray;
    }
    static forEach(map, callback) {
        const keys = Object.keys(map);
        for (const a of keys) {
            const index = parseInt(a, 10);
            callback(map[index], index, map);
        }
    }
    static find(map, callback) {
        const keys = Object.keys(map);
        for (const a of keys) {
            const index = parseInt(a, 10);
            if (callback(map[index], index, map)) {
                return map[index];
            }
        }
        return null;
    }
    static toArray(map) {
        const returnArray = [];
        const keys = Object.keys(map);
        for (const a of keys) {
            const index = parseInt(a, 10);
            returnArray.push(map[index]);
        }
        return returnArray;
    }
}
exports.MapHelper = MapHelper;
function mapify(obj, callback) {
    const newObj = {};
    obj.forEach((ele, index, array) => {
        newObj[callback(ele, index, obj)] = ele;
    });
    return newObj;
}
exports.mapify = mapify;


/***/ }),
/* 7 */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
const React = __webpack_require__(0);
const components_1 = __webpack_require__(1);
class UserView extends React.Component {
    constructor(props) {
        super(props);
        this.state = {
            users: this.props.users,
        };
    }
    render() {
        let searchForm = null;
        if (this.props.addSearchOption) {
            const searchIcon = React.createElement("span", { className: "input-group-addon" },
                React.createElement("i", { className: "glyphicon glyphicon-search" }));
            searchForm = React.createElement(components_1.Search, { className: "input-group", addonBefore: searchIcon, placeholder: "Search for students", onChange: (query) => this.handleOnchange(query) });
        }
        return (React.createElement("div", null,
            searchForm,
            React.createElement(components_1.DynamicTable, { header: ["ID", "First name", "Last name", "Email", "StudentID"], data: this.state.users, selector: (item) => [
                    item.id.toString(),
                    item.firstName,
                    item.lastName,
                    item.email,
                    item.personId.toString(),
                ] })));
    }
    handleOnchange(query) {
        query = query.toLowerCase();
        const filteredData = [];
        this.props.users.forEach((user) => {
            if (user.firstName.toLowerCase().indexOf(query) !== -1
                || user.lastName.toLowerCase().indexOf(query) !== -1
                || user.email.toLowerCase().indexOf(query) !== -1
                || user.personId.toString().indexOf(query) !== -1) {
                filteredData.push(user);
            }
        });
        this.setState({
            users: filteredData,
        });
    }
}
exports.UserView = UserView;


/***/ }),
/* 8 */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __awaiter = (this && this.__awaiter) || function (thisArg, _arguments, P, generator) {
    return new (P || (P = Promise))(function (resolve, reject) {
        function fulfilled(value) { try { step(generator.next(value)); } catch (e) { reject(e); } }
        function rejected(value) { try { step(generator["throw"](value)); } catch (e) { reject(e); } }
        function step(result) { result.done ? resolve(result.value) : new P(function (resolve) { resolve(result.value); }).then(fulfilled, rejected); }
        step((generator = generator.apply(thisArg, _arguments || [])).next());
    });
};
Object.defineProperty(exports, "__esModule", { value: true });
class ArrayHelper {
    static join(array1, array2, callback) {
        const returnObj = [];
        for (const ele1 of array1) {
            for (const ele2 of array2) {
                if (callback(ele1, ele2)) {
                    returnObj.push({ ele1, ele2 });
                }
            }
        }
        return returnObj;
    }
    static find(array, predicate) {
        for (let i = 0; i < array.length; i++) {
            const cur = array[i];
            if (predicate.call(array, cur, i, array)) {
                return cur;
            }
        }
        return null;
    }
    static mapAsync(array, callback) {
        return __awaiter(this, void 0, void 0, function* () {
            const newArray = [];
            for (let i = 0; i < array.length; i++) {
                newArray.push(yield callback(array[i], i, array));
            }
            return newArray;
        });
    }
}
exports.ArrayHelper = ArrayHelper;


/***/ }),
/* 9 */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
const React = __webpack_require__(0);
const NavigationHelper_1 = __webpack_require__(3);
class NavHeaderBar extends React.Component {
    componentDidMount() {
        const temp = this.refs.button;
        temp.setAttribute("data-toggle", "collapse");
        temp.setAttribute("data-target", "#" + this.props.id);
        temp.setAttribute("aria-expanded", "false");
    }
    render() {
        return React.createElement("div", { className: "navbar-header" },
            React.createElement("button", { ref: "button", type: "button", className: "navbar-toggle collapsed" },
                React.createElement("span", { className: "sr-only" }, "Toggle navigation"),
                React.createElement("span", { className: "icon-bar" }),
                React.createElement("span", { className: "icon-bar" }),
                React.createElement("span", { className: "icon-bar" })),
            React.createElement("a", { className: "navbar-brand", onClick: (e) => {
                    NavigationHelper_1.NavigationHelper.handleClick(e, () => {
                        this.props.brandClick();
                    });
                }, href: ";/" }, this.props.brandName));
    }
}
exports.NavHeaderBar = NavHeaderBar;


/***/ }),
/* 10 */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
const React = __webpack_require__(0);
const NavigationHelper_1 = __webpack_require__(3);
class NavMenu extends React.Component {
    render() {
        const items = this.props.links.map((v, i) => {
            let active = "";
            if (v.active) {
                active = "active";
            }
            if (v.uri) {
                return React.createElement("li", { key: i, className: active },
                    React.createElement("a", { onClick: (e) => this.handleClick(e, v), href: "/" + v.uri }, v.name));
            }
            else {
                return React.createElement("li", { key: i, className: active },
                    React.createElement("span", { className: "header" }, v.name));
            }
        });
        return React.createElement("ul", { className: "nav nav-list" }, items);
    }
    handleClick(e, link) {
        NavigationHelper_1.NavigationHelper.handleClick(e, () => {
            if (this.props.onClick) {
                this.props.onClick(link);
            }
        });
    }
}
exports.NavMenu = NavMenu;


/***/ }),
/* 11 */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
const React = __webpack_require__(0);
class ProgressBar extends React.Component {
    render() {
        const progressBarStyle = {
            width: this.props.progress + "%",
        };
        return (React.createElement("div", { className: "progress" },
            React.createElement("div", { className: "progress-bar", role: "progressbar", "aria-valuenow": this.props.progress, "aria-valuemin": "0", "aria-valuemax": "100", style: progressBarStyle },
                this.props.progress,
                "%")));
    }
}
exports.ProgressBar = ProgressBar;


/***/ }),
/* 12 */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
const React = __webpack_require__(0);
class HelloView extends React.Component {
    render() {
        return React.createElement("h1", null, "Hello world");
    }
}
exports.HelloView = HelloView;


/***/ }),
/* 13 */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
const React = __webpack_require__(0);
const NavigationHelper_1 = __webpack_require__(3);
class CollapsableNavMenu extends React.Component {
    constructor() {
        super(...arguments);
        this.topItems = [];
    }
    render() {
        const children = this.props.links.map((e, i) => {
            return this.renderTopElement(i, e);
        });
        return React.createElement("ul", { className: "nav nav-list" }, children);
    }
    toggle(index) {
        const animations = [];
        this.topItems.forEach((temp, i) => {
            if (i === index) {
                if (this.collapseIsOpen(temp)) {
                }
                else {
                    animations.push(this.openCollapse(temp));
                }
            }
            else {
                animations.push(this.closeIfOpen(temp));
            }
        });
        setTimeout(() => {
            animations.forEach((e) => {
                e();
            });
        }, 10);
    }
    collapseIsOpen(ele) {
        return ele.classList.contains("in");
    }
    closeIfOpen(ele) {
        if (this.collapseIsOpen(ele)) {
            return this.closeCollapse(ele);
        }
        return () => {
            "do nothing";
        };
    }
    openCollapse(ele) {
        ele.classList.remove("collapse");
        ele.classList.add("collapsing");
        return () => {
            ele.style.height = ele.scrollHeight + "px";
            setTimeout(() => {
                ele.classList.remove("collapsing");
                ele.classList.add("collapse");
                ele.classList.add("in");
                ele.style.height = null;
            }, 350);
        };
    }
    closeCollapse(ele) {
        ele.style.height = ele.clientHeight + "px";
        ele.classList.add("collapsing");
        ele.classList.remove("collapse");
        ele.classList.remove("in");
        return () => {
            ele.style.height = null;
            setTimeout(() => {
                ele.classList.remove("collapsing");
                ele.classList.add("collapse");
                ele.style.height = null;
            }, 350);
        };
    }
    handleClick(e, link) {
        NavigationHelper_1.NavigationHelper.handleClick(e, () => {
            if (this.props.onClick) {
                this.props.onClick(link);
            }
        });
    }
    renderChilds(index, link) {
        const isActive = link.active ? "active" : "";
        if (link.uri) {
            return React.createElement("li", { key: index, className: isActive },
                React.createElement("a", { onClick: (e) => this.handleClick(e, link), href: "/" + link.uri }, link.name));
        }
        else {
            return React.createElement("li", { key: index, className: isActive },
                React.createElement("span", { className: "header" }, link.name));
        }
    }
    renderTopElement(index, links) {
        const isActive = links.item.active ? "active" : "";
        const subClass = "nav nav-sub collapse " + (links.item.active ? "in" : "");
        let children = [];
        if (links.children) {
            children = links.children.map((e, i) => {
                return this.renderChilds(i, e);
            });
        }
        return React.createElement("li", { key: index, className: isActive },
            React.createElement("a", { onClick: (e) => {
                    this.toggle(index);
                    this.handleClick(e, links.item);
                }, href: "/" + links.item.uri },
                links.item.name,
                React.createElement("span", { style: { float: "right" } },
                    React.createElement("span", { className: "glyphicon glyphicon-menu-down" }))),
            React.createElement("ul", { ref: (ele) => {
                    if (ele) {
                        this.topItems[index] = ele;
                    }
                }, className: subClass }, children));
    }
}
exports.CollapsableNavMenu = CollapsableNavMenu;


/***/ }),
/* 14 */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __awaiter = (this && this.__awaiter) || function (thisArg, _arguments, P, generator) {
    return new (P || (P = Promise))(function (resolve, reject) {
        function fulfilled(value) { try { step(generator.next(value)); } catch (e) { reject(e); } }
        function rejected(value) { try { step(generator["throw"](value)); } catch (e) { reject(e); } }
        function step(result) { result.done ? resolve(result.value) : new P(function (resolve) { resolve(result.value); }).then(fulfilled, rejected); }
        step((generator = generator.apply(thisArg, _arguments || [])).next());
    });
};
Object.defineProperty(exports, "__esModule", { value: true });
const React = __webpack_require__(0);
const ReactDOM = __webpack_require__(15);
const components_1 = __webpack_require__(1);
const managers_1 = __webpack_require__(33);
const ErrorPage_1 = __webpack_require__(38);
const HelpPage_1 = __webpack_require__(39);
const HomePage_1 = __webpack_require__(41);
const StudentPage_1 = __webpack_require__(42);
const TeacherPage_1 = __webpack_require__(44);
const AdminPage_1 = __webpack_require__(45);
const NavBarLogin_1 = __webpack_require__(46);
const NavBarMenu_1 = __webpack_require__(47);
const LoginPage_1 = __webpack_require__(48);
const ServerProvider_1 = __webpack_require__(49);
class AutoGrader extends React.Component {
    constructor(props) {
        super();
        this.userMan = props.userManager;
        this.navMan = props.navigationManager;
        const curUser = this.userMan.getCurrentUser();
        this.state = {
            activePage: undefined,
            topLinks: [],
            curUser,
            currentContent: React.createElement("div", null, "No Content Available"),
        };
        (() => __awaiter(this, void 0, void 0, function* () {
            this.setState({ topLinks: yield this.generateTopLinksFor(curUser) });
        }))();
        this.navMan.onNavigate.addEventListener((e) => this.handleNavigation(e));
        this.userMan.onLogin.addEventListener((e) => __awaiter(this, void 0, void 0, function* () {
            console.log("Sign in");
            this.setState({
                curUser: e.user,
                topLinks: yield this.generateTopLinksFor(e.user),
            });
        }));
        this.userMan.onLogout.addEventListener((e) => __awaiter(this, void 0, void 0, function* () {
            console.log("Sign out");
            this.setState({
                curUser: null,
                topLinks: yield this.generateTopLinksFor(null),
            });
        }));
    }
    handleNavigation(e) {
        return __awaiter(this, void 0, void 0, function* () {
            this.subPage = e.subPage;
            const newContent = yield this.renderTemplate(e.page, e.page.template);
            const tempLink = this.state.topLinks.slice();
            this.checkLinks(tempLink);
            this.setState({ activePage: e.page, topLinks: tempLink, currentContent: newContent });
        });
    }
    generateTopLinksFor(user) {
        return __awaiter(this, void 0, void 0, function* () {
            if (user) {
                const basis = [];
                if (yield this.userMan.isTeacher(user)) {
                    basis.push({ name: "Teacher", uri: "app/teacher/", active: false });
                }
                basis.push({ name: "Student", uri: "app/student/", active: false });
                if (this.userMan.isAdmin(user)) {
                    basis.push({ name: "Admin", uri: "app/admin", active: false });
                }
                basis.push({ name: "Help", uri: "app/help", active: false });
                return basis;
            }
            else {
                return [{ name: "Help", uri: "app/help", active: false }];
            }
        });
    }
    componentDidMount() {
        const curUrl = location.pathname;
        if (curUrl === "/") {
            this.navMan.navigateToDefault();
        }
        else {
            this.navMan.navigateTo(curUrl);
        }
    }
    render() {
        if (this.state.activePage) {
            return this.state.currentContent;
        }
        else {
            return React.createElement("h1", null, "404 not found");
        }
    }
    handleClick(link) {
        if (link.uri) {
            this.navMan.navigateTo(link.uri);
        }
        else {
            console.warn("Warning! Empty link detected", link);
        }
    }
    renderActiveMenu(page, menu) {
        return __awaiter(this, void 0, void 0, function* () {
            if (page) {
                return yield page.renderMenu(menu);
            }
            return "";
        });
    }
    renderActivePage(page, subPage) {
        return __awaiter(this, void 0, void 0, function* () {
            if (page) {
                return yield page.renderContent(subPage);
            }
            return React.createElement("h1", null, "404 Page not found");
        });
    }
    checkLinks(links) {
        this.navMan.checkLinks(links);
    }
    renderTemplate(page, name) {
        return __awaiter(this, void 0, void 0, function* () {
            let body;
            const content = yield this.renderActivePage(page, this.subPage);
            const loginLink = [
                { name: "Github", uri: "app/login/login/github" },
                { name: "Gitlab", uri: "app/login/login/gitlab" },
            ];
            switch (name) {
                case "frontpage":
                    body = (React.createElement(components_1.Row, { className: "container-fluid" },
                        React.createElement("div", { className: "col-xs-12" }, content)));
                default:
                    body = (React.createElement(components_1.Row, { className: "container-fluid" },
                        React.createElement("div", { className: "col-md-2 col-sm-3 col-xs-12" }, yield this.renderActiveMenu(page, 0)),
                        React.createElement("div", { className: "col-md-10 col-sm-9 col-xs-12" }, content)));
            }
            return (React.createElement("div", null,
                React.createElement(components_1.NavBar, { id: "top-bar", isFluid: false, isInverse: true, onClick: (link) => this.handleClick(link), brandName: "Auto Grader" },
                    React.createElement(NavBarMenu_1.NavBarMenu, { links: this.state.topLinks, onClick: (link) => this.handleClick(link) }),
                    React.createElement(NavBarLogin_1.NavBarLogin, { user: this.state.curUser, links: loginLink, onClick: (link) => this.handleClick(link) })),
                body));
        });
    }
}
function main() {
    return __awaiter(this, void 0, void 0, function* () {
        const DEBUG_BROWSER = "DEBUG_BROWSER";
        const DEBUG_SERVER = "DEBUG_SERVER";
        let curRunning;
        curRunning = DEBUG_SERVER;
        const tempData = new managers_1.TempDataProvider();
        let userMan;
        let courseMan;
        let navMan;
        if (curRunning === DEBUG_SERVER) {
            const serverData = new ServerProvider_1.ServerProvider();
            userMan = new managers_1.UserManager(serverData);
            courseMan = new managers_1.CourseManager(tempData);
            navMan = new managers_1.NavigationManager(history);
        }
        else {
            userMan = new managers_1.UserManager(tempData);
            courseMan = new managers_1.CourseManager(tempData);
            navMan = new managers_1.NavigationManager(history);
            const user = yield userMan.tryLogin("test@testersen.no", "1234");
        }
        window.debugData = { tempData, userMan, courseMan, navMan };
        navMan.setDefaultPath("app/home");
        yield navMan.registerPage("app/home", new HomePage_1.HomePage());
        yield navMan.registerPage("app/student", new StudentPage_1.StudentPage(userMan, navMan, courseMan));
        yield navMan.registerPage("app/teacher", new TeacherPage_1.TeacherPage(userMan, navMan, courseMan));
        yield navMan.registerPage("app/admin", new AdminPage_1.AdminPage(navMan, userMan, courseMan));
        yield navMan.registerPage("app/help", new HelpPage_1.HelpPage(navMan));
        yield navMan.registerPage("app/login", new LoginPage_1.LoginPage(navMan, userMan));
        navMan.registerErrorPage(404, new ErrorPage_1.ErrorPage());
        navMan.onNavigate.addEventListener((e) => {
            console.log(e);
        });
        ReactDOM.render(React.createElement(AutoGrader, { userManager: userMan, navigationManager: navMan }), document.getElementById("root"));
    });
}
main();


/***/ }),
/* 15 */
/***/ (function(module, exports) {

module.exports = ReactDOM;

/***/ }),
/* 16 */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
const React = __webpack_require__(0);
const NavHeaderBar_1 = __webpack_require__(9);
class NavBar extends React.Component {
    render() {
        return React.createElement("nav", { className: this.renderNavBarClass() },
            React.createElement("div", { className: this.renderIsFluid() },
                React.createElement(NavHeaderBar_1.NavHeaderBar, { id: this.props.id, brandName: this.props.brandName, brandClick: () => this.handleClick({ name: "Home", uri: "/" }) }),
                React.createElement("div", { className: "collapse navbar-collapse", id: this.props.id }, this.props.children)));
    }
    handleClick(link) {
        if (this.props.onClick) {
            this.props.onClick(link);
        }
    }
    renderIsFluid() {
        let name = "container";
        if (this.props.isFluid) {
            name += "-fluid";
        }
        return name;
    }
    renderNavBarClass() {
        let name = "navbar navbar-absolute-top";
        if (this.props.isInverse) {
            name += " navbar-inverse";
        }
        else {
            name += " navbar-default";
        }
        return name;
    }
}
exports.NavBar = NavBar;


/***/ }),
/* 17 */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
const React = __webpack_require__(0);
class NavMenuFormatable extends React.Component {
    render() {
        const items = this.props.items.map((v, i) => {
            return React.createElement("li", { key: i },
                React.createElement("a", { href: "#", onClick: () => { this.handleItemClick(v); } }, this.renderObj(v)));
        });
        return React.createElement("ul", { className: "nav nav-pills nav-stacked" }, items);
    }
    renderObj(item) {
        if (this.props.formater) {
            return this.props.formater(item);
        }
        return item.toString();
    }
    handleItemClick(item) {
        if (this.props.onClick) {
            this.props.onClick(item);
        }
    }
}
exports.NavMenuFormatable = NavMenuFormatable;


/***/ }),
/* 18 */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
const React = __webpack_require__(0);
class DynamicTable extends React.Component {
    render() {
        const footer = this.props.footer;
        const rows = this.props.data.map((v, i) => {
            return this.renderRow(v, i);
        });
        const tableFooter = footer ? React.createElement("tfoot", null,
            React.createElement("tr", null, this.renderCells(footer))) : null;
        return (React.createElement("table", { className: this.props.onRowClick ? "table table-hover" : "table" },
            React.createElement("thead", null,
                React.createElement("tr", null, this.renderCells(this.props.header, true))),
            React.createElement("tbody", null, rows),
            tableFooter));
    }
    renderCells(values, th = false) {
        return values.map((v, i) => {
            if (th) {
                return React.createElement("th", { key: i }, v);
            }
            return React.createElement("td", { key: i }, v);
        });
    }
    renderRow(item, i) {
        return (React.createElement("tr", { key: i, onClick: (e) => this.handleRowClick(item) }, this.renderCells(this.props.selector(item))));
    }
    handleRowClick(item) {
        if (this.props.onRowClick) {
            this.props.onRowClick(item);
        }
    }
}
exports.DynamicTable = DynamicTable;


/***/ }),
/* 19 */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
const React = __webpack_require__(0);
function Row(props) {
    return React.createElement("div", { className: props.className ? "row " + props.className : "row" }, props.children);
}
exports.Row = Row;


/***/ }),
/* 20 */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
const React = __webpack_require__(0);
const LabResultView_1 = __webpack_require__(21);
class StudentLab extends React.Component {
    render() {
        const testCases = [
            { name: "Test Case 1", score: 60, points: 100, weight: 1 },
            { name: "Test Case 2", score: 50, points: 100, weight: 1 },
            { name: "Test Case 3", score: 40, points: 100, weight: 1 },
            { name: "Test Case 4", score: 30, points: 100, weight: 1 },
            { name: "Test Case 5", score: 20, points: 100, weight: 1 },
        ];
        const labInfo = {
            lab: this.props.assignment.name,
            course: this.props.course.name,
            score: 50,
            weight: 100,
            test_cases: testCases,
            pass_tests: 10,
            fail_tests: 20,
            exec_time: 0.33,
            build_time: new Date(2017, 5, 25),
            build_id: 10,
        };
        if (this.props.student) {
            labInfo.student = this.props.student;
        }
        return React.createElement(LabResultView_1.LabResultView, { labInfo: labInfo });
    }
}
exports.StudentLab = StudentLab;


/***/ }),
/* 21 */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
const React = __webpack_require__(0);
const components_1 = __webpack_require__(1);
class LabResultView extends React.Component {
    render() {
        return (React.createElement("div", { className: "col-md-9 col-sm-9 col-xs-12" },
            React.createElement("div", { className: "result-content", id: "resultview" },
                React.createElement("section", { id: "result" },
                    React.createElement(components_1.LabResult, { course_name: this.props.labInfo.course, lab: this.props.labInfo.lab, progress: this.props.labInfo.score, student: this.props.labInfo.student }),
                    React.createElement(components_1.LastBuild, { test_cases: this.props.labInfo.test_cases, score: this.props.labInfo.score, weight: this.props.labInfo.weight }),
                    React.createElement(components_1.LastBuildInfo, { pass_tests: this.props.labInfo.pass_tests, fail_tests: this.props.labInfo.fail_tests, exec_time: this.props.labInfo.exec_time, build_time: this.props.labInfo.build_time, build_id: this.props.labInfo.build_id }),
                    React.createElement(components_1.Row, null,
                        React.createElement("div", { className: "col-lg-12" },
                            React.createElement("div", { className: "well" },
                                React.createElement("code", { id: "logs" }, "# There is no build for this lab yet."))))))));
    }
}
exports.LabResultView = LabResultView;


/***/ }),
/* 22 */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
const React = __webpack_require__(0);
const NavigationHelper_1 = __webpack_require__(3);
class NavDropdown extends React.Component {
    constructor() {
        super();
        this.state = {
            isOpen: false,
        };
    }
    render() {
        const children = this.props.items.map((item, index) => {
            return React.createElement("li", { key: index },
                React.createElement("a", { href: "/" + item.uri, onClick: (e) => {
                        NavigationHelper_1.NavigationHelper.handleClick(e, () => {
                            this.toggleOpen();
                            this.props.itemClick(item, index);
                        });
                    } }, item.name));
        });
        return React.createElement("div", { className: this.getButtonClass() },
            React.createElement("button", { className: "btn btn-default dropdown-toggle", type: "button", onClick: () => this.toggleOpen() },
                this.renderActive(),
                React.createElement("span", { className: "caret" })),
            React.createElement("ul", { className: "dropdown-menu" }, children));
    }
    getButtonClass() {
        if (this.state.isOpen) {
            return "button open";
        }
        else {
            return "button";
        }
    }
    toggleOpen() {
        const newState = !this.state.isOpen;
        this.setState({ isOpen: newState });
    }
    renderActive() {
        if (this.props.items.length === 0) {
            return "";
        }
        let curIndex = this.props.selectedIndex;
        if (curIndex >= this.props.items.length || curIndex < 0) {
            curIndex = 0;
        }
        return this.props.items[curIndex].name;
    }
}
exports.NavDropdown = NavDropdown;


/***/ }),
/* 23 */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
const React = __webpack_require__(0);
const components_1 = __webpack_require__(1);
class LabResult extends React.Component {
    render() {
        let labHeading;
        if (this.props.student) {
            labHeading = React.createElement("h3", null,
                this.props.student.firstName + " " + this.props.student.lastName,
                ": ",
                this.props.lab);
        }
        else {
            labHeading = React.createElement("div", null,
                React.createElement("h1", null, this.props.course_name),
                React.createElement("p", { className: "lead" },
                    "Your progress on ",
                    React.createElement("strong", null,
                        React.createElement("span", { id: "lab-headline" }, this.props.lab))));
        }
        return (React.createElement(components_1.Row, null,
            React.createElement("div", { className: "col-lg-12" },
                labHeading,
                React.createElement(components_1.ProgressBar, { progress: this.props.progress })),
            React.createElement("div", { className: "col-lg-6" },
                React.createElement("p", null,
                    React.createElement("strong", { id: "status" }, "Status: Nothing built yet."))),
            React.createElement("div", { className: "col-lg-6" },
                React.createElement("p", null,
                    React.createElement("strong", { id: "pushtime" }, "Code delievered: - ")))));
    }
}
exports.LabResult = LabResult;


/***/ }),
/* 24 */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
const React = __webpack_require__(0);
const components_1 = __webpack_require__(1);
class LastBuild extends React.Component {
    render() {
        return (React.createElement(components_1.Row, null,
            React.createElement("div", { className: "col-lg-12" },
                React.createElement(components_1.DynamicTable, { header: ["Test name", "Score", "Weight"], data: this.props.test_cases, selector: (item) => [item.name, item.score.toString() + "/"
                            + item.points.toString() + " pts", item.weight.toString() + " pts"], footer: ["Total score", this.props.score.toString() + "%", this.props.weight.toString() + "%"] }))));
    }
}
exports.LastBuild = LastBuild;


/***/ }),
/* 25 */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
const React = __webpack_require__(0);
const components_1 = __webpack_require__(1);
class LastBuildInfo extends React.Component {
    render() {
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
                            React.createElement("button", { type: "button", id: "rebuild", className: "btn btn-primary", onClick: () => this.handleClick() }, "Rebuild")))))));
    }
    handleClick() {
        console.log("Rebuilding...");
    }
}
exports.LastBuildInfo = LastBuildInfo;


/***/ }),
/* 26 */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
const React = __webpack_require__(0);
const components_1 = __webpack_require__(1);
class CoursesOverview extends React.Component {
    render() {
        const courses = this.props.course_overview.map((val, key) => {
            return React.createElement(components_1.CoursePanel, { key: key, course: val.course, labs: val.labs, navMan: this.props.navMan });
        });
        let added = 0;
        let index = 1;
        let l = courses.length;
        for (index; index < l; index++) {
            if (index % 2 === 0) {
                courses.splice(index + added, 0, React.createElement("div", { className: "visible-md-block visible-sm-block clearfix" }));
                l += 1;
                added += 1;
            }
            if (index % 4 === 0) {
                courses.splice(index + added, 0, React.createElement("div", { className: "visible-lg-block clearfix" }));
                l += 1;
                added += 1;
            }
        }
        return (React.createElement("div", null,
            React.createElement("h1", null, "Your Courses"),
            React.createElement(components_1.Row, null, courses)));
    }
}
exports.CoursesOverview = CoursesOverview;


/***/ }),
/* 27 */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
const React = __webpack_require__(0);
const components_1 = __webpack_require__(1);
class CoursePanel extends React.Component {
    render() {
        const pathPrefix = "app/student/course/" + this.props.course.id + "/lab/";
        return (React.createElement("div", { className: "col-lg-3 col-md-6 col-sm-6" },
            React.createElement("div", { className: "panel panel-primary" },
                React.createElement("div", { className: "panel-heading clickable", onClick: () => this.handleCourseClick() }, this.props.course.name),
                React.createElement("div", { className: "panel-body" },
                    React.createElement(components_1.DynamicTable, { header: ["Labs", "Score", "Weight"], data: this.props.labs, selector: (item) => [item.name, "50%", "100%"], onRowClick: (lab) => this.handleRowClick(pathPrefix, lab) })))));
    }
    handleRowClick(pathPrefix, lab) {
        if (lab) {
            this.props.navMan.navigateTo(pathPrefix + lab.id);
        }
    }
    handleCourseClick() {
        const uri = "app/student/course/" + this.props.course.id;
        this.props.navMan.navigateTo(uri);
    }
}
exports.CoursePanel = CoursePanel;


/***/ }),
/* 28 */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
const React = __webpack_require__(0);
const ProgressBar_1 = __webpack_require__(11);
class SingleCourseOverview extends React.Component {
    render() {
        const labs = this.props.courseAndLabs.labs.map((v, k) => {
            return (React.createElement("li", { key: k, className: "list-group-item" },
                React.createElement("strong", null, v.name),
                React.createElement(ProgressBar_1.ProgressBar, { progress: Math.floor((Math.random() * 100) + 1) })));
        });
        return (React.createElement("div", null,
            React.createElement("h1", null, this.props.courseAndLabs.course.name),
            React.createElement("div", null,
                React.createElement("ul", { className: "list-group" }, labs))));
    }
}
exports.SingleCourseOverview = SingleCourseOverview;


/***/ }),
/* 29 */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
const React = __webpack_require__(0);
class Button extends React.Component {
    render() {
        return (React.createElement("button", { className: this.props.className ? "btn " + this.props.className : "btn", type: this.props.type ? this.props.type : "", onClick: () => this.handleOnclick() }, this.props.text ? this.props.text : ""));
    }
    handleOnclick() {
        if (this.props.onClick) {
            this.props.onClick();
        }
    }
}
exports.Button = Button;


/***/ }),
/* 30 */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
const React = __webpack_require__(0);
const components_1 = __webpack_require__(1);
class CourseForm extends React.Component {
    constructor(props) {
        super(props);
        this.state = {
            name: "",
            tag: "",
            year: "",
        };
    }
    render() {
        return (React.createElement("form", { className: this.props.className ? this.props.className : "", onSubmit: (e) => this.handleFormSubmit(e) },
            React.createElement("div", { className: "form-group" },
                React.createElement("label", { className: "control-label col-sm-2", htmlFor: "name" }, "Course Name:"),
                React.createElement("div", { className: "col-sm-10" },
                    React.createElement("input", { type: "text", className: "form-control", id: "name", placeholder: "Enter course name", name: "name", value: this.state.name, onChange: (e) => this.handleInputChange(e) }))),
            React.createElement("div", { className: "form-group" },
                React.createElement("label", { className: "control-label col-sm-2", htmlFor: "tag" }, "Course Tag:"),
                React.createElement("div", { className: "col-sm-10" },
                    React.createElement("input", { type: "text", className: "form-control", id: "tag", placeholder: "Enter course tag", name: "tag", value: this.state.tag, onChange: (e) => this.handleInputChange(e) }))),
            React.createElement("div", { className: "form-group" },
                React.createElement("label", { className: "control-label col-sm-2", htmlFor: "tag" }, "Year/Semester:"),
                React.createElement("div", { className: "col-sm-10" },
                    React.createElement("input", { type: "text", className: "form-control", id: "tag", placeholder: "Enter year/semester", name: "year", value: this.state.year, onChange: (e) => this.handleInputChange(e) }))),
            React.createElement("div", { className: "form-group" },
                React.createElement("div", { className: "col-sm-offset-2 col-sm-10" },
                    React.createElement(components_1.Button, { className: "btn btn-primary", text: "Submit", type: "submit" })))));
    }
    handleFormSubmit(e) {
        e.preventDefault();
        const errors = this.courseValidate();
        this.props.onSubmit(this.state, errors);
    }
    handleInputChange(e) {
        const target = e.target;
        const value = target.type === "checkbox" ? target.checked : target.value;
        const name = target.name;
        this.setState({
            [name]: value,
        });
    }
    courseValidate() {
        const errors = [];
        if (this.state.name === "") {
            errors.push("Course Name cannot be blank");
        }
        if (this.state.tag === "") {
            errors.push("Course Tag cannot be blank.");
        }
        if (this.state.year === "") {
            errors.push("Year/Semester cannot be blank.");
        }
        return errors;
    }
}
exports.CourseForm = CourseForm;


/***/ }),
/* 31 */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
const React = __webpack_require__(0);
const components_1 = __webpack_require__(1);
class Results extends React.Component {
    constructor(props) {
        super(props);
        this.state = {
            assignment: this.props.labs[0],
            selectedStudent: this.props.students[0],
            students: this.props.students,
        };
    }
    render() {
        let studentLab = null;
        if (this.props.students.length > 0) {
            studentLab = React.createElement(components_1.StudentLab, { course: this.props.course, assignment: this.state.assignment, student: this.state.selectedStudent });
        }
        const searchIcon = React.createElement("span", { className: "input-group-addon" },
            React.createElement("i", { className: "glyphicon glyphicon-search" }));
        return (React.createElement("div", null,
            React.createElement("h1", null,
                "Result: ",
                this.props.course.name),
            React.createElement(components_1.Row, null,
                React.createElement("div", { className: "col-lg6 col-md-6 col-sm-12" },
                    React.createElement(components_1.Search, { className: "input-group", addonBefore: searchIcon, placeholder: "Search for students", onChange: (query) => this.handleOnchange(query) }),
                    React.createElement(components_1.DynamicTable, { header: this.getResultHeader(), data: this.state.students, selector: (item) => this.getResultSelector(item) })),
                React.createElement("div", { className: "col-lg-6 col-md-6 col-sm-12" }, studentLab))));
    }
    getResultHeader() {
        let headers = ["Name", "Slipdays"];
        headers = headers.concat(this.props.labs.map((e) => e.name));
        return headers;
    }
    getResultSelector(student) {
        let selector = [student.firstName + " " + student.lastName, "5"];
        selector = selector.concat(this.props.labs.map((e) => React.createElement("a", { className: "lab-result-cell", onClick: () => this.handleOnclick(student, e), href: "#" }, Math.floor((Math.random() * 100) + 1).toString() + "%")));
        return selector;
    }
    handleOnclick(std, lab) {
        this.setState({
            selectedStudent: std,
            assignment: lab,
        });
    }
    handleOnchange(query) {
        query = query.toLowerCase();
        const filteredData = [];
        this.props.students.forEach((std) => {
            if (std.firstName.toLowerCase().indexOf(query) !== -1
                || std.lastName.toLowerCase().indexOf(query) !== -1
                || std.email.toLowerCase().indexOf(query) !== -1) {
                filteredData.push(std);
            }
        });
        this.setState({
            students: filteredData,
        });
    }
}
exports.Results = Results;


/***/ }),
/* 32 */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
const React = __webpack_require__(0);
class Search extends React.Component {
    constructor(props) {
        super(props);
        this.state = {
            query: "",
        };
    }
    render() {
        let addOn = null;
        if (this.props.addonBefore) {
            addOn = this.props.addonBefore;
        }
        return (React.createElement("div", { className: this.props.className ? this.props.className : "" },
            addOn,
            React.createElement("input", { className: "form-control", type: "text", placeholder: this.props.placeholder ? this.props.placeholder : "", onChange: (e) => this.onChange(e), value: this.state.query })));
    }
    onChange(e) {
        this.setState({
            query: e.target.value,
        });
        if (this.props.onChange) {
            this.props.onChange(e.target.value);
        }
    }
}
exports.Search = Search;


/***/ }),
/* 33 */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

function __export(m) {
    for (var p in m) if (!exports.hasOwnProperty(p)) exports[p] = m[p];
}
Object.defineProperty(exports, "__esModule", { value: true });
__export(__webpack_require__(34));
__export(__webpack_require__(35));
__export(__webpack_require__(36));
__export(__webpack_require__(37));


/***/ }),
/* 34 */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __awaiter = (this && this.__awaiter) || function (thisArg, _arguments, P, generator) {
    return new (P || (P = Promise))(function (resolve, reject) {
        function fulfilled(value) { try { step(generator.next(value)); } catch (e) { reject(e); } }
        function rejected(value) { try { step(generator["throw"](value)); } catch (e) { reject(e); } }
        function step(result) { result.done ? resolve(result.value) : new P(function (resolve) { resolve(result.value); }).then(fulfilled, rejected); }
        step((generator = generator.apply(thisArg, _arguments || [])).next());
    });
};
Object.defineProperty(exports, "__esModule", { value: true });
const map_1 = __webpack_require__(6);
const models_1 = __webpack_require__(4);
class CourseManager {
    constructor(courseProvider) {
        this.courseProvider = courseProvider;
    }
    addUserToCourse(user, course) {
        return __awaiter(this, void 0, void 0, function* () {
            return this.courseProvider.addUserToCourse(user, course);
        });
    }
    getCourse(id) {
        return __awaiter(this, void 0, void 0, function* () {
            const a = (yield this.getCourses())[id];
            if (a) {
                return a;
            }
            return null;
        });
    }
    getCourses() {
        return __awaiter(this, void 0, void 0, function* () {
            return map_1.MapHelper.toArray(yield this.courseProvider.getCourses());
        });
    }
    getRelationsFor(user, state) {
        return __awaiter(this, void 0, void 0, function* () {
            const cLinks = [];
            for (const c of yield this.courseProvider.getCoursesStudent()) {
                if (user.id === c.personId && (state === undefined || c.state === models_1.CourseStudentState.accepted)) {
                    cLinks.push(c);
                }
            }
            return cLinks;
        });
    }
    getCoursesFor(user, state) {
        return __awaiter(this, void 0, void 0, function* () {
            const cLinks = [];
            for (const c of yield this.courseProvider.getCoursesStudent()) {
                if (user.id === c.personId && (state === undefined || c.state === models_1.CourseStudentState.accepted)) {
                    cLinks.push(c);
                }
            }
            const courses = [];
            const tempCourses = yield this.getCourses();
            for (const link of cLinks) {
                const c = tempCourses[link.courseId];
                if (c) {
                    courses.push(c);
                }
            }
            return courses;
        });
    }
    getUserIdsForCourse(course, state) {
        return __awaiter(this, void 0, void 0, function* () {
            const users = [];
            for (const c of yield this.courseProvider.getCoursesStudent()) {
                if (course.id === c.courseId && (state === undefined || c.state === models_1.CourseStudentState.accepted)) {
                    users.push(c);
                }
            }
            return users;
        });
    }
    getAssignment(course, assignmentId) {
        return __awaiter(this, void 0, void 0, function* () {
            const temp = yield this.courseProvider.getAssignments(course.id);
            if (temp[assignmentId]) {
                return temp[assignmentId];
            }
            return null;
        });
    }
    getAssignments(courseId) {
        return __awaiter(this, void 0, void 0, function* () {
            if (models_1.isCourse(courseId)) {
                courseId = courseId.id;
            }
            return map_1.MapHelper.toArray(yield this.courseProvider.getAssignments(courseId));
        });
    }
    changeUserState(link, state) {
        return __awaiter(this, void 0, void 0, function* () {
            return this.courseProvider.changeUserState(link, state);
        });
    }
    createNewCourse(courseData) {
        return __awaiter(this, void 0, void 0, function* () {
            return this.courseProvider.createNewCourse(courseData);
        });
    }
}
exports.CourseManager = CourseManager;


/***/ }),
/* 35 */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __awaiter = (this && this.__awaiter) || function (thisArg, _arguments, P, generator) {
    return new (P || (P = Promise))(function (resolve, reject) {
        function fulfilled(value) { try { step(generator.next(value)); } catch (e) { reject(e); } }
        function rejected(value) { try { step(generator["throw"](value)); } catch (e) { reject(e); } }
        function step(result) { result.done ? resolve(result.value) : new P(function (resolve) { resolve(result.value); }).then(fulfilled, rejected); }
        step((generator = generator.apply(thisArg, _arguments || [])).next());
    });
};
Object.defineProperty(exports, "__esModule", { value: true });
const event_1 = __webpack_require__(5);
const NavigationHelper_1 = __webpack_require__(3);
const ViewPage_1 = __webpack_require__(2);
function isILinkCollection(item) {
    if (item.item) {
        return true;
    }
    return false;
}
exports.isILinkCollection = isILinkCollection;
class NavigationManager {
    constructor(history) {
        this.onNavigate = event_1.newEvent("NavigationManager.onNavigate");
        this.pages = {};
        this.errorPages = [];
        this.defaultPath = "";
        this.currentPath = "";
        this.browserHistory = history;
        window.addEventListener("popstate", (e) => {
            this.navigateTo(location.pathname, true);
        });
    }
    setDefaultPath(path) {
        this.defaultPath = path;
    }
    navigateTo(path, preventPush) {
        if (path === "/") {
            this.navigateToDefault();
            return;
        }
        const parts = NavigationHelper_1.NavigationHelper.getParts(path);
        let curPage = this.pages;
        this.currentPath = parts.join("/");
        if (!preventPush) {
            this.browserHistory.pushState({}, "Autograder", "/" + this.currentPath);
        }
        for (let i = 0; i < parts.length; i++) {
            const a = parts[i];
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
                const cur = curPage[a];
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
    }
    navigateToDefault() {
        this.navigateTo(this.defaultPath);
    }
    navigateToError(statusCode) {
        this.onNavigate({ target: this, page: this.getErrorPage(statusCode), subPage: "", uri: statusCode.toString() });
    }
    registerPage(path, page) {
        return __awaiter(this, void 0, void 0, function* () {
            const parts = NavigationHelper_1.NavigationHelper.getParts(path);
            if (parts.length === 0) {
                throw Error("Can't add page to index element");
            }
            page.setPath(parts.join("/"));
            let curObj = this.pages;
            for (let i = 0; i < parts.length - 1; i++) {
                const a = parts[i];
                if (a.length === 0) {
                    continue;
                }
                let temp = curObj[a];
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
            yield page.init();
        });
    }
    registerErrorPage(statusCode, page) {
        this.errorPages[statusCode] = page;
    }
    checkLinks(links, viewPage) {
        let checkUrl = this.currentPath;
        if (viewPage && viewPage.pagePath === checkUrl) {
            checkUrl += "/" + viewPage.navHelper.defaultPage;
        }
        const long = NavigationHelper_1.NavigationHelper.getParts(checkUrl);
        for (const l of links) {
            if (!l.uri) {
                continue;
            }
            const short = NavigationHelper_1.NavigationHelper.getParts(l.uri);
            l.active = this.checkPartEqual(long, short);
        }
    }
    checkLinkCollection(links, viewPage) {
        let checkUrl = this.currentPath;
        if (viewPage && viewPage.pagePath === checkUrl) {
            checkUrl += "/" + viewPage.navHelper.defaultPage;
        }
        const long = NavigationHelper_1.NavigationHelper.getParts(checkUrl);
        for (const l of links) {
            if (!l.item.uri) {
                continue;
            }
            const short = NavigationHelper_1.NavigationHelper.getParts(l.item.uri);
            l.item.active = this.checkPartEqual(long, short);
            if (l.children) {
                this.checkLinks(l.children, viewPage);
            }
        }
    }
    refresh() {
        this.navigateTo(this.currentPath);
    }
    checkPartEqual(long, short) {
        if (short.length > long.length) {
            return false;
        }
        for (let i = 0; i < short.length; i++) {
            if (short[i] !== long[i]) {
                return false;
            }
        }
        return true;
    }
    getErrorPage(statusCode) {
        if (this.errorPages[statusCode]) {
            return this.errorPages[statusCode];
        }
        throw Error("Status page: " + statusCode + " is not defined");
    }
}
exports.NavigationManager = NavigationManager;


/***/ }),
/* 36 */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __awaiter = (this && this.__awaiter) || function (thisArg, _arguments, P, generator) {
    return new (P || (P = Promise))(function (resolve, reject) {
        function fulfilled(value) { try { step(generator.next(value)); } catch (e) { reject(e); } }
        function rejected(value) { try { step(generator["throw"](value)); } catch (e) { reject(e); } }
        function step(result) { result.done ? resolve(result.value) : new P(function (resolve) { resolve(result.value); }).then(fulfilled, rejected); }
        step((generator = generator.apply(thisArg, _arguments || [])).next());
    });
};
Object.defineProperty(exports, "__esModule", { value: true });
const Models = __webpack_require__(4);
const map_1 = __webpack_require__(6);
class TempDataProvider {
    constructor() {
        this.addLocalAssignments();
        this.addLocalCourses();
        this.addLocalCourseStudent();
        this.addLocalUsers();
    }
    getAllUser() {
        return __awaiter(this, void 0, void 0, function* () {
            return this.localUsers;
        });
    }
    getCourses() {
        return __awaiter(this, void 0, void 0, function* () {
            return this.localCourses;
        });
    }
    getCoursesStudent() {
        return __awaiter(this, void 0, void 0, function* () {
            return this.localCourseStudent;
        });
    }
    getAssignments(courseId) {
        return __awaiter(this, void 0, void 0, function* () {
            const temp = [];
            map_1.MapHelper.forEach(this.localAssignments, (a, i) => {
                if (a.courseId === courseId) {
                    temp[i] = a;
                }
            });
            return temp;
        });
    }
    tryLogin(username, password) {
        return __awaiter(this, void 0, void 0, function* () {
            const user = map_1.MapHelper.find(this.localUsers, (u) => u.email.toLocaleLowerCase() === username.toLocaleLowerCase());
            if (user && user.password === password) {
                return user;
            }
            return null;
        });
    }
    tryRemoteLogin(provider) {
        return __awaiter(this, void 0, void 0, function* () {
            let lookup = "test@testersen.no";
            if (provider === "gitlab") {
                lookup = "bob@bobsen.no";
            }
            const user = map_1.MapHelper.find(this.localUsers, (u) => u.email.toLocaleLowerCase() === lookup);
            return new Promise((resolve, reject) => {
                setTimeout(() => {
                    resolve(user);
                }, 500);
            });
        });
    }
    logout(user) {
        return __awaiter(this, void 0, void 0, function* () {
            return true;
        });
    }
    addUserToCourse(user, course) {
        return __awaiter(this, void 0, void 0, function* () {
            this.localCourseStudent.push({
                courseId: course.id,
                personId: user.id,
                state: Models.CourseStudentState.pending,
            });
            return true;
        });
    }
    createNewCourse(course) {
        return __awaiter(this, void 0, void 0, function* () {
            const courses = map_1.MapHelper.toArray(this.localCourses);
            course.id = courses.length;
            const courseData = course;
            courses.push(courseData);
            this.localCourses = map_1.mapify(courses, (ele) => ele.id);
            return true;
        });
    }
    changeUserState(link, state) {
        return __awaiter(this, void 0, void 0, function* () {
            link.state = state;
            return true;
        });
    }
    addLocalUsers() {
        this.localUsers = map_1.mapify([
            {
                id: 999,
                firstName: "Test",
                lastName: "Testersen",
                email: "test@testersen.no",
                personId: 9999,
                password: "1234",
                isAdmin: true,
            },
            {
                id: 1000,
                firstName: "Admin",
                lastName: "Admin",
                email: "admin@admin",
                personId: 1000,
                password: "1234",
                isAdmin: true,
            },
            {
                id: 1,
                firstName: "Per",
                lastName: "Pettersen",
                email: "per@pettersen.no",
                personId: 1234,
                password: "1234",
                isAdmin: false,
            },
            {
                id: 2,
                firstName: "Bob",
                lastName: "Bobsen",
                email: "bob@bobsen.no",
                personId: 1234,
                password: "1234",
                isAdmin: false,
            },
            {
                id: 3,
                firstName: "Petter",
                lastName: "Pan",
                email: "petter@pan.no",
                personId: 1234,
                password: "1234",
                isAdmin: false,
            },
        ], (ele) => ele.id);
    }
    addLocalAssignments() {
        this.localAssignments = map_1.mapify([
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
            {
                id: 5,
                courseId: 1,
                name: "Lab 2",
                start: new Date(2017, 5, 1),
                deadline: new Date(2017, 5, 25),
                end: new Date(2017, 5, 30),
            },
            {
                id: 6,
                courseId: 1,
                name: "Lab 3",
                start: new Date(2017, 5, 1),
                deadline: new Date(2017, 5, 25),
                end: new Date(2017, 5, 30),
            },
            {
                id: 7,
                courseId: 2,
                name: "Lab 1",
                start: new Date(2017, 5, 1),
                deadline: new Date(2017, 5, 25),
                end: new Date(2017, 5, 30),
            },
            {
                id: 8,
                courseId: 2,
                name: "Lab 2",
                start: new Date(2017, 5, 1),
                deadline: new Date(2017, 5, 25),
                end: new Date(2017, 5, 30),
            },
            {
                id: 9,
                courseId: 3,
                name: "Lab 1",
                start: new Date(2017, 5, 1),
                deadline: new Date(2017, 5, 25),
                end: new Date(2017, 5, 30),
            },
            {
                id: 10,
                courseId: 4,
                name: "Lab 1",
                start: new Date(2017, 5, 1),
                deadline: new Date(2017, 5, 25),
                end: new Date(2017, 5, 30),
            },
        ], (ele) => ele.id);
    }
    addLocalCourses() {
        this.localCourses = map_1.mapify([
            {
                id: 0,
                name: "Object Oriented Programming",
                tag: "DAT100",
                year: "Spring 2017",
            },
            {
                id: 1,
                name: "Algorithms and Datastructures",
                tag: "DAT200",
                year: "Spring 2017",
            },
            {
                id: 2,
                name: "Databases",
                tag: "DAT220",
                year: "Spring 2017",
            },
            {
                id: 3,
                name: "Communication Technology",
                tag: "DAT230",
                year: "Spring 2017",
            },
            {
                id: 4,
                name: "Operating Systems",
                tag: "DAT320",
                year: "Spring 2017",
            },
        ], (ele) => ele.id);
    }
    addLocalCourseStudent() {
        this.localCourseStudent = [
            { courseId: 0, personId: 999, state: 1 },
            { courseId: 1, personId: 999, state: 1 },
            { courseId: 0, personId: 1, state: 0 },
            { courseId: 0, personId: 2, state: 0 },
        ];
    }
}
exports.TempDataProvider = TempDataProvider;


/***/ }),
/* 37 */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __awaiter = (this && this.__awaiter) || function (thisArg, _arguments, P, generator) {
    return new (P || (P = Promise))(function (resolve, reject) {
        function fulfilled(value) { try { step(generator.next(value)); } catch (e) { reject(e); } }
        function rejected(value) { try { step(generator["throw"](value)); } catch (e) { reject(e); } }
        function step(result) { result.done ? resolve(result.value) : new P(function (resolve) { resolve(result.value); }).then(fulfilled, rejected); }
        step((generator = generator.apply(thisArg, _arguments || [])).next());
    });
};
Object.defineProperty(exports, "__esModule", { value: true });
const event_1 = __webpack_require__(5);
const map_1 = __webpack_require__(6);
class UserManager {
    constructor(userProvider) {
        this.onLogin = event_1.newEvent("UserManager.onLogin");
        this.onLogout = event_1.newEvent("UserManager.onLogout");
        this.userProvider = userProvider;
    }
    getCurrentUser() {
        return this.currentUser;
    }
    tryLogin(username, password) {
        return __awaiter(this, void 0, void 0, function* () {
            const result = yield this.userProvider.tryLogin(username, password);
            if (result) {
                this.currentUser = result;
                this.onLogin({ target: this, user: this.currentUser });
            }
            return result;
        });
    }
    tryRemoteLogin(provider) {
        return __awaiter(this, void 0, void 0, function* () {
            const result = yield this.userProvider.tryRemoteLogin(provider);
            if (result) {
                this.currentUser = result;
                this.onLogin({ target: this, user: this.currentUser });
            }
            return result;
        });
    }
    logout() {
        return __awaiter(this, void 0, void 0, function* () {
            if (this.currentUser) {
                yield this.userProvider.logout(this.currentUser);
                this.currentUser = null;
                this.onLogout({ target: this });
            }
        });
    }
    isAdmin(user) {
        return user.isAdmin;
    }
    isTeacher(user) {
        return __awaiter(this, void 0, void 0, function* () {
            return user.id > 100;
        });
    }
    getAllUser() {
        return __awaiter(this, void 0, void 0, function* () {
            return map_1.MapHelper.toArray(yield this.userProvider.getAllUser());
        });
    }
    getUsers(ids) {
        return __awaiter(this, void 0, void 0, function* () {
            const returnUsers = [];
            const allUsers = yield this.userProvider.getAllUser();
            ids.forEach((ele) => {
                const temp = allUsers[ele];
                if (temp) {
                    returnUsers.push(temp);
                }
            });
            return returnUsers;
        });
    }
    getUser(id) {
        return __awaiter(this, void 0, void 0, function* () {
            throw new Error("Not implemented error");
        });
    }
}
exports.UserManager = UserManager;


/***/ }),
/* 38 */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __awaiter = (this && this.__awaiter) || function (thisArg, _arguments, P, generator) {
    return new (P || (P = Promise))(function (resolve, reject) {
        function fulfilled(value) { try { step(generator.next(value)); } catch (e) { reject(e); } }
        function rejected(value) { try { step(generator["throw"](value)); } catch (e) { reject(e); } }
        function step(result) { result.done ? resolve(result.value) : new P(function (resolve) { resolve(result.value); }).then(fulfilled, rejected); }
        step((generator = generator.apply(thisArg, _arguments || [])).next());
    });
};
Object.defineProperty(exports, "__esModule", { value: true });
const React = __webpack_require__(0);
const ViewPage_1 = __webpack_require__(2);
class ErrorPage extends ViewPage_1.ViewPage {
    constructor() {
        super();
        this.pages = {};
        this.navHelper.defaultPage = "404";
        this.navHelper.registerFunction("404", (navInfo) => __awaiter(this, void 0, void 0, function* () {
            return React.createElement("div", null,
                React.createElement("h1", null, "404 Page not found"),
                React.createElement("p", null, "The page you where looking for does not exist"));
        }));
    }
    renderContent(page) {
        return __awaiter(this, void 0, void 0, function* () {
            let content = yield this.navHelper.navigateTo(page);
            if (!content) {
                content = yield this.navHelper.navigateTo("404");
            }
            if (!content) {
                throw new Error("There is a problem with the navigation");
            }
            return content;
        });
    }
}
exports.ErrorPage = ErrorPage;


/***/ }),
/* 39 */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __awaiter = (this && this.__awaiter) || function (thisArg, _arguments, P, generator) {
    return new (P || (P = Promise))(function (resolve, reject) {
        function fulfilled(value) { try { step(generator.next(value)); } catch (e) { reject(e); } }
        function rejected(value) { try { step(generator["throw"](value)); } catch (e) { reject(e); } }
        function step(result) { result.done ? resolve(result.value) : new P(function (resolve) { resolve(result.value); }).then(fulfilled, rejected); }
        step((generator = generator.apply(thisArg, _arguments || [])).next());
    });
};
Object.defineProperty(exports, "__esModule", { value: true });
const React = __webpack_require__(0);
const ViewPage_1 = __webpack_require__(2);
const HelpView_1 = __webpack_require__(40);
class HelpPage extends ViewPage_1.ViewPage {
    constructor(navMan) {
        super();
        this.pages = {};
        this.navMan = navMan;
        this.navHelper.defaultPage = "help";
        this.navHelper.registerFunction("help", this.help);
    }
    help(info) {
        return __awaiter(this, void 0, void 0, function* () {
            return React.createElement(HelpView_1.HelpView, null);
        });
    }
}
exports.HelpPage = HelpPage;


/***/ }),
/* 40 */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
const React = __webpack_require__(0);
const components_1 = __webpack_require__(1);
class HelpView extends React.Component {
    render() {
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
    }
}
exports.HelpView = HelpView;


/***/ }),
/* 41 */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __awaiter = (this && this.__awaiter) || function (thisArg, _arguments, P, generator) {
    return new (P || (P = Promise))(function (resolve, reject) {
        function fulfilled(value) { try { step(generator.next(value)); } catch (e) { reject(e); } }
        function rejected(value) { try { step(generator["throw"](value)); } catch (e) { reject(e); } }
        function step(result) { result.done ? resolve(result.value) : new P(function (resolve) { resolve(result.value); }).then(fulfilled, rejected); }
        step((generator = generator.apply(thisArg, _arguments || [])).next());
    });
};
Object.defineProperty(exports, "__esModule", { value: true });
const React = __webpack_require__(0);
const ViewPage_1 = __webpack_require__(2);
class HomePage extends ViewPage_1.ViewPage {
    constructor() {
        super();
    }
    renderContent(page) {
        return __awaiter(this, void 0, void 0, function* () {
            return React.createElement("h1", null, "Welcome to autograder");
        });
    }
}
exports.HomePage = HomePage;


/***/ }),
/* 42 */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __awaiter = (this && this.__awaiter) || function (thisArg, _arguments, P, generator) {
    return new (P || (P = Promise))(function (resolve, reject) {
        function fulfilled(value) { try { step(generator.next(value)); } catch (e) { reject(e); } }
        function rejected(value) { try { step(generator["throw"](value)); } catch (e) { reject(e); } }
        function step(result) { result.done ? resolve(result.value) : new P(function (resolve) { resolve(result.value); }).then(fulfilled, rejected); }
        step((generator = generator.apply(thisArg, _arguments || [])).next());
    });
};
Object.defineProperty(exports, "__esModule", { value: true });
const React = __webpack_require__(0);
const components_1 = __webpack_require__(1);
const ViewPage_1 = __webpack_require__(2);
const HelloView_1 = __webpack_require__(12);
const UserView_1 = __webpack_require__(7);
const helper_1 = __webpack_require__(8);
const CollapsableNavMenu_1 = __webpack_require__(13);
const EnrollmentView_1 = __webpack_require__(43);
class StudentPage extends ViewPage_1.ViewPage {
    constructor(users, navMan, courseMan) {
        super();
        this.selectedCourse = null;
        this.selectedAssignment = null;
        this.courses = [];
        this.foundId = -1;
        this.navMan = navMan;
        this.userMan = users;
        this.courseMan = courseMan;
        this.navHelper.defaultPage = "index";
        this.navHelper.onPreNavigation.addEventListener((e) => this.setupNavData(e));
        this.navHelper.registerFunction("index", this.index);
        this.navHelper.registerFunction("course/{courseid}", this.course);
        this.navHelper.registerFunction("course/{courseid}/lab/{labid}", this.courseWithLab);
        this.navHelper.registerFunction("course/{coruseid}/{page}", this.courseMissing);
        this.navHelper.registerFunction("enroll", this.enroll);
        this.navHelper.registerFunction("user", this.getUsers);
        this.navHelper.registerFunction("hello", (navInfo) => Promise.resolve(React.createElement(HelloView_1.HelloView, null)));
    }
    getUsers(navInfo) {
        return __awaiter(this, void 0, void 0, function* () {
            yield this.setupData();
            return React.createElement(UserView_1.UserView, { users: yield this.userMan.getAllUser() });
        });
    }
    index(navInfo) {
        return __awaiter(this, void 0, void 0, function* () {
            yield this.setupData();
            const courseOverview = yield this.getCoursesWithAssignments();
            return (React.createElement(components_1.CoursesOverview, { course_overview: courseOverview, navMan: this.navMan }));
        });
    }
    enroll(navInfo) {
        return __awaiter(this, void 0, void 0, function* () {
            yield this.setupData();
            return React.createElement("div", null,
                React.createElement("h1", null, "Enrollment page"),
                React.createElement(EnrollmentView_1.EnrollmentView, { courses: yield this.courseMan.getCourses(), studentCourses: yield this.getRelations(), curUser: this.userMan.getCurrentUser(), onEnrollmentClick: (user, course) => {
                        this.courseMan.addUserToCourse(user, course);
                        this.navMan.refresh();
                    } }));
        });
    }
    course(navInfo) {
        return __awaiter(this, void 0, void 0, function* () {
            yield this.setupData();
            this.selectCourse(navInfo.params.courseid);
            if (this.selectedCourse) {
                const courseAndLabs = yield this.getLabs();
                if (courseAndLabs) {
                    return (React.createElement(components_1.SingleCourseOverview, { courseAndLabs: courseAndLabs }));
                }
            }
            return React.createElement("h1", null, "404 not found");
        });
    }
    courseWithLab(navInfo) {
        return __awaiter(this, void 0, void 0, function* () {
            yield this.setupData();
            this.selectCourse(navInfo.params.courseid);
            if (this.selectedCourse) {
                yield this.selectAssignment(navInfo.params.labid);
                if (this.selectedAssignment) {
                    return React.createElement(components_1.StudentLab, { course: this.selectedCourse, assignment: this.selectedAssignment });
                }
            }
            return React.createElement("div", null, "404 not found");
        });
    }
    courseMissing(navInfo) {
        return __awaiter(this, void 0, void 0, function* () {
            return React.createElement("div", null,
                "The page ",
                navInfo.params.page,
                " is not yet implemented");
        });
    }
    renderMenu(key) {
        return __awaiter(this, void 0, void 0, function* () {
            if (key === 0) {
                const coursesLinks = yield helper_1.ArrayHelper.mapAsync(this.courses, (course, i) => __awaiter(this, void 0, void 0, function* () {
                    const allLinks = [];
                    allLinks.push({ name: "Labs" });
                    const labs = yield this.getLabsfor(course);
                    allLinks.push(...labs.map((lab, ind) => {
                        return { name: lab.name, uri: this.pagePath + "/course/" + course.id + "/lab/" + lab.id };
                    }));
                    allLinks.push({ name: "Group Labs" });
                    allLinks.push({ name: "Settings" });
                    allLinks.push({ name: "Members", uri: this.pagePath + "/course/" + course.id + "/members" });
                    allLinks.push({ name: "Coruse Info", uri: this.pagePath + "/course/" + course.id + "/info" });
                    return {
                        item: { name: course.tag, uri: this.pagePath + "/course/" + course.id },
                        children: allLinks,
                    };
                }));
                const settings = [
                    { name: "Join course", uri: this.pagePath + "/enroll" },
                ];
                this.navMan.checkLinkCollection(coursesLinks, this);
                this.navMan.checkLinks(settings, this);
                return [
                    React.createElement("h4", { key: 0 }, "Courses"),
                    React.createElement(CollapsableNavMenu_1.CollapsableNavMenu, { key: 1, links: coursesLinks, onClick: (link) => this.handleClick(link) }),
                    React.createElement("h4", { key: 2 }, "Settings"),
                    React.createElement(components_1.NavMenu, { key: 3, links: settings, onClick: (link) => this.handleClick(link) }),
                ];
            }
            return [];
        });
    }
    setupNavData(data) {
        return __awaiter(this, void 0, void 0, function* () {
            yield this.setupData();
        });
    }
    setupData() {
        return __awaiter(this, void 0, void 0, function* () {
            this.courses = yield this.getCourses();
        });
    }
    selectCourse(courseId) {
        this.selectedCourse = null;
        const course = parseInt(courseId, 10);
        if (!isNaN(course)) {
            this.selectedCourse = helper_1.ArrayHelper.find(this.courses, (e, i) => {
                if (e.id === course) {
                    this.foundId = i;
                    return true;
                }
                return false;
            });
        }
    }
    selectAssignment(labIdString) {
        return __awaiter(this, void 0, void 0, function* () {
            this.selectedAssignment = null;
            const labId = parseInt(labIdString, 10);
            if (this.selectedCourse && !isNaN(labId)) {
                const lab = yield this.courseMan.getAssignment(this.selectedCourse, labId);
                if (lab) {
                    this.selectedAssignment = lab;
                }
            }
        });
    }
    handleClick(link) {
        if (link.uri) {
            this.navMan.navigateTo(link.uri);
        }
    }
    getRelations() {
        return __awaiter(this, void 0, void 0, function* () {
            const curUsr = this.userMan.getCurrentUser();
            if (curUsr) {
                return this.courseMan.getRelationsFor(curUsr);
            }
            return [];
        });
    }
    getCourses() {
        return __awaiter(this, void 0, void 0, function* () {
            const curUsr = this.userMan.getCurrentUser();
            if (curUsr) {
                return this.courseMan.getCoursesFor(curUsr, 1);
            }
            return [];
        });
    }
    getLabsfor(course) {
        return __awaiter(this, void 0, void 0, function* () {
            return this.courseMan.getAssignments(course);
        });
    }
    getLabs() {
        return __awaiter(this, void 0, void 0, function* () {
            const curUsr = this.userMan.getCurrentUser();
            if (curUsr && !this.selectedCourse) {
                this.selectedCourse = (yield this.courseMan.getCoursesFor(curUsr))[0];
            }
            if (this.selectedCourse) {
                const labs = yield this.courseMan.getAssignments(this.selectedCourse);
                return { course: this.selectedCourse, labs };
            }
            return null;
        });
    }
    getCoursesWithAssignments() {
        return __awaiter(this, void 0, void 0, function* () {
            const courseLabs = [];
            if (this.courses.length === 0) {
                this.courses = yield this.getCourses();
            }
            if (this.courses.length > 0) {
                for (const crs of this.courses) {
                    const labs = yield this.courseMan.getAssignments(crs);
                    const cl = { course: crs, labs };
                    courseLabs.push(cl);
                }
                return courseLabs;
            }
            return [];
        });
    }
}
exports.StudentPage = StudentPage;


/***/ }),
/* 43 */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
const React = __webpack_require__(0);
const components_1 = __webpack_require__(1);
const models_1 = __webpack_require__(4);
const helper_1 = __webpack_require__(8);
class EnrollmentView extends React.Component {
    render() {
        return React.createElement(components_1.DynamicTable, { data: this.props.courses, header: ["Course tag", "Course Name", "Action"], selector: (course) => this.createEnrollmentRow(this.props.studentCourses, course) });
    }
    createEnrollmentRow(studentCourses, course) {
        const base = [course.tag, course.name];
        const curUser = this.props.curUser;
        if (!curUser) {
            return base;
        }
        const temp = helper_1.ArrayHelper.find(studentCourses, (a) => a.courseId === course.id);
        if (temp) {
            if (temp.state === models_1.CourseStudentState.accepted) {
                base.push("Enrolled");
            }
            else if (temp.state === models_1.CourseStudentState.pending) {
                base.push("Pending");
            }
            else {
                base.push(React.createElement("div", null,
                    React.createElement("button", { onClick: () => { this.props.onEnrollmentClick(curUser, course); }, className: "btn btn-primary" }, "Enroll"),
                    React.createElement("span", { style: { padding: "7px", verticalAlign: "middle" }, className: "bg-danger" }, "Rejected")));
            }
        }
        else {
            base.push(React.createElement("button", { onClick: () => { this.props.onEnrollmentClick(curUser, course); }, className: "btn btn-primary" }, "Enroll"));
        }
        return base;
    }
}
exports.EnrollmentView = EnrollmentView;


/***/ }),
/* 44 */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __awaiter = (this && this.__awaiter) || function (thisArg, _arguments, P, generator) {
    return new (P || (P = Promise))(function (resolve, reject) {
        function fulfilled(value) { try { step(generator.next(value)); } catch (e) { reject(e); } }
        function rejected(value) { try { step(generator["throw"](value)); } catch (e) { reject(e); } }
        function step(result) { result.done ? resolve(result.value) : new P(function (resolve) { resolve(result.value); }).then(fulfilled, rejected); }
        step((generator = generator.apply(thisArg, _arguments || [])).next());
    });
};
Object.defineProperty(exports, "__esModule", { value: true });
const React = __webpack_require__(0);
const components_1 = __webpack_require__(1);
const ViewPage_1 = __webpack_require__(2);
const HelloView_1 = __webpack_require__(12);
const UserView_1 = __webpack_require__(7);
const CollapsableNavMenu_1 = __webpack_require__(13);
const models_1 = __webpack_require__(4);
const helper_1 = __webpack_require__(8);
class TeacherPage extends ViewPage_1.ViewPage {
    constructor(userMan, navMan, courseMan) {
        super();
        this.courses = [];
        this.pages = {};
        this.navMan = navMan;
        this.userMan = userMan;
        this.courseMan = courseMan;
        this.navHelper.defaultPage = "course";
        this.navHelper.registerFunction("course/{course}", this.course);
        this.navHelper.registerFunction("course/{course}/members", this.courseUsers);
        this.navHelper.registerFunction("course/{course}/results", this.results);
        this.navHelper.registerFunction("course/{course}/{page}", this.course);
        this.navHelper.registerFunction("user", (navInfo) => __awaiter(this, void 0, void 0, function* () {
            return React.createElement(UserView_1.UserView, { users: yield userMan.getAllUser() });
        }));
        this.navHelper.registerFunction("user", (navInfo) => __awaiter(this, void 0, void 0, function* () {
            return React.createElement(HelloView_1.HelloView, null);
        }));
    }
    init() {
        return __awaiter(this, void 0, void 0, function* () {
            this.courses = yield this.getCourses();
            this.navHelper.defaultPage = "course/" + (this.courses.length > 0 ? this.courses[0].id.toString() : "");
        });
    }
    course(info) {
        return __awaiter(this, void 0, void 0, function* () {
            this.courses = yield this.getCourses();
            const courseId = parseInt(info.params.course, 10);
            const course = yield this.courseMan.getCourse(courseId);
            if (course) {
                if (info.params.page) {
                    return React.createElement("h3", null,
                        "You are know on page ",
                        info.params.page.toUpperCase(),
                        " in course ",
                        info.params.course);
                }
                return React.createElement("h1", null,
                    "Teacher Course ",
                    info.params.course);
            }
            return React.createElement("div", null, "404 Page not found");
        });
    }
    results(info) {
        return __awaiter(this, void 0, void 0, function* () {
            const courseId = parseInt(info.params.course, 10);
            const course = yield this.courseMan.getCourse(courseId);
            if (course) {
                const courseStds = yield this.courseMan.getUserIdsForCourse(course, models_1.CourseStudentState.accepted);
                const students = yield this.userMan.getUsers(courseStds.map((e) => e.personId));
                const labs = yield this.courseMan.getAssignments(courseId);
                return React.createElement(components_1.Results, { course: course, students: students, labs: labs });
            }
            return React.createElement("div", null, "404 Page not found");
        });
    }
    courseUsers(info) {
        return __awaiter(this, void 0, void 0, function* () {
            this.courses = yield this.getCourses();
            const courseId = parseInt(info.params.course, 10);
            const course = yield this.courseMan.getCourse(courseId);
            if (course) {
                const userIds = yield this.courseMan.getUserIdsForCourse(course);
                const users = yield this.userMan.getUsers(userIds.map((e) => e.personId));
                const all = helper_1.ArrayHelper.join(userIds, users, (e1, e2) => e1.personId === e2.id);
                const acceptedUsers = [];
                const pendingUsers = [];
                all.forEach((ele, id) => {
                    switch (ele.ele1.state) {
                        case models_1.CourseStudentState.accepted:
                            acceptedUsers.push(ele.ele2);
                            break;
                        case models_1.CourseStudentState.pending:
                            pendingUsers.push(ele);
                            break;
                    }
                });
                return React.createElement("div", null,
                    React.createElement("h3", null,
                        "Users for ",
                        course.name,
                        " (",
                        course.tag,
                        ")"),
                    React.createElement(UserView_1.UserView, { users: acceptedUsers }),
                    React.createElement("h3", null,
                        "Pending users for ",
                        course.name,
                        " (",
                        course.tag,
                        ")"),
                    this.createPendingTable(pendingUsers));
            }
            return React.createElement("div", null, "404 Page not found");
        });
    }
    createPendingTable(pendingUsers) {
        return React.createElement(components_1.DynamicTable, { data: pendingUsers, header: ["ID", "First name", "Last name", "Email", "StudenID", "Action"], selector: (ele) => [
                ele.ele2.id.toString(),
                ele.ele2.firstName,
                ele.ele2.lastName,
                ele.ele2.email,
                ele.ele2.personId.toString(),
                React.createElement("span", null,
                    React.createElement("button", { onClick: (e) => {
                            this.courseMan.changeUserState(ele.ele1, models_1.CourseStudentState.accepted);
                            this.navMan.refresh();
                        }, className: "btn btn-primary" }, "Accept"),
                    React.createElement("button", { onClick: (e) => {
                            this.courseMan.changeUserState(ele.ele1, models_1.CourseStudentState.rejected);
                            this.navMan.refresh();
                        }, className: "btn btn-danger" }, "Reject")),
            ] });
    }
    generateCollectionFor(link) {
        return {
            item: link,
            children: [
                { name: "Results", uri: link.uri + "/results" },
                { name: "Groups", uri: link.uri + "/groups" },
                { name: "Members", uri: link.uri + "/members" },
                { name: "Settings", uri: link.uri + "/settings" },
                { name: "Course Info", uri: link.uri + "/courseinfo" },
            ],
        };
    }
    renderMenu(menu) {
        return __awaiter(this, void 0, void 0, function* () {
            const curUser = this.userMan.getCurrentUser();
            if (curUser && this.isTeacher(curUser)) {
                if (menu === 0) {
                    const courses = yield this.courseMan.getCoursesFor(curUser);
                    const labLinks = [];
                    courses.forEach((e) => {
                        labLinks.push(this.generateCollectionFor({
                            name: e.tag,
                            uri: this.pagePath + "/course/" + e.id,
                        }));
                    });
                    const settings = [];
                    this.navMan.checkLinkCollection(labLinks, this);
                    this.navMan.checkLinks(settings, this);
                    return [
                        React.createElement("h4", { key: 0 }, "Courses"),
                        React.createElement(CollapsableNavMenu_1.CollapsableNavMenu, { key: 1, links: labLinks, onClick: (link) => this.handleClick(link) }),
                        React.createElement("h4", { key: 2 }, "Settings"),
                        React.createElement(components_1.NavMenu, { key: 3, links: settings, onClick: (link) => this.handleClick(link) }),
                    ];
                }
            }
            return [];
        });
    }
    renderContent(page) {
        const _super = name => super[name];
        return __awaiter(this, void 0, void 0, function* () {
            const curUser = this.userMan.getCurrentUser();
            if (!curUser) {
                return React.createElement("h1", null, "You are not logged in");
            }
            else if (this.isTeacher(curUser)) {
                return yield _super("renderContent").call(this, page);
            }
            return React.createElement("h1", null, "404 page not found");
        });
    }
    handleClick(link) {
        if (link.uri) {
            this.navMan.navigateTo(link.uri);
        }
    }
    getCourses() {
        return __awaiter(this, void 0, void 0, function* () {
            const curUsr = this.userMan.getCurrentUser();
            if (curUsr) {
                return yield this.courseMan.getCoursesFor(curUsr);
            }
            return [];
        });
    }
    isTeacher(curUser) {
        return __awaiter(this, void 0, void 0, function* () {
            return this.userMan.isTeacher(curUser);
        });
    }
}
exports.TeacherPage = TeacherPage;


/***/ }),
/* 45 */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __awaiter = (this && this.__awaiter) || function (thisArg, _arguments, P, generator) {
    return new (P || (P = Promise))(function (resolve, reject) {
        function fulfilled(value) { try { step(generator.next(value)); } catch (e) { reject(e); } }
        function rejected(value) { try { step(generator["throw"](value)); } catch (e) { reject(e); } }
        function step(result) { result.done ? resolve(result.value) : new P(function (resolve) { resolve(result.value); }).then(fulfilled, rejected); }
        step((generator = generator.apply(thisArg, _arguments || [])).next());
    });
};
Object.defineProperty(exports, "__esModule", { value: true });
const React = __webpack_require__(0);
const components_1 = __webpack_require__(1);
const ViewPage_1 = __webpack_require__(2);
const CourseView_1 = __webpack_require__(50);
const UserView_1 = __webpack_require__(7);
class AdminPage extends ViewPage_1.ViewPage {
    constructor(navMan, userMan, courseMan) {
        super();
        this.navMan = navMan;
        this.userMan = userMan;
        this.courseMan = courseMan;
        this.flashMessages = null;
        this.navHelper.defaultPage = "users";
        this.navHelper.registerFunction("users", this.users);
        this.navHelper.registerFunction("courses", this.courses);
        this.navHelper.registerFunction("labs", this.labs);
        this.navHelper.registerFunction("courses/new", this.newCourse);
    }
    users(info) {
        return __awaiter(this, void 0, void 0, function* () {
            const allUsers = yield this.userMan.getAllUser();
            return React.createElement("div", null,
                React.createElement("h1", null, "All Users"),
                React.createElement(UserView_1.UserView, { users: allUsers, addSearchOption: true }));
        });
    }
    courses(info) {
        return __awaiter(this, void 0, void 0, function* () {
            const allCourses = yield this.courseMan.getCourses();
            return React.createElement("div", null,
                React.createElement(components_1.Button, { className: "btn btn-primary pull-right", text: "+Create New", onClick: () => this.handleNewCourse() }),
                React.createElement("h1", null, "All Courses"),
                React.createElement(CourseView_1.CourseView, { courses: allCourses }));
        });
    }
    labs(info) {
        return __awaiter(this, void 0, void 0, function* () {
            const allCourses = yield this.courseMan.getCourses();
            const tables = [];
            for (let i = 0; i < allCourses.length; i++) {
                const e = allCourses[i];
                const labs = yield this.courseMan.getAssignments(e);
                tables.push(React.createElement("div", { key: i },
                    React.createElement("h3", null,
                        "Labs for ",
                        e.name,
                        " (",
                        e.tag,
                        ")"),
                    React.createElement(components_1.DynamicTable, { header: ["ID", "Name", "Start", "Deadline", "End"], data: labs, selector: (lab) => [
                            lab.id.toString(),
                            lab.name,
                            lab.start.toDateString(),
                            lab.deadline.toDateString(),
                            lab.end.toDateString(),
                        ] })));
            }
            return React.createElement("div", null, tables);
        });
    }
    newCourse(info) {
        return __awaiter(this, void 0, void 0, function* () {
            let flashHolder = React.createElement("div", null);
            if (this.flashMessages) {
                const errors = [];
                for (const fm of this.flashMessages) {
                    errors.push(React.createElement("li", null, fm));
                }
                flashHolder = React.createElement("div", { className: "alert alert-danger" },
                    React.createElement("h4", null,
                        errors.length,
                        " errors prohibited Course from being saved: "),
                    React.createElement("ul", null, errors));
            }
            return (React.createElement("div", null,
                React.createElement("h1", null, "Create New Course"),
                flashHolder,
                React.createElement(components_1.CourseForm, { className: "form-horizontal", onSubmit: (formData, errors) => this.createNewCourse(formData, errors) })));
        });
    }
    renderMenu(index) {
        return __awaiter(this, void 0, void 0, function* () {
            if (index === 0) {
                const links = [
                    { name: "All Users", uri: this.pagePath + "/users" },
                    { name: "All Courses", uri: this.pagePath + "/courses" },
                    { name: "All Labs", uri: this.pagePath + "/labs" },
                ];
                this.navMan.checkLinks(links, this);
                return [
                    React.createElement("h4", { key: 0 }, "Admin Menu"),
                    React.createElement(components_1.NavMenu, { key: 1, links: links, onClick: (e) => {
                            if (e.uri) {
                                this.navMan.navigateTo(e.uri);
                            }
                        } }),
                ];
            }
            return [];
        });
    }
    handleNewCourse(flashMessage) {
        if (flashMessage) {
            this.flashMessages = flashMessage;
        }
        this.navMan.navigateTo(this.pagePath + "/courses/new");
    }
    createNewCourse(fd, errors) {
        if (errors.length === 0) {
            this.courseMan.createNewCourse(fd);
            this.flashMessages = null;
            this.navMan.navigateTo(this.pagePath + "/courses");
        }
        else {
            this.handleNewCourse(errors);
        }
    }
}
exports.AdminPage = AdminPage;


/***/ }),
/* 46 */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
const React = __webpack_require__(0);
const NavMenu_1 = __webpack_require__(10);
class NavBarLogin extends React.Component {
    constructor() {
        super();
        this.state = {
            loginOpen: false,
        };
    }
    render() {
        if (this.props.user) {
            return React.createElement("div", { className: "navbar-right" },
                React.createElement("button", { className: "btn btn-primary navbar-btn", onClick: () => { this.handleClick({ name: "Logout", uri: "app/login/logout" }); } }, "Log out"));
        }
        let links = this.props.links;
        if (!links) {
            links = [
                { name: "Missing links" },
            ];
        }
        let isHidden = "hidden";
        if (this.state.loginOpen) {
            isHidden = "";
        }
        return React.createElement("div", { className: "navbar-right" },
            React.createElement("button", { onClick: () => this.toggleMenu(), className: "btn btn-primary navbar-btn" }, "Login"),
            React.createElement("div", { className: "nav-box " + isHidden },
                React.createElement(NavMenu_1.NavMenu, { links: links, onClick: (link) => this.handleClick(link) })));
    }
    toggleMenu() {
        this.setState({ loginOpen: !this.state.loginOpen });
    }
    handleClick(link) {
        this.setState({ loginOpen: false });
        if (this.props.onClick) {
            this.props.onClick(link);
        }
    }
}
exports.NavBarLogin = NavBarLogin;


/***/ }),
/* 47 */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
const React = __webpack_require__(0);
const NavigationHelper_1 = __webpack_require__(3);
class NavBarMenu extends React.Component {
    render() {
        const items = this.props.links.map((link, i) => {
            let active = "";
            if (link.active) {
                active = "active";
            }
            return React.createElement("li", { className: active, key: i },
                React.createElement("a", { href: "/" + link.uri, onClick: (e) => {
                        NavigationHelper_1.NavigationHelper.handleClick(e, () => {
                            this.handleClick(link);
                        });
                    } }, link.name));
        });
        return React.createElement("ul", { className: "nav navbar-nav" }, items);
    }
    handleClick(link) {
        if (this.props.onClick) {
            this.props.onClick(link);
        }
    }
}
exports.NavBarMenu = NavBarMenu;


/***/ }),
/* 48 */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __awaiter = (this && this.__awaiter) || function (thisArg, _arguments, P, generator) {
    return new (P || (P = Promise))(function (resolve, reject) {
        function fulfilled(value) { try { step(generator.next(value)); } catch (e) { reject(e); } }
        function rejected(value) { try { step(generator["throw"](value)); } catch (e) { reject(e); } }
        function step(result) { result.done ? resolve(result.value) : new P(function (resolve) { resolve(result.value); }).then(fulfilled, rejected); }
        step((generator = generator.apply(thisArg, _arguments || [])).next());
    });
};
Object.defineProperty(exports, "__esModule", { value: true });
const React = __webpack_require__(0);
const ViewPage_1 = __webpack_require__(2);
class LoginPage extends ViewPage_1.ViewPage {
    constructor(navMan, userMan) {
        super();
        this.navMan = navMan;
        this.userMan = userMan;
        this.navHelper.defaultPage = "index";
        this.navHelper.registerFunction("index", this.index);
        this.navHelper.registerFunction("login/{provider}", this.login);
        this.navHelper.registerFunction("logout", this.logout);
    }
    index(info) {
        return __awaiter(this, void 0, void 0, function* () {
            return React.createElement("div", null, "Quickly hide, you should not be here! Someone is going to get mad...");
        });
    }
    login(info) {
        return __awaiter(this, void 0, void 0, function* () {
            const temp = this.userMan.tryRemoteLogin(info.params.provider);
            temp.then((result) => {
                if (result) {
                    console.log("Sucessful login of: ", result);
                    this.navMan.navigateToDefault();
                }
                else {
                    console.log("Failed");
                }
            });
            return Promise.resolve(React.createElement("div", null, "Logging in please wait"));
        });
    }
    logout(info) {
        return __awaiter(this, void 0, void 0, function* () {
            this.userMan.logout();
            return React.createElement("div", null, "Logged out");
        });
    }
}
exports.LoginPage = LoginPage;


/***/ }),
/* 49 */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

var __awaiter = (this && this.__awaiter) || function (thisArg, _arguments, P, generator) {
    return new (P || (P = Promise))(function (resolve, reject) {
        function fulfilled(value) { try { step(generator.next(value)); } catch (e) { reject(e); } }
        function rejected(value) { try { step(generator["throw"](value)); } catch (e) { reject(e); } }
        function step(result) { result.done ? resolve(result.value) : new P(function (resolve) { resolve(result.value); }).then(fulfilled, rejected); }
        step((generator = generator.apply(thisArg, _arguments || [])).next());
    });
};
Object.defineProperty(exports, "__esModule", { value: true });
function request(url) {
    return __awaiter(this, void 0, void 0, function* () {
        const req = new XMLHttpRequest();
        return new Promise((resolve, reject) => {
            req.onreadystatechange = () => {
                if (req.readyState === 4) {
                    if (req.status === 200) {
                        console.log(req);
                        resolve(req.responseText);
                    }
                    else {
                        reject(req);
                    }
                }
            };
            req.open("GET", url, true);
            req.send();
        });
    });
}
class ServerProvider {
    tryLogin(username, password) {
        return __awaiter(this, void 0, void 0, function* () {
            throw new Error("Method not implemented.");
        });
    }
    logout(user) {
        return __awaiter(this, void 0, void 0, function* () {
            throw new Error("Method not implemented.");
        });
    }
    getAllUser() {
        return __awaiter(this, void 0, void 0, function* () {
            throw new Error("Method not implemented.");
        });
    }
    tryRemoteLogin(provider) {
        return __awaiter(this, void 0, void 0, function* () {
            let requestString = null;
            switch (provider) {
                case "github":
                    requestString = "/auth/github";
                    break;
                case "gitlab":
                    requestString = "/auth/gitlab";
                    break;
            }
            if (requestString) {
                window.location.assign(requestString);
                return null;
            }
            else {
                return null;
            }
        });
    }
}
exports.ServerProvider = ServerProvider;


/***/ }),
/* 50 */
/***/ (function(module, exports, __webpack_require__) {

"use strict";

Object.defineProperty(exports, "__esModule", { value: true });
const React = __webpack_require__(0);
const components_1 = __webpack_require__(1);
class CourseView extends React.Component {
    constructor(props) {
        super(props);
        this.state = {
            courses: this.props.courses,
        };
    }
    render() {
        const searchIcon = React.createElement("span", { className: "input-group-addon" },
            React.createElement("i", { className: "glyphicon glyphicon-search" }));
        return (React.createElement("div", null,
            React.createElement(components_1.Search, { className: "input-group", addonBefore: searchIcon, placeholder: "Search for courses", onChange: (query) => this.handleOnchange(query) }),
            React.createElement(components_1.DynamicTable, { header: ["ID", "Name", "Tag", "Year/Semester"], data: this.state.courses, selector: (e) => [e.id.toString(), e.name, e.tag, e.year] })));
    }
    handleOnchange(query) {
        query = query.toLowerCase();
        const filteredData = [];
        this.props.courses.forEach((course) => {
            if (course.name.toLowerCase().indexOf(query) !== -1
                || course.tag.toLowerCase().indexOf(query) !== -1
                || course.year.toLowerCase().indexOf(query) !== -1) {
                filteredData.push(course);
            }
        });
        this.setState({
            courses: filteredData,
        });
    }
}
exports.CourseView = CourseView;


/***/ })
/******/ ]);
//# sourceMappingURL=bundle.js.map
var DummyUserProvider = (function () {
    function DummyUserProvider() {
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
    DummyUserProvider.prototype.getAllUser = function () {
        return this.localData;
    };
    DummyUserProvider.prototype.tryLogin = function (username, password) {
        for (var _i = 0, _a = this.localData; _i < _a.length; _i++) {
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
    return DummyUserProvider;
}());
var AssignmentManager = (function () {
    function AssignmentManager() {
    }
    return AssignmentManager;
}());
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
var NavigationManager = (function () {
    function NavigationManager() {
        this.pages = {};
        this.errorPages = [];
        this.onNavigate = newEvent("NavigationManager.onNavigate");
        this.currentPath = "";
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
        this.currentPath = path;
    };
    NavigationManager.prototype.navigateTo = function (path) {
        var parts = this.getParts(path);
        var curPage = this.pages;
        for (var i = 0; i < parts.length; i++) {
            var a = parts[i];
            if (isViewPage(curPage)) {
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
        if (isViewPage(curPage)) {
            this.onNavigate({ target: this, page: curPage, uri: path, subPage: "" });
            return;
        }
        else {
            this.onNavigate({ target: this, page: this.getErrorPage(404), subPage: "", uri: path });
        }
    };
    NavigationManager.prototype.navigateToDefault = function () {
        this.navigateTo(this.currentPath);
    };
    NavigationManager.prototype.navigateToError = function (statusCode) {
        this.onNavigate({ target: this, page: this.getErrorPage(statusCode), subPage: "", uri: statusCode.toString() });
    };
    NavigationManager.prototype.registerPage = function (path, page) {
        var parts = this.getParts(path);
        if (parts.length === 0) {
            throw Error("Can't add page to index element");
        }
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
            else if (!isViewPage(temp)) {
                temp = curObj[a];
            }
            if (isViewPage(temp)) {
                throw Error("Can't assign a IPageContainer to a ViewPage");
            }
            curObj = temp;
        }
        curObj[parts[parts.length - 1]] = page;
    };
    NavigationManager.prototype.registerErrorPage = function (statusCode, page) {
        this.errorPages[statusCode] = page;
    };
    return NavigationManager;
}());

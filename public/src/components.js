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
            React.createElement("a", { className: "navbar-brand", href: "#" }, this.props.brandName));
    };
    return NavHeaderBar;
}(React.Component));
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
                React.createElement("a", { href: "#", onClick: function () { return _this.handleClick(v); } }, v.name));
        });
        return React.createElement("nav", { className: this.renderNavBarClass() },
            React.createElement("div", { className: this.renderIsFluid() },
                React.createElement(NavHeaderBar, { id: this.props.id, brandName: this.props.brandName }),
                React.createElement("div", { className: "collapse navbar-collapse", id: this.props.id },
                    React.createElement("ul", { className: "nav navbar-nav" }, items))));
    };
    return NavBar;
}(React.Component));
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
                React.createElement("a", { href: "#", onClick: function () { if (_this.props.onClick)
                        _this.props.onClick(v); } }, v.name));
        });
        return React.createElement("ul", { className: "nav nav-pills nav-stacked" }, items);
    };
    return NavMenu;
}(React.Component));
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
function Row(props) {
    return React.createElement("div", { className: "row " + props.className ? props.className : "" }, props.children);
}

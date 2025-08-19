import React from "react";
import { Link, useNavigate } from "react-router-dom";
const NavBarLink = ({ link: { text, to, icons, jsx } }) => {
    const navigate = useNavigate();
    const iconElements = [];
    icons?.forEach((icon) => {
        if (icon) {
            iconElements.push(React.createElement("div", { key: icon.text, id: "icon", className: `${icon.classname} ml-2` }, icon.text));
        }
    });
    return (React.createElement("li", null,
        React.createElement("button", { type: "button", onClick: () => navigate(to), className: "navbar-link-btn", style: { background: "none", border: "none", padding: 0, width: "100%" } },
            React.createElement("div", { className: "col", id: "title" },
                React.createElement(Link, { to: to }, text)),
            React.createElement("div", { className: "col" },
                iconElements,
                jsx ?? null))));
};
export default NavBarLink;

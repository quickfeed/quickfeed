import React from "react";
import { useActions, useAppState } from "../overmind";
import { Link } from "react-router-dom";
import NavFavorites from "./NavFavorites";
import NavBarUser from "./navbar/NavBarUser";
import NavBarActiveCourse from "./navbar/NavBarActiveCourse";
const NavBar = () => {
    const state = useAppState();
    const actions = useActions().global;
    let hamburger = null;
    if (state.isLoggedIn) {
        const hamburgerColor = state.showFavorites ? "open" : "closed";
        const classname = `clickable hamburger ${hamburgerColor}`;
        hamburger = React.createElement("span", { onClick: () => actions.toggleFavorites(), className: classname }, "\u2630");
    }
    return (React.createElement("nav", { className: "navbar navbar-top navbar-expand-sm flexbox", id: "main" },
        React.createElement("div", { className: "nav-child brand" },
            hamburger,
            React.createElement(Link, { to: "/", style: { color: "#d4d4d4", fontWeight: "bold" } }, "QuickFeed")),
        React.createElement(NavBarActiveCourse, null),
        React.createElement(NavBarUser, null),
        React.createElement(NavFavorites, null)));
};
export default NavBar;

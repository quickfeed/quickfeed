import React from "react";
import AboutButton from "../navbar-buttons/AboutButton";
import AdminButton from "../navbar-buttons/AdminButton";
import { useAppState } from "../../overmind";
import ProfileButton from "../navbar-buttons/ProfileButton";
import LogoutButton from "../navbar-buttons/LogoutButton";
import StreamStatus from "./StreamStatus";
const NavBarUser = () => {
    const { self, isLoggedIn } = useAppState();
    if (!isLoggedIn) {
        return (React.createElement("a", { href: "/auth/github", className: "flex-user signIn mr-2" }, "Sign In"));
    }
    return (React.createElement("div", { className: "flex-user" },
        React.createElement(StreamStatus, null),
        React.createElement("ul", { className: "nav-item dropdown" },
            React.createElement("i", { className: "fa fa-chevron-down text-white mr-1 chevron-icon" }),
            React.createElement("img", { className: "rounded-circle", src: self.AvatarURL, id: "avatar" }),
            React.createElement("ul", { className: "dropdown-menu dropdown-menu-center bg-dark" },
                React.createElement(ProfileButton, null),
                React.createElement(AboutButton, null),
                React.createElement(AdminButton, null),
                React.createElement(LogoutButton, null)))));
};
export default NavBarUser;

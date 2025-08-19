import React from "react";
import { useActions } from "../../overmind";
const LogoutButton = () => {
    const actions = useActions().global;
    return (React.createElement("li", null,
        React.createElement("a", { href: "/logout", className: "sidebar-items-link dropdown-item bg-dark", style: { color: "#d4d4d4" }, onClick: () => actions.logout() }, "Log out")));
};
export default LogoutButton;

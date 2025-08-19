import React from "react";
import { Link } from "react-router-dom";
import { useAppState } from "../../overmind";
const AdminButton = () => {
    const { self } = useAppState();
    if (self.IsAdmin) {
        return (React.createElement("li", null,
            React.createElement(Link, { to: "/admin", className: "sidebar-items-link dropdown-item bg-dark", style: { color: "#d4d4d4" } }, "Admin")));
    }
    return null;
};
export default AdminButton;

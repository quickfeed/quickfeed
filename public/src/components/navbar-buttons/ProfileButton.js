import React from "react";
import { Link } from "react-router-dom";
const ProfileButton = () => {
    return (React.createElement("li", null,
        React.createElement(Link, { to: "/profile", className: "sidebar-items-link dropdown-item bg-dark", style: { color: "#d4d4d4" } }, "Profile")));
};
export default ProfileButton;

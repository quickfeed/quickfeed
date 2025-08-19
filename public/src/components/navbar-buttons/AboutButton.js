import React from "react";
import { Link } from "react-router-dom";
const AboutButton = () => {
    return (React.createElement("li", { key: "about" },
        React.createElement(Link, { to: "/about", className: "sidebar-items-link dropdown-item bg-dark", style: { color: "#d4d4d4" } }, "About")));
};
export default AboutButton;

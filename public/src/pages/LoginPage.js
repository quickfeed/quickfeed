import React, { useEffect } from "react";
import AboutPage from "./AboutPage";
const LoginPage = () => {
    useEffect(() => {
        document.body.classList.add("login-page");
        return () => {
            document.body.classList.remove("login-page");
        };
    }, []);
    return (React.createElement("div", { className: "loginContainer" },
        React.createElement("h1", { className: "loginWelcomeHeader" }, "Welcome to QuickFeed"),
        React.createElement("p", { className: "lead mt-5 mb-5", style: { textAlign: "center", marginBottom: "50px" } }, "To get started with QuickFeed, please sign in with your GitHub account."),
        React.createElement("section", { id: "loginBox" },
            React.createElement("div", { className: "loginBox" },
                React.createElement("i", { className: "fa fa-5x fa-github align-middle ms-auto", id: "github icon" }),
                React.createElement("h4", null, "Sign in with GitHub"),
                React.createElement("p", { className: "text-secondary" }, " to continue to QuickFeed "),
                React.createElement("a", { href: "/auth/github", className: "loginButton" }, " Sign in "))),
        React.createElement(AboutPage, null)));
};
export default LoginPage;

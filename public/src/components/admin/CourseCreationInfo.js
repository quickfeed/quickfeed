import React from "react";
const CourseCreationInfo = () => {
    return (React.createElement("div", { className: "jumbotron" },
        React.createElement("h1", { className: "display-4" }, "Create Course"),
        React.createElement("p", { className: "lead" }, "For each new semester of a course, QuickFeed requires a new GitHub organization. This is to keep the student roster for the different runs of the course separate."),
        React.createElement("p", null,
            React.createElement("a", { className: "badge-pill badge-success", href: "https://github.com/organizations/plan", target: "_blank", rel: "noopener noreferrer" }, "Create an organization"),
            " for your course. The course organization must allow private repositories."),
        React.createElement("p", null,
            "Add the ",
            React.createElement("a", { className: "badge-pill badge-info", href: process.env.QUICKFEED_APP_URL, target: "_blank", rel: "noopener noreferrer" }, "QuickFeed application"),
            " to your GitHub organization to create a course."),
        React.createElement("p", null, "QuickFeed will create the following repositories for you:"),
        React.createElement("ul", null,
            React.createElement("li", null, "info"),
            React.createElement("li", null, "assignments"),
            React.createElement("li", null, "tests")),
        React.createElement("p", null,
            React.createElement("span", null, "Please refer to the "),
            React.createElement("a", { className: "badge-pill badge-primary", href: "https://github.com/quickfeed/quickfeed/blob/master/doc/teacher.md", target: "_blank", rel: "noopener noreferrer" }, "documentation"),
            React.createElement("span", null, " for further instructions on how to work with the various repositories.")),
        React.createElement("p", null,
            React.createElement("span", null, "After you have installed the QuickFeed application, enter the name of the organization in the field below to find the created course."))));
};
export default CourseCreationInfo;

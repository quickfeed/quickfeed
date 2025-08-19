import React from "react";
const SignupText = () => {
    return (React.createElement("blockquote", { className: "blockquote card-body", style: { color: "red" } },
        React.createElement("p", null, "Fill in the form below to complete signup."),
        React.createElement("p", null,
            "Use your ",
            React.createElement("i", null, "real name"),
            " as it appears on Canvas."),
        React.createElement("p", null, "If your name does not match any names on Canvas, you will not be granted access.")));
};
export default SignupText;

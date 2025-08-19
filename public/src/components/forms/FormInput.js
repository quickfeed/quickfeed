import React from "react";
const FormInput = ({ prepend, name, placeholder, defaultValue, onChange, type, children }) => {
    return (React.createElement("div", { className: "input-group mb-3" },
        React.createElement("div", { className: "input-group-prepend" },
            React.createElement("div", { className: "input-group-text" }, prepend)),
        React.createElement("input", { className: "form-control", name: name, type: type ?? "text", placeholder: placeholder, defaultValue: defaultValue, onChange: onChange }),
        children));
};
export default FormInput;

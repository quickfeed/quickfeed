import React from "react";
const UnstyledButton = (props) => {
    return (React.createElement("button", { type: "button", style: { color: "black" }, onClick: props.onClick, className: "btn btn-link p-0" }, props.children));
};
export default UnstyledButton;

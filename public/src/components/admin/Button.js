import React from "react";
export var ButtonType;
(function (ButtonType) {
    ButtonType["BADGE"] = "badge badge";
    ButtonType["BUTTON"] = "btn btn";
    ButtonType["OUTLINE"] = "btn btn-outline";
    ButtonType["UNSTYLED"] = "btn btn-link p-0";
})(ButtonType || (ButtonType = {}));
const Button = ({ children, text, color, type, className, onClick }) => {
    return (React.createElement("button", { className: `${type}-${color} ${className ?? ""}`, onClick: onClick },
        children,
        text));
};
export default Button;

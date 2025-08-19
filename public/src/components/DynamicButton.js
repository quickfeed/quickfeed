import React, { useState } from "react";
import { Color } from "../Helpers";
const DynamicButton = ({ text, onClick, color, type, className }) => {
    const [isPending, setIsPending] = useState(false);
    const handleClick = async () => {
        if (isPending) {
            return;
        }
        setIsPending(true);
        await onClick();
        setIsPending(false);
    };
    const buttonClass = `${type}-${isPending ? Color.GRAY : color} ${className ?? ""}`;
    const content = isPending
        ? React.createElement("span", { className: "spinner-border spinner-border-sm", role: "status", "aria-hidden": "true" })
        : text;
    return (React.createElement("button", { type: "button", disabled: isPending, className: buttonClass, onClick: handleClick }, content));
};
export default DynamicButton;

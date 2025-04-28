import React from "react"
import { Color } from "../../Helpers"

export enum ButtonType {
    BADGE = "badge badge",
    BUTTON = "btn btn",
    OUTLINE = "btn btn-outline",
    UNSTYLED = "btn btn-link p-0",
}

export type ButtonProps = {
    children?: React.ReactNode,
    text: string,
    color: Color,
    type: ButtonType,
    className?: string,
    onClick: () => void | Promise<void>,
}

const Button = ({ children, text, color, type, className, onClick }: ButtonProps) => {
    return (
        <button className={`${type}-${color} ${className ?? ""}`} onClick={onClick}>
            {children}
            {text}
        </button>
    )
}

export default Button

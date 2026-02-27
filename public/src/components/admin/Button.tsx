import React from "react"
import { Color, ButtonColorClasses } from "../../Helpers"

export enum ButtonType {
    SOLID = "btn", // Default is solid
    OUTLINE = "btn-outline",
    GHOST = "btn-ghost",
    DASH = "btn-dash",
    LINK = "btn-link",
    SOFT = "btn-soft",
}

export type ButtonProps = {
    children?: React.ReactNode,
    text: string,
    color: Color,
    type?: ButtonType,
    className?: string,
    onClick: () => void | Promise<void>,
}

const Button = ({ children, text, color, type, className, onClick }: ButtonProps) => {
    const colorClass = ButtonColorClasses[color]
    return (
        <button className={`btn ${type ?? ""} ${colorClass} ${className ?? ""}`} onClick={onClick}>
            {children}
            {text}
        </button>
    )
}

export default Button

import React from "react"
import { Color } from "../../Helpers"

export const ButtonColorClasses: Record<Color, string> = {
    [Color.RED]: "btn-error",
    [Color.BLUE]: "btn-primary",
    [Color.GREEN]: "btn-success",
    [Color.YELLOW]: "btn-warning",
    [Color.GRAY]: "btn-neutral",
    [Color.WHITE]: "btn-ghost",
    [Color.BLACK]: "btn-neutral",
}

export enum ButtonType {
    SOLID = "", // Default is solid
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
    disabled?: boolean,
}

const Button = ({ children, text, color, type = ButtonType.SOLID, className, onClick, disabled }: ButtonProps) => {
    const colorClass = ButtonColorClasses[color]
    return (
        <button className={`btn ${type} ${colorClass} ${className ?? ""}`} onClick={onClick} disabled={disabled}>
            {children}
            {text}
        </button>
    )
}

export default Button

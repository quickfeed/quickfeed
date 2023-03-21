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
    onclick: () => void,
    color: Color,
    type: ButtonType,
    classname?: string,
}

const Button = ({ children, text, onclick, color, type, classname }: ButtonProps): JSX.Element => {
    return (
        <button className={`${type}-${color}${classname ? " " + classname : ""}`} onClick={onclick}>
            {children}
            {text}
        </button>
    )
}

export default Button

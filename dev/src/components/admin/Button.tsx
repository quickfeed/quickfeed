import React from "react"

export enum Color {
    RED = "danger",
    BLUE = "primary",
    GREEN = "success",
    YELLOW = "warning",
    GRAY = "secondary",
}

export enum ButtonType {
    BADGE = "badge badge",
    BUTTON = "btn btn"
}

export type ButtonProps = {
    text: string,
    onclick: () => void,
    color: Color,
    type: ButtonType,
    classname?: string,
}

const Button = ({ text, onclick, color, type, classname }: ButtonProps): JSX.Element => {
    return (
        <span className={`${type}-${color}` + " clickable " + classname} onClick={onclick}>
            {text}
        </span>
    )
}

export default Button

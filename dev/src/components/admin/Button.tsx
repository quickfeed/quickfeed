import React from "react"

type ColorKeys = "RED" | "BLUE" | "GREEN" | "YELLOW" | "GRAY"
type Color = {[color in ColorKeys]: string}
export const ComponentColor: Color = {
    RED: "danger",
    BLUE: "primary",
    GREEN: "success",
    YELLOW: "warning",
    GRAY: "secondary"
}

type TypeKeys = "BADGE" | "BUTTON"
type ButtonType = {[type in TypeKeys]: string}
export const ButtonType: ButtonType = {
    BADGE: "badge badge",
    BUTTON: "btn btn"
}

const Button = ({text, onclick, color, type}: {text: string, onclick: () => void, color: string, type: string}): JSX.Element => {
    //const text = user.getIsadmin() ? "Demote" : "Promote"
    return (
        <span className={`${type}-${color}` + " clickable"} onClick={onclick}>
            {text}
        </span>
    )
}

export default Button

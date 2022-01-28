import React from "react"
import { useState } from "react"
import { Color } from "../Helpers"
import { ButtonType } from "./admin/Button"


export type DynamicButtonProps = {
    text: string,
    onClick: () => Promise<unknown>,
    color: Color,
    type: ButtonType,
    className?: string,
}

// DynamicButton will display a spinner while the onClick function is running.
const DynamicButton = ({ text, onClick, color, type, className }: DynamicButtonProps) => {
    const [isPending, setIsPending] = useState<boolean>(false)

    const handleClick = async () => {
        setIsPending(true)
        await onClick()
        setIsPending(false)
    }

    const buttonClass = isPending ? `${type}-${Color.GRAY} ${className ?? ""}` : `${type}-${color} ${className ?? ""}`
    const content = isPending
        ? <span className="spinner-border spinner-border-sm" role="status" aria-hidden="true"></span>
        : text

    return (
        <button disabled={isPending} className={buttonClass} onClick={handleClick} >
            {content}
        </button >
    )
}

export default DynamicButton

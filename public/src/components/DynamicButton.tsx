import React, { useState } from "react"
import { Color } from "../Helpers"
import { ButtonType } from "./admin/Button"


export type DynamicButtonProps = {
    text: string,
    onClick: () => Promise<unknown>,
    color: Color,
    type: ButtonType,
    className?: string,
}

/** DynamicButton will display a spinner while the onClick function is running.
 *  This is useful for buttons that perform an action that takes a while to complete.
 *  The button will be disabled while the onClick function is running.
 */
const DynamicButton = ({ text, onClick, color, type, className }: DynamicButtonProps) => {
    const [isPending, setIsPending] = useState<boolean>(false)

    const handleClick = async () => {
        if (isPending) {
            // Disable double clicks
            return
        }
        setIsPending(true)
        await onClick()
        setIsPending(false)
    }

    const buttonClass = `${type}-${isPending ? Color.GRAY : color} ${className ?? ""}`
    const content = isPending
        ? <span className="spinner-border spinner-border-sm" role="status" aria-hidden="true" />
        : text

    return (
        <button type="button" disabled={isPending} className={buttonClass} onClick={handleClick}>
            {content}
        </button>
    )
}

export default DynamicButton

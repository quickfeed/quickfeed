import { useState } from "react"
import { Color } from "../Helpers"
import type { ButtonType } from "./admin/Button"


export type DynamicButtonProps = {
    text: string,
    onClick: () => Promise<void>,
    color: Color,
    type?: ButtonType,
    className?: string,
    disabled?: boolean,
}

/** DynamicButton will display a spinner while the onClick function is running.
 *  This is useful for buttons that perform an action that takes a while to complete.
 *  The button will be disabled while the onClick function is running.
 */
const DynamicButton = ({ text, onClick, color, type, className, disabled }: DynamicButtonProps) => {
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

    const buttonClass = `btn ${type ?? ""} btn-${isPending ? Color.GRAY : color} ${className ?? ""}`
    const content = isPending
        ? <span className="loading loading-spinner" role="status" aria-hidden="true" />
        : text

    return (
        <button type="button" disabled={isPending || disabled} className={buttonClass} onClick={handleClick}>
            {content}
        </button>
    )
}

export default DynamicButton

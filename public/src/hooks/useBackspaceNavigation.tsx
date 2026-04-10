import { useEffect } from "react"
import { useLocation, useNavigate } from "react-router"

/**
 * useBackspaceNavigation sets up a keyboard shortcut that navigates to the root path
 * when the user presses Backspace, unless they are currently typing in an input field.
 *
 * @param root - The root path to navigate to when Backspace is pressed
 */
export const useBackspaceNavigation = (root: string) => {
    const location = useLocation()
    const navigate = useNavigate()

    useEffect(() => {
        const handleKeyDown = (event: KeyboardEvent) => {
            // Only trigger if Backspace is pressed and we're not already at root
            if (event.key === "Backspace" && location.pathname !== root) {
                // Check if the user is typing in an input field
                const target = event.target as HTMLElement
                const isInputField =
                    target.tagName === "INPUT" ||
                    target.tagName === "TEXTAREA" ||
                    target.isContentEditable

                // If not in an input field, navigate to root
                if (!isInputField) {
                    event.preventDefault()
                    navigate(root)
                }
            }
        }

        window.addEventListener("keydown", handleKeyDown)
        return () => window.removeEventListener("keydown", handleKeyDown)
    }, [location.pathname, root, navigate])
}

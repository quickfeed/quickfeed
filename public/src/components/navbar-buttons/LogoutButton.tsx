import React from "react"
import { useActions } from "../../overmind"

const LogoutButton = () => {
    const actions = useActions()
    return (
        <li>
            <a href="/logout" className="sidebar-items-link dropdown-item bg-dark" style={{ color: "#d4d4d4" }} onClick={() => actions.logout()}>Log out</a> {/* skipcq: JS-0417 */}
        </li>
    )
}

export default LogoutButton

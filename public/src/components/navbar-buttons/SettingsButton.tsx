import React from "react"
import { Link } from "react-router-dom"


const SettingsButton = () => {
    return (
        <li>
            <Link to="/settings" className="sidebar-items-link dropdown-item bg-dark" style={{ color: "#d4d4d4" }}>
                Settings
            </Link>
        </li>
    )
}

export default SettingsButton

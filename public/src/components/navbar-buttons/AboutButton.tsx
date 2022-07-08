import React from "react"
import { Link } from "react-router-dom"

const AboutButton = (): JSX.Element => {
    return (
        <li key="about">
            <Link to="/about" className="sidebar-items-link dropdown-item bg-dark" style={{ color: "#d4d4d4" }}>
                About
            </Link>
        </li>
    )
}
export default AboutButton

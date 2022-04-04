import React from "react"
import { Link } from "react-router-dom"

const AboutButton = (): JSX.Element => {
    return (
        <li key="about">
            <a className="dropdown-item bg-dark" >
                <Link to="/about" className="sidebar-items-link" style={{ color: "#d4d4d4" }}>
                    About
                </Link>
            </a>
        </li>
    )
}
export default AboutButton

import React from "react"
import { Link } from "react-router-dom"
import { User } from "../../../proto/ag/ag_pb"


const ProfileButton = () => {
    return (
        <li>
            <Link to="/profile" className="sidebar-items-link dropdown-item bg-dark" style={{ color: "#d4d4d4" }}>
                Profile
            </Link>
        </li>
    )
}

export default ProfileButton

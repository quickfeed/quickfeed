import React from "react"
import { Link } from "react-router-dom"
import { useAppState } from "../../overmind"

const AdminButton = () => {
    const { self } = useAppState()
    if (self.isAdmin) {
        return (
            <li>
                <Link to="/admin" className="sidebar-items-link dropdown-item bg-dark" style={{ color: "#d4d4d4" }}>
                    Admin
                </Link>
            </li>
        )
    }
    return null
}
export default AdminButton

import React from "react"
import { Link } from "react-router-dom"
import { useAppState } from "../../overmind"

const AdminButton = () => {
    const { self } = useAppState()
    if (self.getIsadmin()) {
        return (
            <li>
                <a className="dropdown-item bg-dark" >
                    <Link to="/admin" className="sidebar-items-link" style={{ color: "#d4d4d4" }}>
                        Admin
                    </Link>
                </a>
            </li>
        )
    }
    return null
}
export default AdminButton

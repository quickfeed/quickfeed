import React, { useState } from "react"
import { Link, useHistory } from "react-router-dom"
import { useActions, useAppState } from "../../overmind"


const NavBarFooter = (): JSX.Element => {
    const logout = useActions().logout
    const { self, isLoggedIn } = useAppState()
    const history = useHistory()

    const [hidden, setHidden] = useState<boolean>(true)

    const logoutButton = (
        <li hidden={hidden}>
            <a href="/logout" className="Sidebar-items-link" onClick={() => logout()}>Log out</a>
        </li>
    )

    const profileButton = isLoggedIn
        ? (
            <li onClick={() => history.push("/profile")} onMouseEnter={() => setHidden(false)}>
                <div><img src={self.getAvatarurl()} id="avatar"></img></div>
            </li>
        )
        : null


    const aboutButton = (
        <li key="about" hidden={hidden}>
            <Link to="/about" className="Sidebar-items-link">
                About
            </Link>
        </li>
    )


    const adminButton = self.getIsadmin()
        ? (
            <li hidden={hidden}>
                <Link to="/admin" className="Sidebar-items-link">
                    Admin
                </Link>
            </li>
        )
        : null


    return (
        <div className="SidebarFooter" onMouseLeave={() => setHidden(true)}>
            {aboutButton}
            {adminButton}
            {profileButton}
            {logoutButton}
        </div>
    )
}

export default NavBarFooter

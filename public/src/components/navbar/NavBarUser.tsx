import React from "react"
import AboutButton from "../navbar-buttons/AboutButton"
import AdminButton from "../navbar-buttons/AdminButton"
import { useAppState } from "../../overmind"
import ProfileButton from "../navbar-buttons/ProfileButton"
import LogoutButton from "../navbar-buttons/LogoutButton"
import StreamStatus from "./StreamStatus"
import { Link } from "react-router-dom"

const NavBarUser = () => {
    const { self, isLoggedIn, unreadNotifications } = useAppState()

    if (!isLoggedIn) {
        return (
            <a href="/auth/github" className="flex-user signIn mr-2">Sign In</a>
        )
    }

    return (
        <div className="flex-user">

            <Link to="/notifications" className="mr-2" style={{ position: "relative" }}>
                <i className="fa fa-bell notification-icon" />
                {/* Only show the red circle if there are unread notifications */}
                {unreadNotifications > 0 &&
                    <span className="notification-badge"> {unreadNotifications} </span>
                }
            </Link>

            <ul className="nav-item dropdown">
                <i className="fa fa-chevron-down text-white mr-1 chevron-icon" />
                <img className="rounded-circle" src={self.AvatarURL} id="avatar" />
                <ul className="dropdown-menu dropdown-menu-center bg-dark">
                    <ProfileButton />
                    <AboutButton />
                    <AdminButton />
                    <LogoutButton />
                </ul>
            </ul>

            <StreamStatus />

        </div >
    )
}

export default NavBarUser

import React from "react"
import AboutButton from "../navbar-buttons/AboutButton"
import AdminButton from "../navbar-buttons/AdminButton"
import { useAppState } from "../../overmind"
import ProfileButton from "../navbar-buttons/ProfileButton"
import LogoutButton from "../navbar-buttons/LogoutButton"
import StreamStatus from "./StreamStatus"

const NavBarUser = () => {
    const { self, isLoggedIn } = useAppState()

    if (!isLoggedIn) {
        return (
            <a href="/auth/github" className="flex-user signIn mr-2">Sign In</a>
        )
    }

    return (
        <div className="flex items-center">

            <StreamStatus />
            <ul className="nav-item dropdown flex items-center">
                <i className="fa fa-chevron-down text-white mr-1 chevron-icon" />
                <img className="rounded-circle" src={self.AvatarURL} id="avatar" />
                <ul className="dropdown-menu dropdown-menu-center bg-dark">
                    <ProfileButton />
                    <AboutButton />
                    <AdminButton />
                    <LogoutButton />
                </ul>
            </ul>

        </div>
    )
}

export default NavBarUser

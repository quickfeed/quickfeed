import React from "react"
import AboutButton from "../navbar-buttons/AboutButton"
import AdminButton from "../navbar-buttons/AdminButton"
import { useAppState } from "../../overmind"
import ProfileButton from "../navbar-buttons/ProfileButton"
import LogoutButton from "../navbar-buttons/LogoutButton"
import StreamStatus from "./StreamStatus"
import SettingsButton from "../navbar-buttons/SettingsButton"

const NavBarUser = (): JSX.Element => {
    const { self, isLoggedIn } = useAppState()

    if (!isLoggedIn) {
        return (
            <ul>
                <a href="/auth/github" className="signIn" style={{ textAlign: "right", color: "#d4d4d4", marginRight: "55px" }}>
                    <i className="fa fa-2x fa-github align-middle ms-auto " id="github" />
                </a>
            </ul>
        )
    }

    return (
        <div className="flex-user">

            <StreamStatus />
            <ul className="nav-item dropdown">
                <img className="rounded-circle" src={self.AvatarURL} id="avatar" />
                <ul className="dropdown-menu dropdown-menu-center bg-dark">
                    <ProfileButton />
                    <AboutButton />
                    <AdminButton />
                    <SettingsButton />
                    <LogoutButton />
                </ul>
            </ul>

        </div>
    )
}

export default NavBarUser

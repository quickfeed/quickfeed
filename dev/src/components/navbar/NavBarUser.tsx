import React from "react"
import AboutButton from "../navbar-buttons/AboutButton"
import AdminButton from "../navbar-buttons/AdminButton"
import LoginButton from "../navbar-buttons/LoginButton"
import { useAppState } from "../../overmind"
import ProfileButton from "../navbar-buttons/ProfileButton"

const NavBarUser = (): JSX.Element => {
    const { self } = useAppState()
    const state = useAppState()

    if (!state.isLoggedIn) {
        return <></>
    }

    return (
        <div className="navbar-collapse ml-auto" id="main_nav">
            <ul className="navbar-nav ml-auto">
                <li className="nav-item dropdown ml-auto">
                    <img className="rounded-circle" src={self.getAvatarurl()} id="avatar"
                        style={{ height: "40px", borderRadius: "50%" }} />
                    <ul className="dropdown-menu dropdown-menu-center bg-dark">
                        <ProfileButton />
                        <AboutButton />
                        <AdminButton />
                        <LoginButton />
                    </ul>
                </li>
            </ul>
        </div>
    )
}

export default NavBarUser

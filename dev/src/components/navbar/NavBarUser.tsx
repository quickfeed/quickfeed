import React from "react"
import AboutButton from "../navbar-buttons/AboutButton"
import AdminButton from "../navbar-buttons/AdminButton"
import LoginButton from "../navbar-buttons/LoginButton"
import { useActions, useAppState } from "../../overmind"

const NavBarUser = (): JSX.Element => {
    const { self } = useAppState()
    const state = useAppState()

    return (

        <div className="navbar-collapse ml-auto" id="main_nav">
            <ul className="navbar-nav ml-auto">
                <li className="nav-item dropdown ml-auto">
                    <a href="/auth/github" style={{ textAlign: "center", paddingTop: "15px", marginLeft: "40px" }}>
                        <img className="rounded-circle" src={self.getAvatarurl()} id="avatar" style={{ height: "40px", borderRadius: "50%" }} />
                    </a>
                    {state.isLoggedIn &&
                        <ul className="dropdown-menu dropdown-menu-center bg-dark">
                            <AboutButton />
                            <AdminButton />
                            <LoginButton />
                        </ul>
                    }
                </li>
            </ul>
        </div>

    )
}

export default NavBarUser

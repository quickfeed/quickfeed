import React from "react"
import { useActions, useAppState } from "../overmind"
import { Link } from "react-router-dom"
import NavFavorites from "./NavFavorites"
import NavBarUser from "./navbar/NavBarUser"


const NavBar = (): JSX.Element => {
    const state = useAppState()
    const actions = useActions()

    const hamburger = state.isLoggedIn ? <span onClick={() => actions.toggleFavorites()} className="ml-3 hamburger">☰</span> : null

    return (
        <nav className="navbar navbar-expand-lg" style={{ backgroundColor: "#222", color: "#d4d4d4" }} id="main" >
            {!state.showFavorites &&
                <div className="navbar-brand clickable" style={{ marginLeft: "30px", fontWeight: "bold", fontSize: "30px" }}>
                    <Link to="/" style={{ color: "#d4d4d4" }}>QuickFeed</Link>
                    {hamburger}
                </div>
            }
            {!state.isLoggedIn &&
                <div className="navbar-collapse ml-auto">
                    <a href="/auth/github" className="nav-item ml-auto ms-auto signIn" style={{ textAlign: "right", color: "#d4d4d4", marginRight: "55px" }}>
                        <i className="fa fa-2x fa-github align-middle ms-auto " id="github" />
                    </a>
                </div>
            }
            {state.isLoggedIn &&
                <ul className="ms-auto ml-auto list-unstyled" style={{ marginRight: "55px", paddingTop: "15px" }}>
                    <NavBarUser />
                </ul>
            }
            {state.showFavorites &&
                <NavFavorites />
            }
        </nav>
    )
}

export default NavBar

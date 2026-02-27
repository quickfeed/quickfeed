import React from "react"
import { useActions, useAppState } from "../overmind"
import { Link } from "react-router-dom"
import NavFavorites from "./NavFavorites"
import NavBarUser from "./navbar/NavBarUser"
import NavBarActiveCourse from "./navbar/NavBarActiveCourse"


const NavBar = () => {
    const state = useAppState()
    const actions = useActions().global

    let hamburger = null
    if (state.isLoggedIn) {
        const hamburgerColor = state.showFavorites ? "open" : "closed" // Green / White
        const classname = `clickable hamburger ${hamburgerColor}`
        hamburger = <span onClick={() => actions.toggleFavorites()} className={classname}>☰</span>
    }

    return (
        <nav className="navbar bg-base-300 navbar-top navbar-expand-sm flexbox" id="main" >
            <div className="nav-child brand">
                {hamburger}
                <Link to="/">QuickFeed</Link>
            </div>
            <NavBarActiveCourse />
            <NavBarUser />
            <NavFavorites />
        </nav>
    )
}

export default NavBar

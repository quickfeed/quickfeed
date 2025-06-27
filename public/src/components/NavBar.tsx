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
        <nav className="navbar navbar-top navbar-expand-sm flexbox" id="main" >
            <div className="nav-child brand">
                {hamburger}
                <Link to="/" style={{ color: "#d4d4d4", fontWeight: "bold" }}>QuickFeed</Link>
            </div>
            <NavBarActiveCourse />
            <NavBarUser />
            <NavFavorites />
        </nav>
    )
}

export default NavBar

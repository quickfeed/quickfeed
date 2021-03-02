import React, { Component } from "react";
import { useActions, useOvermind } from "../overmind";
import { Link } from 'react-router-dom'
import { ToggleSwitch } from "./ToggleSwitch";


const NavBar = () => {
    const { state, actions } = useOvermind() 

    const checkUserLoggedIn = () => {
        if (state.user.id > 0) {
            return <a href="/logout"><button>Logout</button></a>
        }
        return <a href="/auth/github"><button>Login</button></a>
    }
   
    const goHome = () => {
        state.currentPage = "home"
    }

    const changeTheme = () => {
        actions.changeTheme()
        window.localStorage.setItem("theme", state.theme)
    }

    return (
        <nav className="navbar">
            
            <Link to="/">
                <button className="navbar-brand">Autograder</button>
            </Link>
            <Link to="/info">
                <button>Info</button>
            </Link>
            <Link to="/profile">
                <button>Profile</button>
            </Link>
            <button onClick={() => changeTheme()}>{state.theme}</button>
            {checkUserLoggedIn()}
        </nav>
    )
    
}

export default NavBar
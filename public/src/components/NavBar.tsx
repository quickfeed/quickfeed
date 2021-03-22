import React, { Component } from "react";
import { useActions, useOvermind } from "../overmind";
import { Link } from 'react-router-dom'
import { ToggleSwitch } from "./ToggleSwitch";


const NavBar = () => {
    const { state, actions } = useOvermind() 

    const checkUserLoggedIn = () => {
        if (state.user.id > 0) {
            return <a href="/logout" className="login">Log out</a>
        }
        return <a href="/auth/github" className="login"><i className="fa fa-2x fa-github" id="github"></i></a>
    }

    const changeTheme = () => {
        actions.changeTheme()
        window.localStorage.setItem("theme", state.theme)
        document.body.className = state.theme
    }

    return (
        <nav className="navigator">
            <div className="container">
            <Link to="/">
                <span className="navbar-brand">Autograder</span>
            </Link>
            
            {
                // TODO: Figure out how to handle this
                // Currently only show this link if the user has an enrollment, regardless of status
                state.enrollments.length > 0 
                ?             
                <Link to="/courses" className="navigator-item">
                    Courses
                </Link>
                : ""
            }

            <Link to="/info" className="navigator-item">
                Info
            </Link>
            <Link to="/profile" className="navigator-item">
                Profile
            </Link>
            <span onClick={() => changeTheme()}><i className={state.theme === "light" ? "fa fa-sun-o" : "fa fa-moon-o"} style={{color: "white"}}></i></span>
            {checkUserLoggedIn()}
            </div>
        </nav>
    )
    
}

export default NavBar
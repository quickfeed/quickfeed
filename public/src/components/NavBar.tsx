import React, { Component } from 'react'
import { useActions, useOvermind } from '../overmind'
import { Link } from 'react-router-dom'


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
        window.localStorage.setItem('theme', state.theme)
        document.body.className = state.theme
    }

    return (
        <nav className="navigator">
            <div className="container">
            <Link to="/">
                <span className="navbar-brand">Autograder</span>
            </Link>
            <Link to="/info" className="navigator-item">
                Info
            </Link>
            <Link to="/profile" className="navigator-item">
                Profile
            </Link>
            <span onClick={() => changeTheme()}><i className={state.theme === 'light' ? 'fa fa-sun-o' : 'fa fa-moon-o'} style={{color: 'white'}} id="themeButton"></i></span>
            {checkUserLoggedIn()}
            </div>
        </nav>
    )
    
}

export default NavBar
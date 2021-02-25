import React, { Component } from "react";
import { useOvermind } from "../overmind";

const NavBar = () => {
    const { state, actions } = useOvermind() 

    const checkUserLoggedIn = () => {
        if (state.user.id > 0) {
            return <a href="/logout"><button>Logout</button></a>
        }
        return <a href="/auth/github"><button>Login</button></a>
    }
   
    const users = () => {
        actions.getUsers()
        actions.getCourses()
    }

    return (
        <nav className="navbar">
            <button className="navbar-brand" onClick={users}>Autograder</button>
            {checkUserLoggedIn()}
        </nav>
    )
    
}

export default NavBar
import React, { Component, useState } from "react";
import { useActions, useOvermind } from "../overmind";
import { Link } from 'react-router-dom'
import { ToggleSwitch } from "./ToggleSwitch";
import { act } from "react-dom/test-utils";


const NavBar = () => {
    const { state, actions } = useOvermind() 

    const [active, setActive] = useState(false)

    const checkUserLoggedIn = () => {
        if (state.user.id > 0) {
            return <li><div id="title"><a href="/logout">Log out</a></div></li>
        }
        return <li><a href="/auth/github"><i className="fa fa-2x fa-github" id="github"></i></a></li>
    }

    const changeTheme = () => {
        actions.changeTheme()
        window.localStorage.setItem("theme", state.theme)
        document.body.className = state.theme
    }

    const courses = () => {
        return state.enrollments.map(enrollment => {
            return (
            <li className={active ? "active" : "inactive"}>
                <div id="title"><Link to={`/course/` + enrollment.getCourseid()}>{enrollment.getCourse()?.getCode()}</Link></div>
            </li>)
        })
    }
    return (
        <nav className="navigator">
            <ul className="SidebarList">
            <li>
                <Link to="/">
                    <span className="logo">Autograder</span>
                </Link>
            </li>
            
        
                {state.user.id > 0 ? 
                <li>
                    <div id="icon"><img src={state.user.avatarurl} id="avatar"></img></div>
                        
                        <div id="title">{state.user.name}</div>
                </li>
                 : ""}

            

            {
                state.enrollments.length > 0 
                ?             
                <li onClick={() => setActive(!active)}>
                    <div id="title">
                        <Link to="/courses">
                            Courses
                        </Link>
                    </div>
                </li>

                : ""
            }
            {
                state.enrollments.length > 0 ?
                courses()
                : ""
            }
            <li>
            <div id="title">
                <Link to="/info">
                    Info
                </Link>
            </div>
            </li>

            <li>
                <div id="title">
                <Link to="/profile">
                    Profile
                </Link>
                </div>
            </li>

            <li>
                <span onClick={() => changeTheme()}>
                    <i className={state.theme === "light" ? "icon fa fa-sun-o" : "icon fa fa-moon-o"} style={{color: "white"}}></i>
                </span>
            </li>
            {checkUserLoggedIn()}
            </ul>
        </nav>
    )
    
}

export default NavBar
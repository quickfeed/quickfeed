import React, { useState } from "react";
import { useOvermind } from "../overmind";
import { Link } from 'react-router-dom'
import NavBarLabs from "./NavBarLabs";


const NavBar = () => {
    const { state, actions } = useOvermind() 

    const [active, setActive] = useState(false)

    const checkUserLoggedIn = () => {
        if (state.user.id > 0) {
            return <li><div id="title"><a href="/logout">Log out</a></div></li>
        }
        return <div><li>Log in with</li><li><a href="/auth/github"><i className="fa fa-2x fa-github" id="github"></i></a></li></div>
    }

    // Generates dropdown items related to Courses
    const CourseItems = (): JSX.Element[] => {
        let links: JSX.Element[] = []

            
            links.push(
            <li key={0} onClick={() => { setActive(!active); actions.setActiveCourse(-1);}}>
                <div id="title">
                    <Link to="/courses">
                        Courses
                    </Link>
                </div>
            </li>
            )
        if (state.enrollments.length > 0) {
            state.enrollments.map((enrollment) =>{
                if(enrollment.getStatus()>=2) {
                    links.push(
                        <div>
                            <li key={enrollment.getCourseid()} className={active ? "active" : "inactive"} onClick={() => actions.setActiveCourse(enrollment.getCourseid())}>
                                <div id="title">
                                    <Link to={`/course/` + enrollment.getCourseid()}>
                                        {enrollment.getCourse()?.getCode()}
                                    </Link>
                                </div> 
                            </li>
                            <div className={active ? "activelabs" : "inactive"}>
                                {state.activeCourse === enrollment.getCourseid() ? <NavBarLabs /> : ""}
                            </div>
                        </div>
                    )
                }
            })

        }
        return links
    }
    return (
        <nav className="navigator">
            <ul className="SidebarList">
                <li className="logo">
                    <Link to="/">
                        Autograder
                    </Link>
                </li>
                
            
                {state.user.id > 0 ? 
                <li>
                    <Link to="/profile">
                        
                        <div id="icon"><img src={state.user.avatarurl} id="avatar"></img></div>    
                        <div id="title">{state.user.name}</div>
                    </Link>
                </li>
                    : ""}

                
                
                {state.user.id > 0 ? CourseItems() : ""}
                <li>
                    <div id="title">
                        <Link to="/info">
                            About
                        </Link>
                    </div>
                </li>

                <li>
                    <span onClick={() => actions.changeTheme()}>
                        <i className={state.theme === "light" ? "icon fa fa-sun-o" : "icon fa fa-moon-o"} style={{color: "white"}}></i>
                    </span>
                </li>
                {checkUserLoggedIn()}
            </ul>
        </nav>
    )
    
}

export default NavBar
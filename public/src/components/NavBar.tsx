import React, { useState } from "react";
import { useOvermind } from "../overmind";
import { Link } from 'react-router-dom'
import NavBarLabs from "./NavBarLabs";
import { Enrollment } from "../proto/ag_pb";
import NavBarTeacher from "./NavBarTeacher";



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
            <div>
                <li key={0} onClick={() => { setActive(!active); actions.setActiveCourse(-1);}}>
                    <div id="title">
                            Courses &nbsp;&nbsp;
                        <i className={active ? "icon fa fa-caret-down fa-lg" : "icon fa fa-caret-down fa-rotate-90 fa-lg"}></i>
                    </div>
                </li>
                <li className={active ? "active" : "inactive"}>
                    <div id="title">
                        <Link to="/courses">
                            View all courses
                        </Link>
                    </div>
                </li>
            </div>
            )
        if (state.enrollments.length > 0) {
            state.enrollments.map((enrollment) =>{
                if (enrollment.getStatus() >= 2 && enrollment.getState() === Enrollment.DisplayState.FAVORITE) {
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
                                {state.activeCourse === enrollment.getCourseid() && enrollment.getStatus() === Enrollment.UserStatus.STUDENT ? <NavBarLabs /> : ""}
                                {state.activeCourse === enrollment.getCourseid() && enrollment.getStatus() === Enrollment.UserStatus.TEACHER ? <NavBarTeacher  courseID={enrollment.getCourseid()}/> : ""}
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
                <li key="logo" className="logo">
                    <Link to="/">
                        Autograder
                    </Link>
                </li>
                
            
                {state.user.id > 0 ? 
                <li key="profile">
                    <Link to="/profile">
                        
                        <div id="icon"><img src={state.user.avatarurl} id="avatar"></img></div>    
                        <div id="title">{state.user.name}</div>
                    </Link>
                </li>
                    : ""}

                
                
                {state.user.id > 0 ? CourseItems() : ""}
                <li key="about">
                    <div id="title">
                        <Link to="/info">
                            About
                        </Link>
                    </div>
                </li>

                <li key="theme">
                    <span onClick={() => actions.changeTheme()}>
                        <i className={state.theme === "light" ? "icon fa fa-sun-o fa-lg" : "icon fa fa-moon-o fa-lg"} style={{color: "white"}}></i>
                    </span>
                </li>
                {checkUserLoggedIn()}

                { state.user.isadmin ? 
                    <li key="admin">
                        <div id="title">
                        <Link to="/admin">
                            Admin
                        </Link>
                        </div>
                    </li>
                    : ""
                }
            </ul>
        </nav>
    )
    
}

export default NavBar
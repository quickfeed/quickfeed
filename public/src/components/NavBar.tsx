import React, { useEffect, useState } from "react";
import { useOvermind } from "../overmind";
import { Link, useHistory } from 'react-router-dom'
import NavBarLabs from "./NavBarLabs";
import { Enrollment } from "../proto/ag_pb";
import NavBarTeacher from "./NavBarTeacher";



const NavBar = () => {
    const { state, actions } = useOvermind() 
    const history = useHistory()

    const [active, setActive] = useState(0)
    const [showCourses, setShowCourses] = useState(false)
    useEffect(() => {
        if (state.activeCourse > 0) {
            setActive(state.activeCourse)
            setShowCourses(false)
        }
    }, [state.activeCourse])

    const LogInButton = () => {
        if (state.user.id > 0) {
            return <li><div id="title"><a href="/logout">Log out</a></div></li>
        }
        return <li><a href="/auth/github">Log in with<i className="fa fa-2x fa-github" id="github"></i></a></li>
    }

    // Generates dropdown items related to Courses
    const CourseItems = (): JSX.Element[] => {
        let links: JSX.Element[] = []
        console.log(state.activeCourse)
            
        links.push(
            <div>
                <li key={0} onClick={() => { setShowCourses(!showCourses); actions.setActiveCourse(-1)}}>
                    <div id="title">
                            Courses &nbsp;&nbsp;
                        <i className={showCourses ? "icon fa fa-caret-down fa-lg" : "icon fa fa-caret-down fa-rotate-90 fa-lg"}></i>
                    </div>
                </li>
                <li className={showCourses ? "active" : "inactive"}>
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
                        <React.Fragment>
                            <li key={enrollment.getCourseid()} className={showCourses || active === enrollment.getCourseid()  ? "active" : "inactive"}  onClick={() => {history.push(`/course/` + enrollment.getCourseid()); setShowCourses(false)}}>
                                <div id="title">
                                        {enrollment.getCourse()?.getCode()}
                                </div> 
                            </li>
                            <div className={ state.activeCourse === enrollment.getCourseid()  ? "activelabs" : "inactive"}>
                                {state.activeCourse === enrollment.getCourseid() && enrollment.getStatus() === Enrollment.UserStatus.STUDENT ? <NavBarLabs /> : ""}
                                {state.activeCourse === enrollment.getCourseid() && enrollment.getStatus() === Enrollment.UserStatus.TEACHER ? <NavBarTeacher  courseID={enrollment.getCourseid()}/> : ""}
                            </div>
                        </React.Fragment>
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
                
            
                

                
                
                {state.user.id > 0 ? CourseItems() : ""}

                <div className="SidebarFooter">
                
                <li key="about">
                    <div id="title">
                        <Link to="/about">
                            About
                        </Link>
                    </div>
                </li>

                <li key="theme" onClick={() => actions.changeTheme()}>
                        <i className={state.theme === "light" ? "icon fa fa-sun-o fa-lg fa-flip-horizontal" : "icon fa fa-moon-o fa-lg flip-horizontal"} style={{color: "white"}}></i>  
                </li>
                {LogInButton()}

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

                {state.user.id > 0 ? 
                    <li key="profile" onClick={() => history.push("/profile")}>
                        <div id="title"><img src={state.user.avatarurl} id="avatar"></img></div>    
                    </li>
                : ""}

                </div>
            </ul>
        </nav>
    )
    
}

export default NavBar
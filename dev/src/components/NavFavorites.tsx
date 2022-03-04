import React, { Component } from "react";
import { useAppState,  useActions } from "../overmind";
import { Link } from 'react-router-dom'
import NavBarFooter from "./navbar/NavBarFooter";
import NavBarCourse from "./navbar/NavBarCourse";
import { isEnrolled, isFavorite, isVisible } from "../Helpers";

const NavFavorites = ():JSX.Element  =>{
    const state = useAppState()
    const actions = useActions()
    
    const visible = state.enrollments.filter(enrollment => isEnrolled(enrollment) && isVisible(enrollment))
    
    const courses = visible.map((enrollment) => {
        return <NavBarCourse key={enrollment.getId()} enrollment={enrollment} />
    })

    const onCourseClick  = () => {
        actions.setActiveFavorite(!state.showFavorites)
    }
    
    return (
        <nav className="navigator">
            <ul key="list" className="SidebarList">
                <li key="logo" className="logo" style={{paddingLeft: "40px"}}>
                    <Link to="/" >
                        QuickFeed 
                    </Link>
                    <a onClick={() => { onCourseClick()}}>✖</a>
                </li>
                {courses}
                {state.isLoggedIn &&
                    <li key="all" className="">
                        <Link to="/courses" className="Sidebar-items-link">
                            View all courses
                        </Link>
                    </li>
                }
            </ul>
        </nav>
    )
}
export default NavFavorites
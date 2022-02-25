import React, { Component } from "react";
import { useAppState,  useActions } from "../overmind";
import { Link } from 'react-router-dom'
import NavBarFooter from "./navbar/NavBarFooter";
import NavBarCourse from "./navbar/NavBarCourse";
import { isEnrolled, isFavorite } from "../Helpers";
import { Statemachine } from "overmind";


// interface Close {
//     isShow: () => void;
// }

const NavFavorites = ():JSX.Element  =>{
    const state = useAppState()
    const actions = useActions()

    const favorites = state.enrollments.filter(enrollment => isEnrolled(enrollment) && isFavorite(enrollment))

    const courses = favorites.map((enrollment) => {
            return <NavBarCourse key={enrollment.getId()} enrollment={enrollment} />
    })

    //var active = state.showFavorites

    const onCourseClick  = () => {
        if (state.showFavorites) {
            actions.setActiveFavorite(false);

        }else{
            actions.setActiveFavorite(true);

        }
    }
    
   


    return (
        <nav className="navigator">
            <ul key="list" className="SidebarList">
                <li key="logo" className="logo">
                    <Link to="/">
                        QuickFeed 
                    </Link>
                    <a onClick={() => { onCourseClick()}}>X</a>
                    
                </li>
                {!state.isLoggedIn &&
                    <li>
                        <a href="/auth/github" style={{ textAlign: "center", paddingTop: "15px" }}>
                            Sign in with <i className="fa fa-2x fa-github align-middle ml-2" id="github" />
                        </a>
                    </li>
                }

                {courses}
                {state.isLoggedIn &&
                    <li key="all" className="">
                        <Link to="/courses" className="Sidebar-items-link">
                            View all courses
                        </Link>
                    </li>}
            </ul>
        </nav>
    )

}

export default NavFavorites
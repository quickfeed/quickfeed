import React from "react";
import { useActions, useAppState } from "../overmind";
import { Link, useHistory } from 'react-router-dom'
import { Enrollment } from "../../proto/ag/ag_pb";
import NavBarLabs from "./navbar/NavBarLabs";
import NavBarTeacher from "./navbar/NavBarTeacher";
import NavBarFooter from "./navbar/NavBarFooter";
import { Status } from "../consts";
import NavBarCourse from "./navbar/NavBarCourse";


//TODO Review the NavBar behaviour. 
const NavBar = (): JSX.Element => {
    const state = useAppState()
    
    const favorites = state.enrollments.filter(enrollment => enrollment.getStatus() >= Enrollment.UserStatus.STUDENT && enrollment.getState() == Enrollment.DisplayState.FAVORITE)

    const courses = favorites.map((enrollment) =>{
            if (enrollment.getState() >= Enrollment.DisplayState.VISIBLE) {
                return (
                    <NavBarCourse key={enrollment.getId()} enrollment={enrollment} />
                )
            }
    })

    return (
        <nav className="navigator">
            <ul key="list" className="SidebarList">
                <li key="logo" className="logo">
                    <Link to="/">
                        QuickFeed
                    </Link>
                </li>

                {courses}
                <li key="all" className="">
                    <Link to="/courses" className="Sidebar-items-link">
                        View all courses
                    </Link>
                </li>
                <NavBarFooter key="foot" />
            </ul>
        </nav>
    )
    
}

export default NavBar
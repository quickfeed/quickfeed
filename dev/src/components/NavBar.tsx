import React from "react"
import { useAppState } from "../overmind"
import { Link } from 'react-router-dom'
import NavBarFooter from "./navbar/NavBarFooter"
import NavBarCourse from "./navbar/NavBarCourse"
import { isEnrolled, isFavorite, isVisible } from "../Helpers"


//TODO Review the NavBar behaviour.
const NavBar = (): JSX.Element => {
    const state = useAppState()

    const visible = state.enrollments.filter(enrollment => isEnrolled(enrollment) && isVisible(enrollment))

    const courses = visible.map((enrollment) => {
        return <NavBarCourse key={enrollment.getId()} enrollment={enrollment} />
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
                {state.isLoggedIn &&
                    <li key="all" className="">
                        <Link to="/courses" className="Sidebar-items-link">
                            View all courses
                        </Link>
                    </li>}
                <NavBarFooter key="foot" />
            </ul>
        </nav>
    )

}

export default NavBar

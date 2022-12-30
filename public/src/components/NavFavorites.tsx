import React from "react"
import { useAppState } from "../overmind"
import { Link } from "react-router-dom"
import NavBarCourse from "./navbar/NavBarCourse"
import { isEnrolled, isVisible } from "../Helpers"

const NavFavorites = (): JSX.Element => {
    const state = useAppState()

    const visible = state.enrollments.filter(enrollment => isEnrolled(enrollment) && isVisible(enrollment))

    const courses = visible.map((enrollment) => {
        return <NavBarCourse key={enrollment.ID.toString()} enrollment={enrollment} />
    })

    return (
        <nav className={`navigator ${state.showFavorites ? "" : "hidden"}`}>
            <ul key="list" className="sidebarList">
                {courses}
                {state.isLoggedIn &&
                    <li key="all" className="courseLink">
                        <Link to="/courses" className="sidebar-items-link">
                            View all courses
                        </Link>
                    </li>
                }
            </ul>
        </nav>
    )
}

export default NavFavorites

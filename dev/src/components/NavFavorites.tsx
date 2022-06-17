import React from "react"
import { useAppState, useActions } from "../overmind"
import { Link } from "react-router-dom"
import NavBarCourse from "./navbar/NavBarCourse"
import { isEnrolled, isVisible } from "../Helpers"

const NavFavorites = (): JSX.Element => {
    const state = useAppState()
    const actions = useActions()

    const visible = state.enrollments.filter(enrollment => isEnrolled(enrollment) && isVisible(enrollment))

    const courses = visible.map((enrollment) => {
        return <NavBarCourse key={enrollment.id} enrollment={enrollment} />
    })

    return (
        <nav className="navigator">
            <ul key="list" className="sidebarList">
                <li key="logo" className="logo" style={{ paddingLeft: "40px" }}>
                    <Link to="/" >
                        QuickFeed
                    </Link>
                    <a onClick={() => actions.toggleFavorites()} role="button" className="closeButton">✖</a>
                </li>
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

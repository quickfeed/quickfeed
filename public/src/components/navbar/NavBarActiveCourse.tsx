import React from "react"
import { Link } from "react-router-dom"
import { useAppState } from "../../overmind"
import CourseFavoriteButton from "../CourseFavoriteButton"
import ToggleSwitch from "./ToggleSwitch"



const NavBarActiveCourse = (): JSX.Element | null => {
    const activeCourse = useAppState((state) => state.activeCourse ? (state.courses.find((c) => c.ID === state.activeCourse) ?? null) : null)
    const enrollment = useAppState((state) => state.enrollmentsByCourseID[state.activeCourse?.toString() ?? ""] ?? null)
    const {isLoggedIn} = useAppState()
    
    if (!isLoggedIn || !activeCourse || !enrollment) {
        return null
    }

    return (
        <div className="nav-child course">
            <Link to={`/course/${activeCourse?.ID}`} className="nav-link">{activeCourse?.name}</Link>
            <CourseFavoriteButton enrollment={enrollment} style={{ "paddingRight": "20px" }} />
            <ToggleSwitch />
        </div>
    )
}

export default NavBarActiveCourse

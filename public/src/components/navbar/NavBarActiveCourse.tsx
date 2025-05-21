import React from "react"
import { useAppState } from "../../overmind"
import CourseFavoriteButton from "../CourseFavoriteButton"
import ToggleSwitch from "./ToggleSwitch"
import Breadcrumbs from "./Breadcrumbs"
import { useLocation } from "react-router"

const NavBarActiveCourse = () => {
    const location = useLocation()
    const activeCourse = useAppState((state) => state.activeCourse ? (state.courses.find((c) => c.ID === state.activeCourse) ?? null) : null)
    const enrollment = useAppState((state) => state.enrollmentsByCourseID[state.activeCourse?.toString() ?? ""] ?? null)
    const { isLoggedIn } = useAppState()

    if (!isLoggedIn || !activeCourse || !enrollment || location.pathname === "/") {
        return null
    }

    return (
        <div className="nav-child course">
            <Breadcrumbs />
            <CourseFavoriteButton enrollment={enrollment} style={{ "paddingRight": "20px" }} />
            <ToggleSwitch />
        </div>
    )
}

export default NavBarActiveCourse

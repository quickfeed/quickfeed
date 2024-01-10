import React from "react"
import { hasTeacher } from "../../Helpers"
import { useActions, useAppState } from "../../overmind"
import CourseFavoriteButton from "../CourseFavoriteButton"
import ToggleSwitch from "./ToggleSwitch"
import Breadcrumbs from "./Breadcrumbs"



const NavBarActiveCourse = (): JSX.Element | null => {
    const activeCourse = useAppState((state) => state.activeCourse ? (state.courses.find((c) => c.ID === state.activeCourse) ?? null) : null)
    const enrollment = useAppState((state) => state.enrollmentsByCourseID[state.activeCourse?.toString() ?? ""] ?? null)
    const { isLoggedIn, status } = useAppState()
    const actions = useActions()

    if (!isLoggedIn || !activeCourse || !enrollment) {
        return null
    }

    let toggleSwitch = null
    if (hasTeacher(status[activeCourse.ID.toString()])) {
        toggleSwitch = <ToggleSwitch click={() => { actions.changeView(enrollment.courseID) }} />
    }

    return (
        <div className="nav-child course">
            <Breadcrumbs />
            <CourseFavoriteButton enrollment={enrollment} style={{ "paddingRight": "20px" }} />
            {toggleSwitch}
        </div>
    )
}

export default NavBarActiveCourse

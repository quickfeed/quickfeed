import React from "react"
import { useAppState } from "../overmind"
import CourseFavoriteButton from "./CourseFavoriteButton"
import RoleSwitch from "./teacher/RoleSwitch"
import { useCourseID } from "../hooks/useCourseID"


// TODO: Maybe add route specific information, ex. if user is viewing a lab, show that in the banner. Could use state in components to display.
// TODO(jostein): This information could possibly be shown in the navbar.
const CourseBanner = () => {
    const state = useAppState()
    const enrollment = state.enrollmentsByCourseID[useCourseID().toString()]

    return (
        <div className="jumbotron">
            <div className="centerblock container">
                <h1>
                    {enrollment.course?.name}
                    <CourseFavoriteButton enrollment={enrollment} style={{ "padding": "20px" }} />
                </h1>
                <RoleSwitch enrollment={enrollment} />
            </div>
        </div>
    )
}

export default CourseBanner

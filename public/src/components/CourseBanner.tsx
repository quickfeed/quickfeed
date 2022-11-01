import React from "react"
import { getCourseID } from "../Helpers"
import { useAppState } from "../overmind"
import CourseFavoriteButton from "./CourseFavoriteButton"
import RoleSwitch from "./teacher/RoleSwitch"


// TODO: Maybe add route specific information, ex. if user is viewing a lab, show that in the banner. Could use state in components to display.
// TODO(jostein): This information could possibly be shown in the navbar.
const CourseBanner = (): JSX.Element => {
    const state = useAppState()
    const enrollment = state.enrollmentsByCourseID[getCourseID().toString()]

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

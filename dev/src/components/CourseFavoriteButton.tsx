import React from "react"
import { Enrollment } from "../../proto/ag/ag_pb"
import { isVisible } from "../Helpers"
import { useActions } from "../overmind"


// CourseFavoriteButton is a component that displays a button to toggle the favorite status of a course.
const CourseFavoriteButton = ({ enrollment, style }: { enrollment: Enrollment.AsObject, style: React.CSSProperties }) => {
    const actions = useActions()

    return (
        <span style={style}>
            <i className={isVisible(enrollment) ? 'fa fa-star' : "fa fa-star-o"}
                onClick={() => actions.setEnrollmentState(enrollment)}></i>
        </span>
    )
}

export default CourseFavoriteButton

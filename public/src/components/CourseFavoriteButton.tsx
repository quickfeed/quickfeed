import React from "react"
import { Enrollment } from "../../proto/qf/types_pb"
import { isVisible } from "../Helpers"
import { useActions } from "../overmind"


// CourseFavoriteButton is a component that displays a button to toggle the favorite status of a course.
const CourseFavoriteButton = ({ enrollment, style }: { enrollment: Enrollment, style: React.CSSProperties }) => {
    const actions = useActions()

    return (
        // TODO: Consider creating a tooltip component.
        <span style={style} title="Favorite or unfavorite this course. Favorite courses will appear on your dashboard.">
            <i className={isVisible(enrollment) ? 'fa fa-star' : "fa fa-star-o"}
                onClick={() => actions.setEnrollmentState(enrollment)} />
        </span>
    )
}

export default CourseFavoriteButton

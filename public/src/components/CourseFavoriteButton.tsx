import React from "react"
import { Enrollment } from "../../proto/qf/types_pb"
import { isVisible } from "../Helpers"
import { useActions, useAppState } from "../overmind"


// CourseFavoriteButton is a component that displays a button to toggle the favorite status of a course.
const CourseFavoriteButton = ({ enrollment, style }: { enrollment: Enrollment, style: React.CSSProperties }) => {
    const actions = useActions().global
    // Calling useAppState to ensure Overmind tracks this component within NavbarActiveCourse
    // Not having this will cause the component to not re-render when star is clicked
    // TODO(jostein): This is a workaround, but I've not found out *why* the component does not re-render otherwise.
    useAppState()

    const starIcon = isVisible(enrollment) ? 'fa fa-star' : "fa fa-star-o"
    return (
        // TODO: Consider creating a tooltip component.
        <span style={style} title="Favorite or unfavorite this course. Favorite courses will appear on your dashboard.">
            <i role="button" aria-hidden="true" className={starIcon} onClick={() => actions.setEnrollmentState(enrollment)} />
        </span>
    )
}

export default CourseFavoriteButton

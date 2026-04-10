import React from "react"
import { Enrollment } from "../../proto/qf/types_pb"
import { isVisible } from "../Helpers"
import { useActions, useAppState } from "../overmind"


// CourseFavoriteButton is a component that displays a button to toggle the favorite status of a course.
const CourseFavoriteButton = ({ enrollment, className }: { enrollment: Enrollment, className?: string }) => {
    const actions = useActions().global
    // Calling useAppState to ensure Overmind tracks this component within NavbarActiveCourse
    // Not having this will cause the component to not re-render when star is clicked
    // TODO(jostein): This is a workaround, but I've not found out *why* the component does not re-render otherwise.
    useAppState()

    const starred = isVisible(enrollment)
    return (
        <button
            className={`font-mono text-sm cursor-pointer tooltip tooltip-bottom ${className ?? ''}`}
            data-tip="Favorite or unfavorite this course. Favorite courses will appear on your dashboard."
            onClick={() => actions.setEnrollmentState(enrollment)}
        >
            {starred
                ? <span className="font-semibold text-md"><i className="fa fa-star" /></span>
                : <span className="opacity-40"><span className="font-semibold"><i className="fa fa-star" /></span></span>
            }
        </button>
    )
}

export default CourseFavoriteButton

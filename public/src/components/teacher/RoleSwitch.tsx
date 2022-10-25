import React from "react"
import { Enrollment } from "../../../gen/qf/types_pb"
import { hasTeacher, isTeacher } from "../../Helpers"
import { useActions, useAppState } from "../../overmind"


// RoleSwitch is a component that displays a button to switch between teacher and student roles.
const RoleSwitch = ({ enrollment }: { enrollment: Enrollment }) => {
    const state = useAppState()
    const actions = useActions()

    if (hasTeacher(state.status[enrollment.courseID.toString()])) {
        return (
            <span className="clickable" onClick={() => actions.changeView(enrollment.courseID)}>
                {isTeacher(enrollment) ? "Switch to Student View" : "Switch to Teacher View"}
            </span>
        )
    }
    return null
}

export default RoleSwitch

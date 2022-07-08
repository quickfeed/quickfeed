import React from "react"
import { Enrollment } from "../../../proto/qf/types_pb"
import { hasTeacher, isTeacher } from "../../Helpers"
import { useActions, useAppState } from "../../overmind"


// RoleSwitch is a component that displays a button to switch between teacher and student roles.
const RoleSwitch = ({ enrollment }: { enrollment: Enrollment.AsObject }) => {
    const state = useAppState()
    const actions = useActions()

    if (hasTeacher(state.status[enrollment.courseid])) {
        return (
            <span className="clickable" onClick={() => actions.changeView(enrollment.courseid)}>
                {isTeacher(enrollment) ? "Switch to Student View" : "Switch to Teacher View"}
            </span>
        )
    }
    return null
}

export default RoleSwitch

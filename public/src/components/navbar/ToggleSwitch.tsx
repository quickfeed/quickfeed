import React, { useEffect } from "react"
import { useNavigate } from "react-router"
import { Enrollment_UserStatus } from "../../../proto/qf/types_pb"
import { hasTeacher } from "../../Helpers"
import { useActions, useAppState } from "../../overmind"

const ToggleSwitch = () => {
    const { activeCourse, enrollmentsByCourseID, status } = useAppState()
    const actions = useActions().global
    const navigate = useNavigate()
    const [enrollmentStatus, setEnrollmentStatus] =
        React.useState<boolean>(false)

    useEffect(() => {
        if (activeCourse && enrollmentsByCourseID[activeCourse.toString()]) {
            setEnrollmentStatus(isTeacher())
        }
    })

    const isTeacher = () => {
        return (
            enrollmentsByCourseID[activeCourse.toString()].status ===
            Enrollment_UserStatus.TEACHER
        )
    }

    const switchView = () => {
        actions.changeView().then(() => {
            setEnrollmentStatus(isTeacher())
            navigate(`/course/${activeCourse}`)
        })
    }

    if (!hasTeacher(status[activeCourse.toString()])) {
        return null
    }

    return (
        <button
            onClick={switchView}
            className="font-mono text-md cursor-pointer tooltip tooltip-bottom"
            data-tip="Toggle between student and teacher view"
        >
            {enrollmentStatus
                ? <span><span className="text-primary font-semibold">#</span> <span className="text-base-content/60">teacher</span></span>
                : <span><span className="text-primary font-semibold">$</span> <span className="text-base-content/60">student</span></span>
            }
        </button>
    )
}

export default ToggleSwitch

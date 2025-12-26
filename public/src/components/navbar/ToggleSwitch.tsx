import React, { useEffect } from "react"
import { Enrollment_UserStatus } from "../../../proto/qf/types_pb"
import { useActions, useAppState } from "../../overmind"
import { hasTeacher } from "../../Helpers"
import { useNavigate } from "react-router"

const ToggleSwitch = () => {
    const { activeCourse, enrollmentsByCourseID, status } = useAppState()
    const actions = useActions().global
    const navigate = useNavigate()
    const [enrollmentStatus, setEnrollmentStatus] =
        React.useState<boolean>(false)
    const [text, setText] = React.useState<string>("")

    useEffect(() => {
        if (activeCourse && enrollmentsByCourseID[activeCourse.toString()]) {
            updateStatus(isTeacher())
        }
    })

    const isTeacher = () => {
        return (
            enrollmentsByCourseID[activeCourse.toString()].status ===
            Enrollment_UserStatus.TEACHER
        )
    }

    const updateStatus = (isTeacher: boolean) => {
        setEnrollmentStatus(isTeacher)
        setText(isTeacher ? "T" : "S")
    }

    const switchView = () => {
        actions.changeView().then(() => {
            updateStatus(isTeacher())
            navigate(`/course/${activeCourse}`)
        })
    }

    if (!hasTeacher(status[activeCourse.toString()])) {
        return null
    }

    return (
        <label className="switch" data-toggle="tooltip" title="Toggle between student and teacher view">
            <input type="checkbox" readOnly checked={enrollmentStatus} />
            <span className="slider round" onClick={switchView}>
                <span className="toggleCircle">{text}</span>
            </span>
        </label>
    )
}

export default ToggleSwitch

import React, { useCallback, useEffect } from "react"
import { Enrollment_UserStatus } from "../../../proto/qf/types_pb"
import { useActions, useAppState } from "../../overmind"
import { hasTeacher } from "../../Helpers"
import { useHistory } from "react-router"

const ToggleSwitch = () => {
    const { activeCourse, enrollmentsByCourseID, status } = useAppState()
    const actions = useActions()
    const navigate = useHistory()
    const [enrollmentStatus, setEnrollmentStatus] = React.useState<boolean>(false)
    const [text, setText] = React.useState<string>("")
    const isTeacher = enrollmentsByCourseID[activeCourse.toString()].status === Enrollment_UserStatus.TEACHER

    useEffect(() => {
        if (activeCourse && enrollmentsByCourseID[activeCourse.toString()]) {
            updateStatus(isTeacher)
        }
    })

    const updateStatus = (isTeacher: boolean) => {
        setEnrollmentStatus(isTeacher)
        setText(isTeacher ? "T" : "S")
    }

    const switchView = useCallback(() => {
        actions.changeView().then(() => {
            updateStatus(isTeacher)
            navigate.push("/course/" + activeCourse)
        })
    }, [actions, activeCourse, isTeacher, navigate])

    if (!hasTeacher(status[activeCourse.toString()])) {
        return null
    }

    return (
        <label className="switch" data-toggle="tooltip" title="Toggle between student and teacher view">
            <input type="checkbox" readOnly checked={enrollmentStatus} />
            <span className="slider round" onClick={switchView}> {/* skipcq: JS-0417 */}
                <span className="toggle">{text}</span>
            </span>
        </label>
    )
}

export default ToggleSwitch

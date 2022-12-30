import React from 'react'
import { Enrollment_UserStatus } from '../../../proto/qf/types_pb'
import { useAppState } from '../../overmind'

const ToggleSwitch = ({ click}: {click: () => void}) => {
    const {activeCourse, enrollmentsByCourseID} = useAppState()

    const isTeacher = React.useMemo(() => {
        return enrollmentsByCourseID[activeCourse?.toString() ?? ""]?.status === Enrollment_UserStatus.TEACHER ?? false
    }, [activeCourse, enrollmentsByCourseID])

    const text = isTeacher ? "T" : "S"

    return (
        <label className="switch">
            <input type="checkbox" readOnly checked={isTeacher} />
            <span className="slider round" onClick={click} >
                <span className="toggle">{text}</span>
            </span>
        </label>
    )
}

export default ToggleSwitch
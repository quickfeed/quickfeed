import React, { useEffect } from 'react'
import { Enrollment_UserStatus } from '../../../proto/qf/types_pb'
import { useActions, useAppState } from '../../overmind'
import { hasTeacher } from '../../Helpers'
import { useHistory } from 'react-router'

const ToggleSwitch = () => {
    const {activeCourse, enrollmentsByCourseID, status} = useAppState()
    const actions = useActions()
    const navigate = useHistory()
    const [enrollmentStatus, setEnrollmentStatus] = React.useState<boolean>(false)
    const [text, setText] = React.useState<string>("")

    useEffect(() => {
        if (activeCourse && enrollmentsByCourseID[activeCourse.toString()]) {
            updateStatus(isTeacher());
        }
    });

    const isTeacher = () => {
        return enrollmentsByCourseID[activeCourse.toString()].status === Enrollment_UserStatus.TEACHER;
    }

    const updateStatus = (isTeacher: boolean) => {
        setEnrollmentStatus(isTeacher);
        setText(isTeacher ? "T" : "S");
    }

    const switchView = () => {
        actions.changeView(activeCourse).then(() => {
            updateStatus(isTeacher());
            navigate.push("/course/" + activeCourse);
        })
    }
    
    if  (!hasTeacher(status[activeCourse.toString()])) {
        return null
    }

    return (
        <label className="switch">
            <input type="checkbox" readOnly checked={enrollmentStatus} />
            <span className="slider round" onClick={switchView} >
                <span className="toggle">{text}</span>
            </span>
        </label>
    )
}

export default ToggleSwitch
import React, { useCallback } from "react"
import { useHistory } from "react-router"
import { Enrollment } from "../../../proto/qf/types_pb"
import { Status } from "../../consts"
import { isStudent, isTeacher } from "../../Helpers"
import { useActions, useAppState } from "../../overmind"
import NavBarLabs from "./NavBarLabs"
import NavBarTeacher from "./NavBarTeacher"


const NavBarCourse = ({ enrollment }: { enrollment: Enrollment }) => {
    const state = useAppState()
    const actions = useActions()
    const history = useHistory()

    // Determines if a dropdown should be shown for the course
    const active = state.activeCourse === enrollment.courseID
    const course = state.courses.find(c => c.ID === enrollment.courseID)

    const navigateTo = useCallback(() => {
        if (active) {
            // Collapse active course dropdown
            actions.setActiveCourse(BigInt(0))
            history.push("/")
        } else {
            history.push(`/course/${enrollment.courseID}`)
            actions.setActiveCourse(enrollment.courseID)
        }
    }, [actions, active, enrollment.courseID, history])

    return (
        <>
            <div role="button" onClick={navigateTo} aria-hidden="true"> {/* skipcq: JS-0417 */}
                <li className="activeClass">
                    <div className="col" id="title">
                        {course?.code}
                    </div>
                    <div className="col" title="icon">
                        <i className={active ? " icon fa fa-caret-down fa-lg float-right" : " icon fa fa-caret-down fa-rotate-90 fa-lg float-right"} />
                    </div>
                </li>
            </div>
            <div className={active ? Status.ActiveLab : Status.Inactive}>
                {active && isStudent(enrollment) ? <NavBarLabs /> : null}
                {active && isTeacher(enrollment) ? <NavBarTeacher /> : null}
            </div>
        </>
    )
}

export default NavBarCourse

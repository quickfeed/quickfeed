import React from "react"
import { useHistory } from "react-router"
import { Enrollment } from "../../../proto/qf/types/types_pb"
import { Status } from "../../consts"
import { isStudent, isTeacher } from "../../Helpers"
import { useActions, useAppState } from "../../overmind"
import NavBarLabs from "./NavBarLabs"
import NavBarTeacher from "./NavBarTeacher"


const NavBarCourse = ({ enrollment }: { enrollment: Enrollment.AsObject }): JSX.Element => {
    const state = useAppState()
    const actions = useActions()
    const history = useHistory()
    // Determines if a dropdown should be shown for the course
    const active = state.activeCourse === enrollment.courseid

    const navigateTo = (courseID: number) => {
        if (active) {
            // Collapse active course dropdown
            actions.setActiveCourse(0)
            history.push("/")
        } else {
            history.push("/course/" + courseID)
            actions.setActiveCourse(courseID)
        }
    }

    return (
        <>
            <li role={"button"} onClick={() => navigateTo(enrollment.courseid)} className="activeClass">
                <div className="col" id="title">
                    {enrollment.course?.code}
                </div>
                <div className="col" title="icon">
                    <i className={active ? " icon fa fa-caret-down fa-lg float-right" : " icon fa fa-caret-down fa-rotate-90 fa-lg float-right"} />
                </div>
            </li>
            <div className={active ? Status.ActiveLab : Status.Inactive}>
                {active && isStudent(enrollment) ? <NavBarLabs /> : null}
                {active && isTeacher(enrollment) ? <NavBarTeacher /> : null}
            </div>
        </>
    )
}

export default NavBarCourse

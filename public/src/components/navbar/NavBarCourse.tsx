import React from "react"
import { useNavigate } from "react-router"
import { Enrollment } from "../../../proto/qf/types_pb"
import { Status } from "../../consts"
import { isStudent, isTeacher } from "../../Helpers"
import { useActions, useAppState } from "../../overmind"
import NavBarLabs from "./NavBarLabs"
import NavBarTeacher from "./NavBarTeacher"


const NavBarCourse = ({ enrollment }: { enrollment: Enrollment }) => {
    const state = useAppState()
    const actions = useActions()
    const navigate = useNavigate()
    // Determines if a dropdown should be shown for the course
    const active = state.activeCourse === enrollment.courseID
    const course = state.courses.find(c => c.ID === enrollment.courseID)

    const navigateTo = (courseID: bigint) => {
        if (active) {
            // Collapse active course dropdown
            actions.setActiveCourse(BigInt(0))
            navigate("/")
        } else {
            actions.setActiveCourse(courseID)
            navigate(`/course/${courseID}`)
        }
    }

    return (
        <>
            <div role="button" onClick={() => navigateTo(enrollment.courseID)} aria-hidden="true">
                <li className="activeClass">
                    <div className="col" id="title">
                        {course?.code}
                    </div>
                    <div className="col" title="icon">
                        <i className={`icon fa fa-caret-down fa-lg float-right ${active ? "" : "fa-rotate-90"}`} />
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

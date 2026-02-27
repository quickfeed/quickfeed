import React from "react"
import { useNavigate } from "react-router"
import { Enrollment } from "../../../proto/qf/types_pb"
import { isStudent, isTeacher } from "../../Helpers"
import { useActions, useAppState } from "../../overmind"
import NavBarLabs from "./NavBarLabs"
import NavBarTeacher from "./NavBarTeacher"


const NavBarCourse = ({ enrollment }: { enrollment: Enrollment }) => {
    const state = useAppState()
    const actions = useActions().global
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
        <li className="w-full">
            <button
                onClick={() => navigateTo(enrollment.courseID)}
                className="flex justify-between items-center w-full h-16 px-4 font-bold hover:bg-base-100 cursor-pointer rounded-none"
            >
                <span>{course?.code}</span>
                <i className={`fa fa-caret-down fa-lg transition-transform ${active ? "" : "rotate-90"}`} />
            </button>
            {active && (
                <ul className="menu p-0 bg-base-200 w-full m-0">
                    {isStudent(enrollment) && <NavBarLabs />}
                    {isTeacher(enrollment) && <NavBarTeacher />}
                </ul>
            )}
        </li>
    )
}

export default NavBarCourse

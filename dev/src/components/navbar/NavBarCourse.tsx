import React from "react";
import { useHistory } from "react-router";
import { Enrollment } from "../../../proto/ag/ag_pb";
import { Status } from "../../consts";
import { useActions, useAppState } from "../../overmind";
import NavBarLabs from "./NavBarLabs";
import NavBarTeacher from "./NavBarTeacher";

const NavBarCourse = ({enrollment}: {enrollment: Enrollment}): JSX.Element => {
    const state = useAppState()
    const actions = useActions()
    const history = useHistory()
    // If selected, used to show teacher / student navbar for that course
    const selected = state.activeCourse === enrollment.getCourseid()
    
    const onCourseClick = (enrollment: Enrollment) => {
        if (selected) {
            // Unselect the active course
            actions.setActiveCourse(0)
            history.push("/")
        }
        else {
            history.push(`/course/` + enrollment.getCourseid())
            actions.setActiveCourse(enrollment.getCourseid())
        }
    }


    return (
        <>
            <li onClick={() => {onCourseClick(enrollment)}}>
                <div className="col" id="title">
                    {enrollment.getCourse()?.getCode()}
                </div> 
                <div className="col" title="icon">
                    <i className={selected ? "icon fa fa-caret-down fa-lg float-right" : "icon fa fa-caret-down fa-rotate-90 fa-lg float-right"}></i>
                </div>
            </li>
            <div className={ selected ? Status.ActiveLab : Status.Inactive}>
                {selected && enrollment.getStatus() === Enrollment.UserStatus.STUDENT ? <NavBarLabs /> : null}
                {selected && enrollment.getStatus() === Enrollment.UserStatus.TEACHER ? <NavBarTeacher /> : null}
            </div>
        </>
    )
}

export default NavBarCourse
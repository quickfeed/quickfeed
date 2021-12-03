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
    
    const onCourseClick = (enrollment: Enrollment) => {
        if (enrollment.getCourseid() === state.activeCourse) {
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
            <li key={`code-${enrollment.getId()}`} className=""  onClick={() => {onCourseClick(enrollment)}}>
                <div className="col" id="title">
                    {enrollment.getCourse()?.getCode()}
                </div> 
                <div className="col" title="icon">
                    <i className={state.activeCourse === enrollment.getCourseid() ? "icon fa fa-caret-down fa-lg float-right" : "icon fa fa-caret-down fa-rotate-90 fa-lg float-right"}></i>
                </div>
            </li>
            <div key={`links-${enrollment.getId()}`} className={ state.activeCourse === enrollment.getCourseid()  ? Status.ActiveLab : Status.Inactive}>
                {state.activeCourse === enrollment.getCourseid() && enrollment.getStatus() === Enrollment.UserStatus.STUDENT ? <NavBarLabs key={`labs-${enrollment.getId()}`} /> : null}
                {state.activeCourse === enrollment.getCourseid() && enrollment.getStatus() === Enrollment.UserStatus.TEACHER ? <NavBarTeacher key={`teacher-${enrollment.getId()}`}  courseID={enrollment.getCourseid()}/> : null}
            </div>
        </>
    )
}

export default NavBarCourse
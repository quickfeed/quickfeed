import React from "react";
import { useActions, useAppState } from "../overmind";
import { Link, useHistory } from 'react-router-dom'
import { Enrollment } from "../../proto/ag/ag_pb";
import NavBarLabs from "./navbar/NavBarLabs";
import NavBarTeacher from "./navbar/NavBarTeacher";
import NavBarFooter from "./navbar/NavBarFooter";
import { Status } from "../consts";


//TODO Review the NavBar behaviour. 
//! Source of a key error
const NavBar = (): JSX.Element => {
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

    const CourseItems = (): JSX.Element[] | null => {
        const links: JSX.Element[] = []
        if (state.self.getId() <= 0) {
            return null
        }
        const favorites = state.enrollments.filter(enrollment => enrollment.getStatus() >= Enrollment.UserStatus.STUDENT && enrollment.getState() == Enrollment.DisplayState.FAVORITE)

        favorites.map((enrollment) =>{
            if (enrollment.getState() >= Enrollment.DisplayState.VISIBLE)
                links.push(
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
                })

        
        return links
    }
    return (
        <nav className="navigator">
            <ul key="list" className="SidebarList">
                <li key="logo" className="logo">
                    <Link to="/">
                        QuickFeed
                    </Link>
                </li>

                {CourseItems()}
                <li key="all" className="">
                    <Link to="/courses" className="Sidebar-items-link">
                        View all courses
                    </Link>
                </li>
                
                
                <NavBarFooter key="foot" />
            </ul>
        </nav>
    )
    
}

export default NavBar
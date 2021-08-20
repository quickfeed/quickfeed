import React, { useEffect, useState } from "react";
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

    const [active, setActive] = useState(0)
    const [showCourses, setShowCourses] = useState(false)
    
    useEffect(() => {
        if (state.activeCourse > 0) {
            setActive(state.activeCourse)
            setShowCourses(false)
        }
    }, [state.activeCourse])

    
    const onCourseClick = (enrollment: Enrollment) => {
        history.push(`/course/` + enrollment.getCourseid())
        setShowCourses(false)
        actions.setActiveCourse(enrollment.getCourseid())
    }

    const CourseItems = (): JSX.Element[] | null => {
        const links: JSX.Element[] = []
        if (state.self.getId() <= 0) {
            return null
        }
        const favorites = state.enrollments.filter(enrollment => enrollment.getStatus() >= Enrollment.UserStatus.STUDENT && enrollment.getState() == Enrollment.DisplayState.FAVORITE)

        favorites.map((enrollment) =>{
                links.push(
                    <>
                        <li key={`code-${enrollment.getId()}`} className={showCourses || active === enrollment.getCourseid() ? Status.Active : Status.Inactive}  onClick={() => {onCourseClick(enrollment)}}>
                            <div>
                                {enrollment.getCourse()?.getCode()}
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

                <li key="courses" onClick={() => { setShowCourses(!showCourses); actions.setActiveCourse(-1)}}>
                    <div id="title">
                            Courses &nbsp;&nbsp;
                        <i className={showCourses ? "icon fa fa-caret-down fa-lg" : "icon fa fa-caret-down fa-rotate-90 fa-lg"}></i>
                    </div>
                </li>
                <li key="all" className={showCourses ? Status.Active : Status.Inactive}>
                    <Link to="/courses" className="Sidebar-items-link">
                        View all courses
                    </Link>
                </li>
                
                
                {CourseItems()}
                <NavBarFooter key="foot" />
            </ul>
        </nav>
    )
    
}

export default NavBar
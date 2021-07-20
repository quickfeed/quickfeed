import React, { useEffect, useState } from "react";
import { useOvermind } from "../overmind";
import { Link, useHistory } from 'react-router-dom'
import { Enrollment } from "../../proto/ag/ag_pb";
import NavBarLabs from "./navbar/NavBarLabs";
import NavBarTeacher from "./navbar/NavBarTeacher";
import NavBarFooter from "./navbar/NavBarFooter";
import { Status } from "../consts";



const NavBar = () => {
    const { state, actions } = useOvermind() 
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

    // Generates dropdown items related to Courses
    const CourseItems: Function = (): JSX.Element[] => {
        const links: JSX.Element[] = []
        if (state.self.getId() <= 0) {
            return links
        }
        const favorites = state.enrollments.filter(enrollment => enrollment.getStatus() >= Enrollment.UserStatus.STUDENT && enrollment.getState() == Enrollment.DisplayState.FAVORITE)
        links.push(
            <div>
                <li onClick={() => { setShowCourses(!showCourses); actions.setActiveCourse(-1)}}>
                    <div id="title">
                            Courses &nbsp;&nbsp;
                        <i className={showCourses ? "icon fa fa-caret-down fa-lg" : "icon fa fa-caret-down fa-rotate-90 fa-lg"}></i>
                    </div>
                </li>
                <li className={showCourses ? Status.Active : Status.Inactive}>
                    <Link to="/courses" className="Sidebar-items-link">
                        View all courses
                    </Link>
                </li>
            </div>
        )

        favorites.map((enrollment) =>{
                links.push(
                    <React.Fragment key={enrollment.getId()}>
                        <li className={showCourses || active === enrollment.getCourseid() ? Status.Active : Status.Inactive}  onClick={() => {onCourseClick(enrollment)}}>
                            <div>
                                {enrollment.getCourse()?.getCode()}
                            </div> 
                        </li>
                        <div className={ state.activeCourse === enrollment.getCourseid()  ? Status.ActiveLab : Status.Inactive}>
                            {state.activeCourse === enrollment.getCourseid() && enrollment.getStatus() === Enrollment.UserStatus.STUDENT ? <NavBarLabs /> : null}
                            {state.activeCourse === enrollment.getCourseid() && enrollment.getStatus() === Enrollment.UserStatus.TEACHER ? <NavBarTeacher  courseID={enrollment.getCourseid()}/> : null}
                        </div>
                    </React.Fragment>
                    )
                })

        
        return links
    }
    return (
        <nav className="navigator">
            <ul className="SidebarList">
                <li key="logo" className="logo">
                    <Link to="/">
                        QuickFeed
                    </Link>
                </li>
                
                <CourseItems />
                <NavBarFooter />
            </ul>
        </nav>
    )
    
}

export default NavBar
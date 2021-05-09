import React, { useEffect, useState } from "react";
import { useOvermind } from "../overmind";
import { Link, useHistory } from 'react-router-dom'
import NavBarLabs from "./NavBarLabs";
import { Enrollment } from "../../proto/ag_pb";
import NavBarTeacher from "./NavBarTeacher";
import NavBarFooter from "./NavBarFooter";



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

    

    // Generates dropdown items related to Courses
    const CourseItems: Function = (): JSX.Element[] => {
        let links: JSX.Element[] = []
        if (state.user.id <= 0) {
            return links
        }
        const favorites = state.enrollments.filter(enrollment => enrollment.getStatus() >= Enrollment.UserStatus.STUDENT && enrollment.getState() == Enrollment.DisplayState.FAVORITE)
        links.push(
            <div>
                <li key={"courses"} onClick={() => { setShowCourses(!showCourses); actions.setActiveCourse(-1)}}>
                    <div id="title">
                            Courses &nbsp;&nbsp;
                        <i className={showCourses ? "icon fa fa-caret-down fa-lg" : "icon fa fa-caret-down fa-rotate-90 fa-lg"}></i>
                    </div>
                </li>
                <li key={"allCourses"} className={showCourses ? "active" : "inactive"}>
                    <div id="title">
                        <Link to="/courses">
                            View all courses
                        </Link>
                    </div>
                </li>
            </div>
        )

        favorites.map((enrollment) =>{
                links.push(
                    <React.Fragment>
                        <li key={enrollment.getCourseid()} className={showCourses || active === enrollment.getCourseid()  ? "active" : "inactive"}  onClick={() => {history.push(`/course/` + enrollment.getCourseid()); setShowCourses(false)}}>
                            <div id="title">
                                {enrollment.getCourse()?.getCode()}
                            </div> 
                        </li>
                        <div className={ state.activeCourse === enrollment.getCourseid()  ? "activelabs" : "inactive"}>
                            {state.activeCourse === enrollment.getCourseid() && enrollment.getStatus() === Enrollment.UserStatus.STUDENT ? <NavBarLabs /> : ""}
                            {state.activeCourse === enrollment.getCourseid() && enrollment.getStatus() === Enrollment.UserStatus.TEACHER ? <NavBarTeacher  courseID={enrollment.getCourseid()}/> : ""}
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
                        Autograder
                    </Link>
                </li>
                
                <CourseItems />
                <NavBarFooter />
            </ul>
        </nav>
    )
    
}

export default NavBar
import React, { useEffect } from "react"
import { useActions, useAppState } from "../overmind"
import { Enrollment } from "../../proto/ag/ag_pb"
import CourseCard from "./CourseCard"
import { AlertType } from "../Helpers"
import Card from "./Card"
import Button, { ButtonType, ComponentColor } from "./admin/Button"
import { useHistory } from "react-router"
interface overview {
    home: boolean
}

/** This component should list user courses, and available courses and allow enrollment */
const Courses = (overview: overview): JSX.Element => {
    const state = useAppState()
    const actions = useActions()
    const history = useHistory()

    useEffect(() => {
        if (state.enrollments.filter(enrollment => 
                enrollment.getState() == Enrollment.DisplayState.FAVORITE && 
                enrollment.getStatus() >= Enrollment.UserStatus.STUDENT).length == 0) 
            { actions.alert({text: "Favorite a course to make it appear on the dashboard and in the sidebar", type: AlertType.INFO}) }
        actions.setActiveCourse(-1)
    }, [])

    // Notify user if there are no courses (should only ever happen with a fresh database on backend)
    // Display shortcut buttons for admins to create new course or managing (promoting) users
    if (state.courses.length == 0) {
        return (
            <div className="container centered">
                <h3>There are currently no available courses.</h3>
                {state.self.getIsadmin() ? 
                    <>
                        <div>
                            <Button classname="mr-3" text="Go to course creation" color={ComponentColor.GREEN} type={ButtonType.BUTTON} onclick={() => history.push("/admin/create")} /> 
                            <Button text="Manage users" color={ComponentColor.BLUE} type={ButtonType.BUTTON} onclick={() => history.push("/admin/manage")} /> 
                        </div>
                    </>
                    : null}
            </div>
        )
    }

    // push to seperate arrays, for layout purposes. Favorite - Student - Teacher - Pending
    const upDateArrays = () => {
        const favorite:   JSX.Element[] = []
        const student:    JSX.Element[] = []
        const teacher:    JSX.Element[] = []
        const pending:    JSX.Element[] = []
        const crsArr:     JSX.Element[] = []
        state.courses.map(course => {       
            const enrol = state.enrollmentsByCourseId[course.getId()]
            if (enrol){
                if (enrol.getState() == Enrollment.DisplayState.FAVORITE){
                    // add to favorite list.
                    favorite.push(
                        <CourseCard key={course.getId()} course= {course} enrollment={enrol} status={enrol.getStatus()}/>
                    )
                } else {
                    switch (enrol.getStatus()){
                        //pending
                        case Enrollment.UserStatus.PENDING:
                            //color orange
                            pending.push(
                                <CourseCard key={course.getId()} course={course} enrollment={enrol} status={enrol.getStatus()}/>
                            )
                            break
                              
                        case Enrollment.UserStatus.STUDENT:
                            // Student
                            //color blue
                            student.push(
                                <CourseCard key={course.getId()} course={course} enrollment={enrol} status={enrol.getStatus()}/>
                            )
                            break
                        case Enrollment.UserStatus.TEACHER:
                            // color green
                            // Teacher
                            teacher.push(
                                <CourseCard key={course.getId()} course={course} enrollment={enrol} status={enrol.getStatus()}/>
                            )
                            break
                        default:
                            console.log("Something went wrong")
                            break
                    }
                }              
            }
            else {
                crsArr.push(
                    <CourseCard key={course.getId()} course= {course} enrollment={new Enrollment} status={Enrollment.UserStatus.NONE}/>
                )
            }
        })
        
        // If overview.home == true, only render favorited courses.
        if (overview.home) {
            return (
                <>
                {favorite.length > 0 &&
                    <div className="container-fluid">       
                        <div className="card-deck course-card-row favorite-row">
                            {favorite}    
                        </div>
                    </div>
                }
                </>
            )
        }

        return (
            <div className="box container-fluid">
                {favorite.length > 0 &&
                <div className="container-fluid">
                    <h2>Favorites</h2>
                    <div className="card-deck course-card-row favorite-row">
                        {favorite}
                    </div>
                </div>
                }
                    
                {(student.length > 0 || teacher.length > 0) &&
                    <div className="container-fluid">
                        <h2>My Courses</h2>
                        <div className="card-deck course-card-row">
                            {teacher}
                            {student}
                        </div>
                    </div>
                }
                {pending.length > 0 &&
                    <div className="container-fluid">
                        {(student.length==0 && teacher.length==0) &&
                            <h2>My Courses</h2>
                        }
                        <div className="card-deck">
                        {pending}
                        </div>
                    </div>
                }
                
                {crsArr.length > 0 &&
                    <React.Fragment>
                    <h2>Available Courses</h2>
                    <div className="card-deck course-card-row">
                    {crsArr}
                    </div>
                    </React.Fragment>
                }
            </div>
        
            
        )
    } 
    return upDateArrays()
        
}

export default Courses
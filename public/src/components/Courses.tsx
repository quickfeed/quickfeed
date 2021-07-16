import React, { useEffect, useState } from "react"
import { useOvermind } from "../overmind"
import { Course, Enrollment } from "../../proto/ag/ag_pb"
import CourseCard from "./CourseCard"
import { AlertType } from "../Helpers"




const EnrollmentStatus = {
    0: "None",
    1: "Pending",
    2: "Enrolled",
    3: "Teacher"
}

interface overview {
    home: boolean
}

/** This component should list user courses, and available courses and allow enrollment */
const Courses = (overview: overview) => {
    const {state, actions} = useOvermind()

    useEffect(() => {
        if (state.enrollments.filter(enrollment => 
                enrollment.getState() == Enrollment.DisplayState.FAVORITE && 
                enrollment.getStatus() >= Enrollment.UserStatus.STUDENT).length == 0) 
            { actions.alert({text: "Favorite a course to make it appear on the dashboard and in the sidebar", type: AlertType.INFO}) }
        actions.setActiveCourse(-1)
    }, [])
    // TODO: UserCourses contains elements describing a course that a user has an enrollment in, regardless of status currently. Need to figure out what UserStatus.NONE is used for
    // push to seperate arrays, for layout purposes. Favorite - Student - Teacher - Pending
    function upDateArrays(){
        let favorite:   JSX.Element[] = []
        let student:    JSX.Element[] = []
        let teacher:    JSX.Element[] = []
        let pending:    JSX.Element[] = []
        let crsArr:     JSX.Element[] = []
        let enrolArr = state.enrollments
        state.courses.map(course => {       
            let enrol = enrolArr.find(enrol => course.getId() == enrol.getCourseid())
            if (enrol){
                if (enrol.getState() == 3){
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
                                <CourseCard key={course.getId()} course= {course} enrollment={enrol} status={enrol.getStatus()}/>
                            )
                            break
                              
                        case Enrollment.UserStatus.STUDENT:
                            // Student
                            //color blue
                            student.push(
                                <CourseCard key={course.getId()} course= {course} enrollment={enrol} status={enrol.getStatus()}/>
                            )
                            break
                        case Enrollment.UserStatus.TEACHER:
                            // color green
                            // Teacher
                            teacher.push(
                                <CourseCard key={course.getId()} course= {course} enrollment={enrol} status={enrol.getStatus()}/>
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
                <React.Fragment>
                {favorite.length > 0 &&
                    <div className="container-fluid">       
                        <div className="card-deck course-card-row favorite-row">
                            {favorite}    
                        </div>
                    </div>
                }
                </React.Fragment>
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
                {(student.length == 0 && teacher.length == 0 && pending.length == 0 && favorite.length == 0) &&
                    <div className="container-fluid">
                        <h1>Seems Like you aren't enrolled in any courses </h1>
                        <h1>Find you course in the list below Maybe make this into an alert?</h1>
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
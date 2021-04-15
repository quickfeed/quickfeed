import React, { useEffect } from "react"
import { Link } from "react-router-dom"
import { useOvermind, useState } from "../overmind"
import { Course, Enrollment } from "../proto/ag_pb"
import CourseCard from "./CourseCard"




const EnrollmentStatus = {
    0: "None",
    1: "Pending",
    2: "Enrolled",
    3: "Teacher"
}

/** This component should list user courses, and available courses and allow enrollment */
const Courses = () => {
    const {state, actions} = useOvermind()

    useEffect(() => {
        // TODO: getCoursesByUser returns courses a user has an enrollment in. I thought a UserStatus = 0 (NONE) would be default, but apparently not.
        //
        actions.setActiveCourse(-1)
    })
    // TODO: UserCourses contains elements describing a course that a user has an enrollment in, regardless of status currently. Need to figure out what UserStatus.NONE is used for
    
    let crsArr:Course[] = []
    // push to seperate arrays, for layout purposes. Favorite - Student - Teacher - Pending
    function upDateArrays(){
        let favorite: JSX.Element[] = []
        let student: JSX.Element[] = []
        let teacher: JSX.Element[] = []
        let pending: JSX.Element[] = []
        let courseArr = state.courses
        state.enrollments.map(enrol => {
               
            let course = courseArr.find(course => course.getId() == enrol.getCourseid())
            if (course){
                courseArr =courseArr.filter(item => item !== course)
                if (enrol.getState()==3){
                    // add to favorite list.
                    favorite.push(
                        <CourseCard key={course.getId()} course= {course} enrollment={enrol} status={enrol.getStatus()}/>
                    )
                }else{
                    switch (enrol.getStatus()){
                        //pending
                        case 1:
                            //color orange
                            pending.push(
                                <CourseCard key={course.getId()} course= {course} enrollment={enrol} status={enrol.getStatus()}/>
                            )
                            break
                              
                        case 2:
                            // Student
                            //color blue
                            student.push(
                                <CourseCard key={course.getId()} course= {course} enrollment={enrol} status={enrol.getStatus()}/>
                            )
                            break
                        case 3:
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
        })
        crsArr = courseArr
        // create enroll modal, to enroll to new courses.
        return (
            <div>
                <h1>Favorites</h1>
                <div className="card-deck row favorite-row">
                    {favorite}
                    
                </div>
                <h1>Courses</h1>
                
                {(student.length>0 || teacher.length>0) &&
                    <div className="card-deck row">
                        {teacher}
                        {student}
                    </div>    
                }
                {pending.length>0 &&
                    <div className="card-deck row">
                    {pending}
                    </div>
                }  
            </div>
        
            
        )
    } 
    return upDateArrays()
        
}

export default Courses
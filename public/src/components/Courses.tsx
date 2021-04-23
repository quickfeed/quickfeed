import React, { useEffect, useState } from "react"
import { useOvermind } from "../overmind"
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
    const [displayModal, setDisplayModal] = useState(false) 
    useEffect(() => {
        // TODO: getCoursesByUser returns courses a user has an enrollment in. I thought a UserStatus = 0 (NONE) would be default, but apparently not.
        //
        actions.setActiveCourse(-1)
    })
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
                if (enrol.getState()==3){
                    // add to favorite list.
                    favorite.push(
                        <CourseCard key={course.getId()} course= {course} enrollment={enrol} status={enrol.getStatus()}/>
                    )
                }else{
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
        
        // create enroll modal, to enroll to new courses.
        return (
            <div className="container-fluid">
                {favorite.length >0 &&
                <div className="container-fluid">
                    <h1>Favorites</h1>
                    <div className="card-deck row favorite-row">
                        {favorite}
                        
                    </div>
                </div>
                }
                    
                {(student.length>0 || teacher.length>0) &&
                    <div className="container-fluid">
                        <h1>My Courses</h1>
                        <div className="card-deck row">
                            {teacher}
                            {student}
                        </div>
                    </div>
                }
                {pending.length>0 &&
                    <div className="container-fluid">
                        <div className="card-deck row">
                        {pending}
                        </div>
                    </div>
                }
                {(student.length==0 && teacher.length==0 && pending.length==0) &&
                    <div className="container-fluid">
                        <h1>Seems Like you aren't enrolled in any courses </h1>
                        <h1>Find you course in the list below Maybe make this into an alert?</h1>
                    </div>
                }
                <h2>All courses  // Enrol in a new Course</h2>
                {crsArr.length >0 &&
                    <div className="card-deck row">
                    {crsArr}
                    </div>
                }
            </div>
        
            
        )
    } 
    return upDateArrays()
        
}

export default Courses
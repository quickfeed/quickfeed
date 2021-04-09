import React, { useEffect } from "react"
import { Link } from "react-router-dom"
import { useOvermind, useState } from "../overmind"
import { Course } from "../proto/ag_pb"


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
        actions.getCoursesByUser()
        console.log(state.userCourses)
    }, [])
    // TODO: UserCourses contains elements describing a course that a user has an enrollment in, regardless of status currently. Need to figure out what UserStatus.NONE is used for
    const UserCourses = state.userCourses.map(course => {
        if (course.getEnrolled() >= 2) {
            return (
                <li className="list-group-item course" key={course.getId()}>
                    <div className="courseCode">{course.getCode()}</div>
                    <div className="courseName">{course.getName()}</div>
                    <div className="courseStatus">{EnrollmentStatus[course.getEnrolled()]}</div>
                </li>
            )
        }

        return (
            <li className="list-group-item course" key={course.getId()}>
                <div className="courseCode">{course.getCode()}</div>
                <div className="courseName">{course.getName()}</div>
                <div className="enrollmentStatus" onClick={() => {actions.enroll(course.getId())}}>{EnrollmentStatus[course.getEnrolled()]}</div>
            </li>
        )
    })

    return (
        <div className="box">
            <div className="card well" style={{width: "70%"}}>
                    <ul className="list-group list-group-flush">
                    {UserCourses}
                    </ul>
            </div>
        </div>
    )
}

export default Courses
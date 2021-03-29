import React, { useEffect } from "react"
import { Link } from "react-router-dom"
import { useOvermind, useState } from "../overmind"
import { Course } from "../proto/ag_pb"

const Courses = () => {
    const {state, actions} = useOvermind()

    useEffect(() => {
        // TODO: getCoursesByUser returns courses a user has an enrollment in. I thought a UserStatus = 0 (NONE) would be default, but apparently not.
        // 
        actions.getCoursesByUser()
        console.log(state.userCourses)
    }, [])
    // TODO: UserCourses contains elements describing a course that a user has an enrollment in, regardless of status currently. Need to figure out what UserStatus.NONE is used for
    const UserCourses = state.courses.map(course => {
        return (
            <h5 key={course.getId()}>
                {course.getName()} {course.getCode()} {course.getEnrolled() }
            </h5>
        )
    })

    return (
        <div>
            {UserCourses}
        </div>
    )
}

export default Courses
import React, { useEffect } from "react"
import { Redirect } from "react-router"
import { getCourseID, isEnrolled, isTeacher } from "../Helpers"
import { useActions, useAppState } from "../overmind"
import StudentPage from "./StudentPage"
import TeacherPage from "./TeacherPage"

/** This component is mainly used to determine which view (Student or Teacher) to render, based on enrollment status. */
const CoursePage = (): JSX.Element => {
    const state = useAppState()
    const setActiveCourse = useActions().setActiveCourse
    const courseID = getCourseID()
    const enrollment = state.enrollmentsByCourseId[courseID]

    useEffect(() => {
        setActiveCourse(courseID)
    }, [courseID])

    if (state.enrollmentsByCourseId[courseID] && isEnrolled(enrollment)) {
        if (isTeacher(enrollment)) {
            return <TeacherPage />
        }
        return <StudentPage />
    }
    else {
        return <Redirect to={"/"} />
    }
}

export default CoursePage
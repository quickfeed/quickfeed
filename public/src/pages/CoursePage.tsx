import React, { useEffect } from "react"
import { Redirect } from "react-router"
import { isEnrolled, isTeacher } from "../Helpers"
import { useActions, useAppState } from "../overmind"
import StudentPage from "./StudentPage"
import TeacherPage from "./TeacherPage"
import { useCourseID } from "../hooks/useCourseID"


/** The CoursePage component renders a Student or Teacher view
 *  depending on the active course and the user's enrollment status. */
const CoursePage = () => {
    const state = useAppState()
    const actions = useActions()
    const courseID = useCourseID()
    const enrollment = state.enrollmentsByCourseID[courseID.toString()]

    useEffect(() => {
        if (!state.showFavorites) {
            actions.toggleFavorites()
        }
        actions.setActiveCourse(courseID)
        actions.getCourseData({ courseID })
    }, [courseID])

    if (state.enrollmentsByCourseID[courseID.toString()] && isEnrolled(enrollment)) {
        if (isTeacher(enrollment)) {
            return <TeacherPage />
        }
        return <StudentPage />
    } else {
        return <Redirect to={"/"} />
    }
}

export default CoursePage

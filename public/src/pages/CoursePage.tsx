import React, { useEffect } from "react"
import { Redirect } from "react-router"
import { getCourseID, isEnrolled, isTeacher } from "../Helpers"
import { useActions, useAppState } from "../overmind"
import StudentPage from "./StudentPage"
import TeacherPage from "./TeacherPage"


/** The CoursePage component renders a Student or Teacher view
 *  depending on the active course and the user's enrollment status. */
const CoursePage = (): JSX.Element => {
    const state = useAppState()
    const actions = useActions()
    const courseID = getCourseID()
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

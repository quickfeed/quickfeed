import React, { useLayoutEffect } from "react"
import { Redirect } from "react-router"
import { getCourseID, isEnrolled, isTeacher } from "../Helpers"
import { useActions, useAppState } from "../overmind"
import StudentPage from "./StudentPage"
import TeacherPage from "./TeacherPage"


/** The CoursePage component renders a Student or Teacher view
 *  depending on the active course and the user's enrollment status. */
const CoursePage = () => {
    const state = useAppState()
    const actions = useActions()
    const courseID = getCourseID()
    const enrollment = state.enrollmentsByCourseID[courseID.toString()]

    useLayoutEffect(() => {
        if (!state.showFavorites) {
            actions.toggleFavorites()
        }
        actions.setActiveCourse(courseID)
        actions.getCourseData({ courseID })
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [actions, courseID]) // Having state.showFavorites in the dependency array locks the sidebar as open.

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

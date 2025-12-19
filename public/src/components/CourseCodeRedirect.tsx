import React from "react"
import { useParams, Navigate } from "react-router-dom"
import { useAppState } from "../overmind"


export const CourseCodeRedirect = () => {
    const { code = "" } = useParams()
    const state = useAppState()

    // find course with the given code
    // multiple courses can have the same code, so we take the one with the highest year
    const courses = state.courses.filter(course => course.code.toLowerCase() === code.toLowerCase())
    if (courses.length === 0) {
        // no course found with the given code
        return <Navigate to="/" replace />
    }

    // find the course with the highest year
    const course = courses.reduce((prev, current) => {
        return (prev.year > current.year) ? prev : current
    })

    return <Navigate to={`/course/${course.ID}`} replace />
}

export default CourseCodeRedirect

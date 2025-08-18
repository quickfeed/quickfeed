import React from "react"
import { Navigate } from "react-router-dom"
import { useAppState } from "../overmind"
import CourseCard from "./CourseCard"
import { create } from "@bufbuild/protobuf"
import { EnrollmentSchema } from "../../proto/qf/types_pb"


export const Enroll = ({ courseID }: { courseID: bigint }) => {
    const state = useAppState()
    const course = state.courses.find(c => c.ID === courseID)
    const enrollment = state.enrollmentsByCourseID[courseID.toString()] || create(EnrollmentSchema)
    if (!course) {
        // If no course is found, we can return a placeholder or an error message.
        return <Navigate to="/" replace />
    }
    return (
        <div className="box centered">
            <h3>Enroll in {course.name}</h3>
            <CourseCard course={course} enrollment={enrollment} />
        </div>
    )
}

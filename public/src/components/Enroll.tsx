import React from "react"
import { Navigate } from "react-router-dom"
import { useAppState } from "../overmind"
import CourseCard from "./CourseCard"
import { create } from "@bufbuild/protobuf"
import { EnrollmentSchema } from "../../proto/qf/types_pb"
import { isPending } from "../Helpers"


export const Enroll = ({ courseID }: { courseID: bigint }) => {
    const state = useAppState()
    const course = state.courses.find(c => c.ID === courseID)
    const enrollment = state.enrollmentsByCourseID[courseID.toString()] || create(EnrollmentSchema)
    if (!course) {
        // If no course is found, we can return a placeholder or an error message.
        return <Navigate to="/" replace />
    }

    if (isPending(enrollment)) {
        // If the user is already pending enrollment, inform them that they are pending
        return (
            <div className="box centered">
                <h3>You have already requested to enroll in {course.name}.</h3>
                <p>Please wait for the course administrator to approve your enrollment.</p>
                <CourseCard course={course} enrollment={enrollment} />
            </div>
        )
    }

    return (
        <div className="box centered">
            <h3>Enroll in {course.name}</h3>
            <CourseCard course={course} enrollment={enrollment} />
        </div>
    )
}

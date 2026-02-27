import React, { useCallback } from 'react'
import { useNavigate } from 'react-router'
import { EnrollmentStatus, hasEnrolled, hasNone, hasPending } from '../Helpers'
import { useActions, useAppState } from '../overmind'
import { Course, Enrollment } from '../../proto/qf/types_pb'
import CourseFavoriteButton from './CourseFavoriteButton'
import Badge from './Badge'


interface CardProps {
    course: Course,
    enrollment: Enrollment
    unavailable?: boolean
}

// Map enrollment status to DaisyUI badge color classes
// These must be complete class names for Tailwind's purge to detect them
const CardColorClasses: Record<number, string> = {
    0: "bg-info text-info-content", // NONE - shouldn't appear
    1: "bg-secondary text-secondary-content", // PENDING
    2: "bg-primary text-primary-content", // STUDENT
    3: "bg-success text-success-content", // TEACHER
}

const CourseCard = ({ course, enrollment, unavailable }: CardProps) => {
    const actions = useActions().global
    const navigate = useNavigate()
    const status = enrollment.status
    const state = useAppState()

    const handleEnroll = useCallback(() => actions.enroll(course.ID), [actions, course.ID])
    const CourseEnrollmentButton = () => {
        // Always show 'Go to Course' if enrolled or pending, even if unavailable
        if (hasNone(status)) {
            // Show enroll if not unavailable, or if user is admin
            if (!unavailable || (state.self.IsAdmin)) {
                return <button className="btn btn-primary course-button" onClick={handleEnroll}>Enroll</button>
            }
            // Otherwise, hide button
            return null
        } else if (hasPending(status)) {
            // Show pending if not unavailable, or if user is admin
            if (!unavailable || (state.self.IsAdmin)) {
                return <button className="btn btn-secondary course-button disabled">Pending</button>
            }
            // Otherwise, hide button
            return null
        }
        // Always show Go to Course, even if unavailable
        return <button className="btn btn-primary course-button" onClick={() => navigate(`/course/${enrollment.courseID}`)}>Go to Course</button>
    }

    const CourseEnrollmentStatus = () => {
        if (!hasEnrolled(status)) {
            return null
        }
        return (
            <div className="flex items-center gap-2">
                <CourseFavoriteButton enrollment={enrollment} />
                <span className="text-sm">{EnrollmentStatus[status]}</span>
            </div>
        )
    }

    // Get color classes based on enrollment status, fallback to secondary if unavailable
    const headerColorClasses = unavailable
        ? "bg-secondary text-secondary-content"
        : CardColorClasses[status] || "bg-base-100"

    return (
        <div className="card w-full shadow-xl bg-base-200 overflow-hidden">
            {/* Colored Header Section */}
            <div className={`${headerColorClasses} p-6`}>
                <div className="flex justify-between items-center">
                    <h2 className="text-2xl font-bold">{course.code}</h2>
                    <div className="flex items-center gap-2">
                        {unavailable && <Badge color="yellow" text="Unavailable" type="solid" />}
                        <CourseEnrollmentStatus />
                    </div>
                </div>
            </div>

            {/* Neutral Body Section */}
            <div className="card-body">
                <h3 className="card-title text-xl">{course.name}</h3>
                <p className="text-base-content/70">{course.tag} {course.year}</p>

                <div className="card-actions justify-end mt-4">
                    <CourseEnrollmentButton />
                </div>
            </div>
        </div>
    )
}

export default CourseCard

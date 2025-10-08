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

const CardColor = [
    "info", // "NONE in enrollment. Shouldn't ever appear."
    "secondary",
    "primary",
    "success"
]

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
            <div className="d-flex align-items-center">
                <CourseFavoriteButton enrollment={enrollment} style={{ marginLeft: 'auto' }} />
                <p className="mb-0 ml-2 text-white">{EnrollmentStatus[status]}</p>
            </div>
        )
    }

    const color = unavailable ? "secondary" : CardColor[status]
    return (
        <div className="card course-card mb-4 shadow-sm">
            <div className={`card-header bg-${color} text-white d-flex justify-content-between align-items-center`}>
                <span>{course.code}</span>
                {unavailable && <Badge color="yellow" text="Unavailable" />}
                <CourseEnrollmentStatus />
            </div>

            <div className="card-body">
                <h5 className="card-title">{course.name}</h5>
                <p className="card-text text-muted">{course.tag} {course.year}</p>
                <CourseEnrollmentButton />
            </div>
        </div>
    )
}

export default CourseCard

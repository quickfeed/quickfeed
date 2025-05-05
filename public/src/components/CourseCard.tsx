import React, { useCallback } from 'react'
import { useHistory } from 'react-router'
import { EnrollmentStatus, hasEnrolled, hasNone, hasPending } from '../Helpers'
import { useActions } from '../overmind'
import { Course, Enrollment } from '../../proto/qf/types_pb'
import CourseFavoriteButton from './CourseFavoriteButton'


interface CardProps {
    course: Course,
    enrollment: Enrollment
}

const CardColor = [
    "info", // "NONE in enrollment. Shouldn't ever appear."
    "secondary",
    "primary",
    "success"
]

const CourseCard = ({ course, enrollment }: CardProps) => {
    const actions = useActions()
    const history = useHistory()
    const status = enrollment.status

    const handleEnroll = useCallback(() => actions.enroll(course.ID), [actions, course.ID])
    const CourseEnrollmentButton = () => {
        if (hasNone(status)) {
            return <button className="btn btn-primary course-button" onClick={handleEnroll}>Enroll</button>
        } else if (hasPending(status)) {
            return <button className="btn btn-secondary course-button disabled">Pending</button>
        }
        return <button className="btn btn-primary course-button" onClick={() => history.push(`/course/${enrollment.courseID}`)}>Go to Course</button>
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

    return (
        <div className="card course-card mb-4 shadow-sm">
            <div className={`card-header bg-${CardColor[status]} text-white d-flex justify-content-between align-items-center`}>
                <span>{course.code}</span>
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

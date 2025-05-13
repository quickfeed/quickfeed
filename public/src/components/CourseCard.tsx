import React from 'react'
import { useNavigate } from 'react-router'
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
    const navigate = useNavigate()
    const status = enrollment.status

    const CourseEnrollmentButton = () => {
        if (hasNone(status)) {
            return <div className="btn btn-primary course-button" onClick={() => actions.enroll(course.ID)}>Enroll</div>
        } else if (hasPending(status)) {
            return <div className="btn btn-secondary course-button disabled">Pending</div>
        }
        return <div className="btn btn-primary course-button" onClick={() => navigate(`/course/${enrollment.courseID}`)}>Go to Course</div>
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

import React from 'react'
import { useHistory } from 'react-router'
import { EnrollmentStatus, hasEnrolled, hasNone, hasPending } from '../Helpers'
import { useActions } from '../overmind'
import { Course, Enrollment } from '../../gen/qf/types_pb'
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

const CourseCard = ({ course, enrollment }: CardProps): JSX.Element => {
    const actions = useActions()
    const history = useHistory()
    const status = enrollment.status

    const CourseEnrollmentButton = (): JSX.Element => {
        if (hasNone(status)) {
            return <div className="btn btn-primary course-button" onClick={() => actions.enroll(course.ID)}>Enroll</div>
        } else if (hasPending(status)) {
            return <div className="btn btn-secondary course-button disabled">Pending</div>
        }
        return <div className="btn btn-primary course-button" onClick={() => history.push("/course/" + enrollment.courseID.toString())}>Go to Course</div>
    }

    const CourseEnrollmentStatus = (): JSX.Element | null => {
        if (hasEnrolled(status)) {
            return (
                <>
                    <CourseFavoriteButton enrollment={enrollment} style={{ "float": "right" }} />
                    <p className="float-sm-right mr-2">{EnrollmentStatus[status]}</p>
                </>
            )
        }
        return null
    }

    return (
        <div className="col-sm-4">
            <div className="card" style={{ maxWidth: "35rem", marginBottom: "10px", minHeight: "205px" }}>
                <div className={"card-header bg-" + CardColor[status] + " text-white"}>
                    {course.code}
                    <CourseEnrollmentStatus />
                </div>

                <div className="card-body position-relative">
                    <h5 className="card-title">{course.name} - {course.tag}/{course.year}</h5>
                    <CourseEnrollmentButton />
                </div>
            </div>
        </div>
    )
}

export default CourseCard

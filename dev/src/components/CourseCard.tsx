import * as React from 'react'
import { useHistory } from 'react-router'
import { EnrollmentStatus, hasEnrolled, hasNone, hasPending, isVisible } from '../Helpers'
import { useActions } from '../overmind'
import { Course, Enrollment } from '../../proto/ag/ag_pb'


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
    const status = enrollment.getStatus()

    return (
        <div className="col-sm-4">
            <div className="card" style={{ maxWidth: "35rem", marginBottom: "10px", minHeight: "205px" }}>
                <div className={"card-header bg-" + CardColor[status] + " text-white"}>
                    {course.getCode()}
                    {hasEnrolled(status) &&
                        <>
                            <span className="float-right">
                                <i className={isVisible(enrollment) ? 'fa fa-star' : "fa fa-star-o"}
                                    onClick={() => actions.setEnrollmentState(enrollment)}></i>
                            </span>
                            <p className="float-sm-right mr-2">{EnrollmentStatus[status]}</p>
                        </>
                    }
                </div>

                <div className="card-body position-relative">
                    <h5 className="card-title">{course.getName()} - {course.getTag()}/{course.getYear()}</h5>
                    {hasNone(status) ?
                        <div className="btn btn-primary course-button" onClick={() => actions.enroll(course.getId())}>Enroll</div>
                        : hasPending(status) ?
                            <div className="btn btn-secondary course-button disabled">Pending</div>
                            :
                            <div className="btn btn-primary course-button" onClick={() => history.push("/course/" + enrollment.getCourseid())}>Go to Course</div>
                    }
                </div>
            </div>
        </div>
    )
}

export default CourseCard

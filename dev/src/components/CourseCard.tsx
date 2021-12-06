import * as React from 'react';
import { useHistory } from 'react-router';
import { EnrollmentStatus } from '../Helpers';
import { useActions } from '../overmind';
import { Course, Enrollment } from '../../proto/ag/ag_pb';

// TODO Should be exported to a seperate file 

interface CardProps {
    course : Course,
    enrollment: Enrollment
    status: number 
}

const CardColor = [
    "info", // "NONE in enrollment. Shouldn't ever appear."
    "secondary",
    "primary",
    "success"
]

const CourseCard = ({course, enrollment, status}: CardProps): JSX.Element => {
    const actions = useActions()
    const history = useHistory()

    return (
        <div className="col-sm-4">
            <div className="card" style= {{maxWidth: "35rem", marginBottom:"10px",minHeight:"205px"}}>
                <div className={"card-header bg-"+CardColor[status]+" text-white"}>
                    {course.getCode()}
                    {enrollment.getStatus() > Enrollment.UserStatus.NONE && 
                    <>
                        <span className="float-right">
                            <i className={enrollment.getState() === Enrollment.DisplayState.VISIBLE ? 'fa fa-star-o' : "fa fa-star "} 
                                onClick={() => actions.setEnrollmentState(enrollment)}></i>
                        </span>
                        <p className="float-sm-right mr-2">{enrollment ? EnrollmentStatus[enrollment?.getStatus()]  : ''}</p>
                    </>
                    }
                </div>
                
                <div className="card-body position-relative">
                    <h5 className="card-title">{course.getName()} - {course.getTag()}/{course.getYear()}</h5>
                    { status === Enrollment.UserStatus.NONE ? 
                        <div className="btn btn-primary course-button" onClick={() => actions.enroll(course.getId())}>Enroll</div>
                    : status === Enrollment.UserStatus.PENDING ?
                        <div className="btn btn-secondary course-button disabled">Pending</div>
                    :
                        <div className="btn btn-primary course-button" onClick={() => history.push("/course/"+enrollment.getCourseid())}>Go to Course</div>
                    }
                </div>
            </div>
        </div>
    )
   
}
export default CourseCard
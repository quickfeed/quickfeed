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

const cardcolor = [
    "info", // "NONE in enrollment. Shouldn't ever appear."
    "secondary",
    "primary",
    "success"
]

const CourseCard = (props: CardProps) => {
    const actions = useActions()
    const history = useHistory()

    return (
        <div className="col-sm-4">
            <div className="card" style= {{maxWidth: "35rem", marginBottom:"10px",minHeight:"205px"}}>
                <div className={"card-header bg-"+cardcolor[props.status]+" text-white"}>
                    {props.course.getCode()}
                    <span className="float-right "><i className={props.enrollment.getState() === Enrollment.DisplayState.VISIBLE ? 'fa fa-star-o' : "fa fa-star "} onClick={() => actions.setEnrollmentState(props.enrollment)}></i></span>
                    <p className="float-sm-right">{props.enrollment ? EnrollmentStatus[props.enrollment?.getStatus()]  : ''}</p>
                </div>
                
                <div className="card-body">
                    <h5 className="card-title">{props.course.getName()} - {props.course.getYear()}/{props.course.getTag()}</h5>
                    { props.status === Enrollment.UserStatus.NONE ? 
                        <div className="btn btn-primary float-down-left" onClick={() => actions.enroll(props.course.getId())}>Enroll to Course</div>
                    : props.status === Enrollment.UserStatus.PENDING ?
                        null
                    :
                        <div className="btn btn-primary float-down-left" onClick={() => history.push("/course/"+props.enrollment.getCourseid())}>Go to Course</div>
                    }
                </div>

            </div>
        </div>
    )
   
}
export default CourseCard
import * as React from 'react';
import { EnrollmentStatus } from '../Helpers';
import { Course, Enrollment } from '../proto/ag_pb';

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
    return (
        <div className="col-sm-4">
            <div className="card" style= {{maxWidth: "35rem", marginBottom:"10px"}}>
                <div className={"card-header bg-"+cardcolor[props.status]+" text-white"}>
                    {props.course.getCode()}
                    <span className="float-right "><i className='fa fa-star-o'></i></span>
                    <span className="float-right "><i className="fa fa-star "></i></span>
                    <p className="float-sm-right">{props.enrollment ? EnrollmentStatus[props.enrollment?.getStatus()]  : ''}</p>
                </div>
                
                <div className="card-body">
                    <h5 className="card-title">{props.course.getName()} - {props.course.getYear()}/{props.course.getTag()}</h5>
                    <p className="card-text">placeholder, don't know what to put here</p>
                </div>

            </div>
        </div>
    )
   
}
export default CourseCard
import * as React from 'react';
import { EnrollmentStatus } from '../Helpers';
import { useActions } from '../overmind';
import { Course, Enrollment } from '../proto/ag_pb';

// TODO Should be exported to a seperate file 

interface CardProps {
    course : Course,
    enrollment: Enrollment,
    status: number,

}

const cardcolor = [
    "info", // "NONE in enrollment. Shouldn't ever appear."
    "secondary",
    "primary",
    "success"
]

const CourseCard = (props: CardProps) => {
    const actions = useActions()
    if(props.status===Enrollment.DisplayState.UNSET){
        return(
            <div className="col-sm-4">
                <div className="card border-secondary" style= {{maxWidth: "35rem", marginBottom:"10px",minHeight:"250px"}}>
                    <div className={"card-header bg-"+cardcolor[props.status]+" text-white"}>
                        {props.course.getCode()}
                    </div>
                    
                    <div className="card-body">
                        <h5 className="card-title">{props.course.getName()}</h5>
                        <h5 className="card-title">{props.course.getYear()}/{props.course.getTag()}</h5>
                        <p className="card-text">enroll button here.</p>
                    </div>
    
                </div>
            </div>
        )
    }
    return (
        <div className="col-sm-4">
            <div className="card border-secondary" style= {{maxWidth: "35rem", marginBottom:"10px",minHeight:"250px"}}>
                <div className={"card-header bg-"+cardcolor[props.status]+" text-white"}>
                    {props.course.getCode()}
                    <span className="float-right "><i className={props.enrollment.getState() === Enrollment.DisplayState.VISIBLE ? 'fa fa-star-o' : "fa fa-star "} onClick={() => actions.setEnrollmentState(props.enrollment)}></i></span>
                    <p className="float-sm-right">{props.enrollment ? EnrollmentStatus[props.enrollment?.getStatus()]  : ''}</p>
                </div>
                
                <div className="card-body">
                    <h5 className="card-title">{props.course.getName()}</h5>
                    <h5 className="card-title">{props.course.getYear()}/{props.course.getTag()}</h5>
                    <p className="card-text">placeholder, don't know what to put here</p>
                </div>

            </div>
        </div>
    )

   
}
export default CourseCard
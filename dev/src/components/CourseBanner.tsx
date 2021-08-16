import React from "react";
import { Enrollment } from "../../proto/ag/ag_pb";
import { useActions, useAppState } from "../overmind";


// TODO: Maybe add route specific information, ex. if user is viewing a lab, show that in the banner. Could use state in components to display.

const CourseBanner = ({enrollment}: {enrollment: Enrollment}) => {
    const status = useAppState().status
    const actions = useActions()
    const style = enrollment.getState() === Enrollment.DisplayState.VISIBLE ? 'fa fa-star-o' : "fa fa-star "
    return (
        <div className="jumbotron">
            <div className="centerblock container">
                <h1>{enrollment.getCourse()?.getName()} 
                    <span style={{"paddingLeft": "20px"}}>
                        <i  className={style} 
                            onClick={() => actions.setEnrollmentState(enrollment)}>
                        </i>
                    </span>
                </h1>
                {status[enrollment.getCourseid()] === Enrollment.UserStatus.TEACHER && 
                    <span style={{"cursor": "pointer"}} onClick={() => actions.changeView(enrollment.getCourseid())}>
                        {enrollment.getStatus() === Enrollment.UserStatus.TEACHER ? "Switch to Student View" : "Switch to Teacher View" }
                    </span>
                }
            </div>
        </div>
    )
}

export default CourseBanner
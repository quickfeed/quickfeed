import React from "react";
import { Enrollment } from "../../proto/ag/ag_pb";
import { useOvermind } from "../overmind";


const CourseBanner = ({enrollment}: {enrollment: Enrollment}) => {
    const {actions} = useOvermind()
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
            </div>
        </div>
    )
}

export default CourseBanner
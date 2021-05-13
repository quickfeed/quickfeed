import React, { useEffect } from "react"
import { RouteComponentProps } from "react-router-dom"
import { Enrollment } from "../../proto/ag_pb"
import { useOvermind } from "../overmind"
import { CourseLabs } from "./CourseLabs"
import CourseUtilityLinks from "./CourseUtilityLinks"


interface MatchProps {
    id: string
}

/* */
const CourseOverview = (props: RouteComponentProps<MatchProps>) => {
    const { state, actions } = useOvermind()
    const courseID = Number(props.match.params.id)
    
    useEffect(() => {
    }, [props])
    

    return (
        <div className="box">
            <h1>{state.enrollmentsByCourseId[courseID].getCourse()?.getName()} <span className=""><i className={state.enrollmentsByCourseId[courseID].getState() === Enrollment.DisplayState.VISIBLE ? 'fa fa-star-o' : "fa fa-star "} onClick={() => actions.setEnrollmentState(state.enrollmentsByCourseId[courseID])}></i></span></h1>
            
            <div className="row">
                <div className="col-md-9" >
                    <CourseLabs crsid={courseID}/>
                </div>
                <CourseUtilityLinks courseID={courseID} />
            </div>

        </div>
    )
}

export default CourseOverview
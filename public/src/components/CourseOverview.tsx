import React, { useEffect } from "react"
import { RouteComponentProps } from "react-router-dom"
import { Enrollment } from "../../proto/ag/ag_pb"
import { getCourseID } from "../Helpers"
import { useOvermind } from "../overmind"
import { CourseLabs } from "./CourseLabs"
import CourseUtilityLinks from "./CourseUtilityLinks"


interface MatchProps {
    id: string
}

/* */
const CourseOverview = () => {
    const { state, actions } = useOvermind()
    const courseID = getCourseID()
    const style = state.enrollmentsByCourseId[courseID].getState() === Enrollment.DisplayState.VISIBLE ? 'fa fa-star-o' : "fa fa-star "
    
    useEffect(() => {
    }, [])
    
    return (
        <div className="box">
            <h1>{state.enrollmentsByCourseId[courseID].getCourse()?.getName()} 
                <span>
                    <i  className={style} 
                        onClick={() => actions.setEnrollmentState(state.enrollmentsByCourseId[courseID])}>
                    </i>
                </span>
            </h1>
            
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
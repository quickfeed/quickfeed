import React, { useEffect } from "react"
import { Link, RouteComponentProps } from "react-router-dom"
import { Enrollment, Repository } from "../../proto/ag_pb"
import { useOvermind } from "../overmind"
import { CourseLabs } from "./CourseLabs"
import SubmissionsTable from "./SubmissionsTable"


interface MatchProps {
    id: string
}


const CourseOverview = (props: RouteComponentProps<MatchProps>) => {

    const { state, actions } = useOvermind()
    let courseID = Number(props.match.params.id)
    useEffect(() => {

    }, [props])

    return (
        <div className="box">
            <h1>{state.enrollmentsByCourseId[courseID].getCourse()?.getName()} <span className=""><i className={state.enrollmentsByCourseId[courseID].getState() === Enrollment.DisplayState.VISIBLE ? 'fa fa-star-o' : "fa fa-star "} onClick={() => actions.setEnrollmentState(state.enrollmentsByCourseId[courseID])}></i></span></h1>
            
            <div className="row">
                <div className="col-md-9" >
                    <CourseLabs crsid={courseID}/>
                </div>
                <div className="col-sm-3" >
                    
                    <div className="list-group">
                        <div className="list-group-item list-group-item-action active text-center"><h6><strong>Utility</strong></h6></div>
                        <a href={state.repositories[courseID][Repository.Type.USER]} className="list-group-item list-group-item-action">User Repository</a>
                        {state.repositories[courseID][Repository.Type.GROUP] !== "" ?(
                        <a href={state.repositories[courseID][Repository.Type.GROUP]} className="list-group-item list-group-item-action overflow-ellipses" style={{textAlign:"left"}}>Group Repository ({state.enrollmentsByCourseId[courseID].getGroup()?.getName()})</a>
                        ):(
                            <Link to={"/course/" + courseID + "/group"} className="list-group-item list-group-item-action list-group-item-success">Create a Group</Link>
                        )}
                        <a href={state.repositories[courseID][Repository.Type.ASSIGNMENTS]} className="list-group-item list-group-item-action">Assignments</a>

                        <a href={state.repositories[courseID][Repository.Type.COURSEINFO]} className="list-group-item list-group-item-action">Course Info</a>
                        {state.enrollmentsByCourseId[courseID].hasGroup() ? <Link to={"/course/" + courseID + "/group"} className="list-group-item list-group-item-action">View Group</Link> : ""}
                    </div>
                </div>
            </div>

        </div>
    )
}

export default CourseOverview
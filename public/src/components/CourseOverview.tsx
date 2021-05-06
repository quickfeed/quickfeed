import React, { useEffect } from "react"
import { Link, RouteComponentProps } from "react-router-dom"
import { useOvermind } from "../overmind"
import { Repository } from "../proto/ag_pb"
import SubmissionsTable from "./SubmissionsTable"


interface MatchProps {
    id: string
}


const CourseOverview = (props: RouteComponentProps<MatchProps>) => {

    const { state } = useOvermind()
    let courseID = Number(props.match.params.id)
    useEffect(() => {

    }, [props])

    return (
        <div className="box">
            <h1>{state.enrollmentsByCourseId[courseID].getCourse()?.getName()}</h1>
            <div className="Links">
                <a href={state.repositories[courseID][Repository.Type.USER]}>User Repository</a>
                <a href={state.repositories[courseID][Repository.Type.GROUP]}>Group Repository</a>
                <a href={state.repositories[courseID][Repository.Type.COURSEINFO]}>Course Info</a>
                <a href={state.repositories[courseID][Repository.Type.ASSIGNMENTS]}>Assignments</a>
                <Link to={"/course/" + courseID + "/group"} >Group</Link>
            </div>
            
            <SubmissionsTable courseID={courseID} />
            

        </div>
    )
}

export default CourseOverview
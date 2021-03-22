import React, { useEffect, useState } from "react"
import { RouteComponentProps, Route, useRouteMatch } from "react-router"
import { Link } from "react-router-dom"
import { getFormattedDeadline } from "../Helpers"
import { useOvermind } from "../overmind"

import { Courses, Enrollment, Repositories, Repository } from "../proto/ag_pb"
import Lab from "./Lab"


interface MatchProps {
    id: string
}

const Course = (props: RouteComponentProps<MatchProps>) => {
    const { state, actions } = useOvermind()
    const { url } = useRouteMatch()
    const [enrollment, setEnrollment] = useState(new Enrollment())
    let courseID = Number(props.match.params.id)


    useEffect(() => {
        const enrol = actions.getEnrollmentByCourseId(courseID)
        if (enrol !== null) {
            setEnrollment(enrol)
        }
    }, [])

    /**if(state.isLoading){
        return(
            <h1>Loading icon here...</h1>
        )
    }*/


    if (enrollment.getId() !== 0 && typeof state.assignments[courseID] !== 'undefined'){
        return (
        <div className="box">
            <h1>Welcome to {enrollment.getCourse()?.getName()}, {enrollment.getUser()?.getName()}! You are a {enrollment.getStatus() == Enrollment.UserStatus.STUDENT ? ("student") : ("teacher")}</h1>
            {
                state.assignments[courseID].map(assignment => {
                    return (
                        <h2 key={assignment.getId()}><Link to={`/course/${courseID}/${assignment.getId()}`}>{assignment.getName()}</Link> Deadline: {getFormattedDeadline(assignment.getDeadline())} </h2>
                    )
                })
            }
            
            <Route path={`${url}/:lab`}>
                <Lab crsID={courseID}></Lab>
            </Route>
            <div className="Links">
            <a href={state.repositories[courseID][Repository.Type.USER]}>User Repository</a>
            <a href={state.repositories[courseID][Repository.Type.GROUP]}>Group Repository</a>
            <a href={state.repositories[courseID][Repository.Type.COURSEINFO]}>Course Info</a>
            <a href={state.repositories[courseID][Repository.Type.ASSIGNMENTS]}>Assignments</a>
            </div>
        </div>)
    }
    return <h1>Loading</h1>
}

export default Course
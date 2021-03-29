import React, { useEffect, useState } from 'react'
import { RouteComponentProps, Route, useRouteMatch } from 'react-router'
import { Link } from 'react-router-dom'
import { getFormattedDeadline } from '../Helpers'
import { useOvermind } from '../overmind'

import { Courses, Enrollment, Repositories, Repository } from "../proto/ag_pb"
import Lab from "./Lab"
import LandingPageLabTable from "./LandingPageLabTable"


interface MatchProps {
    id: string
}


const Course = (props: RouteComponentProps<MatchProps>) => {
    const { state, actions } = useOvermind()
    const { url } = useRouteMatch()
    const [isLoading , setLoading] = useState(true)
    const [enrollment, setEnrollment] = useState(new Enrollment())
    let courseID = Number(props.match.params.id)
    actions.setActiveCourse(courseID)

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
            <h1>{enrollment.getCourse()?.getName()}</h1>
            <div className="Links">
            <a href={state.repositories[courseID][Repository.Type.USER]}>User Repository</a>
            <a href={state.repositories[courseID][Repository.Type.GROUP]}>Group Repository</a>
            <a href={state.repositories[courseID][Repository.Type.COURSEINFO]}>Course Info</a>
            <a href={state.repositories[courseID][Repository.Type.ASSIGNMENTS]}>Assignments</a>
            </div>
            <LandingPageLabTable courseID={courseID} />
            
            <Route path={`${url}/:lab`}>
                <Lab crsID={courseID}></Lab>
            </Route>

        </div>)
    }
    return <h1>Loading</h1>
}

export default Course
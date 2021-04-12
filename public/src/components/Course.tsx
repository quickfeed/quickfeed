import React, { useEffect, useLayoutEffect, useState } from "react"
import { RouteComponentProps, Route, useRouteMatch } from "react-router"
import { Link } from "react-router-dom"
import { getFormattedDeadline } from "../Helpers"
import { useOvermind } from "../overmind"

import { Courses, Enrollment, Repositories, Repository } from "../proto/ag_pb"
import LandingPageLabTable from "./LandingPageLabTable"


interface MatchProps {
    id: string
}


const Course = (props: RouteComponentProps<MatchProps>) => {
    const { state, actions } = useOvermind()
    const [enrollment, setEnrollment] = useState(new Enrollment())
    let courseID = Number(props.match.params.id)


    useEffect(() => {
        courseID = Number(props.match.params.id)
        let enrol = actions.getEnrollmentByCourseId(courseID)
        actions.setActiveCourse(courseID)
        if (enrol !== null) {
            setEnrollment(enrol)
        }
    }, [props])

    if (state.courses){
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
            

        </div>)
    }
    return <h1>Loading</h1>
}

export default Course
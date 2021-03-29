import React, { useEffect, useState } from 'react'
import { RouteComponentProps, Route, useRouteMatch } from 'react-router'
import { Link } from 'react-router-dom'
import { getFormattedDeadline } from '../Helpers'
import { useOvermind } from '../overmind'

import { Courses, Enrollment } from '../proto/ag_pb'
import Lab from './Lab'


interface MatchProps {
    id: string
}

const Course = (props: RouteComponentProps<MatchProps>) => {
    const { state, actions } = useOvermind()
    const { url } = useRouteMatch()
    const [isLoading , setLoading] = useState(true)
    const [enrollment, setEnrollment] = useState(new Enrollment())
    let crsID = Number(props.match.params.id)
    useEffect(() => {
        state.isLoading = true
        actions.getEnrollmentsByUser()
        .then(success => {
            if (success){
                const enrol = actions.getEnrollmentByCourseId(Number(props.match.params.id))
                if (enrol !== null) {
                    setEnrollment(enrol)
                    actions.getSubmissions(Number(props.match.params.id))
                    actions.getAssignmentsByCourse(Number(props.match.params.id))
                    console.log(state.assignments)
                }
                
            }
        })
        .finally(() => {
            actions.loading()
        })
    }, [])
    if(state.isLoading){
        return(
            <h1>Loading icon here...</h1>
        )
    }
    /*
    const getSubmissions = state.submissions[crsID].map(submission => {
        return (
            <div>
                <h1>{submission.getScore()} / 100</h1>
                <code>{submission.getBuildinfo()}</code>
            </div>
        )
    })

    const listAssignments = state.assignments[Number(props.match.params.id)].map(assignment => {
        return (
            <h2 key={assignment.getId()}><Link to={`/course/${props.match.params.id}/${assignment.getId()}`}>{assignment.getName()}</Link> Deadline: {getFormattedDeadline(assignment.getDeadline())} </h2>
        )
    })*/


    if (enrollment.getId() !== 0 && typeof state.assignments[crsID] !== 'undefined'){
        return (
        <div className="box">
            <h1>Welcome to {enrollment.getCourse()?.getName()}, {enrollment.getUser()?.getName()}! You are a {enrollment.getStatus() == Enrollment.UserStatus.STUDENT ? ('student') : ('teacher')}</h1>
            {
                state.assignments[Number(props.match.params.id)].map(assignment => {
                    return (
                        <h2 key={assignment.getId()}><Link to={`/course/${props.match.params.id}/${assignment.getId()}`}>{assignment.getName()}</Link> Deadline: {getFormattedDeadline(assignment.getDeadline())} </h2>
                    )
                })
            }
            <Route path={`${url}/:lab`}>
                <Lab crsID={crsID}></Lab>
            </Route>
        </div>)
    }
    return <h1>Loading</h1>
}

export default Course
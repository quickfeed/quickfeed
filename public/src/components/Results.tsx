import React, { useEffect, useState } from "react"
import { Redirect, useParams } from "react-router-dom"
import { Enrollment } from "../../proto/ag/ag_pb"
import { useOvermind } from "../overmind"
import ResultItem from "./teacher/ResultItem"


const Results = () => {

    const {state, actions} = useOvermind()
    const course = useParams<{id?: string}>()
    const courseID = Number(course.id)

    const [query, setQuery] = useState<string>("")

    useEffect(() => {
        if (courseID && !state.courseSubmissions[courseID]) {
            actions.getAllCourseSubmissions(courseID)
        }
    }, [state.courseSubmissions, query, setQuery])


    const TableAssignmentsHead = state.assignments[courseID].map(assignment => {
        return <td>{assignment.getName()}</td>
    })

    // TODO: Allow admin to view
    if (!state.cSubs[courseID] || state.enrollmentsByCourseId[courseID].getStatus() !== Enrollment.UserStatus.TEACHER) {
        return <h1>Nothing</h1>
    }

    const UserResults = state.cSubs[courseID].map(user => {
        if (user.enrollment && user.submissions) {
            return <ResultItem enrollment={user.enrollment} submissionsLink={user.submissions} query={query} />
        }
    })

    return (
        <React.Fragment>
        <input onChange={e => setQuery(e.target.value.toLowerCase())}></input>
        <table>
            <thead>
                <td>Name</td>
                <td>Group</td>
                {TableAssignmentsHead}
            </thead>
            {UserResults}
        </table>
         </React.Fragment>
    )
}

export default Results
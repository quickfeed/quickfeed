import React, { useEffect } from "react"
import { Enrollment } from "../../proto/ag/ag_pb"
import { getCourseID } from "../Helpers"
import { useActions, useAppState } from "../overmind"
import Lab from "./Lab"
import Search from "./Search"
import ResultItem from "./teacher/ResultItem"


const Results = () => {
    const state = useAppState()
    const actions = useActions()
    const { getAllCourseSubmissions } = useActions()
    const courseID = getCourseID()

    useEffect(() => {
        if (courseID && !state.courseSubmissions[courseID]) {
            getAllCourseSubmissions(courseID)
        }
        return actions.setActiveSubmission(undefined)
    }, [state.courseSubmissions])


    const TableAssignmentsHead = state.assignments[courseID].map(assignment => {
        return <td>{assignment.getName()}</td>
    })

    // TODO: Allow admin to view
    if (!state.courseSubmissions[courseID] || state.enrollmentsByCourseId[courseID].getStatus() !== Enrollment.UserStatus.TEACHER) {
        return <h1>Nothing</h1>
    }

    const UserResults = state.courseSubmissions[courseID].map(user => {
        if (user.enrollment && user.submissions) {
            return <ResultItem enrollment={user.enrollment} submissionsLink={user.submissions} />
        }
    })

    return (
        <div className="box">
        <Search />
        <div className="row">
            <div className="col">
            <table className="table table-curved table-striped">
                <thead className="thead-dark">
                    <td>Name</td>
                    <td>Group</td>
                    {TableAssignmentsHead}
                </thead>
                <tbody>
                    {UserResults}
                </tbody>
            </table>
            </div>
            <div className="col reviewLab">
                {state.activeSubmission ?
                    <Lab teacherSubmission={state.activeSubmission} /> : null
                }  
            </div>
        </div>
        </div>

    )
}

export default Results
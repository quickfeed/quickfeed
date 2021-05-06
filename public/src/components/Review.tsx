import React, { useEffect, useState } from "react"
import { useHistory } from "react-router"
import { sortByField } from "../Helpers"
import { useOvermind } from "../overmind"
import { Enrollment, EnrollmentLink, User } from "../proto/ag_pb"
import Lab from "./Lab"


const Review = () => {
    const {state, actions} = useOvermind()

    const history = useHistory()
    useEffect(() => {
            actions.getAllCourseSubmissions(4)

    }, [state.courseSubmissions])


    if (state.courseSubmissions[4]) {
        const s = sortByField(state.courseSubmissions[4], [EnrollmentLink.prototype.getEnrollment, Enrollment.prototype.getUser], User.prototype.getEmail, true).map((l: EnrollmentLink) => {
            
            return (
            <div className="card well" style={{width: "400px", marginBottom: "5px"}}>
            <div className="card-header">{l.getEnrollment()?.getUser()?.getEmail()} - {l.getEnrollment()?.getStatus()}</div>
            <ul className="list-group list-group-flush">
                
                {l.getSubmissionsList().map(s => {
                    return (
                    <React.Fragment>
                    <li className="list-group-item">{s.getAssignment()?.getName()} - {s.getSubmission()?.getScore()} / 100</li>
                    
                    </React.Fragment>
                    )
                })
                }
            </ul>
            </div>
            )
        })
    

        return (
            <div>
                {s}
            </div>
        )
    }
    return (
        <div>Loading</div>
    )
}

export default Review
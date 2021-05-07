import React, { useEffect, useState } from "react"
import { useHistory } from "react-router"
import { sortByField } from "../Helpers"
import { useOvermind } from "../overmind"
import { Enrollment, EnrollmentLink, User } from "../../proto/ag_pb"


const Review = () => {
    const {state, actions} = useOvermind()

    useEffect(() => {
            actions.getAllCourseSubmissions(4)

    }, [state.courseSubmissions])


    if (state.courseSubmissions[4]) {
        const submissions = sortByField(state.courseSubmissions[4], [EnrollmentLink.prototype.getEnrollment], Enrollment.prototype.setStatus, false).map((link: EnrollmentLink) => {
            
            return (
            <div className="card well" style={{width: "400px", marginBottom: "5px"}}>
            <div className="card-header">{link.getEnrollment()?.getUser()?.getEmail()} - {link.getEnrollment()?.getSlipdaysremaining()}</div>
            <ul className="list-group list-group-flush">
                
                {link.getSubmissionsList().map(submissionLink => {
                    return (
                    <React.Fragment>
                    <li className="list-group-item">{submissionLink.getAssignment()?.getName()} - {submissionLink.getSubmission()?.getScore()} / 100</li>
                    
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
                {submissions}
            </div>
        )
    }
    return (
        <div>Loading</div>
    )
}

export default Review
import React from "react"
import { state } from "../overmind/state"

interface lab {
    id: number
    courseID: number
}

const LabResultTable = ({id, courseID}: lab) => {
    console.log(id, courseID)

    const LabResult = (): JSX.Element => {
        if (state.submissions !== undefined) {
        console.log(state.submissions)
        let submission = state.submissions[courseID]?.find(s => s.getAssignmentid() == id)
        console.log(submission, "<----------")
        return (<tr>Hei</tr>)
        }
        return (<div></div>)
    }

    return (
        <div>
            <table className="table table-curved" id="LandingPageTable">
                <thead>
                    <tr>
                        <tr>Assignment</tr>
                        <tr>Progress</tr>
                        <tr>Deadline</tr>
                        <tr>Due in</tr>
                        <tr>Status</tr>
                        <tr>Grouplab</tr>
                    </tr>
                </thead>
                <tbody>
                    {LabResult()}
                </tbody>
            </table>
        </div>
    )
}

export default LabResultTable
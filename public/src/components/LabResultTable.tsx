import React, { useEffect } from "react"
import { getBuildInfo, getScoreObjects } from "../Helpers"
import { useOvermind } from "../overmind"
import { state } from "../overmind/state"

interface lab {
    id: number
    courseID: number
}

const LabResultTable = ({id, courseID}: lab) => {
    const {state} = useOvermind()

    const LabResult = (): JSX.Element => {

            
            let submission = state.submissions[courseID]?.find(s => s.getAssignmentid() === id)
            let assignment = state.assignments[courseID]?.find(a => a.getId() == id)
            console.log(state.submissions[courseID])
            if (submission && assignment) {
                console.log("Found")
                const buildInfo = getBuildInfo(submission.getBuildinfo())
                const scoreObjects = getScoreObjects(submission.getScoreobjects())
            return (
            <table className="table table-curved">
                <thead>
                    <th>Lab information</th>
                </thead>
                <tr className="clickable-row">
                    <th>Status</th>
                    <td>{submission.getStatus()}</td>
                </tr>
                <tr>
                    <th>Delivered</th>
                    <td>{buildInfo.builddate}</td>
                </tr>
                <tr>
                    <th>Deadline</th>
                    <td>{assignment.getDeadline()}</td>
                </tr>
                <tr>
                    <th>Tests passed</th>
                    <td>{submission.getScoreobjects()}</td>
                </tr>
            </table>
            )
        }
        return (<div></div>)
    }

    return (
    <div>
        test
        {LabResult()}
    </div>
    )
}

export default LabResultTable
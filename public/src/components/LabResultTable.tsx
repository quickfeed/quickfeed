import React, { useEffect } from "react"
import { getBuildInfo, getScoreObjects, IScoreObjects } from "../Helpers"
import { useOvermind } from "../overmind"
import { state } from "../overmind/state"
import { ProgressBar } from "./ProgressBar"

interface lab {
    id: number
    courseID: number
}

const LabResultTable = ({id, courseID}: lab) => {
    const {state} = useOvermind()


    const ListScoreObjects = (scoreObjects: IScoreObjects[]) => {
        return scoreObjects.map(scoreObject => {
            return (
                <tr>
                    <th>
                        {scoreObject.TestName}
                    </th>
                    <th>
                        {scoreObject.Score}/{scoreObject.MaxScore}
                    </th>
                    <th>
                        {scoreObject.Weight}
                    </th>
                </tr>
            )
        })
    }

    const LabResult = (): JSX.Element => {

            
            let submission = state.submissions[courseID]?.find(s => s.getAssignmentid() === id)
            let assignment = state.assignments[courseID]?.find(a => a.getId() == id)
            console.log(state.submissions[courseID])
            if (submission && assignment) {
                console.log("Found")
                const buildInfo = getBuildInfo(submission.getBuildinfo())
                const scoreObjects = getScoreObjects(submission.getScoreobjects())
            return (
            <React.Fragment>
            <ProgressBar courseID={courseID} assignmentID={assignment.getId()} submission={submission} type={"lab"} />
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
                {ListScoreObjects(scoreObjects)}
                <tfoot>
                    
                </tfoot>
            </table>
            </React.Fragment>
            )
        }
        return (<div></div>)
    }

    return (
    <div>
        {LabResult()}
    </div>
    )
}

export default LabResultTable
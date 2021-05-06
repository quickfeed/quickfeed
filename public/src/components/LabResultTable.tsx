import React, { useEffect } from "react"
import { getBuildInfo, getScoreObjects, IScoreObjects } from "../Helpers"
import { useOvermind } from "../overmind"
import { ProgressBar } from "./ProgressBar"

interface lab {
    id: number
    courseID: number
}

const LabResultTable = ({id, courseID}: lab) => {
    const {state: {assignments, submissions}} = useOvermind()


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

        let submission = submissions[courseID]?.find(s => s.getAssignmentid() === id)
        let assignment = assignments[courseID]?.find(a => a.getId() == id)
        
        if (submission && assignment) {
            const buildInfo = getBuildInfo(submission.getBuildinfo())
            const scoreObjects = getScoreObjects(submission.getScoreobjects())
            
            return (
                <div className="container" style={{paddingBottom: "20px"}}>
                    <ProgressBar courseID={courseID} assignmentID={assignment.getId()} submission={submission} type={"lab"} />
                    <table className="table table-curved">
                        <thead>
                            <th colSpan={3}>Lab information</th>
                        </thead>
                        <tr className="clickable-row">
                            <th colSpan={2}>Status</th>
                            <td>{submission.getStatus()}</td>
                        </tr>
                        <tr>
                            <th colSpan={2}>Delivered</th>
                            <td>{buildInfo.builddate}</td>
                        </tr>
                        <tr>
                            <th colSpan={2}>Approved</th>
                            <td>{submission.getApproveddate()}</td>
                        </tr>
                        <tr>
                            <th colSpan={2}>Deadline</th>
                            <td>{assignment.getDeadline()}</td>
                        </tr>
                        {ListScoreObjects(scoreObjects)}
                    <tfoot>
                    
                    </tfoot>
                </table>
            </div>
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
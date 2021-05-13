import React from "react"
import { Assignment, Submission } from "../../proto/ag_pb"
import { getBuildInfo, getScoreObjects, IScoreObjects, SubmissionStatus } from "../Helpers"
import { useOvermind } from "../overmind"
import { ProgressBar } from "./ProgressBar"

interface lab {
    submission: Submission
    assignment: Assignment
}

const LabResultTable = ({submission, assignment}: lab) => {
    const {state: {
        enrollmentsByCourseId, }
    } = useOvermind()

    const ScoreObject = ({ scoreObject }: {scoreObject: IScoreObjects}) => {
        const boxShadow = (scoreObject.Score === scoreObject.MaxScore) ? "0 0px 0 #000 inset, 5px 0 0 green inset" : "0 0px 0 #000 inset, 8px 0 0 red inset"
        return (
            <tr>
                <th style={{boxShadow: boxShadow, paddingLeft: "15px"}}>
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
    }

    const LabResult = (): JSX.Element => {
        if (submission && assignment) {
            const buildInfo = getBuildInfo(submission.getBuildinfo())
            const scoreObjects = getScoreObjects(submission.getScoreobjects())
            
            let totalTests = 0
            let passedTests = 0
            scoreObjects.forEach(scoreObject => {
                if (scoreObject.Score === scoreObject.MaxScore) {
                    passedTests++
                } 
                totalTests++
            })

            const boxShadow = (submission.getStatus() === Submission.Status.APPROVED) ? "0 0px 0 #000 inset, 5px 0 0 green inset" : "0 0px 0 #000 inset, 8px 0 0 red inset"
            return (
                <div className="container" style={{paddingBottom: "20px"}}>
                    <div style={{paddingBottom: "10px"}}>
                        <ProgressBar key={"progress-bar"} courseID={assignment.getCourseid()} assignmentIndex={assignment.getOrder() - 1} submission={submission} type={"lab"} />
                    </div>
                    <table className="table table-curved table-striped">
                        <thead className={"thead-dark"}>
                            <tr>
                                <th colSpan={2}>Lab information</th>
                                <th colSpan={1}>{assignment.getName()}</th>
                            </tr>
                        </thead>
                        <tbody>
                        <tr>
                            <th colSpan={2} style={{boxShadow: boxShadow}}>Status</th>
                            <td>{SubmissionStatus[submission.getStatus()]}</td>
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
                        <tr>
                        <th colSpan={2}>Tests Passed</th>
                            <td>{passedTests}/{totalTests}</td>
                        </tr>
                        <tr>
                            <th colSpan={2}>Slip days</th>
                            <td>{enrollmentsByCourseId[assignment.getCourseid()].getSlipdaysremaining()}</td>
                        </tr>
                        <tr className={"thead-dark"}>
                            <th colSpan={1}>Test Name</th>
                            <th colSpan={1}>Score</th>
                            <th colSpan={1}>Weight</th>
                        
                        </tr>
                        {scoreObjects.map((scoreObject, index) => 
                            <ScoreObject key={index} scoreObject={scoreObject} />
                        )}
                        
                        </tbody>
                        <tfoot>
                            <tr>
                                <th>Total Score</th>
                                <th>{submission.getScore()}%</th>
                                <th>100%</th>
                            </tr>
                        </tfoot>
                    </table>
                </div>
            )
        }
        return (<div className="container"> No Submission </div>)
    }

    return (
    <div>
        <LabResult />
    </div>
    )
}

export default LabResultTable
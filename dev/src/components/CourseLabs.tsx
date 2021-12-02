import { useHistory } from "react-router"
import { getFormattedTime, SubmissionStatus } from "../Helpers"
import { useAppState } from "../overmind"
import { Submission } from "../../proto/ag/ag_pb"
import { Progress, ProgressBar } from "./ProgressBar"
import React from "react"

export const CourseLabs = ({courseID}: {courseID: number}): JSX.Element =>  {
    const state = useAppState()
    const history  = useHistory()
    
    const redirectToLab = (assignmentID: number) => {
        history.push(`/course/${courseID}/${assignmentID}`)
    }

    const Labs = (): JSX.Element[] => {
        const labs :JSX.Element[] = []
        let submission: Submission = new Submission()
        
        if (state.assignments[courseID] && state.submissions[courseID]) {
            state.assignments[courseID].forEach(assignment => {
                // Submissions are indexed by the assignment order.    
                if (state.submissions[courseID][assignment.getOrder() - 1]){
                    submission = state.submissions[courseID][assignment.getOrder() - 1]
                }
                
                labs.push(
                    <li key={assignment.getId()} className="list-group-item border"style={{marginBottom:"5px",cursor:"pointer"}} onClick={()=>redirectToLab(assignment.getId())}>
                        
                        <div className="row" >
                            <div className="col-8">
                                <strong>{assignment.getName()}</strong>
                            </div>
                            <div className="col-4 text-center">
                                <strong>Deadline:</strong>
                            </div>
                        </div>
                        <div className="row" >
                            <div className="col-5">
                                <ProgressBar courseID={courseID} assignmentIndex={assignment.getOrder() - 1} submission={submission} type={Progress.LAB}/>
                            </div>
                            <div className="col-3 text-center">
                                {(submission.getStatus() == 0 && submission.getScore() >= assignment.getScorelimit()) ? "Awaiting Approval" : SubmissionStatus[submission.getStatus()]}
                            </div>
                            <div className="col-4 text-center">
                                {getFormattedTime(assignment.getDeadline())}
                            </div>
                        </div>
                    </li>
                )
            })
        }
        return labs
    }
    return (
        <ul className="list-group">
            {Labs()}
        </ul>
    )
}
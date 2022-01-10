import { useHistory } from "react-router"
import { assignmentStatusText, getCourseID, getFormattedTime } from "../Helpers"
import { useAppState } from "../overmind"
import { Submission } from "../../proto/ag/ag_pb"
import ProgressBar, { Progress } from "./ProgressBar"
import React from "react"

/* Displays the a list of assignments and related submissions for a course */
export const CourseLabs = (): JSX.Element => {
    const state = useAppState()
    const history = useHistory()

    const courseID = getCourseID()

    const labs: JSX.Element[] = []
    let submission: Submission = new Submission()

    if (state.assignments[courseID] && state.submissions[courseID]) {
        state.assignments[courseID].forEach(assignment => {
            // Submissions are indexed by the assignment order.    
            if (state.submissions[courseID][assignment.getOrder() - 1]) {
                submission = state.submissions[courseID][assignment.getOrder() - 1]
            }

            labs.push(
                <li key={assignment.getId()} className="list-group-item border clickable mb-2" onClick={() => history.push(`/course/${courseID}/${assignment.getId()}`)}>
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
                            <ProgressBar courseID={courseID} assignmentIndex={assignment.getOrder() - 1} submission={submission} type={Progress.LAB} />
                        </div>
                        <div className="col-3 text-center">
                            {assignmentStatusText(assignment, submission)}
                        </div>
                        <div className="col-4 text-center">
                            {getFormattedTime(assignment.getDeadline())}
                        </div>
                    </div>
                </li>
            )
        })
    }
    return (
        <ul className="list-group">
            {labs}
        </ul>
    )
}

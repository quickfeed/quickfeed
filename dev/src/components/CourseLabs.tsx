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

    const redirectTo = (assignmentID: number) => {
        history.push(`/course/${courseID}/${assignmentID}`)
    }

    if (state.assignments[courseID] && state.submissions[courseID]) {
        state.assignments[courseID].forEach(assignment => {
            const assignmentIndex = assignment.getOrder() - 1
            // Submissions are indexed by the assignment order.
            const submission = state.submissions[courseID][assignmentIndex] ?? new Submission()

            labs.push(
                <li key={assignment.getId()} className="list-group-item border clickable mb-2" onClick={() => redirectTo(assignment.getId())}>
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
                            <ProgressBar courseID={courseID} assignmentIndex={assignmentIndex} submission={submission} type={Progress.LAB} />
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

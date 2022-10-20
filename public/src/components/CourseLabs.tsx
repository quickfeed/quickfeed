import { useHistory } from "react-router"
import { assignmentStatusText, getCourseID, getFormattedTime } from "../Helpers"
import { useAppState } from "../overmind"
import { Submission } from "../../gen/qf/types_pb"
import ProgressBar, { Progress } from "./ProgressBar"
import React from "react"


/* Displays the a list of assignments and related submissions for a course */
const CourseLabs = (): JSX.Element => {
    const state = useAppState()
    const history = useHistory()
    const courseID = getCourseID().toString()
    const labs: JSX.Element[] = []

    const redirectTo = (assignmentID: bigint) => {
        history.push(`/course/${courseID}/${assignmentID.toString()}`)
    }

    if (state.assignments[courseID] && state.submissions[courseID]) {
        state.assignments[courseID].forEach(assignment => {
            const assignmentIndex = assignment.order - 1
            // Submissions are indexed by the assignment order.
            const submission = state.submissions[courseID][assignmentIndex] ?? new Submission()

            labs.push(
                <li key={assignment.ID.toString()} className="list-group-item border clickable mb-2 labList" onClick={() => redirectTo(assignment.ID)}>
                    <div className="row" >
                        <div className="col-8">
                            <strong>{assignment.name}</strong>
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
                            {getFormattedTime(assignment.deadline)}
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

export default CourseLabs

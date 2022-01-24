import React from "react"
import { useHistory } from "react-router"
import { assignmentStatusText, getFormattedTime, SubmissionStatus, timeFormatter } from "../../Helpers"
import { useAppState } from "../../overmind"
import { Assignment, Submission } from "../../../proto/ag/ag_pb"
import ProgressBar, { Progress } from "../ProgressBar"


/* SubmissionsTable is a component that displays a table of assignments and their submissions for all courses. */
const SubmissionsTable = (): JSX.Element => {
    const state = useAppState()
    const history = useHistory()

    const sortedAssignments = () => {
        const assignments: Assignment[] = []
        for (const courseID in state.assignments) {
            assignments.push(...state.assignments[courseID])
        }
        assignments.sort((a, b) => {
            if (b.getDeadline() > a.getDeadline()) {
                return -1
            }
            if (a.getDeadline() > b.getDeadline()) {
                return 1
            }
            return 0
        })
        return assignments
    }

    const SubmissionsTable = (): JSX.Element[] => {
        const table: JSX.Element[] = []
        sortedAssignments().forEach(assignment => {
            const courseID = assignment.getCourseid()
            const submissions = state.submissions[courseID]
            if (!submissions) {
                return
            }
            // Submissions are indexed by the assignment order - 1.
            const submission = submissions[assignment.getOrder() - 1] ?? new Submission()
            if (submission.getStatus() !== Submission.Status.APPROVED) {
                const deadline = timeFormatter(assignment.getDeadline())
                if (deadline.daysUntil > 3 && submission.getScore() >= assignment.getScorelimit()) {
                    deadline.className = "table-success"
                }
                if (!deadline.message) {
                    return
                }
                const course = state.courses.find(course => course.getId() === courseID)
                table.push(
                    <tr key={assignment.getId()} className={"clickable-row " + deadline.className}
                        onClick={() => history.push(`/course/${courseID}/${assignment.getId()}`)}>
                        <th scope="row">{course?.getCode()}</th>
                        <td>
                            {assignment.getName()}
                            {assignment.getIsgrouplab() ?
                                <span className="badge ml-2 float-right"><i className="fa fa-users" title="Group Assignment"></i></span> : null}
                        </td>
                        <td><ProgressBar assignmentIndex={assignment.getOrder() - 1} courseID={courseID} submission={submission} type={Progress.OVERVIEW} /></td>
                        <td>{getFormattedTime(assignment.getDeadline())}</td>
                        <td>{deadline.message ? deadline.message : '--'}</td>
                        <td className={SubmissionStatus[submission.getStatus()]}>
                            {assignmentStatusText(assignment, submission)}
                        </td>
                    </tr>
                )
            }
        })
        return table
    }

    return (
        <div>
            <table className="table rounded-lg table-bordered table-hover" id="LandingPageTable">
                <thead >
                    <tr>
                        <th scope="col">Course</th>
                        <th scope="col">Assignment</th>
                        <th scope="col">Progress</th>
                        <th scope="col">Deadline</th>
                        <th scope="col">Due in</th>
                        <th scope="col">Status</th>
                    </tr>
                </thead>
                <tbody>
                    {SubmissionsTable()}
                </tbody>
            </table>
        </div>
    )
}

export default SubmissionsTable

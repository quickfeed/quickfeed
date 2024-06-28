import React from "react"
import { useHistory } from "react-router"
import { assignmentStatusText, getFormattedTime, getStatusByUser, isApproved, SubmissionStatus, timeFormatter } from "../../Helpers"
import { useAppState } from "../../overmind"
import { Assignment, Submission } from "../../../proto/qf/types_pb"
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
            if (a.deadline && b.deadline) {
                return a.deadline.toDate().getTime() - b.deadline.toDate().getTime()
            }
            return 0
        })
        return assignments
    }

    const NewSubmissionsTable = (): JSX.Element[] => {
        const table: JSX.Element[] = []
        sortedAssignments().forEach(assignment => {
            const courseID = assignment.CourseID
            const submissions = state.submissions[courseID.toString()]
            if (!submissions) {
                return
            }
            // Submissions are indexed by the assignment order - 1.
            const submission = submissions[assignment.order - 1] ?? new Submission()
            const status = getStatusByUser(submission, state.self.ID)
            if (!isApproved(status) && assignment.deadline) {
                const deadline = timeFormatter(assignment.deadline)
                if (deadline.daysUntil > 3 && submission.score >= assignment.scoreLimit) {
                    deadline.className = "table-success"
                }
                if (!deadline.message) {
                    return
                }
                const course = state.courses.find(c => c.ID === courseID)
                table.push(
                    <tr key={assignment.ID.toString()} className={`clickable-row ${deadline.className}`}
                        onClick={() => history.push(`/course/${courseID}/lab/${assignment.ID}`)}>
                        <th scope="row">{course?.code}</th>
                        <td>
                            {assignment.name}
                            {assignment.isGroupLab ?
                                <span className="badge ml-2 float-right"><i className="fa fa-users" title="Group Assignment" /></span> : null}
                        </td>
                        <td><ProgressBar assignmentIndex={assignment.order - 1} courseID={courseID.toString()} submission={submission} type={Progress.OVERVIEW} /></td>
                        <td>{getFormattedTime(assignment.deadline)}</td>
                        <td>{deadline.message ? deadline.message : '--'}</td>
                        <td className={SubmissionStatus[status]}>
                            {assignmentStatusText(assignment, submission, status)}
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
                    {NewSubmissionsTable()}
                </tbody>
            </table>
        </div>
    )
}

export default SubmissionsTable

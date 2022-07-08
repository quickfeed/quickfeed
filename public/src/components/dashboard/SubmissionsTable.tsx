import React from "react"
import { useHistory } from "react-router"
import { assignmentStatusText, getFormattedTime, SubmissionStatus, timeFormatter } from "../../Helpers"
import { useAppState } from "../../overmind"
import { Assignment, Submission } from "../../../proto/qf/types_pb"
import ProgressBar, { Progress } from "../ProgressBar"


/* SubmissionsTable is a component that displays a table of assignments and their submissions for all courses. */
const SubmissionsTable = (): JSX.Element => {
    const state = useAppState()
    const history = useHistory()

    const sortedAssignments = () => {
        const assignments: Assignment.AsObject[] = []
        for (const courseID in Object.keys(state.assignments)) {
            assignments.push(...state.assignments[courseID])
        }
        assignments.sort((a, b) => {
            if (b.deadline > a.deadline) {
                return -1
            }
            if (a.deadline > b.deadline) {
                return 1
            }
            return 0
        })
        return assignments
    }

    const SubmissionsTable = (): JSX.Element[] => {
        const table: JSX.Element[] = []
        sortedAssignments().forEach(assignment => {
            const courseID = assignment.courseid
            const submissions = state.submissions[courseID]
            if (!submissions) {
                return
            }
            // Submissions are indexed by the assignment order - 1.
            const submission = submissions[assignment.order - 1] ?? (new Submission()).toObject()
            if (submission.status !== Submission.Status.APPROVED) {
                const deadline = timeFormatter(assignment.deadline)
                if (deadline.daysUntil > 3 && submission.score >= assignment.scorelimit) {
                    deadline.className = "table-success"
                }
                if (!deadline.message) {
                    return
                }
                const course = state.courses.find(course => course.id === courseID)
                table.push(
                    <tr key={assignment.id} className={"clickable-row " + deadline.className}
                        onClick={() => history.push(`/course/${courseID}/${assignment.id}`)}>
                        <th scope="row">{course?.code}</th>
                        <td>
                            {assignment.name}
                            {assignment.isgrouplab ?
                                <span className="badge ml-2 float-right"><i className="fa fa-users" title="Group Assignment"  /></span> : null}
                        </td>
                        <td><ProgressBar assignmentIndex={assignment.order - 1} courseID={courseID} submission={submission} type={Progress.OVERVIEW} /></td>
                        <td>{getFormattedTime(assignment.deadline)}</td>
                        <td>{deadline.message ? deadline.message : '--'}</td>
                        <td className={SubmissionStatus[submission.status]}>
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

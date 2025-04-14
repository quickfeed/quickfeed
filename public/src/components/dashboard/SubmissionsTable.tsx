import React from "react"
import { useHistory } from "react-router"
import { assignmentStatusText, getFormattedTime, getStatusByUser, Icon, isApproved, SubmissionStatus, timeFormatter } from "../../Helpers"
import { useAppState } from "../../overmind"
import { Assignment, Enrollment_UserStatus, SubmissionSchema } from "../../../proto/qf/types_pb"
import ProgressBar, { Progress } from "../ProgressBar"
import { create } from "@bufbuild/protobuf"
import { timestampDate } from "@bufbuild/protobuf/wkt"

/* SubmissionsTable is a component that displays a table of assignments and their submissions for all courses. */
const SubmissionsTable = () => {
    const state = useAppState()
    const history = useHistory()

    const sortedAssignments = () => {
        const assignments: Assignment[] = []
        for (const courseID in state.assignments) {
            assignments.push(...state.assignments[courseID])
        }
        assignments.sort((a, b) => {
            if (a.deadline && b.deadline) {
                return timestampDate(a.deadline).getTime() - timestampDate(b.deadline).getTime()
            }
            return 0
        })
        return assignments
    }

    const table: React.JSX.Element[] = []
    sortedAssignments().forEach(assignment => {
        const courseID = assignment.CourseID
        const submissions = state.submissions.ForAssignment(assignment)
        if (!submissions) {
            return
        }
        if (state.enrollmentsByCourseID[courseID.toString()]?.status !== Enrollment_UserStatus.STUDENT) {
            return
        }
        // Submissions are indexed by the assignment order - 1.
        const submission = submissions.find(sub => sub.AssignmentID === assignment.ID) ?? create(SubmissionSchema)
        const status = getStatusByUser(submission, state.self.ID)
        if (!isApproved(status) && assignment.deadline) {
            const date = timestampDate(assignment.deadline)
            const now = new Date()

            /*
                Only show assignments which are the same year as now and if one month after deadline month.
                This way students will only see assignments which are the current year and the expired assignments
                are shown for a month after the deadline.
            */
            if (date.getFullYear() !== now.getFullYear() || date.getMonth() > now.getMonth() + 1) {
                return
            }

            const deadline = timeFormatter(date, submission.score >= assignment.scoreLimit)
            const course = state.courses.find(c => c.ID === courseID)
            table.push(
                <tr key={assignment.ID.toString()} className={`clickable-row ${deadline.className}`}
                    onClick={() => history.push(`/course/${courseID}/lab/${assignment.ID}`)}>
                    <th scope="row">{course?.code}</th>
                    <td>
                        {assignment.name}
                        {assignment.isGroupLab ?
                            <span className="badge ml-2 float-right"><i className={Icon.GROUP} title="Group Assignment" /></span> : null}
                    </td>
                    <td><ProgressBar courseID={courseID.toString()} submission={submission} type={Progress.OVERVIEW} /></td>
                    <td>{getFormattedTime(assignment.deadline, true)}</td>
                    <td>{deadline.message}</td>
                    <td className={SubmissionStatus[status]}>
                        {assignmentStatusText(assignment, submission, status)}
                    </td>
                </tr>
            )
        }
    })

    return (
        table.length !== 0 ? (
            <div>
                <h2> Assignment Deadlines </h2>
                <table className="table rounded-lg table-bordered table-hover" id="LandingPageTable">
                    <thead>
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
                        {table}
                    </tbody>
                </table>
            </div>
        ) : null
    )
}

export default SubmissionsTable

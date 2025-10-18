import React from "react"
import { useNavigate } from "react-router"
import { assignmentStatusText, getStatusByUser, Icon, isApproved, isExpired, SubmissionStatus, deadlineFormatter } from "../../Helpers"
import { useAppState } from "../../overmind"
import { Assignment, Enrollment_UserStatus, SubmissionSchema } from "../../../proto/qf/types_pb"
import ProgressBar, { Progress } from "../ProgressBar"
import { create } from "@bufbuild/protobuf"
import { timestampDate } from "@bufbuild/protobuf/wkt"

/* SubmissionsTable is a component that displays a table of assignments and their submissions for all courses. */
const SubmissionsTable = () => {
    const state = useAppState()
    const navigate = useNavigate()

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
        const deadline = assignment.deadline
        if (!deadline || isExpired(deadline)) {
            return // ignore expired assignments
        }
        const courseID = assignment.CourseID
        const submissions = state.submissions.ForAssignment(assignment)
        if (submissions.length === 0) {
            return
        }
        if (state.enrollmentsByCourseID[courseID.toString()]?.status !== Enrollment_UserStatus.STUDENT) {
            return
        }
        const submission = submissions.find(sub => sub.AssignmentID === assignment.ID) ?? create(SubmissionSchema)
        if (!assignment.isGroupLab && submission.groupID !== 0n) {
            return // ignore group submissions for individual assignments
        }
        const status = getStatusByUser(submission, state.self.ID)
        if (!isApproved(status)) {
            const deadlineInfo = deadlineFormatter(deadline, assignment.scoreLimit, submission.score)
            const course = state.courses.find(c => c.ID === courseID)
            table.push(
                <tr key={assignment.ID.toString()} className={`clickable-row ${deadlineInfo.className}`}
                    onClick={() => navigate(`/course/${courseID}/lab/${assignment.ID}`)}>
                    <th scope="row">{course?.code}</th>
                    <td>
                        {assignment.name}
                        {assignment.isGroupLab ?
                            <span className="badge ml-2 float-right"><i className={Icon.GROUP} title="Group Assignment" /></span> : null}
                    </td>
                    <td><ProgressBar courseID={courseID.toString()} submission={submission} type={Progress.OVERVIEW} /></td>
                    <td>{deadlineInfo.time}</td>
                    <td>{deadlineInfo.message}</td>
                    <td className={SubmissionStatus[status]}>
                        {assignmentStatusText(assignment, submission, status)}
                    </td>
                </tr>
            )
        }
    })

    if (table.length === 0) {
        return null
    }
    return (
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
    )
}

export default SubmissionsTable

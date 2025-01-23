import React from "react"
import { assignmentStatusText, getFormattedTime, getPassedTestsCount, getStatusByUser, isAllApproved, isManuallyGraded } from "../../Helpers"
import { Assignment, Submission } from "../../../proto/qf/types_pb"
import { useAppState } from "../../overmind"

type SubmissionInfoProps = {
    submission: Submission
    assignment: Assignment
}

const SubmissionInfo = ({ submission, assignment }: SubmissionInfoProps) => {
    const state = useAppState()
    const enrollment = state.selectedEnrollment ?? state.enrollmentsByCourseID[assignment.CourseID.toString()]
    const buildInfo = submission.BuildInfo
    const delivered = getFormattedTime(buildInfo?.SubmissionDate)
    const built = getFormattedTime(buildInfo?.BuildDate)
    const executionTime = buildInfo ? `${buildInfo.ExecTime / BigInt(1000)} seconds` : ""
    
    const status = getStatusByUser(submission, enrollment.userID)
    const className = isAllApproved(submission) ? "passed" : "failed"
    return (
        <table className="table table-curved table-striped">
            <thead className="thead-dark">
                <tr>
                    <th colSpan={2}>Lab information</th>
                    <th>{assignment.name}</th>
                </tr>
            </thead>
            <tbody>
                <tr>
                    <td colSpan={2} className={className}>
                        Status
                    </td>
                    <td>{assignmentStatusText(assignment, submission, status)}</td>
                </tr>
                <tr>
                    <td colSpan={2}>Delivered</td>
                    <td>{delivered}</td>
                </tr>
                <tr>
                    <td colSpan={2}>Built</td>
                    <td>{built}</td>
                </tr>
                {
                    // Only render row if submission has an approved date
                    submission.approvedDate ? (
                        <tr>
                            <td colSpan={2}>Approved</td>
                            <td>{getFormattedTime(submission.approvedDate)}</td>
                        </tr>
                    ) : null
                }
                <tr>
                    <td colSpan={2}>Deadline</td>
                    <td>{getFormattedTime(assignment.deadline, true)}</td>
                </tr>

                {!isManuallyGraded(assignment) ? (
                    <tr>
                        <td colSpan={2}>Tests Passed</td>
                        <td>{getPassedTestsCount(submission.Scores)}</td>
                    </tr>
                ) : null}
                <tr>
                    <td colSpan={2}>Execution time</td>
                    <td>{executionTime}</td>
                </tr>
                <tr>
                    <td colSpan={2}>Slip days</td>
                    <td>{enrollment.slipDaysRemaining}</td>
                </tr>
            </tbody>
        </table>
    )
}

export default SubmissionInfo

import React from "react"
import { assignmentStatusText, getFormattedTime, getPassedTestsCount, getStatusByUser, isAllApproved, isManuallyGraded } from "../../Helpers"
import { Assignment, Submission, UsedSlipDays } from "../../../proto/qf/types_pb"
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

    const isGroupSubmission = submission.groupID > 0n
    const group = isGroupSubmission
        ? (state.groups[assignment.CourseID.toString()]?.find(g => g.ID === submission.groupID)
            ?? state.userGroup[assignment.CourseID.toString()])
        : undefined

    const status = getStatusByUser(submission, enrollment.userID)
    const className = isAllApproved(submission) ? "passed" : "failed"
    return (
        <table className="table table-zebra">
            <thead className="bg-base-300 text-base-content">
                <tr className="text-lg">
                    <th colSpan={2}>Lab information</th>
                    <th>{assignment.name}</th>
                </tr>
            </thead>
            <tbody>
                <tr>
                    <td colSpan={2} className={`${className} pl-3!`}>
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

                {!isManuallyGraded(assignment.reviewers) ? (
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
                    <td colSpan={2}>{isGroupSubmission ? "Group slip days" : "Slip days"}</td>
                    <td>{isGroupSubmission ? group?.slipDaysRemaining : enrollment.slipDaysRemaining}</td>
                </tr>
                <tr>
                    <td colSpan={2}>{isGroupSubmission ? "Used group slip days for this assignment" : "Used slip days for this assignment"}</td>
                    <td>
                        {isGroupSubmission
                            ? usedSlipdaysRows(assignment, group?.usedSlipDays ?? [])
                            : usedSlipdaysRows(assignment, enrollment.usedSlipDays ?? [])}
                    </td>
                </tr>
            </tbody>
        </table>
    )
}

function usedSlipdaysRows(assignment: Assignment, usedSlipDays: UsedSlipDays[]): React.JSX.Element {
    // returns a table row if there exists some used slip days for the assignment, otherwise returns nothing
    if (usedSlipDays.length === 0) {
        return <></>
    }

    const rows = usedSlipDays
        .filter(slipDay => slipDay.assignmentID === assignment.ID)
        .map((slipDay, index) => {
            return (
                <span key={index}>
                    {slipDay.usedDays}
                </span>
            )
        })
    return <>{rows}</>
}

export default SubmissionInfo

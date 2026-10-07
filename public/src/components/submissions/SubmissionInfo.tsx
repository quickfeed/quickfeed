import { RunStatus } from "../../../proto/kit/score/score_pb"
import type { Assignment, Submission, UsedSlipDays } from "../../../proto/qf/types_pb"
import { Submission_Status } from "../../../proto/qf/types_pb"
import { assignmentStatusText, getFormattedTime, getPassedTestsCount, getStatusByUser, isAllApproved, isManuallyGraded, runFailureScoreText, runFailureText, runsTests } from "../../Helpers"
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
                {buildInfo && buildInfo.Status !== RunStatus.SUCCESS ? (
                    <tr>
                        <td colSpan={3} className="failed pl-3!">
                            Last run failed: {runFailureText(buildInfo.Status)} {runFailureScoreText(buildInfo.Status)}
                        </td>
                    </tr>
                ) : null}
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

                {runsTests(assignment) ? (
                    <tr>
                        <td colSpan={2}>Tests Passed</td>
                        <td>{getPassedTestsCount(submission.Scores)}</td>
                    </tr>
                ) : null}
                {runsTests(assignment) && isManuallyGraded(assignment.reviewers) ? (
                    <ScoreRows submission={submission} assignment={assignment} status={status} isTeacher={state.isTeacher} />
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

/** ScoreRows shows how the test and review scores of a tested and reviewed
 *  assignment make up the submission's score. A student does not see the
 *  review score until the submission is graded. */
const ScoreRows = ({ submission, assignment, status, isTeacher }: { submission: Submission, assignment: Assignment, status: Submission_Status, isTeacher: boolean }) => {
    const reviewWeight = assignment.reviewWeight
    const reviewHidden = !isTeacher && status === Submission_Status.NONE
    return (
        <>
            <tr>
                <td colSpan={2}>Test score</td>
                <td title={`Tests count for ${100 - reviewWeight}% of the score`}>{submission.testScore}% (weight {100 - reviewWeight}%)</td>
            </tr>
            <tr>
                <td colSpan={2}>Review score</td>
                <td title={`Reviews count for ${reviewWeight}% of the score`}>
                    {reviewHidden ? "Not graded yet" : `${submission.reviewScore}% (weight ${reviewWeight}%)`}
                </td>
            </tr>
            {reviewHidden ? null : (
                <tr>
                    <td colSpan={2}>Total score</td>
                    <td>{submission.score}%</td>
                </tr>
            )}
        </>
    )
}

function usedSlipdaysRows(assignment: Assignment, usedSlipDays: UsedSlipDays[]): React.ReactNode {
    // returns a table row if there exists some used slip days for the assignment, otherwise returns nothing
    if (usedSlipDays.length === 0) {
        return null
    }

    return usedSlipDays
        .filter(slipDay => slipDay.assignmentID === assignment.ID)
        .map(slipDay => (
            <span key={slipDay.ID}>
                {slipDay.usedDays}
            </span>
        ))
}

export default SubmissionInfo

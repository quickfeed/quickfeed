import React, { useCallback } from "react"
import { Assignment, Submission, Submission_Status } from "../../proto/qf/types_pb"
import { assignmentStatusText, getFormattedTime, getPassedTestsCount, isManuallyGraded } from "../Helpers"
import { useAppState } from "../overmind"
import ProgressBar, { Progress } from "./ProgressBar"
import SubmissionScore from "./SubmissionScore"

interface lab {
    submission: Submission
    assignment: Assignment
}

type ScoreSort = "name" | "score" | "weight"

const LabResultTable = ({ submission, assignment }: lab): JSX.Element => {
    const state = useAppState()

    const [sortKey, setSortKey] = React.useState<ScoreSort>("name")
    const [sortAscending, setSortAscending] = React.useState<boolean>(true)

    const sortScores = () => {
        const sortBy = sortAscending ? 1 : -1
        const scores = submission.clone().Scores
        return scores.sort((a, b) => {
            switch (sortKey) {
                case "name":
                    return sortBy * (a.TestName.localeCompare(b.TestName))
                case "score":
                    return sortBy * (a.Score - b.Score)
                case "weight":
                    return sortBy * (a.Weight - b.Weight)
                default:
                    return 0
            }
        })
    }

    const handleSort = useCallback((event: React.MouseEvent<HTMLTableCellElement>) => {
        const key = event.currentTarget.dataset["key"] as ScoreSort
        if (sortKey === key) {
            setSortAscending(!sortAscending)
        } else {
            setSortKey(key)
            setSortAscending(true)
        }
    }, [sortKey, sortAscending])

    const sortedScores = React.useMemo(sortScores, [submission, sortKey, sortAscending])

    if (submission && assignment) {
        const enrollment = state.activeEnrollment ?? state.enrollmentsByCourseID[assignment.CourseID.toString()]
        const buildInfo = submission.BuildInfo
        const delivered = getFormattedTime(buildInfo?.SubmissionDate)
        const built = getFormattedTime(buildInfo?.BuildDate)
        const executionTime = buildInfo ? `${buildInfo.ExecTime / BigInt(1000)} seconds` : ""

        const className = (submission.status === Submission_Status.APPROVED) ? "passed" : "failed"
        return (
            <div className="pb-2">
                <div className="pb-2">
                    <ProgressBar key={"progress-bar"} courseID={assignment.CourseID.toString()} assignmentIndex={assignment.order - 1} submission={submission} type={Progress.LAB} />
                </div>
                <table className="table table-curved table-striped">
                    <thead className={"thead-dark"}>
                        <tr>
                            <th colSpan={2}>Lab information</th>
                            <th colSpan={1}>{assignment.name}</th>
                        </tr>
                    </thead>
                    <tbody>
                        <tr>
                            <td colSpan={2} className={className}>Status</td>
                            <td>{assignmentStatusText(assignment, submission)}</td>
                        </tr>
                        <tr>
                            <td colSpan={2}>Delivered</td>
                            <td>{delivered}</td>
                        </tr>
                        <tr>
                            <td colSpan={2}>Built</td>
                            <td>{built}</td>
                        </tr>
                        { // Only render row if submission has an approved date
                            submission.approvedDate ?
                                <tr>
                                    <td colSpan={2}>Approved</td>
                                    <td>{getFormattedTime(submission.approvedDate)}</td>
                                </tr>
                                : null
                        }
                        <tr>
                            <td colSpan={2}>Deadline</td>
                            <td>{assignment.deadline ? getFormattedTime(assignment.deadline) : "N/A"}</td>
                        </tr>

                        {!isManuallyGraded(assignment) ?
                            <tr>
                                <td colSpan={2}>Tests Passed</td>
                                <td>{getPassedTestsCount(submission.Scores)}</td>
                            </tr>
                            : null
                        }
                        <tr>
                            <td colSpan={2}>Execution time</td>
                            <td>{executionTime}</td>
                        </tr>
                        <tr>
                            <td colSpan={2}>Slip days</td>
                            <td>{
                                enrollment.slipDaysRemaining
                            }</td>
                        </tr>
                        <tr className={"thead-dark"}>
                            <th colSpan={1} data-key={"name"} role="button" onClick={handleSort}>Test Name</th>
                            <th colSpan={1} data-key={"score"} role="button" onClick={handleSort}>Score</th>
                            <th colSpan={1} data-key={"weight"} role="button" onClick={handleSort}>Weight</th>

                        </tr>
                        {sortedScores.map(score =>
                            <SubmissionScore key={score.ID.toString()} score={score} />
                        )}

                    </tbody>
                    <tfoot>
                        <tr>
                            <th>Total Score</th>
                            <th>{submission.score}%</th>
                            <th>100%</th>
                        </tr>
                    </tfoot>
                </table>
            </div>
        )
    }
    return <div className="container"> No Submission </div>
}

export default LabResultTable

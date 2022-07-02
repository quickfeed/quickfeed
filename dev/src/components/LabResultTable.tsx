import React, { useCallback } from "react"
import { Assignment, Submission } from "../../proto/ag/ag_pb"
import { Converter } from "../convert"
import { assignmentStatusText, getFormattedTime, getPassedTestsCount, isManuallyGraded } from "../Helpers"
import { useAppState } from "../overmind"
import ProgressBar, { Progress } from "./ProgressBar"
import SubmissionScore from "./SubmissionScore"

interface lab {
    submission: Submission.AsObject
    assignment: Assignment.AsObject
}

type ScoreSort = "name" | "score" | "weight"

const LabResultTable = ({ submission, assignment }: lab): JSX.Element => {
    const state = useAppState()

    const [sortKey, setSortKey] = React.useState<ScoreSort>("name")
    const [sortAscending, setSortAscending] = React.useState<boolean>(true)

    const sortScores = () => {
        const sortBy = sortAscending ? 1 : -1
        const scores = Converter.clone(submission.scoresList)
        return scores.sort((a, b) => {
            switch (sortKey) {
                case "name":
                    return sortBy * (a.testname.localeCompare(b.testname))
                case "score":
                    return sortBy * (a.score - b.score)
                case "weight":
                    return sortBy * (a.weight - b.weight)
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
        const enrollment = state.activeEnrollment ?? state.enrollmentsByCourseID[assignment.courseid]
        const buildInfo = submission.buildinfo
        const delivered = buildInfo ? getFormattedTime(buildInfo.builddate) : "N/A"
        const executionTime = buildInfo ? `${buildInfo.exectime / 1000} seconds` : ""

        const className = (submission.status === Submission.Status.APPROVED) ? "passed" : "failed"
        return (
            <div className="pb-2">
                <div className="pb-2">
                    <ProgressBar key={"progress-bar"} courseID={assignment.courseid} assignmentIndex={assignment.order - 1} submission={submission} type={Progress.LAB} />
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
                            <th colSpan={2} className={className}>Status</th>
                            <td>{assignmentStatusText(assignment, submission)}</td>
                        </tr>
                        <tr>
                            <th colSpan={2}>Delivered</th>
                            <td>{delivered}</td>
                        </tr>
                        { // Only render row if submission has an approved date
                            submission.approveddate ?
                                <tr>
                                    <th colSpan={2}>Approved</th>
                                    <td>{getFormattedTime(submission.approveddate)}</td>
                                </tr>
                                : null
                        }
                        <tr>
                            <th colSpan={2}>Deadline</th>
                            <td>{getFormattedTime(assignment.deadline)}</td>
                        </tr>

                        {!isManuallyGraded(assignment) ?
                            <tr>
                                <th colSpan={2}>Tests Passed</th>
                                <td>{getPassedTestsCount(submission.scoresList)}</td>
                            </tr>
                            : null
                        }
                        <tr>
                            <th colSpan={2}>Execution time</th>
                            <td>{executionTime}</td>
                        </tr>
                        <tr>
                            <th colSpan={2}>Slip days</th>
                            <td>{
                                enrollment.slipdaysremaining
                            }</td>
                        </tr>
                        <tr className={"thead-dark"}>
                            <th colSpan={1} data-key={"name"} role="button" onClick={handleSort}>Test Name</th>
                            <th colSpan={1} data-key={"score"} role="button" onClick={handleSort}>Score</th>
                            <th colSpan={1} data-key={"weight"} role="button" onClick={handleSort}>Weight</th>

                        </tr>
                        {sortedScores.map(score =>
                            <SubmissionScore key={score.id} score={score} />
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

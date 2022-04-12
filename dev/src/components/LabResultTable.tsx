import React from "react"
import { Assignment, Submission } from "../../proto/ag/ag_pb"
import { assignmentStatusText, getFormattedTime, getPassedTestsCount, isManuallyGraded } from "../Helpers"
import { useAppState } from "../overmind"
import ProgressBar, { Progress } from "./ProgressBar"
import SubmissionScore from "./SubmissionScore"

interface lab {
    submission: Submission.AsObject
    assignment: Assignment.AsObject
}

const LabResultTable = ({ submission, assignment }: lab): JSX.Element => {
    const state = useAppState()

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
                            <th colSpan={1}>Test Name</th>
                            <th colSpan={1}>Score</th>
                            <th colSpan={1}>Weight</th>

                        </tr>
                        {submission.scoresList.map(score =>
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
    return (<div className="container"> No Submission </div>)
}

export default LabResultTable

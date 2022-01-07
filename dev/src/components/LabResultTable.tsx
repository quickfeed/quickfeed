import { json } from "overmind"
import React from "react"
import { Assignment, Submission } from "../../proto/ag/ag_pb"
import { Score } from "../../proto/kit/score/score_pb"
import { generateStatusText, getFormattedTime, getPassedTestsCount, isManuallyGraded } from "../Helpers"
import { useAppState } from "../overmind"
import ProgressBar, { Progress } from "./ProgressBar"

interface lab {
    submission: Submission
    assignment: Assignment
}

const LabResultTable = ({ submission, assignment }: lab): JSX.Element => {
    const state = useAppState()

    const Score = ({ score }: { score: Score }) => {
        const classname = (score.getScore() === score.getMaxscore()) ? "passed" : "failed"
        return (
            <tr>
                <th className={classname + " pl-4"}>
                    {score.getTestname()}
                </th>
                <th>
                    {score.getScore()}/{score.getMaxscore()}
                </th>
                <th>
                    {score.getWeight()}
                </th>
            </tr>
        )
    }

    const LabResult = (): JSX.Element => {
        if (submission && assignment) {
            const buildInfo = submission.getBuildinfo()
            const delivered = buildInfo ? getFormattedTime(buildInfo.getBuilddate()) : "N/A"
            const executionTime = buildInfo ? `${buildInfo.getExectime() / 1000} seconds` : ""

            const classname = (submission.getStatus() === Submission.Status.APPROVED) ? "passed" : "failed"
            return (
                <div className="pb-2">
                    <div className="pb-2">
                        <ProgressBar key={"progress-bar"} courseID={assignment.getCourseid()} assignmentIndex={assignment.getOrder() - 1} submission={submission} type={Progress.LAB} />
                    </div>
                    <table className="table table-curved table-striped">
                        <thead className={"thead-dark"}>
                            <tr>
                                <th colSpan={2}>Lab information</th>
                                <th colSpan={1}>{assignment.getName()}</th>
                            </tr>
                        </thead>
                        <tbody>
                            <tr>
                                <th colSpan={2} className={classname}>Status</th>
                                <td>{generateStatusText(assignment, submission)}</td>
                            </tr>
                            <tr>
                                <th colSpan={2}>Delivered</th>
                                <td>{delivered}</td>
                            </tr>
                            { // Only render row if submission has an approved date
                                submission.getApproveddate() ?
                                    <tr>
                                        <th colSpan={2}>Approved</th>
                                        <td>{getFormattedTime(submission.getApproveddate())}</td>
                                    </tr>
                                    : null
                            }
                            <tr>
                                <th colSpan={2}>Deadline</th>
                                <td>{getFormattedTime(assignment.getDeadline())}</td>
                            </tr>

                            {!isManuallyGraded(assignment) ?
                                <tr>
                                    <th colSpan={2}>Tests Passed</th>
                                    <td>{getPassedTestsCount(json(submission).getScoresList())}</td>
                                </tr>
                                : null
                            }
                            <tr>
                                <th colSpan={2}>Execution time</th>
                                <td>{executionTime}</td>
                            </tr>
                            <tr>
                                <th colSpan={2}>Slip days</th>
                                <td>{state.enrollmentsByCourseId[assignment.getCourseid()].getSlipdaysremaining()}</td>
                            </tr>
                            <tr className={"thead-dark"}>
                                <th colSpan={1}>Test Name</th>
                                <th colSpan={1}>Score</th>
                                <th colSpan={1}>Weight</th>

                            </tr>
                            {json(submission).getScoresList().map((score, index) =>
                                <Score key={index} score={score} />
                            )}

                        </tbody>
                        <tfoot>
                            <tr>
                                <th>Total Score</th>
                                <th>{submission.getScore()}%</th>
                                <th>100%</th>
                            </tr>
                        </tfoot>
                    </table>
                </div>
            )
        }
        return (<div className="container"> No Submission </div>)
    }

    return (
        <div>
            <LabResult />
        </div>
    )
}

export default LabResultTable
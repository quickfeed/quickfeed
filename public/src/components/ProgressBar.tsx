import React from "react"
import { useAppState } from "../overmind"
import { Assignment, Submission, Submission_Status } from "../../proto/qf/types_pb"
import { getStatusByUser } from "../Helpers"
import SubmissionTypeIcon from "./student/SubmissionTypeIcon"

export enum Progress {
    NAV,
    LAB,
    OVERVIEW
}

type ProgressBarProps = {
    courseID: string,
    submission: Submission,
    type: Progress
}

const ProgressBar = ({ courseID, submission, type }: ProgressBarProps) => {
    const state = useAppState()

    const assignment = state.assignments[courseID]?.find(assignment => assignment.ID === submission.AssignmentID)

    const score = submission.score ?? 0
    const scorelimit = assignment?.scoreLimit ?? 0
    const status = getStatusByUser(submission, state.self.ID)
    const secondaryProgress = scorelimit - score
    // Returns a thin line to be used for labs in the NavBar
    if (type === Progress.NAV) {
        const percentage = 100 - score
        const color = score >= scorelimit ? "green" : "yellow"
        return (
            <div style={{
                position: "absolute",
                borderBottom: `2px solid ${color}`,
                bottom: 0,
                left: 0,
                right: `${percentage}%`,
                opacity: 0.3
            }} />
        )
    }

    let text = ""
    let secondaryText = ""
    if (type === Progress.LAB) {
        text = `${score} %`
        secondaryText = `${secondaryProgress} %`
    }

    // Returns a regular size progress bar to be used for labs
    let color = ""
    if (type > Progress.NAV) {
        switch (status) {
            case Submission_Status.NONE:
                color = "bg-primary"
                break
            case Submission_Status.APPROVED:
                color = "bg-success"
                break
            case Submission_Status.REJECTED:
                color = "bg-danger"
                break
            case Submission_Status.REVISION:
                color = "bg-warning text-dark"
                break
        }
    }

    return (
        <div className="progress">
            <PrimaryProgressBar color={color} score={score} text={text} />
            {secondaryProgress > 0 &&
                <SecondaryProgressBar progress={secondaryProgress} text={secondaryText} />
            }
        </div>
    )
}

export default ProgressBar

// DefaultProgressBar is a function that returns a progress bar for a lab/assignment with no submissions
export const DefaultProgressBar = ({ assignment }: { assignment: Assignment }) => {
    return (
        <div className="row mb-1 py-2 align-items-center text-left">
            <div className="col-8">
                <div className="progress">
                    <PrimaryProgressBar score={0} text={"0 %"} />
                    <SecondaryProgressBar progress={assignment.scoreLimit} text={`${assignment.scoreLimit} %`} />
                </div>
            </div>
            <SubmissionTypeIcon solo={!assignment.isGroupLab} />
            <div className="col-3">
                No submission
            </div>
        </div>
    )
}


const PrimaryProgressBar = ({ color, score, text }: { color?: string, score: number, text: string }) => {
    return <div
        className={`progress-bar ${color}`}
        role="progressbar"
        style={{ width: `${score}%`, transitionDelay: "0.5s" }}
        aria-valuenow={score}
        aria-valuemin={0}
        aria-valuemax={100}
    >
        {text}
    </div>
}

const SecondaryProgressBar = ({ progress, text }: { progress: number, text: string }) => {
    return <div
        className={"progress-bar progressbar-secondary bg-secondary"}
        role="progressbar"
        style={{ width: `${progress}%` }}
        aria-valuemax={100}
    >
        {text}
    </div>
}

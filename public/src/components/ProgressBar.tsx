import React from "react"
import { useAppState } from "../overmind"
import { Submission, Submission_Status } from "../../proto/qf/types_pb"

export enum Progress {
    NAV,
    LAB,
    OVERVIEW
}

type ProgressBarProps = {
    courseID: string,
    assignmentIndex: number,
    submission?: Submission,
    type: Progress
}

const ProgressBar = ({ courseID, assignmentIndex, submission, type }: ProgressBarProps): JSX.Element => {
    const state = useAppState()

    const sub = submission
        ? submission
        : state.submissions[courseID][assignmentIndex]

    const assignment = state.assignments[courseID][assignmentIndex]

    const score = sub?.score ?? 0
    const scorelimit = assignment?.scoreLimit ?? 0
    const status = sub?.status ?? Submission_Status.NONE
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
            <div
                className={"progress-bar " + color}
                role="progressbar"
                style={{ width: score + "%", transitionDelay: "0.5s" }}
                aria-valuenow={score}
                aria-valuemin={0}
                aria-valuemax={100}
            >
                {text}
            </div>
            {secondaryProgress > 0 &&
                <div
                    className={"progress-bar progressbar-secondary bg-secondary"}
                    role="progressbar"
                    style={{ width: secondaryProgress + "%" }}
                    aria-valuemax={100}
                >
                    {secondaryText}
                </div>
            }
        </div>
    )
}

export default ProgressBar

import React from "react"
import { useAppState } from "../overmind"
import { Submission } from "../../proto/ag/ag_pb"

export enum Progress {
    NAV,
    LAB,
    OVERVIEW
}

const ProgressBar = (props: { courseID: number, assignmentIndex: number, submission?: Submission.AsObject, type: Progress }): JSX.Element => {
    const state = useAppState()

    const submission = props.submission
        ? props.submission
        : state.submissions[props.courseID][props.assignmentIndex]

    const assignment = state.assignments[props.courseID][props.assignmentIndex]

    const score = submission?.score ?? 0
    const scorelimit = assignment?.scorelimit ?? 0
    const status = submission?.status ?? Submission.Status.NONE
    const secondaryProgress = scorelimit - score
    // Returns a thin line to be used for labs in the NavBar
    if (props.type === Progress.NAV) {
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
            }}>
            </div>
        )
    }

    let text = ""
    let secondaryText = ""
    if (props.type === Progress.LAB) {
        text = `${score} %`
        secondaryText = `${secondaryProgress} %`
    }

    // Returns a regular size progress bar to be used for labs
    let color = ""
    if (props.type > Progress.NAV) {
        switch (status) {
            case Submission.Status.NONE:
                color = "bg-primary"
                break
            case Submission.Status.APPROVED:
                color = "bg-success"
                break
            case Submission.Status.REJECTED:
                color = "bg-danger"
                break
            case Submission.Status.REVISION:
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

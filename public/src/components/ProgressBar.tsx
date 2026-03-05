import React from "react"
import { useAppState } from "../overmind"
import { Submission, Submission_Status } from "../../proto/qf/types_pb"
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
    const remainingToPass = Math.max(0, scorelimit - score)
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
        secondaryText = `${remainingToPass} %`
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
        <div className="relative w-full h-8 bg-base-300 rounded-lg overflow-hidden">
            <PrimaryProgressBar color={color} score={score} text={text} />
            {remainingToPass > 0 &&
                <SecondaryProgressBar startPosition={score} width={remainingToPass} text={secondaryText} />
            }
        </div>
    )
}

export default ProgressBar

// DefaultProgressBar is a function that returns a progress bar for a lab/assignment with no submissions
export const DefaultProgressBar = ({ scoreLimit, isGroupLab }: { scoreLimit: number, isGroupLab: boolean }) => {
    return (
        <div className="flex items-center gap-4 py-3 px-2 mb-2 rounded-lg">
            <div className="flex-1 min-w-0">
                <div className="relative w-full h-8 bg-base-300 rounded-lg overflow-hidden">
                    <PrimaryProgressBar score={0} text={"0 %"} />
                    <SecondaryProgressBar startPosition={0} width={scoreLimit} text={`${scoreLimit} %`} />
                </div>
            </div>
            <div className="flex-shrink-0 w-10 flex items-center justify-center">
                <SubmissionTypeIcon solo={!isGroupLab} />
            </div>
            <div className="flex-shrink-0 w-40 text-sm font-medium text-right">
                No submission
            </div>
        </div>
    )
}


const PrimaryProgressBar = ({ color, score, text }: { color?: string, score: number, text: string }) => {
    return (
        <div
            className={`absolute top-0 left-0 h-full ${color || 'bg-primary'} flex items-center justify-center text-xs font-semibold text-primary-content transition-all duration-500`}
            role="progressbar"
            style={{ width: `${score}%` }}
            aria-valuenow={score}
            aria-valuemin={0}
            aria-valuemax={100}
        >
            {score > 10 && text}
        </div>
    )
}

const SecondaryProgressBar = ({ startPosition, width, text }: { startPosition: number, width: number, text: string }) => {
    return (
        <div
            className="absolute top-0 h-full bg-base-content/20 flex items-center justify-center text-xs font-semibold text-base-content transition-all duration-300"
            role="progressbar"
            style={{ left: `${startPosition}%`, width: `${width}%` }}
            aria-valuemax={100}
        >
            {width > 10 && text}
        </div>
    )
}

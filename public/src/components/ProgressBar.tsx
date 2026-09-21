import type { Submission } from "../../proto/qf/types_pb"
import { Submission_Status } from "../../proto/qf/types_pb"
import { getEffectiveStatus } from "../Helpers"
import { useAppState } from "../overmind"
import SubmissionTypeIcon from "./student/SubmissionTypeIcon"

type ProgressBarProps = {
    courseID: string,
    submission: Submission,
    showText?: boolean
}

const ProgressBar = ({ courseID, submission, showText = true }: ProgressBarProps) => {
    const state = useAppState()

    const assignment = state.assignments[courseID]?.find(assignment => assignment.ID === submission.AssignmentID)

    const score = submission.score ?? 0
    const scorelimit = assignment?.scoreLimit ?? 0
    const status = getEffectiveStatus(submission, state.self.ID)
    const remainingToPass = Math.max(0, scorelimit - score)

    let text = ""
    let secondaryText = ""
    if (showText) {
        text = `${score} %`
        secondaryText = `${remainingToPass} %`
    }

    let color = ""
    switch (status) {
        case Submission_Status.NONE:
            color = "bg-primary"
            break
        case Submission_Status.APPROVED:
            color = "bg-success"
            break
        case Submission_Status.REJECTED:
            color = "bg-error"
            break
        case Submission_Status.REVISION:
            color = "bg-warning text-dark"
            break
    }

    return (
        <ProgressBarView
            color={color}
            score={score}
            remainingToPass={remainingToPass}
            text={text}
            secondaryText={secondaryText}
        />
    )
}

export default ProgressBar

type ProgressBarViewProps = {
    score: number
    remainingToPass: number
    color?: string
    text?: string
    secondaryText?: string
}

// ProgressBarView is the presentation-only bar. ProgressBar resolves the score,
// the status color and the pass threshold from application state and delegates
// here, so callers without that state, such as the About page previews, can
// render the same bar from plain values.
export const ProgressBarView = ({ score, remainingToPass, color, text = "", secondaryText = "" }: ProgressBarViewProps) => (
    <div className="relative w-full h-8 bg-base-300 rounded-lg overflow-hidden">
        <PrimaryProgressBar color={color} score={score} text={text} />
        {remainingToPass > 0 &&
            <SecondaryProgressBar startPosition={score} width={remainingToPass} text={secondaryText} />
        }
    </div>
)

// DefaultProgressBar is a function that returns a progress bar for a lab/assignment with no submissions
export const DefaultProgressBar = ({ scoreLimit, isGroupLab }: { scoreLimit: number, isGroupLab: boolean }) => {
    return (
        <div className="flex items-center gap-4 py-3 px-2 mb-2 rounded-lg">
            <div className="flex-1 min-w-0">
                <ProgressBarView score={0} remainingToPass={scoreLimit} text="0 %" secondaryText={`${scoreLimit} %`} />
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

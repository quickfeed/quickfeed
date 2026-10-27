import type { Submission } from "../../proto/qf/types_pb"
import { Submission_Status } from "../../proto/qf/types_pb"
import { getEffectiveStatus, isScoreHidden, isTestedAndReviewed, reviewVisible, scoreParts, testPart } from "../Helpers"
import { useAppState } from "../overmind"
import SubmissionTypeIcon from "./student/SubmissionTypeIcon"

type ProgressBarProps = {
    courseID: string,
    submission: Submission,
    showText?: boolean
    // split shows the parts of the score given by tests and by reviews
    split?: boolean
}

const ProgressBar = ({ courseID, submission, showText = true, split = false }: ProgressBarProps) => {
    const state = useAppState()

    const assignment = state.assignments[courseID]?.find(assignment => assignment.ID === submission.AssignmentID)

    const status = getEffectiveStatus(submission, state.self.ID)
    const scorelimit = assignment?.scoreLimit ?? 0
    // A hidden score shows the tests' part of it, followed by the review's part as pending.
    const hidden = assignment !== undefined && isScoreHidden(assignment, state.isTeacher, status)
    const score = hidden ? testPart(submission, assignment) : submission.score ?? 0
    let remainingToPass = status === Submission_Status.APPROVED ? 0 : Math.max(0, scorelimit - score)
    if (hidden) {
        remainingToPass = assignment.reviewWeight
    }
    const parts = split && assignment && isTestedAndReviewed(assignment) && reviewVisible(state.isTeacher, status)
        ? scoreParts(submission, assignment)
        : undefined

    let text = ""
    let secondaryText = ""
    let reviewText = ""
    if (showText) {
        if (parts) {
            text = `Tests ${parts.tests} %`
            reviewText = `Review ${parts.review} %`
        } else if (hidden) {
            text = `Tests ${score} %`
        } else {
            text = `${score} %`
        }
        secondaryText = hidden ? "Review pending" : `${remainingToPass} %`
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
            reviewPart={parts?.review}
            remainingToPass={remainingToPass}
            text={text}
            reviewText={reviewText}
            secondaryText={secondaryText}
        />
    )
}

export default ProgressBar

type ProgressBarViewProps = {
    score: number
    // reviewPart is the part of score given by reviews, shown as its own segment
    reviewPart?: number
    remainingToPass: number
    color?: string
    text?: string
    reviewText?: string
    secondaryText?: string
}

// ProgressBarView is the presentation-only bar. ProgressBar resolves the score,
// the status color and the pass threshold from application state and delegates
// here, so callers without that state, such as the About page previews, can
// render the same bar from plain values.
export const ProgressBarView = ({ score, reviewPart = 0, remainingToPass, color, text = "", reviewText = "", secondaryText = "" }: ProgressBarViewProps) => (
    <div className="relative w-full h-8 bg-base-300 rounded-lg overflow-hidden">
        <PrimaryProgressBar color={color} width={score - reviewPart} score={score} text={text} />
        {reviewPart > 0 &&
            <ReviewProgressBar color={color} startPosition={score - reviewPart} width={reviewPart} text={reviewText} />
        }
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


// PrimaryProgressBar is width wide and reports score, which is larger when a review segment follows it.
const PrimaryProgressBar = ({ color, width, score, text }: { color?: string, width: number, score: number, text: string }) => {
    return (
        <div
            className={`absolute top-0 left-0 h-full ${color || 'bg-primary'} flex items-center justify-center text-xs font-semibold text-primary-content transition-all duration-500`}
            role="progressbar"
            style={{ width: `${width}%` }}
            aria-valuenow={score}
            aria-valuemin={0}
            aria-valuemax={100}
        >
            {width > 10 && text}
        </div>
    )
}

const ReviewProgressBar = ({ color, startPosition, width, text }: { color?: string, startPosition: number, width: number, text: string }) => {
    return (
        <div
            className={`absolute top-0 h-full ${color || 'bg-primary'} opacity-70 border-l-2 border-base-100 flex items-center justify-center text-xs font-semibold text-primary-content transition-all duration-500`}
            style={{ left: `${startPosition}%`, width: `${width}%` }}
            title={text}
        >
            {width > 15 && text}
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

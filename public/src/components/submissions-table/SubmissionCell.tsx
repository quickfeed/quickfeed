import React, { memo, useState } from "react"
import { Enrollment, Group, Submission } from "../../../proto/qf/types_pb"
import { getSubmissionCellColor, Icon } from "../../Helpers"
import { useAppState } from "../../overmind"

/** Represents both possible submissions for a group lab assignment */
interface SubmissionPair {
    individual?: Submission
    group?: Submission
}

interface SubmissionCellProps {
    submissionPair: SubmissionPair
    owner: Enrollment | Group
    onSubmissionClick: (submission: Submission) => void
    review: boolean
}

const SubmissionCell = memo(({ submissionPair, owner, onSubmissionClick, review }: SubmissionCellProps) => {
    const state = useAppState()

    // Check if submissions actually exist (not just truthy, but valid submission objects)
    const hasIndividual = submissionPair.individual !== undefined && submissionPair.individual !== null
    const hasGroup = submissionPair.group !== undefined && submissionPair.group !== null
    const hasBothSubmissions = hasIndividual && hasGroup

    // DEBUG: Log what we're receiving
    console.log("SubmissionCell:", {
        hasIndividual,
        hasGroup,
        hasBothSubmissions,
        individualId: submissionPair.individual?.ID?.toString(),
        groupId: submissionPair.group?.ID?.toString(),
        pairKeys: Object.keys(submissionPair)
    })

    // Local state to toggle between individual and group submission when both exist
    // Only used when both submissions are available
    const [showGroup, setShowGroup] = useState(!state.individualSubmissionView)

    // Determine which submission to display and whether it's a group submission
    let submission: Submission
    let isGroupSubmission: boolean

    if (hasBothSubmissions) {
        // Both exist: use toggle state to determine which to show
        submission = showGroup ? submissionPair.group! : submissionPair.individual!
        isGroupSubmission = showGroup
    } else if (hasGroup) {
        // Only group exists
        submission = submissionPair.group!
        isGroupSubmission = true
    } else {
        // Only individual exists (or neither, but that's handled by MemberRow)
        submission = submissionPair.individual!
        isGroupSubmission = false
    }

    const isSelected = state.selectedSubmission?.ID === submission.ID
    const colorClass = getSubmissionCellColor(submission, owner)
    const selectedClass = isSelected ? "ring-2 ring-primary ring-inset" : ""

    const handleToggle = (e: React.MouseEvent) => {
        e.stopPropagation()
        setShowGroup(!showGroup)
    }

    if (review) {
        return (
            <ReviewCell
                submission={submission}
                onClick={() => onSubmissionClick(submission)}
                isSelected={isSelected}
                colorClass={colorClass}
            />
        )
    }

    // Determine what indicator to show:
    // - Both exist: clickable toggle button
    // - Only group exists: static group icon (for enrollments viewing group lab)
    // - Only individual exists: no icon needed (it's the default/expected case)
    const renderIndicator = () => {
        if (hasBothSubmissions) {
            return (
                <button
                    onClick={handleToggle}
                    className="btn btn-xs btn-ghost px-1 min-h-0 h-5"
                    title={isGroupSubmission ? "Showing group submission (click for individual)" : "Showing individual submission (click for group)"}
                >
                    <i className={isGroupSubmission ? "fa fa-users" : "fa fa-user"} />
                </button>
            )
        }
        if (hasGroup && !hasIndividual) {
            // Only show group icon when it's a group submission without an individual alternative
            return <i className="fa fa-users text-xs opacity-60" title="Group submission" />
        }
        return null
    }

    return (
        <td
            className={`cursor-pointer ${colorClass} ${selectedClass}`}
            onClick={() => onSubmissionClick(submission)}
        >
            <div className="flex items-center justify-center gap-1">
                <span>{submission.score}%</span>
                {renderIndicator()}
            </div>
        </td>
    )
})

SubmissionCell.displayName = "SubmissionCell"

interface ReviewCellProps {
    submission: Submission
    onClick: () => void
    isSelected: boolean
    colorClass: string
}

const ReviewCell = memo(({ submission, onClick, isSelected, colorClass }: ReviewCellProps) => {
    const state = useAppState()

    if (!state.isManuallyGraded(submission)) {
        return (
            <td className="text-base-content/50 text-center" title="Auto graded">
                <i className={Icon.DASH} />
            </td>
        )
    }

    const reviews = state.review.reviews.get(submission.ID) ?? []
    const numReviewers = state.assignments[state.activeCourse.toString()]?.find(
        a => a.ID === submission.AssignmentID
    )?.reviewers ?? 0

    // Check if the current user has any pending reviews for this submission
    const hasPendingReview = reviews.some(r => !r.ready && r.ReviewerID === state.self.ID)

    // Check if the average score meets the minimum for release
    const avgScore = reviews.length > 0
        ? reviews.reduce((acc, r) => acc + r.score, 0) / reviews.length
        : 0
    const willBeReleased = state.review.minimumScore > 0 && avgScore >= state.review.minimumScore

    const selectedClass = isSelected ? "ring-2 ring-primary ring-inset" : ""
    const pendingClass = hasPendingReview ? "shadow-[inset_0_0_0_3px_rgba(255,165,0,0.5)]" : ""
    const releaseClass = willBeReleased ? "opacity-75" : ""

    return (
        <td
            className={`cursor-pointer ${colorClass} ${selectedClass} ${pendingClass} ${releaseClass}`}
            onClick={onClick}
        >
            <div className="flex items-center justify-center gap-1">
                <i
                    className={submission.released ? "fa fa-unlock" : "fa fa-lock"}
                    title={submission.released ? "Released" : "Not released"}
                />
                <span>{reviews.length}/{numReviewers}</span>
            </div>
        </td>
    )
})

ReviewCell.displayName = "ReviewCell"

export default SubmissionCell

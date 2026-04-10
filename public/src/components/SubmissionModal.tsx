import React, { useEffect } from "react"
import { groupRepoLink, userRepoLink } from "../Helpers"
import { useAppState } from "../overmind"
import LabResult from "./LabResult"
import ReviewForm from "./manual-grading/ReviewForm"

interface SubmissionModalProps {
    /** Whether this results view is in review (grading) mode */
    review: boolean
    /** ID of the active course, used to look up group info */
    courseID: bigint
    /** Called when the user requests the modal to close (Esc, ✕ button, or backdrop) */
    onClose: () => void
}

/**
 * Overlay modal shown in wide-table mode when a submission is selected.
 * Reads owner/assignment info from global state and resolves the GitHub link
 * to the specific lab directory. Closes on Esc, the ✕ button, or backdrop click.
 */
const SubmissionModal = ({ review, courseID, onClose }: SubmissionModalProps) => {
    const state = useAppState()
    const assignment = state.selectedAssignment

    // Register Esc key handler for the lifetime of the modal
    useEffect(() => {
        const handleKeyDown = (e: KeyboardEvent) => {
            if (e.key === "Escape") onClose()
        }
        window.addEventListener("keydown", handleKeyDown)
        return () => window.removeEventListener("keydown", handleKeyDown)
    }, [onClose])

    // Resolve the display name and direct GitHub lab link for the submission owner.
    // The owner can be either an individual enrollment or a group.
    let ownerName = ""
    let labLink: string | undefined
    const course = state.courses.find(c => c.ID === state.activeCourse)
    if (state.submissionOwner.type === "ENROLLMENT" && state.selectedEnrollment?.user) {
        const user = state.selectedEnrollment.user
        ownerName = user.Name || "Unknown"
        const repoBase = userRepoLink(user, course)
        if (assignment) {
            labLink = `${repoBase}/tree/main/${assignment.name}`
        }
    } else if (state.submissionOwner.type === "GROUP") {
        const group = state.groups[courseID.toString()]?.find(g => g.ID === state.submissionOwner.id)
        if (group) {
            ownerName = group.name
            const repoBase = groupRepoLink(group, course)
            if (assignment) {
                labLink = `${repoBase}/tree/main/${assignment.name}`
            }
        }
    }

    // Choose the result panel to embed based on mode
    let resultPanel: React.ReactNode
    if (review) {
        resultPanel = <ReviewForm />
    } else {
        resultPanel = <LabResult />
    }

    return (
        <div className="modal modal-open" role="dialog" aria-modal="true">
            <div className="modal-box max-w-4xl w-full max-h-[90vh] overflow-y-auto">

                {/* Header: owner name, lab link, and close controls */}
                <div className="flex justify-between items-center mb-4">
                    <div>
                        {ownerName && <div className="font-semibold">{ownerName}</div>}
                        {labLink && assignment && (
                            <a href={labLink} target="_blank" rel="noopener noreferrer"
                                className="text-xs link link-hover text-primary flex items-center gap-1">
                                <i className="fa fa-external-link" />
                                {assignment.name}
                            </a>
                        )}
                    </div>
                    <div className="flex items-center gap-2">
                        <span className="text-xs text-base-content/50">
                            Press <kbd className="kbd kbd-sm">Esc</kbd> to close
                        </span>
                        <button className="btn btn-sm btn-circle btn-ghost" onClick={onClose}>✕</button>
                    </div>
                </div>

                {resultPanel}
            </div>

            {/* Backdrop click closes the modal */}
            <div className="modal-backdrop" onClick={onClose} />
        </div>
    )
}

export default SubmissionModal

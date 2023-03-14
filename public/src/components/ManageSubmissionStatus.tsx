import React, { useMemo } from "react"
import { Submission_Status } from "../../proto/qf/types_pb"
import { isManuallyGraded } from "../Helpers"
import { useActions, useAppState } from "../overmind"

const ManageSubmissionStatus = (): JSX.Element => {
    const actions = useActions()
    const state = useAppState()
    const assignment = state.selectedAssignment

    const [rebuilding, setRebuilding] = React.useState(false)
    const [updating, setUpdating] = React.useState<Submission_Status>(Submission_Status.NONE)

    const buttons: { text: string, status: Submission_Status, style: string, onClick?: () => void }[] = [
        { text: "Approve", status: Submission_Status.APPROVED, style: "success" },
        { text: "Revision", status: Submission_Status.REVISION, style: "warning" },
        { text: "Reject", status: Submission_Status.REJECTED, style: "danger" },
    ]


    const handleRebuild = async () => {
        if (rebuilding) { return } // Don't allow multiple rebuilds at once
        setRebuilding(true)
        await actions.rebuildSubmission()
        setRebuilding(false)
    }

    const handleSetStatus = async (status: Submission_Status) => {
        if (updating !== Submission_Status.NONE) { return } // Don't allow multiple updates at once
        setUpdating(status)
        await actions.updateSubmission({ owner: state.submissionOwner, submission: state.selectedSubmission, status })
        setUpdating(Submission_Status.NONE)
    }

    if (assignment && !isManuallyGraded(assignment)) {
        // Add rebuild button if the assignment is not manually graded
        buttons.push({ text: rebuilding ? "Rebuilding..." : "Rebuild", status: -1, style: rebuilding ? "secondary" : "primary", onClick: handleRebuild })
    }

    const StatusButtons = useMemo(() =>
        buttons.map((button, index) => {
            if (updating === button.status) {
                // Show spinner while submission is being updated
                return (
                    <button key={index} className={`col btn btn-secondary mr-2`}>
                        <span className="spinner-border spinner-border-sm" role="status" aria-hidden="true"></span>
                        <span className="sr-only">Loading...</span>
                    </button>
                )
            }
            const style = state.selectedSubmission?.status === button.status
                ? `col btn btn-${button.style} mr-2`         // Show solid button if the submission status matches the button status
                : `col btn btn-outline-${button.style} mr-2` // Show outlined button otherwise

            return (
                <button key={index} className={style} onClick={() => button.onClick ? button.onClick() : handleSetStatus(button.status)}>
                    {button.text}
                </button>
            )
        })
        , [buttons, rebuilding, state.selectedSubmission?.status, updating]
    )

    return (
        <div className="row m-auto">
            {StatusButtons}
        </div>
    )
}

export default ManageSubmissionStatus

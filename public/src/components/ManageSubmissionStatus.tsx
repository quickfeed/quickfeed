import React from "react"
import { Submission } from "../../proto/qf/types_pb"
import { isManuallyGraded } from "../Helpers"
import { useActions, useAppState } from "../overmind"

const ManageSubmissionStatus = (): JSX.Element => {
    const actions = useActions()
    const state = useAppState()
    const assignment = state.activeSubmissionLink?.assignment

    const [rebuilding, setRebuilding] = React.useState(false)

    const buttons: { text: string, status: Submission.Status, style: string, onClick?: () => void }[] = [
        { text: "Approve", status: Submission.Status.APPROVED, style: "primary" },
        { text: "Revision", status: Submission.Status.REVISION, style: "warning" },
        { text: "Reject", status: Submission.Status.REJECTED, style: "danger" },
    ]


    const handleRebuild = async () => {
        setRebuilding(true)
        await actions.rebuildSubmission()
        setRebuilding(false)
    }


    if (assignment && !isManuallyGraded(assignment)) {
        buttons.push({ text: rebuilding ? "Rebuilding..." : "Rebuild", status: Submission.Status.NONE, style: rebuilding ? "secondary" : "primary", onClick: handleRebuild})
    }

    const StatusButtons = buttons.map((button, index) => {
        const style = state.currentSubmission?.status === button.status ? `col btn btn-${button.style} mr-2` : `col btn btn-outline-${button.style} mr-2`
        // TODO: Perhaps refactor button into a separate general component to enable reuse
        return (
            <div key={index} className={style} onClick={() => button.onClick ? button.onClick() : actions.updateSubmission(button.status)}>
                {button.text}
            </div>
        )
    })
    return (
        <div className="row m-auto">
            {StatusButtons}
        </div>
    )
}

export default ManageSubmissionStatus

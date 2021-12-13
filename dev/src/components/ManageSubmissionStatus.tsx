import React from "react"
import { Submission } from "../../proto/ag/ag_pb"
import { useActions, useAppState } from "../overmind"

const ManageSubmissionStatus = (): JSX.Element => {

    const actions = useActions()
    const state = useAppState()

    // TODO: Add an "Are you sure you want to <Status> submission" prompt?
    const updateStatus = (status: Submission.Status) => {
        if (state.activeSubmission) {
            actions.updateSubmission(status)
        }
    }

    const buttons: {text: string, status: Submission.Status, style: string, onClick?: () => void}[] = [
        {text: "Approve", status: Submission.Status.APPROVED, style: "primary"}, 
        {text: "Revision", status: Submission.Status.REVISION, style: "warning"},
        {text: "Reject", status: Submission.Status.REJECTED, style: "danger"},
        {text: "Rebuild", status: Submission.Status.NONE, style: "primary", onClick: () => actions.rebuildSubmission()}
    ]

    const StautusButtons = buttons.map((button, index) => {
        const style = state.activeSubmission?.getStatus() === button.status ? `col btn btn-${button.style} mr-2` : `col btn btn-outline-${button.style} mr-2`
        // TODO: Perhaps refactor button into a separate general component to enable reuse
        return (
            <div key={index} className={style} onClick={() => button.onClick ? button.onClick() : updateStatus(button.status)}>
                {button.text}
            </div>
        )
    })
    return (
        <div className="container">
            <div className="row m-auto">
                {StautusButtons}
            </div>
        </div>
    )
}

export default ManageSubmissionStatus
import React from "react"
import { Submission } from "../../proto/ag/ag_pb"
import { getCourseID } from "../Helpers"
import { useActions, useAppState } from "../overmind"


const ManageSubmissionStatus = (): JSX.Element => {

    const actions = useActions()
    const state = useAppState()
    const courseID = getCourseID()

    // TODO: Add an "Are you sure you want to <Status> submission" prompt?
    const updateStatus = (status: Submission.Status) => {
        if (state.activeSubmission) {
            actions.updateSubmission({courseID: courseID, submission: state.activeSubmission, status: status})
        }
    }

    const buttons: {text: string, status: Submission.Status, style: string}[] = [
        {text: "Approved", status: Submission.Status.APPROVED, style: "col btn btn-primary mr-2"}, 
        {text: "Revision", status: Submission.Status.REVISION, style: "col btn btn-warning mr-2"},
        {text: "Rejected", status: Submission.Status.REJECTED, style: "col btn btn-danger mr-2"}
    ]

    const StautusButtons = buttons.map((button, index) => {
        const style = state.activeSubmission?.getStatus() === button.status ? button.style : "col btn border mr-2"
        return (
            <div key={index} className={style} onClick={() => updateStatus(button.status)}>
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
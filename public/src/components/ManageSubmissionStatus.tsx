import React from "react"
import { Submission } from "../../proto/ag/ag_pb"
import { getCourseID } from "../Helpers"
import { useActions, useAppState } from "../overmind"


const ManageSubmissionStatus = () => {

    const actions = useActions()
    const state = useAppState()
    const courseID = getCourseID()

    const updateStatus = (status: Submission.Status) => {
        if (state.activeSubmission) {
            actions.updateSubmission({courseID: courseID, submission: state.activeSubmission, status: status})
        }
    }
    return (
        <>
            <button className={"btn btn-primary"} onClick={() => updateStatus(Submission.Status.APPROVED)}>
                Approve
            </button>

            <button className={"btn btn-warning"} onClick={() => { updateStatus(Submission.Status.REVISION) } }>
                Revision
            </button>

            <button className={"btn btn-danger"} onClick={() => { updateStatus(Submission.Status.REJECTED) } }>
                Reject
            </button>

        </>
    )
}

export default ManageSubmissionStatus
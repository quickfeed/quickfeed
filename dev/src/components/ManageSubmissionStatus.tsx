import React from "react"
import { Submission } from "../../proto/ag/ag_pb"
import { getCourseID } from "../Helpers"
import { useActions, useAppState } from "../overmind"


const ManageSubmissionStatus = (): JSX.Element => {

    const actions = useActions()
    const state = useAppState()
    const courseID = getCourseID()

    const updateStatus = (status: Submission.Status) => {
        if (state.activeSubmission) {
            actions.updateSubmission({courseID: courseID, submission: state.activeSubmission, status: status})
        }
    }
    return (
        <div className="container">
            <div className="row m-auto">
            <div className="col btn btn-primary mr-2" onClick={() => updateStatus(Submission.Status.APPROVED)}>
                Approve
            </div>
            <div className="col btn btn-warning mr-2" onClick={() => { updateStatus(Submission.Status.REVISION) } }>
                Revision
            </div>
            <div className="col btn btn-danger mr-2" onClick={() => { updateStatus(Submission.Status.REJECTED) } }>
                Reject
            </div>
            </div>
        </div>
    )
}

export default ManageSubmissionStatus
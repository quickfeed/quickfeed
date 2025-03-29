import React from "react"
import { useAppState } from "../overmind"
import Lab from "./Lab"
import ManageSubmissionStatus from "./ManageSubmissionStatus"
import { CenteredMessage, KnownMessage } from "./CenteredMessage"

const LabResult = () => {
    const state = useAppState()
    if (!state.selectedSubmission) {
        return <CenteredMessage message={KnownMessage.NoSubmission} />
    }
    const assignment = state.selectedAssignment
    if (!assignment) {
        return <CenteredMessage message={KnownMessage.NoAssignment} />
    }
    return (
        <div className="lab-resize lab-sticky">
            <ManageSubmissionStatus courseID={assignment.CourseID.toString()} reviewers={assignment.reviewers} />
            <div className="reviewLabResult mt-2">
                <Lab />
            </div>
        </div>
    )
}

export default LabResult

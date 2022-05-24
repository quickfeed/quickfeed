import React from "react"
import { useAppState } from "../overmind"
import Lab from "./Lab"
import ManageSubmissionStatus from "./ManageSubmissionStatus"


const LabResult = () => {
    const state = useAppState()

    if (state.currentSubmission) {
        return (
            <div className="lab-resize">
                <ManageSubmissionStatus />
                <div className="reviewLabResult mt-2">
                    <Lab />
                </div>
            </div>
        )
    }
    return (
        null
    )
}

export default LabResult

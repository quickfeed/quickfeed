import React from "react"
import SubmissionLogs from "./SubmissionLogs"
import { Submission } from "../../../proto/qf/types_pb"


const SubmissionSearchResults = ({ submissions }: { submissions: Submission[] }) => {

    if (submissions.length === 0) {
        return <p>No results found</p>
    }

    return (
        <ul>
            {submissions.map(submission => {
                return <SubmissionLogs key={submission.ID.toString()} submission={submission} />
            })}
        </ul>
    )
}

export default SubmissionSearchResults

import React from "react"
import { Submission } from "../../../proto/qf/types_pb"


const SubmissionLogs = ({ submission }: { submission: Submission }) => {
    // show allows us to show the logs if the user clicks on the submission
    const [show, setShow] = React.useState<boolean>(false)

    
    if (!submission.BuildInfo) {
        return null
    }

    return (
        <div>
            <h4 onClick={() => setShow(!show)}>{submission.AssignmentID}</h4>
            <div className={`card bg-light ${show ? "" : "hide"}`}>
                <code className="card-body" style={{ color: "#c7254e" }}>
                    {submission.BuildInfo.BuildLog.split("\n").map((x: string, i: number) => <span key={i} >{x}<br /></span>)}
                </code>
            </div>
        </div>
    )
}

export default SubmissionLogs

import React from "react"
import { SubmissionLink } from "../../../proto/qf/types_pb"


const SubmissionLogs = ({ subLink }: { subLink: SubmissionLink }) => {
    // show allows us to show the logs if the user clicks on the submission
    const [show, setShow] = React.useState<boolean>(false)

    const assignment = subLink.assignment
    const submission = subLink.submission
    if (!assignment || !submission || !submission.BuildInfo) {
        return null
    }

    return (
        <div>
            <h4 onClick={() => setShow(!show)}>{assignment.name}</h4>
            <div className={`card bg-light ${show ? "" : "hide"}`}>
                <code className="card-body" style={{ color: "#c7254e" }}>
                    {submission.BuildInfo.BuildLog.split("\n").map((x: string, i: number) => <span key={i} >{x}<br /></span>)}
                </code>
            </div>
        </div>
    )
}

export default SubmissionLogs

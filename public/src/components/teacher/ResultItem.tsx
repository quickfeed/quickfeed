import React from "react"
import { Enrollment, SubmissionLink } from "../../../proto/ag/ag_pb"



const ResultItem = (data: {enrollment: Enrollment | undefined, submissionsLink: SubmissionLink[], query: string}) => {
    const submissions = data.submissionsLink.map(link => {
        if (link.hasSubmission() && link.hasAssignment()) {
            return (
                <td>
                    {link.getSubmission()?.getScore()}%
                </td>
            )
        }
    })
    let hidden = data.enrollment?.getUser()?.getName().toLowerCase().includes(data.query) || data.enrollment?.getGroup()?.getName().toLowerCase().includes(data.query)
    return (
        <tr hidden={!hidden}>
            <td>
                {data.enrollment?.getUser()?.getName()}
            </td>
            <td>
                {data.enrollment?.getGroup()?.getName()}
            </td>
            {submissions}
        </tr>
    )
}

export default ResultItem
import { json } from "overmind"
import React from "react"
import { Enrollment, Submission, SubmissionLink } from "../../../proto/ag/ag_pb"
import { useActions, useAppState } from "../../overmind"



const ResultItem = ({enrollment, submissionsLink}: {enrollment: Enrollment | undefined, submissionsLink: SubmissionLink[]}) => {
    const query = useAppState().query
    const actions = useActions()
    const submissions = submissionsLink.map(link => {
        const color = link.getSubmission()?.getStatus() === Submission.Status.APPROVED ? "#CCFFCC" : "#FFCCCC"
        if (link.hasSubmission() && link.hasAssignment()) {
            return (
                <td style={{"backgroundColor": color, "cursor": "pointer"}} onClick={() => actions.setActiveSubmission(json(link.getSubmission()))}>
                    {link.getSubmission()?.getScore()}%
                </td>
            )
        } else {
            return (
                <td onClick={() => actions.setActiveSubmission(undefined)}>
                    0%
                </td>
            )
        }
    })

    // TODO: Make a helper function that can resolve hidden.
    const hidden = enrollment?.getUser()?.getName().toLowerCase().includes(query) || enrollment?.getGroup()?.getName().toLowerCase().includes(query)
    return (
        <tr hidden={!hidden}>
            <td className="font-weight-bold">
                {enrollment?.getUser()?.getName()}
            </td>
            <td>
                {enrollment?.getGroup()?.getName()}
            </td>
            {submissions}
        </tr>
    )
}

export default ResultItem
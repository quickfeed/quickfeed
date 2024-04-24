import React from "react"
import { Score } from "../../../proto/kit/score/score_pb"


const SubmissionScore = ({ score }: { score: Score }) => {
    const className = score.Score === score.MaxScore ? "passed" : "failed"
    return (
        <tr>
            <td className={`${className} pl-4`}>
                {score.TestName}
            </td>
            <td>
                {score.Score}/{score.MaxScore}
            </td>
            <td>
                {score.Weight}
            </td>
        </tr>
    )
}

export default SubmissionScore

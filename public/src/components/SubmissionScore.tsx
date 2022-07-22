import React from "react"
import { Score } from "../../proto/kit/score/score_pb"


const SubmissionScore = ({ score }: { score: Score.AsObject }) => {
    const className = score.score === score.maxscore ? "passed" : "failed"
    return (
        <tr>
            <td className={className}>
                <div className="pl-1">
                    {score.testname}
                </div>
            </td>
            <td>
                {score.score}/{score.maxscore}
            </td>
            <td>
                {score.weight}
            </td>
        </tr>
    )
}

export default SubmissionScore

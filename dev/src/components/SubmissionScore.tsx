import React from "react"
import { Score } from "../../proto/kit/score/score_pb"


const SubmissionScore = ({ score }: { score: Score.AsObject }) => {
    const className = score.score === score.maxscore ? "passed" : "failed"
    return (
        <tr>
            <th className={className + " pl-4"}>
                {score.testname}
            </th>
            <th>
                {score.score}/{score.maxscore}
            </th>
            <th>
                {score.weight}
            </th>
        </tr>
    )
}

export default SubmissionScore

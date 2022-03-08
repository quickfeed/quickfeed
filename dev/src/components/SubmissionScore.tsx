import React from "react"
import { Score } from "../../proto/kit/score/score_pb"


const SubmissionScore = ({ score }: { score: Score }) => {
    const className = (score.getScore() === score.getMaxscore()) ? "passed" : "failed"
    return (
        <tr>
            <th className={className + " pl-4"}>
                {score.getTestname()}
            </th>
            <th>
                {score.getScore()}/{score.getMaxscore()}
            </th>
            <th>
                {score.getWeight()}
            </th>
        </tr>
    )
}

export default SubmissionScore

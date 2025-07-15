import React from "react"
import { Score } from "../../../proto/kit/score/score_pb"

const SubmissionScore = ({
    score,
    totalWeight,
}: {
    score: Score
    totalWeight: number
}) => {
    const className = score.Score === score.MaxScore ? "passed" : "failed"
    const percentage = (score.Score / score.MaxScore) * (score.Weight / totalWeight) * 100
    const maxPercentage = (score.MaxScore / score.MaxScore) * (score.Weight / totalWeight) * 100
    const cellColor = percentage === maxPercentage ? "text-success" : "text-danger"

    return (
        <tr>
            <td className={`${className} pl-4`}>{score.TestName}</td>
            <td className="fixed-width-score">
                {score.Score}/{score.MaxScore}
            </td>
            <td className="fixed-width-percent">
                <span className={cellColor}>
                    {percentage.toFixed(1)}%
                </span>
            </td>

            <td className="fixed-width-percent">
                <span
                    style={{ opacity: 0.5 }}
                    title={`Weight: ${score.Weight}`}
                    aria-label={`Max weighted percentage is ${maxPercentage.toFixed(1)} percent, weight ${score.Weight}`}
                >
                    {maxPercentage.toFixed(1)}%
                </span>
            </td>
        </tr>
    )
}

export default SubmissionScore

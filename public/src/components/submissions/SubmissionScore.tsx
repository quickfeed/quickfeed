import React from "react"
import { Score } from "../../../proto/kit/score/score_pb"

const SubmissionScore = ({
    score,
    totalWeight,
}: {
    score: Score
    totalWeight: number
}) => {
    const passed = score.Score === score.MaxScore
    const rowClass = passed ? "passed" : "failed"
    const percentage = (score.Score / score.MaxScore) * (score.Weight / totalWeight) * 100
    const maxPercentage = (score.MaxScore / score.MaxScore) * (score.Weight / totalWeight) * 100
    const cellColor = percentage === maxPercentage ? "text-success" : "text-error"

    return (
        <tr className={rowClass}>
            <td className="pl-3! w-full">{score.TestName}</td>
            <td className="whitespace-nowrap min-w-24 text-right">
                {score.Score}/{score.MaxScore}
            </td>
            <td className="whitespace-nowrap min-w-24 text-right">
                <span className={cellColor}>
                    {percentage.toFixed(1)}%
                </span>
            </td>

            <td className="whitespace-nowrap min-w-24 text-right">
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

import React from "react"
import { Score } from "../../proto/kit/score/score_pb"

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

    return (
        <tr>
            <td className={`${className} pl-4`}>{score.TestName}</td>
            <td className="text-right">
                {score.Score}/{score.MaxScore}
            </td>
            <td className="text-right">
                <span className={percentage === maxPercentage ? "text-success" : "text-danger"}>{percentage.toFixed(1)}%</span>
            </td>
            <td className="text-right">
                <span style={{opacity: 0.5}}  data-toggle="tooltip" title={"Weight: " + score.Weight}>{maxPercentage.toFixed(1)}%</span>
            </td>
        </tr>
    )
}

export default SubmissionScore

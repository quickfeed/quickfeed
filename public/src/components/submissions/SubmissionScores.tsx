import React, { useCallback } from 'react'
import { Submission } from "../../../proto/qf/types_pb"
import SubmissionScore from "./SubmissionScore"
import { ScoreSchema } from "../../../proto/kit/score/score_pb"
import { clone } from "@bufbuild/protobuf"

type ScoreSort = "name" | "score" | "weight" | "percentage"

const SubmissionScores = ({submission}: {submission: Submission}) => {
    const [sortKey, setSortKey] = React.useState<ScoreSort>("name")
    const [sortAscending, setSortAscending] = React.useState<boolean>(true)

    const sortScores = () => {
        const sortBy = sortAscending ? 1 : -1
        const scores = submission.Scores.map(score => clone(ScoreSchema, score))
        const totalWeight = scores.reduce((acc, score) => acc + score.Weight, 0)
        return scores.sort((a, b) => {
            switch (sortKey) {
                case "name":
                    return sortBy * (a.TestName.localeCompare(b.TestName))
                case "score":
                    return sortBy * (a.Score - b.Score)
                case "weight":
                    return sortBy * (a.Weight - b.Weight)
                case "percentage":
                    return sortBy * ((a.Score / a.MaxScore) * (a.Weight / totalWeight) - (b.Score / b.MaxScore) * (b.Weight / totalWeight))
                default:
                    return 0
            }
        })
    }

    const handleSort = useCallback((event: React.MouseEvent<HTMLTableCellElement>) => {
        const key = event.currentTarget.dataset.key as ScoreSort
        if (sortKey === key) {
            setSortAscending(!sortAscending)
        } else {
            setSortKey(key)
            setSortAscending(true)
        }
    }, [sortKey, sortAscending])

    const sortedScores = React.useMemo(sortScores, [submission, sortKey, sortAscending])
    const totalWeight = sortedScores.reduce((acc, score) => acc + score.Weight, 0)
    return (
        <table className="table table-curved table-striped table-hover">
            <thead className="thead-dark">
                <tr>
                    <th colSpan={1} className="col-md-8" data-key={"name"} role="button" onClick={handleSort}>Test Name</th>
                    <th colSpan={1} className="text-right col-md-auto" data-key={"score"} role="button" onClick={handleSort}>Score</th>
                    <th colSpan={1} className="text-right col-md-auto" data-key={"percentage"} role="button" onClick={handleSort}>%</th>
                    <th colSpan={1} className="text-right col-md-auto" data-key={"weight"} data-toggle="tooltip" title={"Maximum % contribution to total score"} role="button" onClick={handleSort}>Max</th>
                </tr>
            </thead>
            <tbody style={{"wordBreak": "break-word"}}>
                {sortedScores.map(score =>
                    <SubmissionScore key={score.ID.toString()} score={score} totalWeight={totalWeight} />
                )}
            </tbody>
            <tfoot>
                <tr>
                    <th colSpan={2}>Total Score</th>
                    <th className="text-right">{submission.score}%</th>
                    <th className="text-right">100%</th>
                </tr>
            </tfoot>
        </table>
    )
}

export default SubmissionScores

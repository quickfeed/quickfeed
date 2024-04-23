import React, { useCallback } from 'react'
import { Submission } from "../../../proto/qf/types_pb"
import SubmissionScore from "./SubmissionScore"

type ScoreSort = "name" | "score" | "weight"

const SubmissionScores = ({submission}: {submission: Submission}) => {
    const [sortKey, setSortKey] = React.useState<ScoreSort>("name")
    const [sortAscending, setSortAscending] = React.useState<boolean>(true)

    const sortScores = () => {
        const sortBy = sortAscending ? 1 : -1
        const scores = submission.clone().Scores
        return scores.sort((a, b) => {
            switch (sortKey) {
                case "name":
                    return sortBy * (a.TestName.localeCompare(b.TestName))
                case "score":
                    return sortBy * (a.Score - b.Score)
                case "weight":
                    return sortBy * (a.Weight - b.Weight)
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

    return (
        <table className="table table-curved table-striped">
                    <thead className="thead-dark">
                        <tr>
                            <th colSpan={1} data-key={"name"} role="button" onClick={handleSort}>Test Name</th>
                            <th colSpan={1} data-key={"score"} role="button" onClick={handleSort}>Score</th>
                            <th colSpan={1} data-key={"weight"} role="button" onClick={handleSort}>Weight</th>
                        </tr>
                    </thead>
                    <tbody style={{"wordBreak": "break-word"}}>
                        {sortedScores.map(score =>
                            <SubmissionScore key={score.ID.toString()} score={score} />
                        )}
                    </tbody>
                    <tfoot>
                        <tr>
                            <th>Total Score</th>
                            <th>{submission.score}%</th>
                            <th>100%</th>
                        </tr>
                    </tfoot>
        </table>
    )
}

export default SubmissionScores
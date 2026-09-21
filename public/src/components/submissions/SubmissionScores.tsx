import { clone } from "@bufbuild/protobuf"
import React, { useCallback } from 'react'
import { ScoreSchema } from "../../../proto/kit/score/score_pb"
import type { Submission } from "../../../proto/qf/types_pb"
import { testFailed } from "../../Helpers"
import ResultsSummary from "./ResultsSummary"
import SubmissionScore, { hasTestRun } from "./SubmissionScore"

type ScoreSort = "name" | "score" | "weight" | "percentage"

/** autoExpandLimit bounds how many failures open by themselves. A submission
 *  where everything failed would otherwise open as the wall of text that
 *  per-test output replaces. */
const autoExpandLimit = 3

/** testHash names the test a link points at, e.g. "#test=TestStack/Push", so
 *  that a teacher can send a student straight to one failure. */
const testHash = (hash: string): string =>
    hash.startsWith("#test=") ? decodeURIComponent(hash.slice("#test=".length)) : ""

const SubmissionScores = ({ submission }: { submission: Submission }) => {
    const [sortKey, setSortKey] = React.useState<ScoreSort>("name")
    const [sortAscending, setSortAscending] = React.useState<boolean>(true)
    const [filter, setFilter] = React.useState("")
    const [failuresOnly, setFailuresOnly] = React.useState(false)

    const [expanded, setExpanded] = React.useState<Set<string>>(() => {
        const linked = testHash(window.location.hash)
        if (linked && submission.Scores.some(s => s.TestName === linked)) {
            return new Set([linked])
        }
        // A student opens a submission to find what broke, so the failures are
        // already open; the tests that passed stay out of the way.
        return new Set(submission.Scores.filter(s => testFailed(s) && hasTestRun(s))
            .slice(0, autoExpandLimit)
            .map(s => s.TestName))
    })

    const handleSort = useCallback((event: React.MouseEvent<HTMLTableCellElement>) => {
        const key = event.currentTarget.dataset.key as ScoreSort
        if (sortKey === key) {
            setSortAscending(prev => !prev)
        } else {
            setSortKey(key)
            setSortAscending(true)
        }
    }, [sortKey])

    const toggle = useCallback((testName: string) => {
        setExpanded(prev => {
            const next = new Set(prev)
            if (next.delete(testName)) {
                if (testHash(window.location.hash) === testName) {
                    window.history.replaceState(null, "", window.location.pathname + window.location.search)
                }
            } else {
                next.add(testName)
                window.history.replaceState(null, "", `#test=${encodeURIComponent(testName)}`)
            }
            return next
        })
    }, [])

    const sortedScores = React.useMemo(() => {
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
    }, [submission, sortKey, sortAscending])

    const needle = filter.trim().toLowerCase()
    const shownScores = sortedScores.filter(score =>
        (!failuresOnly || testFailed(score)) &&
        (needle === "" || score.TestName.toLowerCase().includes(needle)))

    const totalWeight = sortedScores.reduce((acc, score) => acc + score.Weight, 0)
    const failed = sortedScores.filter(testFailed).length
    return (
        <div>
            <ResultsSummary
                failed={failed}
                total={sortedScores.length}
                filter={filter}
                failuresOnly={failuresOnly}
                onFilter={setFilter}
                onFailuresOnly={() => setFailuresOnly(prev => !prev)}
                onExpandAll={() => setExpanded(new Set(sortedScores.filter(hasTestRun).map(s => s.TestName)))}
                onCollapseAll={() => setExpanded(new Set())}
            />
            <table className="table table-zebra">
                <thead className="bg-base-300 text-base-content">
                    <tr className="text-lg">
                        <th colSpan={1} data-key="name" role="button" onClick={handleSort}>Test Name</th>
                        <th colSpan={1} className="fixed-width-percent text-right" data-key="score" role="button" onClick={handleSort}>Score</th>
                        <th colSpan={1} className="fixed-width-percent text-right" data-key="percentage" role="button" onClick={handleSort}>%</th>
                        <th colSpan={1} className="fixed-width-percent text-right" data-key="weight" data-toggle="tooltip" title="Maximum % contribution to total score" role="button" onClick={handleSort}>Max</th>
                    </tr>
                </thead>
                <tbody>
                    {shownScores.map(score =>
                        <SubmissionScore
                            key={score.ID.toString()}
                            score={score}
                            totalWeight={totalWeight}
                            expanded={expanded.has(score.TestName)}
                            onToggle={toggle}
                        />
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
        </div>
    )
}

export default SubmissionScores

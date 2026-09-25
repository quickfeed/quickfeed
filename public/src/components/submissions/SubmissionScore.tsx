import type { Score } from "../../../proto/kit/score/score_pb"
import { TestStatus } from "../../../proto/kit/score/score_pb"
import { testFailed, testStatusText } from "../../Helpers"
import TestOutputPanel from "./TestOutputPanel"

/** hasTestRun reports whether the run recorded anything about this test beyond
 *  its score. Submissions predating per-test attribution have nothing to open,
 *  and offering an empty panel would only invite a pointless click. */
export const hasTestRun = (score: Score): boolean =>
    score.TestDetails !== "" || score.TestOutput !== "" || score.Status !== TestStatus.NOT_RUN

const badgeClass = (score: Score): string => {
    switch (testStatusText(score)) {
        case "Passed":
            return "badge-success"
        case "Skipped":
            return "badge-outline"
        default:
            return "badge-error"
    }
}

const SubmissionScore = ({
    score,
    totalWeight,
    expanded,
    onToggle,
}: {
    score: Score
    totalWeight: number
    expanded: boolean
    onToggle: (testName: string) => void
}) => {
    const rowClass = testFailed(score) ? "failed" : "passed"
    const percentage = (score.Score / score.MaxScore) * (score.Weight / totalWeight) * 100
    const maxPercentage = (score.MaxScore / score.MaxScore) * (score.Weight / totalWeight) * 100
    const cellColor = percentage === maxPercentage ? "text-success" : "text-error"
    const panelID = `test-panel-${score.ID}`
    const openable = hasTestRun(score)

    const name = (
        <span className="flex items-center gap-2 text-left">
            {openable ? <i className={`fas fa-chevron-${expanded ? "down" : "right"} text-xs opacity-60`} /> : null}
            <span className={`badge badge-sm ${badgeClass(score)}`}>{testStatusText(score)}</span>
            <span>{score.TestName}</span>
            {score.Elapsed > 0 ? <span className="opacity-50 text-xs">{score.Elapsed.toFixed(2)}s</span> : null}
        </span>
    )

    return (
        <>
            <tr className={rowClass}>
                <td className="pl-3! w-full">
                    {openable ? (
                        <button
                            type="button"
                            className="btn btn-ghost btn-sm w-full justify-start px-1 font-normal"
                            aria-expanded={expanded}
                            aria-controls={panelID}
                            onClick={() => onToggle(score.TestName)}
                        >
                            {name}
                        </button>
                    ) : name}
                </td>
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
            {openable && expanded ? (
                <tr>
                    <td
                        colSpan={4}
                        className={`bg-base-200 p-4 border-l-4 ${testFailed(score) ? "border-error" : "border-base-300"}`}
                        id={panelID}
                        role="region"
                        aria-label={score.TestName}
                    >
                        <TestOutputPanel score={score} />
                    </td>
                </tr>
            ) : null}
        </>
    )
}

export default SubmissionScore

const plural = (n: number, noun: string) => `${n} ${noun}${n === 1 ? "" : "s"}`

/** summaryText says what the run found. A skipped test neither passed nor
 *  failed, so it is counted apart from both. */
const summaryText = (failed: number, skipped: number, total: number): string => {
    const skippedText = skipped > 0 ? `, ${skipped} skipped` : ""
    if (failed > 0) {
        return `${failed} of ${plural(total, "test")} failed${skippedText}`
    }
    if (skipped === total) {
        return total === 0 ? "No tests" : `All ${plural(total, "test")} skipped`
    }
    return skipped > 0
        ? `${total - skipped} of ${plural(total, "test")} passed${skippedText}`
        : `All ${plural(total, "test")} passed`
}

/** ResultsSummary says what the run found and offers the ways through it: only
 *  the failures, one named test, or everything at once.
 */
const ResultsSummary = ({
    failed,
    skipped,
    total,
    filter,
    failuresOnly,
    onFilter,
    onFailuresOnly,
    onExpandAll,
    onCollapseAll,
}: {
    failed: number
    skipped: number
    total: number
    filter: string
    failuresOnly: boolean
    onFilter: (value: string) => void
    onFailuresOnly: () => void
    onExpandAll: () => void
    onCollapseAll: () => void
}) => (
    <div className="flex flex-wrap items-center justify-between gap-2 px-3 py-2 bg-base-200 rounded-t-lg">
        <span className={`font-semibold ${failed > 0 ? "text-error" : skipped === total ? "" : "text-success"}`}>
            {summaryText(failed, skipped, total)}
        </span>
        <div className="flex flex-wrap items-center gap-2">
            <input
                type="search"
                className="input input-sm input-bordered w-40"
                placeholder="Filter tests"
                aria-label="Filter tests by name"
                value={filter}
                onChange={event => onFilter(event.target.value)}
            />
            <button
                type="button"
                className={`btn btn-sm ${failuresOnly ? "btn-error" : ""}`}
                aria-pressed={failuresOnly}
                disabled={failed === 0}
                onClick={onFailuresOnly}
            >
                Failures only
            </button>
            <button type="button" className="btn btn-sm" onClick={onExpandAll}>Expand all</button>
            <button type="button" className="btn btn-sm" onClick={onCollapseAll}>Collapse all</button>
        </div>
    </div>
)

export default ResultsSummary

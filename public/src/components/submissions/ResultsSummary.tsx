const plural = (n: number, noun: string) => `${n} ${noun}${n === 1 ? "" : "s"}`

/** ResultsSummary says what the run found and offers the ways through it: only
 *  the failures, one named test, or everything at once.
 */
const ResultsSummary = ({
    failed,
    total,
    filter,
    failuresOnly,
    onFilter,
    onFailuresOnly,
    onExpandAll,
    onCollapseAll,
}: {
    failed: number
    total: number
    filter: string
    failuresOnly: boolean
    onFilter: (value: string) => void
    onFailuresOnly: () => void
    onExpandAll: () => void
    onCollapseAll: () => void
}) => (
    <div className="flex flex-wrap items-center justify-between gap-2 px-3 py-2 bg-base-200 rounded-t-lg">
        <span className={`font-semibold ${failed > 0 ? "text-error" : "text-success"}`}>
            {failed > 0
                ? `${failed} of ${plural(total, "test")} failed`
                : `All ${plural(total, "test")} passed`}
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

import { timestampDate } from "@bufbuild/protobuf/wkt"
import { useMemo, useState, type ReactNode } from "react"
import type { CourseLogEntry } from "../../../proto/qf/requests_pb"
import { CourseLogEntry_Level } from "../../../proto/qf/requests_pb"
import { useCourseID } from "../../hooks/useCourseID"
import { useCourseLogs } from "../../hooks/useCourseLogs"
import { CenteredMessage } from "../CenteredMessage"
import LogOutput from "../LogOutput"
import Search from "../Search"

const LEVEL_NAMES: Record<CourseLogEntry_Level, string> = {
    [CourseLogEntry_Level.DEBUG]: "Debug",
    [CourseLogEntry_Level.INFO]: "Info",
    [CourseLogEntry_Level.WARN]: "Warn",
    [CourseLogEntry_Level.ERROR]: "Error",
}

const LEVEL_BADGE_COLOR: Record<CourseLogEntry_Level, string> = {
    [CourseLogEntry_Level.DEBUG]: "badge-ghost",
    [CourseLogEntry_Level.INFO]: "badge-info",
    [CourseLogEntry_Level.WARN]: "badge-warning",
    [CourseLogEntry_Level.ERROR]: "badge-error",
}

// The fixed columns shown for every entry, in display order; a column per
// distinct key found across the loaded entries' fields is appended after
// these. Order here also drives the order of the toggle menu.
const FIXED_COLUMNS = ["time", "level", "repository", "message", "source"] as const
type FixedColumn = typeof FIXED_COLUMNS[number]

const COLUMN_LABELS: Record<FixedColumn, string> = {
    time: "Time",
    level: "Level",
    repository: "Repository",
    message: "Message",
    source: "Source",
}

const pad = (n: number): string => n.toString().padStart(2, "0")

// toLocalDatetimeInput formats date for a <input type="datetime-local"> value, in the
// browser's local time zone; Date#toISOString is always UTC, so it cannot be reused here.
const toLocalDatetimeInput = (date: Date): string =>
    `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}T${pad(date.getHours())}:${pad(date.getMinutes())}`

// entryTime renders an entry's timestamp in a fixed 24-hour, year-month-day
// order (e.g. "2024-02-08 23:59:00"), rather than the browser locale's, which
// could show AM/PM and a locale-dependent date order such as mm/dd/yyyy.
const entryTime = (entry: CourseLogEntry): string => {
    if (!entry.time) {
        return ""
    }
    const d = timestampDate(entry.time)
    return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
}

// entryFields renders an entry's remaining structured attributes, sorted by
// key because a protobuf map has no order of its own and an entry's fields
// would otherwise move around between requests.
const entryFields = (entry: CourseLogEntry): string =>
    Object.entries(entry.fields)
        .sort(([a], [b]) => a.localeCompare(b))
        .map(([key, value]) => `${key}=${value}`)
        .join(" ")

// entryText renders one entry as plain text for the copy and download actions,
// and for the free-text filter; it lists every part regardless of which
// columns are currently hidden, so hiding a column never hides what it filters,
// copies, or downloads.
const entryText = (entry: CourseLogEntry): string => {
    const repository = entry.repository ? `[${entry.repository}]` : ""
    const parts = [entryTime(entry), LEVEL_NAMES[entry.level], repository, entry.message, entryFields(entry), entry.source]
    return parts.filter(Boolean).join(" ")
}

// renderCell renders one entry's value for a given column: the fixed columns
// read their own field, and anything else is looked up in the entry's fields
// map, where a value not present in this particular entry renders as nothing.
const renderCell = (entry: CourseLogEntry, column: string): ReactNode => {
    switch (column) {
        case "time":
            return entryTime(entry) || "N/A"
        case "level":
            return (
                <span className={`badge badge-xs ${LEVEL_BADGE_COLOR[entry.level]}`}>
                    {LEVEL_NAMES[entry.level]}
                </span>
            )
        case "repository":
            return entry.repository
        case "message":
            return (
                <span className="flex flex-wrap items-center gap-2">
                    <span>{entry.message}</span>
                    {entry.truncated && <span className="badge badge-xs badge-warning">truncated</span>}
                </span>
            )
        case "source":
            return entry.source
        default:
            return entry.fields[column] ?? ""
    }
}

/** ColumnsMenu is a checklist dropdown for showing or hiding individual columns
 *  of the log table, e.g. to deactivate a field like branch_ref that is not
 *  relevant right now. Hidden columns are keyed by name, so a column keeps
 *  its shown/hidden state across a Refresh even if it briefly disappears
 *  because no loaded entry currently carries it. */
const ColumnsMenu = ({ columns, hidden, onToggle }: { columns: string[]; hidden: Set<string>; onToggle: (column: string) => void }) => (
    <div className="dropdown dropdown-end">
        <div tabIndex={0} role="button" className="btn btn-sm">
            <i className="fas fa-table-columns" />
            Columns
        </div>
        <ul tabIndex={0} className="dropdown-content menu z-10 mt-2 w-56 max-h-80 overflow-y-auto rounded-box bg-base-100 p-2 shadow">
            {columns.map(column => {
                const label = COLUMN_LABELS[column as FixedColumn] ?? column
                return (
                    <li key={column}>
                        {/* The visible (and so implicit-label) text reads "Show <column>"
                            rather than the bare column name, so it never collides with an
                            identically-named filter above (e.g. the "Repository" select)
                            for anyone finding controls by label text, sighted or not. */}
                        <label className="flex items-center gap-2">
                            <input
                                type="checkbox"
                                className="checkbox checkbox-xs"
                                checked={!hidden.has(column)}
                                onChange={() => onToggle(column)}
                            />
                            <span>Show {label}</span>
                        </label>
                    </li>
                )
            })}
        </ul>
    </div>
)

/** CourseLogs is the teacher-only "Course Logs" page at /course/:id/logs.
 *  It queries GetCourseLog for the current course and lets a teacher narrow
 *  the result by interval, repository, and minimum level, then locally
 *  filter, copy, or download whatever was loaded. Filters other than the
 *  free-text one take effect only on Refresh. */
const CourseLogs = () => {
    const courseID = useCourseID()
    const [notice, setNotice] = useState<string | null>(null)
    const [draft, setDraft] = useState(() => ({
        courseID,
        from: toLocalDatetimeInput(new Date(Date.now() - 24 * 60 * 60 * 1000)),
        to: toLocalDatetimeInput(new Date()),
        toEdited: false,
        repository: "",
        level: CourseLogEntry_Level.DEBUG,
    }))
    if (draft.courseID !== courseID) {
        setDraft({ ...draft, courseID, repository: "", to: draft.toEdited ? draft.to : toLocalDatetimeInput(new Date()) })
        setNotice(null)
    }
    const { from, to, toEdited, repository, level } = draft
    const { result, loading, error, refresh } = useCourseLogs(courseID, { from, to: "", repository, level })
    const [search, setSearch] = useState("")
    const [hiddenColumns, setHiddenColumns] = useState<Set<string>>(new Set())

    const invalidInterval = Boolean(from) && new Date(from) > (toEdited && to ? new Date(to) : new Date())
    const handleRefresh = () => {
        if (invalidInterval) {
            return
        }
        setNotice(null)
        if (!toEdited) {
            setDraft({ ...draft, to: toLocalDatetimeInput(new Date()) })
        }
        refresh({ from, to: toEdited ? to : "", repository, level })
    }

    // The response lists every repository with an entry in the interval, whatever
    // the repository filter, but a selection whose repository fell silent must
    // stay in the list; the select would otherwise sit blank while still
    // filtering on it.
    const repositories = result?.repositories ?? []
    const repositoryOptions = repository && !repositories.includes(repository)
        ? [...repositories, repository].sort((a, b) => a.localeCompare(b))
        : repositories

    // Memoized because it feeds the fieldColumns memo below; without this, a
    // fresh array each render would defeat that memo and recompute the column
    // set on every keystroke in the free-text filter.
    const entries = useMemo(() => result?.entries ?? [], [result])
    const filtered = search
        ? entries.filter(entry => entryText(entry).toLowerCase().includes(search))
        : entries

    // The field columns are derived from every loaded entry, not just the
    // filtered ones, so typing into the free-text filter never makes a column
    // appear or disappear on its own.
    const fieldColumns = useMemo(
        () => Array.from(new Set(entries.flatMap(entry => Object.keys(entry.fields)))).sort((a, b) => a.localeCompare(b)),
        [entries]
    )
    const columns = useMemo(() => [...FIXED_COLUMNS, ...fieldColumns], [fieldColumns])
    const visibleColumns = columns.filter(column => !hiddenColumns.has(column))
    const toggleColumn = (column: string) => {
        setHiddenColumns(prev => {
            const next = new Set(prev)
            if (next.has(column)) {
                next.delete(column)
            } else {
                next.add(column)
            }
            return next
        })
    }

    const logText = () => filtered.map(entryText).join("\n")

    const handleCopy = async () => {
        try {
            // navigator.clipboard is undefined outside a secure context, and
            // writeText rejects when the browser denies clipboard access.
            await navigator.clipboard.writeText(logText())
            setNotice(null)
        } catch {
            setNotice("Could not copy the log; the browser denied access to the clipboard")
        }
    }

    const handleDownload = () => {
        const url = URL.createObjectURL(new Blob([logText()], { type: "text/plain" }))
        const link = document.createElement("a")
        link.href = url
        link.download = `course-${courseID}-log.txt`
        link.click()
        // Revoking the URL before the browser has read it cancels the download
        // the click just started, so leave that to the next tick.
        setTimeout(() => URL.revokeObjectURL(url), 0)
    }

    return (
        <div className="flex flex-col gap-4">
            <div className="card bg-base-200 shadow-sm">
                <div className="card-body gap-3">
                    <div className="grid grid-cols-1 md:grid-cols-4 gap-3">
                        <label className="form-control w-full">
                            <span className="label-text font-semibold">From</span>
                            <input
                                type="datetime-local"
                                // A locale like en-US would otherwise render this in
                                // 12-hour AM/PM time with a mm/dd/yyyy field order; sv-SE
                                // renders 24-hour time in yyyy-mm-dd order in Chromium-based
                                // browsers regardless of the browser's own locale. Firefox
                                // does not honor lang here and keeps its own OS-locale format.
                                lang="sv-SE"
                                className="input input-bordered w-full"
                                value={from}
                                onChange={e => setDraft({ ...draft, from: e.target.value })}
                            />
                        </label>
                        <label className="form-control w-full">
                            <span className="label-text font-semibold">To</span>
                            <input
                                type="datetime-local"
                                lang="sv-SE"
                                className="input input-bordered w-full"
                                value={to}
                                onChange={e => setDraft({ ...draft, toEdited: true, to: e.target.value })}
                            />
                        </label>
                        <label className="form-control w-full">
                            <span className="label-text font-semibold">Repository</span>
                            <select
                                className="select select-bordered w-full"
                                value={repository}
                                onChange={e => setDraft({ ...draft, repository: e.target.value })}
                            >
                                <option value="">All repositories</option>
                                {repositoryOptions.map(repo => <option key={repo} value={repo}>{repo}</option>)}
                            </select>
                        </label>
                        <label className="form-control w-full">
                            <span className="label-text font-semibold">Minimum level</span>
                            <select
                                className="select select-bordered w-full"
                                value={level}
                                onChange={e => setDraft({ ...draft, level: Number(e.target.value) })}
                            >
                                {Object.values(CourseLogEntry_Level).filter((v): v is CourseLogEntry_Level => typeof v === "number").map(value => (
                                    <option key={value} value={value}>{LEVEL_NAMES[value]}</option>
                                ))}
                            </select>
                        </label>
                    </div>
                    {invalidInterval && (
                        <div className="alert alert-error">
                            <span>From is after To; pick a From that precedes the end of the interval.</span>
                        </div>
                    )}
                    <div className="flex items-center gap-2">
                        <button type="button" className="btn btn-primary" onClick={handleRefresh} disabled={loading || invalidInterval}>
                            {loading ? "Refreshing…" : "Refresh"}
                        </button>
                        <Search placeholder="Filter loaded entries" setQuery={setSearch} className="flex-1" />
                    </div>
                </div>
            </div>

            {notice && <div className="alert alert-error"><span>{notice}</span></div>}
            {error && <CenteredMessage message={`Failed to load course logs: ${error}`} />}
            {!error && loading && <CenteredMessage message="Loading course logs…" />}
            {/* truncated reports that the server cut its own result at the limit,
                which the free-text filter neither causes nor undoes; the count is
                therefore the server's, and the warning stands even when the filter
                leaves nothing on screen. */}
            {!error && !loading && result?.truncated && (
                <div className="alert alert-warning">
                    <span>
                        Result limited to the newest {entries.length} entries.
                        Narrow the interval or the filters and click Refresh to see the rest.
                    </span>
                </div>
            )}
            {!error && !loading && result && filtered.length === 0 && (
                <CenteredMessage message="No log entries match the current filters" />
            )}
            {!error && !loading && result && filtered.length > 0 && (
                <LogOutput
                    title="Course Logs"
                    variant="table"
                    fill
                    controls={
                        <div className="flex items-center gap-2">
                            <ColumnsMenu columns={columns} hidden={hiddenColumns} onToggle={toggleColumn} />
                            <button type="button" className="btn btn-sm" onClick={() => void handleCopy()}>Copy</button>
                            <button type="button" className="btn btn-sm" onClick={handleDownload}>Download</button>
                        </div>
                    }
                >
                    <table className="table table-zebra table-xs">
                        <thead className="sticky top-0 z-10 bg-base-300">
                            <tr>
                                {visibleColumns.map(column => (
                                    <th key={column} className="whitespace-nowrap">
                                        {COLUMN_LABELS[column as FixedColumn] ?? column}
                                    </th>
                                ))}
                            </tr>
                        </thead>
                        <tbody>
                            {filtered.map((entry, idx) => (
                                // eslint-disable-next-line react/no-array-index-key
                                <tr key={idx}>
                                    {visibleColumns.map(column => (
                                        <td key={column} className="align-top whitespace-pre-wrap break-words">
                                            {renderCell(entry, column)}
                                        </td>
                                    ))}
                                </tr>
                            ))}
                        </tbody>
                    </table>
                </LogOutput>
            )}
        </div>
    )
}

export default CourseLogs

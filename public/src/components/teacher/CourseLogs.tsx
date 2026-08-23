import { timestampDate, timestampFromDate } from "@bufbuild/protobuf/wkt"
import { useEffect, useState } from "react"
import type { CourseLog, CourseLogEntry } from "../../../proto/qf/requests_pb"
import { CourseLogEntry_Level } from "../../../proto/qf/requests_pb"
import { useCourseID } from "../../hooks/useCourseID"
import { useGrpc } from "../../overmind"
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

// toLocalDatetimeInput formats date for a <input type="datetime-local"> value, in the
// browser's local time zone; Date#toISOString is always UTC, so it cannot be reused here.
const toLocalDatetimeInput = (date: Date): string => {
    const pad = (n: number) => n.toString().padStart(2, "0")
    return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}T${pad(date.getHours())}:${pad(date.getMinutes())}`
}

const entryTime = (entry: CourseLogEntry): string =>
    entry.time ? timestampDate(entry.time).toLocaleString() : ""

// entryFields renders an entry's remaining structured attributes, sorted by
// key because a protobuf map has no order of its own and an entry's fields
// would otherwise move around between requests.
const entryFields = (entry: CourseLogEntry): string =>
    Object.entries(entry.fields)
        .sort(([a], [b]) => a.localeCompare(b))
        .map(([key, value]) => `${key}=${value}`)
        .join(" ")

// entryText renders one entry as plain text for the copy and download actions;
// it must list the same parts, in the same order, as the rendered row, so that
// the screen, the clipboard, the downloaded file, and the free-text filter all
// agree on what an entry says.
const entryText = (entry: CourseLogEntry): string => {
    const repository = entry.repository ? `[${entry.repository}]` : ""
    const parts = [entryTime(entry), LEVEL_NAMES[entry.level], repository, entry.message, entryFields(entry), entry.source]
    return parts.filter(Boolean).join(" ")
}

/** CourseLogs is the teacher-only "Course Logs" page at /course/:id/logs.
 *  It queries GetCourseLog for the current course and lets a teacher narrow
 *  the result by interval, repository, and minimum level, then locally
 *  filter, copy, or download whatever was loaded. Filters other than the
 *  free-text one take effect only on Refresh. */
const CourseLogs = () => {
    const courseID = useCourseID()
    const { api } = useGrpc().global

    const [from, setFrom] = useState(() => toLocalDatetimeInput(new Date(Date.now() - 24 * 60 * 60 * 1000)))
    const [to, setTo] = useState(() => toLocalDatetimeInput(new Date()))
    const [toEdited, setToEdited] = useState(false)
    const [repository, setRepository] = useState("")
    const [level, setLevel] = useState(CourseLogEntry_Level.DEBUG)
    const [search, setSearch] = useState("")

    const [result, setResult] = useState<CourseLog | null>(null)
    const [loading, setLoading] = useState(false)
    const [error, setError] = useState<string | null>(null)
    const [notice, setNotice] = useState<string | null>(null)

    const fetchLogs = async () => {
        setLoading(true)
        setError(null)
        setNotice(null)
        // A To the teacher has not set means "up to now", so leave it out and let
        // the server bound the interval by its own clock; a page left open would
        // otherwise keep querying the interval that ended when it mounted, and
        // Refresh could never show anything logged since. The field is advanced
        // to match, since it stays pre-filled on screen.
        if (!toEdited) {
            setTo(toLocalDatetimeInput(new Date()))
        }
        const response = await api.client.getCourseLog({
            courseID,
            from: from ? timestampFromDate(new Date(from)) : undefined,
            to: toEdited && to ? timestampFromDate(new Date(to)) : undefined,
            repository,
            level,
        })
        setLoading(false)
        if (response.error) {
            setError(response.error.message)
            return
        }
        setResult(response.message)
    }

    useEffect(() => {
        void fetchLogs()
        // Fetch on mount and whenever the route names another course, since the
        // router keeps this page mounted across that change. The filters are
        // deliberately left out, so that they apply only on Refresh.
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [courseID])

    const entries = result?.entries ?? []
    const filtered = search
        ? entries.filter(entry => entryText(entry).toLowerCase().includes(search))
        : entries

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
                                className="input input-bordered w-full"
                                value={from}
                                onChange={e => setFrom(e.target.value)}
                            />
                        </label>
                        <label className="form-control w-full">
                            <span className="label-text font-semibold">To</span>
                            <input
                                type="datetime-local"
                                className="input input-bordered w-full"
                                value={to}
                                onChange={e => { setToEdited(true); setTo(e.target.value) }}
                            />
                        </label>
                        <label className="form-control w-full">
                            <span className="label-text font-semibold">Repository</span>
                            <select
                                className="select select-bordered w-full"
                                value={repository}
                                onChange={e => setRepository(e.target.value)}
                            >
                                <option value="">All repositories</option>
                                {(result?.repositories ?? []).map(repo => <option key={repo} value={repo}>{repo}</option>)}
                            </select>
                        </label>
                        <label className="form-control w-full">
                            <span className="label-text font-semibold">Minimum level</span>
                            <select
                                className="select select-bordered w-full"
                                value={level}
                                onChange={e => setLevel(Number(e.target.value))}
                            >
                                {Object.values(CourseLogEntry_Level).filter((v): v is CourseLogEntry_Level => typeof v === "number").map(value => (
                                    <option key={value} value={value}>{LEVEL_NAMES[value]}</option>
                                ))}
                            </select>
                        </label>
                    </div>
                    <div className="flex items-center gap-2">
                        <button type="button" className="btn btn-primary" onClick={() => void fetchLogs()} disabled={loading}>
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
                    codeClassName="text-base-content"
                    controls={
                        <div className="flex items-center gap-2">
                            <button type="button" className="btn btn-sm" onClick={() => void handleCopy()}>Copy</button>
                            <button type="button" className="btn btn-sm" onClick={handleDownload}>Download</button>
                        </div>
                    }
                >
                    {filtered.map((entry, idx) => {
                        const fields = entryFields(entry)
                        return (
                            // A row is a <span> because it renders inside the <code> element
                            // of LogOutput, which only admits phrasing content.
                            // eslint-disable-next-line react/no-array-index-key
                            <span key={idx} className="flex flex-wrap items-baseline gap-2 py-0.5">
                                <span className="text-base-content/60 shrink-0">{entryTime(entry) || "N/A"}</span>
                                <span className={`badge badge-xs ${LEVEL_BADGE_COLOR[entry.level]} shrink-0`}>
                                    {LEVEL_NAMES[entry.level]}
                                </span>
                                {entry.repository && (
                                    <span className="text-base-content/60 shrink-0">[{entry.repository}]</span>
                                )}
                                <span className="break-words">{entry.message}</span>
                                {fields && (
                                    <span className="text-base-content/70 break-words whitespace-pre-wrap">{fields}</span>
                                )}
                                {entry.source && <span className="text-base-content/40 shrink-0">{entry.source}</span>}
                                {entry.truncated && <span className="badge badge-xs badge-warning shrink-0">truncated</span>}
                            </span>
                        )
                    })}
                </LogOutput>
            )}
        </div>
    )
}

export default CourseLogs

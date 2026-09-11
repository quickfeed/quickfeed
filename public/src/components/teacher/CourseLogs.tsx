import { useState } from "react"
import type { CourseLogEntry } from "../../../proto/qf/requests_pb"
import { CourseLogEntry_Level } from "../../../proto/qf/requests_pb"
import { useCourseID } from "../../hooks/useCourseID"
import { useCourseLogs } from "../../hooks/useCourseLogs"
import { CenteredMessage } from "../CenteredMessage"
import CourseLogTable from "./CourseLogTable"
import { entryTime, LEVEL_NAMES, toLocalDatetimeInput } from "./courseLogFormatting"
import Search from "../Search"

const EMPTY_ENTRIES: CourseLogEntry[] = []

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

    const entries = result?.entries ?? EMPTY_ENTRIES
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
            <CourseLogTable
                entries={entries}
                rows={filtered}
                controls={
                    <>
                        <button type="button" className="btn btn-sm" onClick={() => void handleCopy()}>Copy</button>
                        <button type="button" className="btn btn-sm" onClick={handleDownload}>Download</button>
                    </>
                }
            />
        </div>
    )
}

export default CourseLogs

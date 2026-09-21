import { ConnectError } from "@connectrpc/connect"
import { useState } from "react"
import { CourseLogEntry_Level } from "../../../proto/qf/requests_pb"
import { ConnStatus } from "../../Helpers"
import { useCourseID } from "../../hooks/useCourseID"
import { useCourseLogStream } from "../../hooks/useCourseLogStream"
import { CenteredMessage } from "../CenteredMessage"
import Search from "../Search"
import CourseLogTable from "./CourseLogTable"
import { entryText, LEVEL_NAMES, logText, MAX_WINDOW, parseWindow } from "./courseLogFormatting"

const DEFAULT_WINDOW = "2h"

const STATUS_BADGE: Record<ConnStatus, { color: string, text: string }> = {
    [ConnStatus.CONNECTED]: { color: "badge-success", text: "Live" },
    [ConnStatus.RECONNECTING]: { color: "badge-warning", text: "Reconnecting…" },
    [ConnStatus.DISCONNECTED]: { color: "badge-ghost", text: "Not connected" },
}

/** CourseLogs is the teacher-only "Course Logs" page at /course/:id/logs.
 *  It tails the current course's log: the view is seeded with the entries from
 *  the chosen window and then grows as the server logs more, so a teacher can
 *  push and watch what QuickFeed makes of it. The window, repository, and
 *  level filters reopen the stream on Refresh; the free-text box, the order
 *  toggle, Copy, and Download act on what is already loaded. */
const CourseLogs = () => {
    const courseID = useCourseID()
    const [notice, setNotice] = useState<string | null>(null)
    const [draft, setDraft] = useState(() => ({
        courseID,
        window: DEFAULT_WINDOW,
        repository: "",
        level: CourseLogEntry_Level.DEBUG,
    }))
    if (draft.courseID !== courseID) {
        setDraft({ ...draft, courseID, repository: "" })
        setNotice(null)
    }
    const { window: windowText, repository, level } = draft
    const parsedWindow = parseWindow(windowText)
    const { result, loading, error, status, refresh, loadOlder } = useCourseLogStream(courseID, {
        window: parseWindow(DEFAULT_WINDOW) ?? MAX_WINDOW,
        repository: "",
        level: CourseLogEntry_Level.DEBUG,
    })
    const [search, setSearch] = useState("")

    const handleRefresh = () => {
        if (parsedWindow === null) {
            return
        }
        setNotice(null)
        refresh({ window: parsedWindow, repository, level })
    }

    // The response lists every repository with an entry in the window, whatever
    // the repository filter, but a selection whose repository fell silent must
    // stay in the list; the select would otherwise sit blank while still
    // filtering on it.
    const repositories = result?.repositories ?? []
    const repositoryOptions = repository && !repositories.includes(repository)
        ? [...repositories, repository].sort((a, b) => a.localeCompare(b))
        : repositories

    const entries = result?.entries ?? []
    const filtered = search
        ? entries.filter(entry => entryText(entry).toLowerCase().includes(search))
        : entries

    const handleCopy = async () => {
        try {
            // navigator.clipboard is undefined outside a secure context, and
            // writeText rejects when the browser denies clipboard access.
            await navigator.clipboard.writeText(logText(filtered))
            setNotice(null)
        } catch {
            setNotice("Could not copy the log; the browser denied access to the clipboard")
        }
    }

    // A failed Load older leaves the live view alone: the stream is still
    // running and the loaded entries still stand, so it reports as a notice
    // rather than replacing the page with an error.
    const handleLoadOlder = async () => {
        try {
            await loadOlder()
            setNotice(null)
        } catch (err) {
            setNotice(`Could not load older entries: ${ConnectError.from(err).message}`)
        }
    }

    const handleDownload = () => {
        const url = URL.createObjectURL(new Blob([logText(filtered)], { type: "text/plain" }))
        const link = document.createElement("a")
        link.href = url
        link.download = `course-${courseID}-log.txt`
        link.click()
        // Revoking the URL before the browser has read it cancels the download
        // the click just started, so leave that to the next tick.
        setTimeout(() => URL.revokeObjectURL(url), 0)
    }

    const badge = STATUS_BADGE[status]
    return (
        <div className="flex flex-col gap-4 h-[calc(100dvh_-_var(--navbar-height)_-_3rem)]">
            <div className="card bg-base-200 shadow-sm shrink-0">
                <div className="card-body gap-3">
                    <div className="grid grid-cols-1 md:grid-cols-3 gap-3">
                        <label className="form-control w-full">
                            <span className="label-text font-semibold">Show the last</span>
                            <input
                                type="text"
                                aria-label="Show the last"
                                placeholder="2h"
                                className="input input-bordered w-full"
                                value={windowText}
                                onChange={e => setDraft({ ...draft, window: e.target.value })}
                                onKeyUp={e => e.key === "Enter" && handleRefresh()}
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
                    {parsedWindow === null && (
                        <div className="alert alert-error shrink-0">
                            <span>Write how far back to look, such as 15 min, 4 h, or 3 days.</span>
                        </div>
                    )}
                    <div className="flex items-center gap-2">
                        <button type="button" className="btn btn-primary" onClick={handleRefresh} disabled={parsedWindow === null}>
                            Refresh
                        </button>
                        <span className={`badge ${badge.color}`} title="Log entries appear here as the server records them">
                            {badge.text}
                        </span>
                        <Search placeholder="Filter loaded entries" setQuery={setSearch} className="flex-1" />
                    </div>
                </div>
            </div>

            {notice && <div className="alert alert-error shrink-0"><span>{notice}</span></div>}
            {error && <CenteredMessage message={`Failed to load course logs: ${error}`} />}
            {!error && loading && <CenteredMessage message="Loading course logs…" />}
            {/* truncated reports that the server cut its own result at the limit,
                which the free-text filter neither causes nor undoes; the count is
                therefore the server's, and the warning stands even when the filter
                leaves nothing on screen. */}
            {!error && !loading && result?.truncated && (
                <div className="alert alert-warning shrink-0">
                    <span>
                        The window held more than the newest {entries.length} entries.
                        Narrow the window or the filters and click Refresh to see the rest.
                    </span>
                </div>
            )}
            {!error && !loading && result && filtered.length === 0 && (
                <CenteredMessage message="No log entries match the current filters" />
            )}
            <CourseLogTable
                entries={entries}
                rows={filtered}
                onLoadOlder={() => void handleLoadOlder()}
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

import { ConnectError } from "@connectrpc/connect"
import { useState, type KeyboardEvent } from "react"
import type { CourseLogEntry } from "../../../proto/qf/requests_pb"
import { CourseLogEntry_Level } from "../../../proto/qf/requests_pb"
import { ConnStatus } from "../../Helpers"
import { useCourseID } from "../../hooks/useCourseID"
import type { CourseLogRange } from "../../hooks/useCourseLogStream"
import { useCourseLogStream } from "../../hooks/useCourseLogStream"
import { CenteredMessage } from "../CenteredMessage"
import Search from "../Search"
import CourseLogTable from "./CourseLogTable"
import { DAY, entryText, HOUR, LEVEL_NAMES, logText, MINUTE, RETENTION, toLocalDatetimeInput } from "./courseLogFormatting"

const EMPTY_ENTRIES: CourseLogEntry[] = []

const PRESETS = [
    { label: "15m", duration: 15 * MINUTE },
    { label: "1h", duration: HOUR },
    { label: "4h", duration: 4 * HOUR },
    { label: "24h", duration: 24 * HOUR },
    { label: "3d", duration: 3 * DAY },
]

const DEFAULT_PRESET = HOUR

interface Badge {
    color: string
    text: string
    title?: string
}

const BADGES: Record<"live" | "reconnecting" | "notFollowing" | "notConnected", Badge> = {
    live: { color: "badge-success", text: "Live", title: "New entries appear as they are logged" },
    reconnecting: { color: "badge-warning", text: "Reconnecting…" },
    notFollowing: { color: "badge-ghost", text: "Not following", title: "New entries are not shown; clear To or click Load newer" },
    notConnected: { color: "badge-ghost", text: "Not connected" },
}

interface Draft {
    courseID: bigint
    /** The chosen preset's duration, or null once a date has been picked. */
    preset: number | null
    from: string
    to: string
    repository: string
    level: CourseLogEntry_Level
}

/** rangeProblem says why draft's dates cannot be asked for, or null if they can. */
const rangeProblem = (draft: Draft): string | null => {
    if (draft.preset !== null) {
        return null
    }
    if (!draft.from) {
        return "Pick when the range starts, or choose a preset."
    }
    if (draft.to && new Date(draft.from) > new Date(draft.to)) {
        return "From is after To; pick a From that precedes the end of the range."
    }
    return null
}

const rangeOf = (draft: Draft): CourseLogRange => draft.preset !== null
    ? { kind: "preset", duration: draft.preset }
    : { kind: "dates", from: new Date(draft.from), to: draft.to ? new Date(draft.to) : null }

/** CourseLogs is the teacher-only "Course Logs" page at /course/:id/logs.
 *  It streams the current course's log and lets a teacher narrow it by range,
 *  repository, and minimum level, then locally filter, copy, or download
 *  whatever was loaded. A preset applies at once; picked dates and the other
 *  filters take effect on Enter or Refresh. */
const CourseLogs = () => {
    const courseID = useCourseID()
    const [notice, setNotice] = useState<string | null>(null)
    const [draft, setDraft] = useState<Draft>(() => ({
        courseID,
        preset: DEFAULT_PRESET,
        from: toLocalDatetimeInput(new Date(Date.now() - DEFAULT_PRESET)),
        to: "",
        repository: "",
        level: CourseLogEntry_Level.DEBUG,
    }))
    if (draft.courseID !== courseID) {
        setDraft({ ...draft, courseID, repository: "" })
        setNotice(null)
    }
    const { from, to, repository, level } = draft
    const problem = rangeProblem(draft)
    const { result, loading, error, status, refresh, loadOlder, loadNewer } = useCourseLogStream(courseID, {
        range: { kind: "preset", duration: DEFAULT_PRESET },
        repository: "",
        level: CourseLogEntry_Level.DEBUG,
    })
    const [search, setSearch] = useState("")

    const apply = (next: Draft) => {
        if (rangeProblem(next) !== null) {
            return
        }
        setNotice(null)
        refresh({ range: rangeOf(next), repository: next.repository, level: next.level })
    }

    // A preset is measured from the moment it is applied, so From shows where
    // the range now begins.
    const applyPreset = (duration: number) => {
        const next = { ...draft, preset: duration, from: toLocalDatetimeInput(new Date(Date.now() - duration)), to: "" }
        setDraft(next)
        apply(next)
    }

    const handleRefresh = () => {
        if (draft.preset !== null) {
            applyPreset(draft.preset)
            return
        }
        apply(draft)
    }

    // The response lists every repository with an entry in the range, whatever
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

    // A failed Load older leaves the live view running, so report it as a notice.
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

    const badge = status === ConnStatus.CONNECTED ? BADGES.live
        : status === ConnStatus.RECONNECTING ? BADGES.reconnecting
            : loading || error ? BADGES.notConnected : BADGES.notFollowing
    const now = new Date()
    const pickerBounds = { min: toLocalDatetimeInput(new Date(now.getTime() - RETENTION)), max: toLocalDatetimeInput(now) }
    const onEnter = (e: KeyboardEvent) => e.key === "Enter" && handleRefresh()
    return (
        <div className="flex flex-col gap-4 h-[calc(100dvh_-_var(--navbar-height)_-_3rem)]">
            <div className="card bg-base-200 shadow-sm shrink-0">
                <div className="card-body gap-3">
                    <div className="grid grid-cols-1 md:grid-cols-4 gap-3">
                        <label className="form-control w-full">
                            <span className="label-text font-semibold">From</span>
                            <input
                                type="datetime-local"
                                // The native picker follows a locale, and en-US would render
                                // it in 12-hour AM/PM time with a mm/dd/yyyy field order.
                                // nb-NO gives 24-hour time and day-month-year in Chromium-based
                                // browsers whatever the browser's own locale; the rendered log
                                // timestamps are ours to format, and stay yyyy-mm-dd. Firefox
                                // does not honor lang here and keeps its OS-locale format.
                                lang="nb-NO"
                                className="input input-bordered w-full"
                                {...pickerBounds}
                                value={from}
                                onChange={e => setDraft({ ...draft, preset: null, from: e.target.value })}
                                onKeyUp={onEnter}
                            />
                        </label>
                        <label className="form-control w-full">
                            <span className="label-text font-semibold">To</span>
                            <input
                                type="datetime-local"
                                title="Leave blank to follow new entries"
                                lang="nb-NO"
                                className="input input-bordered w-full"
                                {...pickerBounds}
                                value={to}
                                onChange={e => setDraft({ ...draft, preset: null, to: e.target.value })}
                                onKeyUp={onEnter}
                            />
                        </label>
                        <label className="form-control w-full">
                            <span className="label-text font-semibold">Repository</span>
                            <select
                                className="select select-bordered w-full"
                                value={repository}
                                onChange={e => setDraft({ ...draft, repository: e.target.value })}
                                onKeyUp={onEnter}
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
                                onKeyUp={onEnter}
                            >
                                {Object.values(CourseLogEntry_Level).filter((v): v is CourseLogEntry_Level => typeof v === "number").map(value => (
                                    <option key={value} value={value}>{LEVEL_NAMES[value]}</option>
                                ))}
                            </select>
                        </label>
                    </div>
                    {problem && (
                        <div className="alert alert-error shrink-0">
                            <span>{problem}</span>
                        </div>
                    )}
                    <div className="flex flex-wrap items-center gap-2">
                        <button type="button" className="btn btn-primary" onClick={handleRefresh} disabled={problem !== null}>
                            Refresh
                        </button>
                        <div role="group" aria-label="Show the last" className="join">
                            {PRESETS.map(p => (
                                <button
                                    key={p.label}
                                    type="button"
                                    className={`btn join-item ${draft.preset === p.duration ? "btn-primary btn-soft" : ""}`}
                                    aria-pressed={draft.preset === p.duration}
                                    onClick={() => applyPreset(p.duration)}
                                >
                                    {p.label}
                                </button>
                            ))}
                        </div>
                        <span className={`badge ${badge.color}`} title={badge.title}>
                            {badge.text}
                        </span>
                        <Search placeholder="Filter loaded entries" setQuery={setSearch} className="flex-1 min-w-48" />
                    </div>
                </div>
            </div>

            {notice && <div className="alert alert-error shrink-0"><span>{notice}</span></div>}
            {error && <CenteredMessage message={`Failed to load course logs: ${error}`} />}
            {!error && loading && <CenteredMessage message="Loading course logs…" />}
            {!error && !loading && result && filtered.length === 0 && (
                <CenteredMessage message="No log entries match the current filters" />
            )}
            <CourseLogTable
                entries={entries}
                rows={filtered}
                onLoadOlder={result?.moreOlder ? () => void handleLoadOlder() : undefined}
                onLoadNewer={result?.moreNewer ? loadNewer : undefined}
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

import { useLayoutEffect, useMemo, useRef, useState, type ReactNode } from "react"
import type { CourseLogEntry } from "../../../proto/qf/requests_pb"
import { CourseLogEntry_Level } from "../../../proto/qf/requests_pb"
import LogCard from "../LogCard"
import LogViewerModal from "../LogViewerModal"
import { entryTime, LEVEL_NAMES } from "./courseLogFormatting"

interface Column {
    id: string
    label: string
    /** The cell's value as plain text, for a column holding free text. */
    text?: (entry: CourseLogEntry) => string
    /** Renders the cell. A column with `text` is handed its value ready to
     *  show: the whole of it, or a Show button in its place when it is long.
     *  A column without one is free to render the entry as it likes. */
    render: (entry: CourseLogEntry, value: ReactNode) => ReactNode
}

const LEVEL_BADGE_COLOR: Record<CourseLogEntry_Level, string> = {
    [CourseLogEntry_Level.DEBUG]: "badge-ghost",
    [CourseLogEntry_Level.INFO]: "badge-info",
    [CourseLogEntry_Level.WARN]: "badge-warning",
    [CourseLogEntry_Level.ERROR]: "badge-error",
}

const columns: Column[] = [
    { id: "time", label: "Time", render: entry => entryTime(entry) || "N/A" },
    {
        id: "level", label: "Level", render: entry => (
            <span className={`badge badge-xs ${LEVEL_BADGE_COLOR[entry.level]}`}>
                {LEVEL_NAMES[entry.level]}
            </span>
        ),
    },
    { id: "repository", label: "Repository", text: entry => entry.repository, render: (_, value) => value },
    { id: "repositoryType", label: "Repository type", text: entry => entry.repositoryType, render: (_, value) => value },
    {
        // The store cuts a record's message at 64 KiB, so a message is as
        // free-form as any field and collapses on the same rule.
        id: "message", label: "Message", text: entry => entry.message,
        render: (entry, value) => (
            <span className="flex flex-wrap items-center gap-2">
                <span className="min-w-0">{value}</span>
                {entry.truncated && <span className="badge badge-xs badge-warning">truncated</span>}
            </span>
        ),
    },
    { id: "source", label: "Source", text: entry => entry.source, render: (_, value) => value },
]

// COLLAPSE_AT is the width, in characters, beyond which a value is replaced by
// a Show button. A test run's output is thousands of characters over several
// lines, and left whole it stretches its column until the rest of the table is
// unreadable. Even a first line of it would widen the column for little gain.
const COLLAPSE_AT = 60

const isLong = (value: string): boolean => value.includes("\n") || value.length > COLLAPSE_AT

// FOLLOW_SLACK is how close, in pixels, to the newest end the view must be
// for it to keep following new entries.
const FOLLOW_SLACK = 32

interface CourseLogTableProps {
    entries: CourseLogEntry[]
    rows: CourseLogEntry[]
    controls?: ReactNode
    /** Offered only while the limit left out entries before the oldest loaded. */
    onLoadOlder?: () => void
    /** Offered only while the limit left out entries after the newest loaded. */
    onLoadNewer?: () => void
}

const CourseLogTable = ({ entries, rows, controls, onLoadOlder, onLoadNewer }: CourseLogTableProps) => {
    const [hidden, setHidden] = useState<Set<string>>(new Set())
    const [newestFirst, setNewestFirst] = useState(false)
    const [shown, setShown] = useState<{ title: string, text: string } | null>(null)
    const scroller = useRef<HTMLDivElement>(null)
    // Whether the view sits at the newest end of the log. Recorded on scroll,
    // since new rows have already moved the scroll position when they land.
    const following = useRef(true)
    const handleScroll = () => {
        const el = scroller.current
        if (!el) {
            return
        }
        following.current = newestFirst
            ? el.scrollTop <= FOLLOW_SLACK
            : el.scrollHeight - el.scrollTop - el.clientHeight <= FOLLOW_SLACK
    }

    // Keep the newest entry in view as the log grows, the way tail -f does.
    useLayoutEffect(() => {
        const el = scroller.current
        if (!el || !following.current) {
            return
        }
        el.scrollTop = newestFirst ? 0 : el.scrollHeight
    }, [rows, newestFirst])
    // Derive columns from all loaded entries so searching doesn't change the chips.
    const available = useMemo(() => [
        ...columns,
        ...Array.from(new Set(entries.flatMap(entry => Object.keys(entry.fields))))
            .sort((a, b) => a.localeCompare(b))
            .map((key): Column => ({
                id: `field:${key}`,
                label: columns.some(column => column.id === key) ? `${key} (field)` : key,
                text: entry => entry.fields[key] ?? "",
                render: (_, value) => value,
            })),
    ], [entries])
    const visible = available.filter(column => !hidden.has(column.id))
    // Reversing only affects the view, not Copy or Download.
    const ordered = newestFirst ? [...rows].reverse() : rows
    const toggleColumn = (id: string) => setHidden(previous => {
        const next = new Set(previous)
        if (next.has(id)) {
            next.delete(id)
        } else {
            next.add(id)
        }
        return next
    })

    const paging = (
        <>
            {onLoadOlder && <button type="button" className="btn btn-sm" onClick={onLoadOlder}>Load older</button>}
            {onLoadNewer && <button type="button" className="btn btn-sm" onClick={onLoadNewer}>Load newer</button>}
        </>
    )

    // Stay mounted through loading and empty results to preserve column
    // choices, and keep paging reachable when nothing loaded matches.
    if (rows.length === 0) {
        return (onLoadOlder || onLoadNewer) && (
            <div className="flex justify-center gap-2 shrink-0">{paging}</div>
        )
    }

    return (
        <LogCard
            title="Course Logs"
            className="flex-1 min-h-48"
            controls={
                <div className="flex items-center gap-2">
                    {paging}
                    <button
                        type="button"
                        className="btn btn-sm"
                        aria-pressed={newestFirst}
                        onClick={() => setNewestFirst(!newestFirst)}
                    >
                        <i className={`fas ${newestFirst ? "fa-arrow-up-9-1" : "fa-arrow-down-1-9"}`} />
                        {newestFirst ? "Newest first" : "Oldest first"}
                    </button>
                    {controls}
                </div>
            }
        >
            <div
                role="group"
                aria-label="Columns"
                className="flex flex-wrap items-center gap-1.5 shrink-0 max-h-24 overflow-y-auto border-b border-base-content/10 px-4 py-2"
            >
                <span aria-hidden="true" className="text-xs font-semibold opacity-60 mr-1">
                    <i className="fas fa-table-columns mr-1.5" />
                    Columns
                </span>
                {available.map(column => (
                    <label
                        key={column.id}
                        // chips have a tint when selected, gray with an outline when not
                        className={`btn btn-xs h-auto py-1 leading-none font-normal has-[:focus-visible]:outline has-[:focus-visible]:outline-2 has-[:focus-visible]:outline-offset-2 ${hidden.has(column.id) ? "btn-outline opacity-50" : "btn-soft btn-primary"}`}
                    >
                        <input
                            type="checkbox"
                            className="sr-only"
                            checked={!hidden.has(column.id)}
                            onChange={() => toggleColumn(column.id)}
                        />
                        {/* The accessible name reads "Show <column>" rather than the bare column name. */}
                        <span className="sr-only">Show </span>
                        {column.label}
                    </label>
                ))}
            </div>
            <div ref={scroller} onScroll={handleScroll} className="flex-1 min-h-0 overflow-auto rounded-b-2xl">
                <table className="table table-zebra table-xs">
                    <thead className="sticky top-0 z-10 bg-base-300">
                        <tr>
                            {visible.map(column => (
                                <th key={column.id} className="whitespace-nowrap">{column.label}</th>
                            ))}
                        </tr>
                    </thead>
                    <tbody>
                        {ordered.map((entry, idx) => (
                            // eslint-disable-next-line react/no-array-index-key
                            <tr key={idx}>
                                {visible.map(column => {
                                    const text = column.text?.(entry) ?? ""
                                    const value = isLong(text)
                                        ? (
                                            <button type="button" className="btn btn-xs" onClick={() => setShown({ title: column.label, text })}>
                                                Show
                                            </button>
                                        )
                                        : text
                                    return (
                                        <td key={column.id} className="align-top whitespace-pre-wrap break-words">
                                            {column.render(entry, value)}
                                        </td>
                                    )
                                })}
                            </tr>
                        ))}
                    </tbody>
                </table>
            </div>
            {shown && (
                <LogViewerModal
                    title={shown.title}
                    text={shown.text}
                    // A column label may hold spaces and parentheses, which make an
                    // awkward file name; keep it to what reads well in a shell.
                    filename={`course-log-${shown.title.replace(/[^\w.-]+/g, "-")}`}
                    onClose={() => setShown(null)}
                />
            )}
        </LogCard>
    )
}

export default CourseLogTable

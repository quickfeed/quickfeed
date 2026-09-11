import { useMemo, useState, type ReactNode } from "react"
import type { CourseLogEntry } from "../../../proto/qf/requests_pb"
import { CourseLogEntry_Level } from "../../../proto/qf/requests_pb"
import LogCard from "../LogCard"
import { entryTime, LEVEL_NAMES } from "./courseLogFormatting"

interface Column {
    id: string
    label: string
    render: (entry: CourseLogEntry) => ReactNode
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
    { id: "repository", label: "Repository", render: entry => entry.repository },
    { id: "repositoryType", label: "Repository type", render: entry => entry.repositoryType },
    {
        id: "message", label: "Message", render: entry => (
            <span className="flex flex-wrap items-center gap-2">
                <span>{entry.message}</span>
                {entry.truncated && <span className="badge badge-xs badge-warning">truncated</span>}
            </span>
        ),
    },
    { id: "source", label: "Source", render: entry => entry.source },
]

interface CourseLogTableProps {
    entries: CourseLogEntry[]
    rows: CourseLogEntry[]
    controls?: ReactNode
}

const CourseLogTable = ({ entries, rows, controls }: CourseLogTableProps) => {
    const [hidden, setHidden] = useState<Set<string>>(new Set())
    // Derive columns from all loaded entries so searching doesn't change the chips.
    const available = useMemo(() => [
        ...columns,
        ...Array.from(new Set(entries.flatMap(entry => Object.keys(entry.fields))))
            .sort((a, b) => a.localeCompare(b))
            .map((key): Column => ({
                id: `field:${key}`,
                label: columns.some(column => column.id === key) ? `${key} (field)` : key,
                render: entry => entry.fields[key] ?? "",
            })),
    ], [entries])
    const visible = available.filter(column => !hidden.has(column.id))
    const toggleColumn = (id: string) => setHidden(previous => {
        const next = new Set(previous)
        if (next.has(id)) {
            next.delete(id)
        } else {
            next.add(id)
        }
        return next
    })

    // Stay mounted through loading and empty results to preserve column choices.
    if (rows.length === 0) {
        return null
    }

    return (
        <LogCard
            title="Course Logs"
            className="flex-1 min-h-48"
            controls={<div className="flex items-center gap-2">{controls}</div>}
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
                        className={`btn btn-xs font-normal has-[:focus-visible]:outline has-[:focus-visible]:outline-2 has-[:focus-visible]:outline-offset-2 ${hidden.has(column.id) ? "btn-outline opacity-50" : "btn-soft btn-primary"}`}
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
            <div className="flex-1 min-h-0 overflow-auto rounded-b-2xl">
                <table className="table table-zebra table-xs">
                    <thead className="sticky top-0 z-10 bg-base-300">
                        <tr>
                            {visible.map(column => (
                                <th key={column.id} className="whitespace-nowrap">{column.label}</th>
                            ))}
                        </tr>
                    </thead>
                    <tbody>
                        {rows.map((entry, idx) => (
                            // eslint-disable-next-line react/no-array-index-key
                            <tr key={idx}>
                                {visible.map(column => (
                                    <td key={column.id} className="align-top whitespace-pre-wrap break-words">
                                        {column.render(entry)}
                                    </td>
                                ))}
                            </tr>
                        ))}
                    </tbody>
                </table>
            </div>
        </LogCard>
    )
}

export default CourseLogTable

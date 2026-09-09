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
    // Derive columns from all loaded entries so searching doesn't change the menu.
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
            controls={
                <div className="flex items-center gap-2">
                    <div className="dropdown dropdown-end">
                        <div tabIndex={0} role="button" className="btn btn-sm">
                            <i className="fas fa-table-columns" />
                            Columns
                        </div>
                        <ul tabIndex={0} className="dropdown-content menu z-10 mt-2 w-56 max-h-80 overflow-y-auto rounded-box bg-base-100 p-2 shadow">
                            {available.map(column => (
                                <li key={column.id}>
                                    <label className="flex items-center gap-2">
                                        <input
                                            type="checkbox"
                                            className="checkbox checkbox-xs"
                                            checked={!hidden.has(column.id)}
                                            onChange={() => toggleColumn(column.id)}
                                        />
                                        <span>Show {column.label}</span>
                                    </label>
                                </li>
                            ))}
                        </ul>
                    </div>
                    {controls}
                </div>
            }
        >
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

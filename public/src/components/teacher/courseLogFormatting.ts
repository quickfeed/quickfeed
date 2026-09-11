import { timestampDate } from "@bufbuild/protobuf/wkt"
import type { CourseLogEntry } from "../../../proto/qf/requests_pb"
import { CourseLogEntry_Level } from "../../../proto/qf/requests_pb"

export const LEVEL_NAMES: Record<CourseLogEntry_Level, string> = {
    [CourseLogEntry_Level.DEBUG]: "Debug",
    [CourseLogEntry_Level.INFO]: "Info",
    [CourseLogEntry_Level.WARN]: "Warn",
    [CourseLogEntry_Level.ERROR]: "Error",
}

const pad = (n: number): string => n.toString().padStart(2, "0")

// toLocalDatetimeInput formats date for a <input type="datetime-local"> value, in the
// browser's local time zone; Date#toISOString is always UTC, so it cannot be reused here.
export const toLocalDatetimeInput = (date: Date): string =>
    `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}T${pad(date.getHours())}:${pad(date.getMinutes())}`

// entryTime renders an entry's timestamp in a fixed 24-hour, year-month-day
// order (e.g. "2024-02-08 23:59:00"), rather than the browser locale's, which
// could show AM/PM and a locale-dependent date order such as mm/dd/yyyy.
export const entryTime = (entry: CourseLogEntry): string => {
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
export const entryText = (entry: CourseLogEntry): string => {
    const repository = entry.repository ? `[${entry.repository}]` : ""
    const parts = [entryTime(entry), LEVEL_NAMES[entry.level], repository, entry.message, entryFields(entry), entry.source]
    return parts.filter(Boolean).join(" ")
}

export const logText = (entries: readonly CourseLogEntry[]): string => entries.map(entryText).join("\n")

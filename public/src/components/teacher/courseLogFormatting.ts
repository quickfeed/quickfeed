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

const MINUTE = 60 * 1000
const HOUR = 60 * MINUTE
const DAY = 24 * HOUR

// MAX_WINDOW is the server's course log retention; a wider window returns no
// more than this one does, so the field clamps to it rather than rejecting.
export const MAX_WINDOW = 14 * DAY

const UNITS: Record<string, number> = {
    m: MINUTE, min: MINUTE, mins: MINUTE, minute: MINUTE, minutes: MINUTE,
    h: HOUR, hr: HOUR, hrs: HOUR, hour: HOUR, hours: HOUR,
    d: DAY, day: DAY, days: DAY,
}

const WINDOW = /^(\d+(?:[.,]\d+)?)\s*([a-z]+)$/

// parseWindow parses how far back the log should reach, written the way one
// says it out loud: "15 min", "15m", "1 h", "3 days". It returns the window in
// milliseconds, clamped to MAX_WINDOW, or null when text is not a window.
export const parseWindow = (text: string): number | null => {
    const match = WINDOW.exec(text.trim().toLowerCase())
    if (!match) {
        return null
    }
    const unit = UNITS[match[2]]
    const amount = Number(match[1].replace(",", "."))
    if (!unit || !(amount > 0)) {
        return null
    }
    return Math.min(amount * unit, MAX_WINDOW)
}

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
    // A repository and its type share one bracketed token, so that the type does
    // not read as a word of the message when the line is copied or downloaded.
    const repo = [entry.repository, entry.repositoryType].filter(Boolean).join(" ")
    const repository = repo ? `[${repo}]` : ""
    const parts = [entryTime(entry), LEVEL_NAMES[entry.level], repository, entry.message, entryFields(entry), entry.source]
    return parts.filter(Boolean).join(" ")
}

export const logText = (entries: readonly CourseLogEntry[]): string => entries.map(entryText).join("\n")

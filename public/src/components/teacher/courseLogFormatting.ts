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

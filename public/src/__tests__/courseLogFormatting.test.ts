import { create } from "@bufbuild/protobuf"
import { timestampFromDate } from "@bufbuild/protobuf/wkt"
import { CourseLogEntry_Level, CourseLogEntrySchema } from "../../proto/qf/requests_pb"
import { entryText, entryTime, logText, MAX_WINDOW, parseWindow } from "../components/teacher/courseLogFormatting"

describe("course log timestamps", () => {
    test.each([
        { date: new Date(2026, 0, 2, 0, 3, 4), time: "2026-01-02 00:03:04" },
        { date: new Date(2024, 1, 29, 23, 59, 58), time: "2024-02-29 23:59:58" },
    ])("formats $time using local time and a 24-hour clock", ({ date, time }) => {
        expect(entryTime(create(CourseLogEntrySchema, { time: timestampFromDate(date) }))).toBe(time)
    })

    test("omits a missing timestamp", () => {
        expect(entryTime(create(CourseLogEntrySchema))).toBe("")
    })
})

const MINUTE = 60 * 1000
const HOUR = 60 * MINUTE
const DAY = 24 * HOUR

describe("course log window", () => {
    test.each([
        { text: "15m", want: 15 * MINUTE },
        { text: "15 m", want: 15 * MINUTE },
        { text: "15min", want: 15 * MINUTE },
        { text: "15 min", want: 15 * MINUTE },
        { text: "15 mins", want: 15 * MINUTE },
        { text: "15 minute", want: 15 * MINUTE },
        { text: "90 minutes", want: 90 * MINUTE },
        { text: "1h", want: HOUR },
        { text: "1 h", want: HOUR },
        { text: "4 hr", want: 4 * HOUR },
        { text: "4hrs", want: 4 * HOUR },
        { text: "4 hour", want: 4 * HOUR },
        { text: "12 hours", want: 12 * HOUR },
        { text: "3d", want: 3 * DAY },
        { text: "3 d", want: 3 * DAY },
        { text: "3 day", want: 3 * DAY },
        { text: "3 days", want: 3 * DAY },
        { text: "  2H  ", want: 2 * HOUR },
        { text: "1.5h", want: 1.5 * HOUR },
        { text: "1,5h", want: 1.5 * HOUR },
    ])("reads $text", ({ text, want }) => {
        expect(parseWindow(text)).toBe(want)
    })

    test.each(["", "   ", "banana", "h", "15", "15 weeks", "-2h", "2 h ago", "0h", "0"])("rejects %j", text => {
        expect(parseWindow(text)).toBeNull()
    })

    // A plain object answers for the names Object.prototype carries, so this
    // reads as a unit and multiplies out to NaN, which is not null and reaches
    // Refresh and new Date as though it were a window. "constructor" is the
    // one such name that survives the lowercasing, which is what makes it the
    // case worth pinning rather than the only one the guard has to cover.
    test("rejects an inherited name in place of a unit", () => {
        expect(parseWindow("1 constructor")).toBeNull()
    })

    test("clamps to the server's retention rather than rejecting", () => {
        expect(parseWindow("30 days")).toBe(MAX_WINDOW)
        expect(parseWindow("14 days")).toBe(MAX_WINDOW)
    })
})

describe("course log text", () => {
    test("includes structured fields in key order without losing multiline output or colliding names", () => {
        const entry = create(CourseLogEntrySchema, {
            time: timestampFromDate(new Date(2026, 2, 10, 12, 0, 1)),
            level: CourseLogEntry_Level.ERROR,
            repository: "student-a",
            repositoryType: "USER",
            message: "test run failed",
            fields: { output: "--- FAIL: TestFoo\n    want 1, got 2", message: "details", assignment: "lab1" },
            source: "ci/run_tests.go:120",
        })
        expect(entryText(entry)).toBe("2026-03-10 12:00:01 Error [student-a USER] test run failed assignment=lab1 message=details output=--- FAIL: TestFoo\n    want 1, got 2 ci/run_tests.go:120")
        const reordered = create(CourseLogEntrySchema, {
            ...entry,
            fields: { assignment: "lab1", message: "details", output: "--- FAIL: TestFoo\n    want 1, got 2" },
        })
        expect(entryText(reordered)).toBe(entryText(entry))
    })

    test.each([
        [CourseLogEntry_Level.DEBUG, "Debug"],
        [CourseLogEntry_Level.INFO, "Info"],
        [CourseLogEntry_Level.WARN, "Warn"],
        [CourseLogEntry_Level.ERROR, "Error"],
    ] as const)("formats level %s and omits absent attributes", (level, name) => {
        expect(entryText(create(CourseLogEntrySchema, { level, message: "event" }))).toBe(`${name} event`)
    })

    test("exports only the supplied entries in their existing order", () => {
        const entries = [
            create(CourseLogEntrySchema, { level: CourseLogEntry_Level.WARN, message: "second" }),
            create(CourseLogEntrySchema, { level: CourseLogEntry_Level.INFO, message: "first" }),
        ]
        expect(logText(entries)).toBe("Warn second\nInfo first")
        expect(logText(entries.slice(1))).toBe("Info first")
        expect(logText([])).toBe("")
    })
})

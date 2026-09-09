import { create } from "@bufbuild/protobuf"
import { timestampFromDate } from "@bufbuild/protobuf/wkt"
import { CourseLogEntry_Level, CourseLogEntrySchema } from "../../proto/qf/requests_pb"
import { entryText, entryTime, logText, toLocalDatetimeInput } from "../components/teacher/courseLogFormatting"

describe("course log timestamps", () => {
    test.each([
        { date: new Date(2026, 0, 2, 0, 3, 4), input: "2026-01-02T00:03", time: "2026-01-02 00:03:04" },
        { date: new Date(2024, 1, 29, 23, 59, 58), input: "2024-02-29T23:59", time: "2024-02-29 23:59:58" },
    ])("formats $time using local time and a 24-hour clock", ({ date, input, time }) => {
        expect(toLocalDatetimeInput(date)).toBe(input)
        expect(entryTime(create(CourseLogEntrySchema, { time: timestampFromDate(date) }))).toBe(time)
    })

    test("omits a missing timestamp", () => {
        expect(entryTime(create(CourseLogEntrySchema))).toBe("")
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

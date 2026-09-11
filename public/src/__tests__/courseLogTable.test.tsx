import { create } from "@bufbuild/protobuf"
import { fireEvent, render, screen } from "@testing-library/react"
import { CourseLogEntrySchema } from "../../proto/qf/requests_pb"
import CourseLogTable from "../components/teacher/CourseLogTable"

describe("CourseLogTable", () => {
    test("renders and toggles fixed and structured columns independently", () => {
        const entries = [create(CourseLogEntrySchema, {
            message: "log message",
            source: "ci/tests.go:12",
            fields: { message: "structured message", source: "structured source", "field:message": "nested name" },
        })]
        render(<CourseLogTable entries={entries} rows={entries} />)

        expect(screen.getByText("log message")).toBeTruthy()
        expect(screen.getByText("structured message")).toBeTruthy()
        expect(screen.getByText("structured source")).toBeTruthy()
        expect(screen.getByText("nested name")).toBeTruthy()
        fireEvent.click(screen.getByLabelText("Show Message"))
        expect(screen.queryByText("log message")).toBeNull()
        expect(screen.getByText("structured message")).toBeTruthy()
        fireEvent.click(screen.getByLabelText("Show source (field)"))
        expect(screen.queryByText("structured source")).toBeNull()
        expect(screen.getByText("ci/tests.go:12")).toBeTruthy()
        fireEvent.click(screen.getByLabelText("Show Message"))
        expect(screen.getByText("log message")).toBeTruthy()
    })

    test("keeps columns from entries excluded by search", () => {
        const entries = [
            create(CourseLogEntrySchema, { message: "first", fields: { assignment: "lab1" } }),
            create(CourseLogEntrySchema, { message: "second", fields: { output: "test output" } }),
        ]
        render(<CourseLogTable entries={entries} rows={[entries[0]]} />)
        expect(screen.getByRole("columnheader", { name: "output" })).toBeTruthy()
        expect(screen.getByLabelText("Show output")).toBeTruthy()
        expect(screen.queryByText("test output")).toBeNull()
    })

    test("remembers hidden columns through empty results and changing fields", () => {
        const entries = [create(CourseLogEntrySchema, { fields: { commit: "abc123" } })]
        const { rerender } = render(<CourseLogTable entries={entries} rows={entries} />)
        fireEvent.click(screen.getByLabelText("Show commit"))
        rerender(<CourseLogTable entries={[]} rows={[]} />)
        expect(screen.queryByRole("table")).toBeNull()
        const withoutCommit = [create(CourseLogEntrySchema, { message: "no commit" })]
        rerender(<CourseLogTable entries={withoutCommit} rows={withoutCommit} />)
        expect(screen.queryByLabelText("Show commit")).toBeNull()
        rerender(<CourseLogTable entries={entries} rows={entries} />)
        expect(screen.queryByRole("columnheader", { name: "commit" })).toBeNull()
        expect((screen.getByLabelText("Show commit") as HTMLInputElement).checked).toBe(false)
        fireEvent.click(screen.getByLabelText("Show commit"))
        expect(screen.getByText("abc123")).toBeTruthy()
    })
})

import { create } from "@bufbuild/protobuf"
import { timestampDate, timestampFromDate } from "@bufbuild/protobuf/wkt"
import { Code, ConnectError } from "@connectrpc/connect"
import { fireEvent, render, screen, waitFor, within } from "@testing-library/react"
import { Provider } from "overmind-react"
import { MemoryRouter, Route, Routes, useNavigate } from "react-router"
import type { CourseLog, LogCursor } from "../../proto/qf/requests_pb"
import { CourseLogEntry_Level, CourseLogEntrySchema, CourseLogSchema, LogCursorSchema, LogPositionSchema } from "../../proto/qf/requests_pb"
import { UserSchema } from "../../proto/qf/types_pb"
import CourseLogs from "../components/teacher/CourseLogs"
import TeacherPage from "../pages/TeacherPage"
import { ApiClient } from "../overmind/namespaces/global/effects"
import { MockData } from "./mock_data/mockData"
import { initializeOvermind, mock } from "./TestHelpers"
import { makeStream, type TestStream } from "./testStream"

const teacherState = {
    self: create(UserSchema, { ID: BigInt(1), Name: "Teacher", IsAdmin: false }),
    activeCourse: BigInt(1),
    isTeacher: true,
    courses: MockData.mockedCourses(),
}

/** cursorAt stands in for a cursor the server handed out. */
const cursorAt = (offset: number): LogCursor =>
    create(LogCursorSchema, { date: timestampFromDate(new Date(Date.UTC(2026, 2, 10))), offset: BigInt(offset) })

const entry = (overrides: Partial<{ message: string; level: CourseLogEntry_Level; repository: string; truncated: boolean; source: string; fields: Record<string, string>; cursor: LogCursor }> = {}) =>
    create(CourseLogEntrySchema, {
        time: timestampFromDate(new Date(2026, 2, 10, 12, 0, 0)),
        level: CourseLogEntry_Level.INFO,
        message: "resolved push repository",
        repository: "student-a",
        repositoryType: "USER",
        ...overrides,
    })

const log = (init: Parameters<typeof create<typeof CourseLogSchema>>[1] = {}) => create(CourseLogSchema, init)

type Request = Parameters<ApiClient["client"]["courseLogStream"]>[0]

// The request type carries message *initializers*, not built messages, so
// reading one back needs a round trip through create.
const position = (p: Request["from"]) => create(LogPositionSchema, p).position
/** millis returns p's time in milliseconds, or NaN if p is not a time. */
const millis = (p: Request["from"]): number => {
    const pos = position(p)
    return pos.case === "time" ? timestampDate(pos.value).getTime() : NaN
}
const cursorIn = (p: Request["from"]): LogCursor | undefined => {
    const pos = position(p)
    return pos.case === "cursor" ? pos.value : undefined
}

const HOUR = 60 * 60 * 1000

/** localInput formats date the way a datetime-local input holds it. */
const localInput = (date: Date): string => {
    const pad = (n: number) => n.toString().padStart(2, "0")
    return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}T${pad(date.getHours())}:${pad(date.getMinutes())}`
}

/** mockStream stands in for CourseLogStream. Each opened stream is seeded with
 *  what seed returns for that call, and then left open, as a live tail is. */
const mockStream = (seed: (call: number) => CourseLog | Error | null) => {
    const opened: { request: Request, stream: TestStream<CourseLog> }[] = []
    const api = new ApiClient()
    api.client = {
        ...api.client,
        courseLogStream: mock("courseLogStream", request => {
            const stream = makeStream<CourseLog>()
            opened.push({ request, stream })
            const first = seed(opened.length - 1)
            if (first instanceof Error) {
                stream.fail(first)
            } else if (first) {
                stream.push(first)
            }
            return stream.iterable
        }),
    }
    return { api, opened }
}

/** backlog mocks a stream that always answers with the same backlog. */
const backlog = (message: CourseLog) => mockStream(() => message)

const renderCourseLogs = (api: ApiClient) => {
    const mockedOvermind = initializeOvermind(teacherState, api)
    render(
        <Provider value={mockedOvermind}>
            <MemoryRouter initialEntries={["/course/1/logs"]}>
                <Routes>
                    <Route path="/course/:id/logs" element={<CourseLogs />} />
                </Routes>
            </MemoryRouter>
        </Provider>
    )
    return mockedOvermind
}

/** renderWithNavigation renders the page next to a button that routes to
 *  another course, which keeps CourseLogs mounted and only changes :id. */
const renderWithNavigation = (api: ApiClient) => {
    const GoToCourseTwo = () => {
        const navigate = useNavigate()
        return <button type="button" onClick={() => navigate("/course/2/logs")}>Course 2</button>
    }
    render(
        <Provider value={initializeOvermind(teacherState, api)}>
            <MemoryRouter initialEntries={["/course/1/logs"]}>
                <GoToCourseTwo />
                <Routes>
                    <Route path="/course/:id/logs" element={<CourseLogs />} />
                </Routes>
            </MemoryRouter>
        </Provider>
    )
}

describe("CourseLogs", () => {
    test("renders the backlog the stream opens with", async () => {
        const { api } = backlog(log({
            entries: [entry({ message: "resolved push repository" }), entry({ message: "cloned assignments repository", repository: "student-b" })],
            repositories: ["student-a", "student-b"],
        }))
        renderCourseLogs(api)

        expect(await screen.findByText("resolved push repository")).toBeTruthy()
        expect(screen.getByText("cloned assignments repository")).toBeTruthy()
        // The repository selector lists every repository from the response.
        expect(screen.getByRole("option", { name: "student-a" })).toBeTruthy()
        expect(screen.getByRole("option", { name: "student-b" })).toBeTruthy()
        expect(screen.getByText("Live")).toBeTruthy()
    })

    test("appends entries as they are logged, without a Refresh", async () => {
        const { api, opened } = backlog(log({ entries: [entry()], repositories: ["student-a"] }))
        renderCourseLogs(api)
        await screen.findByText("resolved push repository")

        opened[0].stream.push(log({
            entries: [entry({ message: "running assignment tests", repository: "student-b" })],
            repositories: ["student-b"],
        }))

        // This is the whole point of the page: a teacher pushes and watches.
        expect(await screen.findByText("running assignment tests")).toBeTruthy()
        expect(screen.getByText("resolved push repository")).toBeTruthy()
        // A repository first seen live becomes a filter option.
        await waitFor(() => expect(screen.getByRole("option", { name: "student-b" })).toBeTruthy())
        expect(opened).toHaveLength(1)
    })

    test("opens on the last hour, newest first from the server, and follows", async () => {
        const { api, opened } = backlog(log({ entries: [entry()], repositories: ["student-a"] }))
        renderCourseLogs(api)
        await screen.findByText("resolved push repository")

        // No To means the server keeps the stream open and follows it.
        expect(opened[0].request.to).toBeUndefined()
        expect(opened[0].request.newest).toBe(true)
        const since = Date.now() - millis(opened[0].request.from)
        expect(since).toBeGreaterThanOrEqual(HOUR - 5000)
        expect(since).toBeLessThanOrEqual(HOUR + 5000)
        expect(screen.getByRole("button", { name: "1h" }).getAttribute("aria-pressed")).toBe("true")
    })

    test("a preset applies at once, and shows where its range begins", async () => {
        const { api, opened } = backlog(log({ entries: [entry()], repositories: ["student-a"] }))
        renderCourseLogs(api)
        await screen.findByText("resolved push repository")

        fireEvent.change(screen.getByLabelText("To"), { target: { value: localInput(new Date()) } })
        fireEvent.click(screen.getByRole("button", { name: "3d" }))
        await waitFor(() => expect(opened).toHaveLength(2))
        const since = Date.now() - millis(opened[1].request.from)
        expect(since).toBeGreaterThanOrEqual(3 * 24 * HOUR - 5000)
        expect(since).toBeLessThanOrEqual(3 * 24 * HOUR + 5000)
        expect(opened[1].request.newest).toBe(true)
        // A preset runs up to now, so it clears the To it replaces.
        expect(opened[1].request.to).toBeUndefined()
        expect((screen.getByLabelText("To") as HTMLInputElement).value).toBe("")
        expect((screen.getByLabelText("From") as HTMLInputElement).value).toBe(localInput(new Date(Date.now() - 3 * 24 * HOUR)))
        expect(screen.getByRole("button", { name: "3d" }).getAttribute("aria-pressed")).toBe("true")
        expect(screen.getByRole("button", { name: "1h" }).getAttribute("aria-pressed")).toBe("false")
    })

    test("a picked date waits for Enter, and reads forward from it", async () => {
        const { api, opened } = backlog(log({ entries: [entry()], repositories: ["student-a"] }))
        renderCourseLogs(api)
        await screen.findByText("resolved push repository")

        const picked = new Date(Date.now() - 5 * HOUR)
        picked.setSeconds(0, 0)
        const from = screen.getByLabelText("From")
        fireEvent.change(from, { target: { value: localInput(picked) } })
        // Picking a date leaves the preset behind, but asks for nothing yet.
        expect(screen.getByRole("button", { name: "1h" }).getAttribute("aria-pressed")).toBe("false")
        expect(opened).toHaveLength(1)

        fireEvent.keyUp(from, { key: "Enter" })
        await waitFor(() => expect(opened).toHaveLength(2))
        expect(millis(opened[1].request.from)).toBe(picked.getTime())
        expect(opened[1].request.newest).toBe(false)
        expect(opened[1].request.to).toBeUndefined()
    })

    test("reports a From after To instead of querying it", async () => {
        const { api, opened } = backlog(log({ entries: [entry()], repositories: ["student-a"] }))
        renderCourseLogs(api)
        await screen.findByText("resolved push repository")

        fireEvent.change(screen.getByLabelText("From"), { target: { value: localInput(new Date(Date.now() - HOUR)) } })
        fireEvent.change(screen.getByLabelText("To"), { target: { value: localInput(new Date(Date.now() - 2 * HOUR)) } })

        expect(screen.getByText(/From is after To/)).toBeTruthy()
        const refresh = screen.getByRole("button", { name: "Refresh" }) as HTMLButtonElement
        expect(refresh.disabled).toBe(true)
        fireEvent.click(refresh)
        expect(opened).toHaveLength(1)
    })

    test("shows newest entries last, and reverses them on request", async () => {
        const { api } = backlog(log({
            entries: [entry({ message: "first" }), entry({ message: "second" })],
            repositories: ["student-a"],
        }))
        renderCourseLogs(api)
        await screen.findByText("first")

        const messages = () => screen.getAllByRole("row").slice(1).map(row => row.textContent)
        expect(messages()[0]).toContain("first")
        expect(messages()[1]).toContain("second")

        fireEvent.click(screen.getByRole("button", { name: /Oldest first/ }))
        expect(messages()[0]).toContain("second")
        expect(messages()[1]).toContain("first")
        expect(screen.getByRole("button", { name: /Newest first/ })).toBeTruthy()
    })

    test("loads older entries in front of what is on screen while the range was cut short", async () => {
        const { api, opened } = mockStream(call => call === 0
            ? log({ entries: [entry({ message: "oldest loaded", cursor: cursorAt(200) })], repositories: ["student-a"], truncated: true })
            : log({ entries: [entry({ message: "from before", cursor: cursorAt(100) })], repositories: ["student-a"] }))
        renderCourseLogs(api)
        await screen.findByText("oldest loaded")
        expect(screen.queryByRole("button", { name: "Load newer" })).toBeNull()

        fireEvent.click(screen.getByRole("button", { name: "Load older" }))
        expect(await screen.findByText("from before")).toBeTruthy()

        // Bounded at the oldest entry on screen, so it ends on its own.
        await waitFor(() => expect(opened).toHaveLength(2))
        expect(cursorIn(opened[1].request.to)).toEqual(cursorAt(200))
        const rows = screen.getAllByRole("row").slice(1).map(row => row.textContent)
        expect(rows[0]).toContain("from before")
        // The older page fit, so the range is loaded back to its start.
        await waitFor(() => expect(screen.queryByRole("button", { name: "Load older" })).toBeNull())
    })

    test("loads newer entries while a picked range was cut short, and says it is not following", async () => {
        const pages = [
            log({ entries: [entry({ cursor: cursorAt(50) })], repositories: ["student-a"] }),
            log({ entries: [entry({ message: "first page", cursor: cursorAt(100) })], repositories: ["student-a"], truncated: true }),
            log({ entries: [entry({ message: "second page", cursor: cursorAt(200) })], repositories: ["student-a"] }),
        ]
        const { api, opened } = mockStream(call => pages[call])
        renderCourseLogs(api)
        await screen.findByText("resolved push repository")
        fireEvent.change(screen.getByLabelText("From"), { target: { value: localInput(new Date(Date.now() - 5 * HOUR)) } })
        fireEvent.click(screen.getByRole("button", { name: "Refresh" }))
        await screen.findByText("first page")
        // The server ends a stream whose page was cut short.
        opened[1].stream.close()
        expect(await screen.findByText("Not following")).toBeTruthy()
        expect(screen.queryByRole("button", { name: "Load older" })).toBeNull()

        fireEvent.click(screen.getByRole("button", { name: "Load newer" }))
        expect(await screen.findByText("second page")).toBeTruthy()
        expect(cursorIn(opened[2].request.from)).toEqual(cursorAt(100))
        // That page reached the present, so the view follows again.
        expect(await screen.findByText("Live")).toBeTruthy()
        expect(screen.queryByRole("button", { name: "Load newer" })).toBeNull()
    })

    test("renders the structured fields and source of an entry", async () => {
        const { api } = backlog(log({
            entries: [entry({
                message: "test run failed",
                level: CourseLogEntry_Level.ERROR,
                source: "ci/run_tests.go:120",
                fields: { assignment: "lab1", output: "--- FAIL: TestFoo\n    foo_test.go:12: want 1, got 2" },
            })],
            repositories: ["student-a"],
        }))
        renderCourseLogs(api)

        // Fields render as their own column, headed by the key, with only the
        // value in each row.
        expect(await screen.findByRole("columnheader", { name: "assignment" })).toBeTruthy()
        expect(screen.getByText("lab1")).toBeTruthy()
        expect(screen.getByText("ci/run_tests.go:120")).toBeTruthy()
    })

    test("collapses a multi-line field and shows it in full on request", async () => {
        const output = "--- FAIL: TestFoo\n    foo_test.go:12: want 1, got 2"
        const { api } = backlog(log({
            entries: [entry({ message: "test run failed", level: CourseLogEntry_Level.ERROR, fields: { output } })],
            repositories: ["student-a"],
        }))
        renderCourseLogs(api)
        await screen.findByText("test run failed")

        // Left whole, a test run's output stretches its column until the rest
        // of the table is unreadable; the cell keeps only a Show button.
        expect(screen.queryByText(/--- FAIL: TestFoo/)).toBeFalsy()
        expect(screen.queryByText(/want 1, got 2/)).toBeFalsy()

        fireEvent.click(screen.getByRole("button", { name: "Show" }))
        const dialog = screen.getByRole("dialog")
        expect(dialog.textContent).toContain("want 1, got 2")
        // The overlay is headed by the field it came from.
        expect(dialog.textContent).toContain("output")

        fireEvent.keyDown(window, { key: "Escape" })
        await waitFor(() => expect(screen.queryByRole("dialog")).toBeFalsy())
        expect(screen.queryByText(/want 1, got 2/)).toBeFalsy()
    })

    test("takes focus into the overlay and gives it back on close", async () => {
        const output = "--- FAIL: TestFoo\n    foo_test.go:12: want 1, got 2"
        const { api } = backlog(log({
            entries: [entry({ message: "test run failed", level: CourseLogEntry_Level.ERROR, fields: { output } })],
            repositories: ["student-a"],
        }))
        renderCourseLogs(api)
        await screen.findByText("test run failed")

        const show = screen.getByRole("button", { name: "Show" })
        show.focus()
        fireEvent.click(show)

        // The overlay covers the table it was opened from; focus left behind on
        // the trigger would leave a keyboard user tabbing through rows they can
        // no longer see.
        const dialog = screen.getByRole("dialog")
        await waitFor(() => expect(dialog.contains(document.activeElement)).toBe(true))

        // Tab off the end of the dialog wraps back into it rather than reaching
        // the page behind.
        const focusable = Array.from(dialog.querySelectorAll<HTMLElement>("button"))
        const last = focusable[focusable.length - 1]
        last.focus()
        fireEvent.keyDown(window, { key: "Tab" })
        expect(document.activeElement).toBe(focusable[0])

        fireEvent.keyDown(window, { key: "Escape" })
        await waitFor(() => expect(screen.queryByRole("dialog")).toBeFalsy())
        // Back on the row the teacher was reading.
        expect(document.activeElement).toBe(show)
    })

    test("reports a clipboard the overlay will not be given access to", async () => {
        const output = "--- FAIL: TestFoo\n    foo_test.go:12: want 1, got 2"
        const { api } = backlog(log({
            entries: [entry({ message: "test run failed", level: CourseLogEntry_Level.ERROR, fields: { output } })],
            repositories: ["student-a"],
        }))
        renderCourseLogs(api)
        await screen.findByText("test run failed")
        fireEvent.click(screen.getByRole("button", { name: "Show" }))

        // jsdom leaves navigator.clipboard undefined, as a browser does outside
        // a secure context. The overlay's own Copy must report that the way the
        // page's does: the output is there to select by hand, but only a reader
        // who knows the button did nothing will go and do it.
        const dialog = screen.getByRole("dialog")
        fireEvent.click(within(dialog).getByRole("button", { name: "Copy" }))
        expect(await within(dialog).findByText(/Could not copy/)).toBeTruthy()
    })

    test("collapses a long message, keeping the truncated badge beside it", async () => {
        // The store cuts a record's message at 64 KiB and marks the record, so
        // a message is as free-form as a field and must collapse the same way.
        const message = `cloning failed\n${"x".repeat(400)}`
        const { api } = backlog(log({
            entries: [entry({ message, level: CourseLogEntry_Level.ERROR, truncated: true })],
            repositories: ["student-a"],
        }))
        renderCourseLogs(api)
        // The badge says the record was cut, which the overlay cannot show.
        expect(await screen.findByText("truncated")).toBeTruthy()
        expect(screen.queryByText(/cloning failed/)).toBeFalsy()

        fireEvent.click(screen.getByRole("button", { name: "Show" }))
        const dialog = screen.getByRole("dialog")
        expect(dialog.textContent).toContain("x".repeat(400))
        // The overlay names itself for a screen reader, which otherwise
        // announces it as an unnamed dialog.
        expect(dialog.getAttribute("aria-label")).toBe("Message")
    })

    test("says so when Load older fails, without disturbing the live view", async () => {
        // The first stream is the live tail; the second is the bounded request
        // Load older makes, which here is refused.
        const { api } = mockStream(call => call === 0
            ? log({ entries: [entry({ message: "already loaded", cursor: cursorAt(100) })], repositories: ["student-a"], truncated: true })
            : new ConnectError("reading course log", Code.Internal))
        renderCourseLogs(api)
        await screen.findByText("already loaded")

        fireEvent.click(screen.getByRole("button", { name: "Load older" }))
        await screen.findByText(/Could not load older entries/)
        // The live stream is still running and what it loaded still stands, so
        // the page must not be replaced by an error.
        expect(screen.getByText("already loaded")).toBeTruthy()
        expect(screen.queryByText(/Failed to load course logs/)).toBeFalsy()
    })

    test("leaves a short field value in the cell", async () => {
        const { api } = backlog(log({
            entries: [entry({ fields: { commit: "abc123" } })],
            repositories: ["student-a"],
        }))
        renderCourseLogs(api)
        await screen.findByText("resolved push repository")

        expect(screen.getByText("abc123")).toBeTruthy()
        expect(screen.queryByRole("button", { name: "Show" })).toBeFalsy()
    })

    test("gives repositoryType its own column, and a colliding field key its own", async () => {
        const { api } = backlog(log({
            entries: [entry({
                message: "resolved push repository",
                // "message" collides with a fixed column: the server keys the
                // log message as "msg", so an attribute named "message" does
                // reach the fields map.
                fields: { commit: "abc123", message: "shadowed" },
            })],
            repositories: ["student-a"],
        }))
        renderCourseLogs(api)
        await screen.findByText("resolved push repository")

        // repositoryType is promoted out of fields by the server, so it needs a
        // column of its own or it is not on screen at all.
        expect(screen.getByRole("columnheader", { name: "Repository type" })).toBeTruthy()
        expect(screen.getByText("USER")).toBeTruthy()
        expect(screen.getByRole("columnheader", { name: "commit" })).toBeTruthy()
        // The colliding key gets a column of its own, suffixed so it reads apart
        // from the fixed Message column, which it neither claims nor displaces.
        expect(screen.getAllByRole("columnheader", { name: "Message" })).toHaveLength(1)
        expect(screen.getByRole("columnheader", { name: "message (field)" })).toBeTruthy()
        expect(screen.getByText("shadowed")).toBeTruthy()
    })

    test("renders entry timestamps in 24-hour, yyyy-mm-dd order rather than the locale's", async () => {
        const { api } = backlog(log({ entries: [entry()], repositories: ["student-a"] }))
        renderCourseLogs(api)

        // entry() fixes the time at 2026-03-10 12:00:00 local, so the rendered
        // text is deterministic regardless of the machine's own time zone.
        expect(await screen.findByText("2026-03-10 12:00:00")).toBeTruthy()
    })

    test("shows the Debug badge for a Debug-level entry", async () => {
        const { api } = backlog(log({
            entries: [entry({ message: "verbose detail", level: CourseLogEntry_Level.DEBUG })],
            repositories: ["student-a"],
        }))
        renderCourseLogs(api)

        expect(await screen.findByText("verbose detail")).toBeTruthy()
        // "Debug" also names an option in the "Minimum level" select; scope to the badge.
        expect(screen.getByText("Debug", { selector: "span.badge" })).toBeTruthy()
    })

    test("hides and reshows a column from the Columns menu", async () => {
        const { api } = backlog(log({
            entries: [entry({ message: "resolved push repository", source: "ci/run_tests.go:120" })],
            repositories: ["student-a"],
        }))
        renderCourseLogs(api)
        await screen.findByText("resolved push repository")
        expect(screen.getByText("ci/run_tests.go:120")).toBeTruthy()

        fireEvent.click(screen.getByLabelText("Show Source"))
        expect(screen.queryByText("ci/run_tests.go:120")).toBeFalsy()
        expect(screen.queryByRole("columnheader", { name: "Source" })).toBeFalsy()

        fireEvent.click(screen.getByLabelText("Show Source"))
        expect(screen.getByText("ci/run_tests.go:120")).toBeTruthy()
    })

    test("keeps column choices through Refresh and an empty search", async () => {
        const { api } = backlog(log({ entries: [entry({ fields: { message: "structured detail" } })] }))
        renderCourseLogs(api)
        await screen.findByText("resolved push repository")
        fireEvent.click(screen.getByLabelText("Show Message"))
        fireEvent.click(screen.getByRole("button", { name: "Refresh" }))
        await screen.findByText("structured detail")
        expect(screen.queryByRole("columnheader", { name: "Message" })).toBeNull()
        const search = screen.getByPlaceholderText("Filter loaded entries")
        fireEvent.keyUp(search, { target: { value: "no match" } })
        expect(screen.queryByRole("table")).toBeNull()
        fireEvent.keyUp(search, { target: { value: "resolved push" } })
        expect(screen.getByText("structured detail")).toBeTruthy()
        expect(screen.queryByRole("columnheader", { name: "Message" })).toBeNull()
        fireEvent.click(screen.getByLabelText("Show Message"))
        expect(screen.getByText("resolved push repository")).toBeTruthy()
    })

    test("applies draft filters only on Refresh, including repeated refreshes", async () => {
        const { api, opened } = backlog(log({ entries: [entry()], repositories: ["student-a"] }))
        renderCourseLogs(api)
        await screen.findByText("resolved push repository")
        const picked = new Date(Date.now() - 45 * 60 * 1000)
        picked.setSeconds(0, 0)
        fireEvent.change(screen.getByLabelText("From"), { target: { value: localInput(picked) } })
        fireEvent.change(screen.getByLabelText("Repository"), { target: { value: "student-a" } })
        fireEvent.change(screen.getByLabelText("Minimum level"), { target: { value: CourseLogEntry_Level.ERROR } })
        expect(opened).toHaveLength(1)

        fireEvent.click(screen.getByRole("button", { name: "Refresh" }))
        await waitFor(() => expect(opened).toHaveLength(2))
        expect(opened[1].request).toMatchObject({ repository: "student-a", level: CourseLogEntry_Level.ERROR, newest: false })
        expect(millis(opened[1].request.from)).toBe(picked.getTime())

        fireEvent.click(screen.getByRole("button", { name: "Refresh" }))
        await waitFor(() => expect(opened).toHaveLength(3))
        expect(opened[2].request).toMatchObject({ repository: "student-a", level: CourseLogEntry_Level.ERROR })
    })

    test("a preset takes pending filters along", async () => {
        const { api, opened } = backlog(log({ entries: [entry()], repositories: ["student-a"] }))
        renderCourseLogs(api)
        await screen.findByText("resolved push repository")
        fireEvent.change(screen.getByLabelText("Minimum level"), { target: { value: CourseLogEntry_Level.WARN } })
        fireEvent.click(screen.getByRole("button", { name: "4h" }))
        await waitFor(() => expect(opened).toHaveLength(2))
        expect(opened[1].request.level).toBe(CourseLogEntry_Level.WARN)
    })

    test("shows an error state when the stream fails", async () => {
        const { api } = mockStream(() => new ConnectError("permission denied", Code.PermissionDenied))
        renderCourseLogs(api)

        expect(await screen.findByText(/Failed to load course logs/)).toBeTruthy()
    })

    test("shows an empty state when there are no matching entries", async () => {
        const { api } = backlog(log({ entries: [], repositories: [], truncated: false }))
        renderCourseLogs(api)

        expect(await screen.findByText("No log entries match the current filters")).toBeTruthy()
    })

    test("filters loaded entries locally without reopening the stream", async () => {
        const { api, opened } = backlog(log({
            entries: [entry({ message: "resolved push repository" }), entry({ message: "cloned assignments repository" })],
            repositories: ["student-a"],
        }))
        renderCourseLogs(api)
        await screen.findByText("resolved push repository")
        expect(opened).toHaveLength(1)

        const search = screen.getByPlaceholderText("Filter loaded entries")
        fireEvent.keyUp(search, { target: { value: "cloned" } })

        await waitFor(() => {
            expect(screen.getByText("cloned assignments repository")).toBeTruthy()
            expect(screen.queryByText("resolved push repository")).toBeFalsy()
        })
        // The search filter is applied client-side; it must not reopen the stream.
        expect(opened).toHaveLength(1)
    })

    test("keeps paging within reach when a search hides every loaded entry", async () => {
        const { api } = backlog(log({ entries: [entry({ cursor: cursorAt(100) })], repositories: ["student-a"], truncated: true }))
        renderCourseLogs(api)
        await screen.findByText("resolved push repository")

        // What the search looks for may be in the part not yet loaded.
        fireEvent.keyUp(screen.getByPlaceholderText("Filter loaded entries"), { target: { value: "no such entry" } })
        await screen.findByText("No log entries match the current filters")
        expect(screen.getByRole("button", { name: "Load older" })).toBeTruthy()
    })

    test("offers no paging when the range fit within the limit", async () => {
        const { api } = backlog(log({ entries: [entry({ cursor: cursorAt(100) })], repositories: ["student-a"] }))
        renderCourseLogs(api)
        await screen.findByText("resolved push repository")
        expect(screen.queryByRole("button", { name: "Load older" })).toBeNull()
        expect(screen.queryByRole("button", { name: "Load newer" })).toBeNull()
    })

    test("reports a clipboard the browser will not give access to", async () => {
        const { api } = backlog(log({ entries: [entry()], repositories: ["student-a"] }))
        renderCourseLogs(api)
        await screen.findByText("resolved push repository")

        // jsdom leaves navigator.clipboard undefined, as a browser does outside
        // a secure context; the failure must reach the teacher, not the console.
        fireEvent.click(screen.getByRole("button", { name: "Copy" }))

        expect(await screen.findByText(/Could not copy the log/)).toBeTruthy()
    })

    test("keeps a selected repository listed after it falls silent", async () => {
        // The second stream no longer reports an entry for student-b.
        const { api, opened } = mockStream(call => log({
            entries: [entry()],
            repositories: call === 0 ? ["student-a", "student-b"] : ["student-a"],
        }))
        renderCourseLogs(api)
        await screen.findByText("resolved push repository")

        fireEvent.change(screen.getByLabelText("Repository"), { target: { value: "student-b" } })
        fireEvent.click(screen.getByRole("button", { name: "Refresh" }))
        await waitFor(() => expect(opened).toHaveLength(2))
        expect(opened[1].request.repository).toBe("student-b")

        // The select must still show what it is filtering on, not sit blank.
        await screen.findByText("resolved push repository")
        expect(screen.getByRole("option", { name: "student-b" })).toBeTruthy()
        expect((screen.getByLabelText("Repository") as HTMLSelectElement).value).toBe("student-b")
    })

    test("clears the repository selection when the route names another course", async () => {
        const { api, opened } = backlog(log({ entries: [entry()], repositories: ["student-a"] }))
        renderWithNavigation(api)
        await screen.findByText("resolved push repository")
        fireEvent.change(screen.getByLabelText("Repository"), { target: { value: "student-a" } })
        const picked = localInput(new Date(Date.now() - 6 * HOUR))
        fireEvent.change(screen.getByLabelText("From"), { target: { value: picked } })
        fireEvent.click(screen.getByRole("button", { name: "Course 2" }))

        // A repository belongs to one course; it must not filter another's log.
        await waitFor(() => expect(opened).toHaveLength(2))
        expect(opened[1].request.courseID).toBe(BigInt(2))
        expect(opened[1].request.repository).toBe("")
        // The date the teacher picked is theirs, not the course's.
        expect((screen.getByLabelText("From") as HTMLInputElement).value).toBe(picked)
        expect((screen.getByLabelText("Repository") as HTMLSelectElement).value).toBe("")
    })

    test("reopens the stream when the route names another course", async () => {
        const { api, opened } = backlog(log({ entries: [entry()], repositories: ["student-a"] }))
        // The router keeps CourseLogs mounted when only :id changes, so the page
        // would otherwise keep showing the previous course's entries.
        renderWithNavigation(api)
        await screen.findByText("resolved push repository")
        expect(opened).toHaveLength(1)
        expect(opened[0].request.courseID).toBe(BigInt(1))

        fireEvent.click(screen.getByRole("button", { name: "Course 2" }))

        await waitFor(() => expect(opened).toHaveLength(2))
        expect(opened[1].request.courseID).toBe(BigInt(2))
    })
})

describe("CourseLogs tile", () => {
    test("navigates from the teacher tile to the logs page", async () => {
        const { api } = backlog(log({ entries: [entry()], repositories: ["student-a"], truncated: false }))
        const mockedOvermind = initializeOvermind(teacherState, api)
        render(
            <Provider value={mockedOvermind}>
                <MemoryRouter initialEntries={["/course/1"]}>
                    <Routes>
                        <Route path="/course/:id/*" element={<TeacherPage />} />
                    </Routes>
                </MemoryRouter>
            </Provider>
        )

        fireEvent.click(screen.getByText("Logs"))

        expect(await screen.findByText("resolved push repository")).toBeTruthy()
    })
})

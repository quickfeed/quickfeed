import { create } from "@bufbuild/protobuf"
import { timestampDate, timestampFromDate, TimestampSchema } from "@bufbuild/protobuf/wkt"
import { Code, ConnectError } from "@connectrpc/connect"
import { fireEvent, render, screen, waitFor } from "@testing-library/react"
import { Provider } from "overmind-react"
import { MemoryRouter, Route, Routes, useNavigate } from "react-router"
import type { CourseLog } from "../../proto/qf/requests_pb"
import { CourseLogEntry_Level, CourseLogEntrySchema, CourseLogSchema } from "../../proto/qf/requests_pb"
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

const entry = (overrides: Partial<{ message: string; level: CourseLogEntry_Level; repository: string; truncated: boolean; source: string; fields: Record<string, string> }> = {}) =>
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

// The request type carries timestamp *initializers*, not built messages, so
// reading one back needs a round trip through create.
const millis = (ts: Request["from"]): number => timestampDate(create(TimestampSchema, ts)).getTime()

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

    test("sends the chosen window and leaves the upper bound open", async () => {
        const { api, opened } = backlog(log({ entries: [entry()], repositories: ["student-a"] }))
        renderCourseLogs(api)
        await screen.findByText("resolved push repository")

        // No To means the server keeps the stream open and tails it.
        expect(opened[0].request.to).toBeUndefined()
        const from = millis(opened[0].request.from)
        expect(Date.now() - from).toBeGreaterThanOrEqual(2 * 60 * 60 * 1000 - 5000)

        fireEvent.change(screen.getByLabelText("Show the last"), { target: { value: "3 days" } })
        fireEvent.click(screen.getByRole("button", { name: "Refresh" }))
        await waitFor(() => expect(opened).toHaveLength(2))
        const days = Date.now() - millis(opened[1].request.from)
        expect(days).toBeGreaterThanOrEqual(3 * 24 * 60 * 60 * 1000 - 5000)
    })

    test("reports an unwritable window instead of querying it", async () => {
        const { api, opened } = backlog(log({ entries: [entry()], repositories: ["student-a"] }))
        renderCourseLogs(api)
        await screen.findByText("resolved push repository")
        expect(opened).toHaveLength(1)

        fireEvent.change(screen.getByLabelText("Show the last"), { target: { value: "a fortnight" } })

        // The server would read an empty window as "nothing happened" rather
        // than "your window is not a window".
        expect(screen.getByText(/Write how far back to look/)).toBeTruthy()
        const refresh = screen.getByRole("button", { name: "Refresh" }) as HTMLButtonElement
        expect(refresh.disabled).toBe(true)
        refresh.click()
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

    test("loads an older window in front of what is on screen", async () => {
        const { api, opened } = mockStream(call => call === 0
            ? log({ entries: [entry({ message: "oldest loaded" })], repositories: ["student-a"] })
            : log({ entries: [entry({ message: "from before" })], repositories: ["student-a"] }))
        renderCourseLogs(api)
        await screen.findByText("oldest loaded")

        fireEvent.click(screen.getByRole("button", { name: "Load older" }))
        expect(await screen.findByText("from before")).toBeTruthy()

        // A bounded request, so it ends on its own rather than tailing.
        await waitFor(() => expect(opened).toHaveLength(2))
        expect(opened[1].request.to).toBeDefined()
        const rows = screen.getAllByRole("row").slice(1).map(row => row.textContent)
        expect(rows[0]).toContain("from before")
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

        // The CI output a teacher needs lives in fields; it must be on screen,
        // not only in the copied and downloaded text.
        expect(await screen.findByText(/--- FAIL: TestFoo/)).toBeTruthy()
        // Fields render as their own column, headed by the key, with only the
        // value in each row.
        expect(screen.getByRole("columnheader", { name: "assignment" })).toBeTruthy()
        expect(screen.getByText("lab1")).toBeTruthy()
        expect(screen.getByText("ci/run_tests.go:120")).toBeTruthy()
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
        fireEvent.change(screen.getByLabelText("Show the last"), { target: { value: "45 min" } })
        fireEvent.change(screen.getByLabelText("Repository"), { target: { value: "student-a" } })
        fireEvent.change(screen.getByLabelText("Minimum level"), { target: { value: CourseLogEntry_Level.ERROR } })
        expect(opened).toHaveLength(1)

        fireEvent.click(screen.getByRole("button", { name: "Refresh" }))
        await waitFor(() => expect(opened).toHaveLength(2))
        expect(opened[1].request).toMatchObject({ repository: "student-a", level: CourseLogEntry_Level.ERROR })
        const window = Date.now() - millis(opened[1].request.from)
        expect(window).toBeGreaterThanOrEqual(45 * 60 * 1000 - 5000)
        expect(window).toBeLessThanOrEqual(45 * 60 * 1000 + 5000)

        fireEvent.click(screen.getByRole("button", { name: "Refresh" }))
        await waitFor(() => expect(opened).toHaveLength(3))
        expect(opened[2].request).toMatchObject({ repository: "student-a", level: CourseLogEntry_Level.ERROR })
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

    test("reports the server-side cutoff, not the locally filtered count", async () => {
        const { api } = backlog(log({
            entries: [entry({ message: "resolved push repository" }), entry(), entry()],
            repositories: ["student-a"],
            truncated: true,
        }))
        renderCourseLogs(api)

        await waitFor(() => expect(screen.getByText(/more than the newest 3 entries/)).toBeTruthy())

        // Filtering locally changes neither the cutoff nor the count it reports,
        // and the warning must survive a filter that matches nothing.
        const search = screen.getByPlaceholderText("Filter loaded entries")
        fireEvent.keyUp(search, { target: { value: "no such entry" } })

        await waitFor(() => expect(screen.getByText("No log entries match the current filters")).toBeTruthy())
        expect(screen.getByText(/more than the newest 3 entries/)).toBeTruthy()
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

    test("offers Load older when the window came back empty", async () => {
        const { api, opened } = mockStream(() => log({ entries: [], repositories: [] }))
        renderCourseLogs(api)
        await screen.findByText(/No log entries match the current filters/)

        // A window that held nothing is exactly when a teacher reaches further
        // back, so the action cannot be hidden along with the table.
        fireEvent.click(screen.getByRole("button", { name: "Load older" }))
        await waitFor(() => expect(opened).toHaveLength(2))
        // A bounded request, so it ends on its own rather than tailing.
        expect(opened[1].request.to).toBeTruthy()
    })

    test("reports a clipboard the browser will not give access to", async () => {
        const { api } = backlog(log({ entries: [entry()], repositories: ["student-a"] }))
        renderCourseLogs(api)
        await screen.findByText("resolved push repository")

        // jsdom leaves navigator.clipboard undefined, as a browser does outside
        // a secure context; the failure must reach the teacher, not the console.
        screen.getByRole("button", { name: "Copy" }).click()

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
        screen.getByRole("button", { name: "Refresh" }).click()
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
        fireEvent.change(screen.getByLabelText("Show the last"), { target: { value: "6 h" } })
        screen.getByRole("button", { name: "Course 2" }).click()

        // A repository belongs to one course; it must not filter another's log.
        await waitFor(() => expect(opened).toHaveLength(2))
        expect(opened[1].request.courseID).toBe(BigInt(2))
        expect(opened[1].request.repository).toBe("")
        // The window the teacher typed is theirs, not the course's.
        expect((screen.getByLabelText("Show the last") as HTMLInputElement).value).toBe("6 h")
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

        screen.getByRole("button", { name: "Course 2" }).click()

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

        screen.getByText("Logs").click()

        expect(await screen.findByText("resolved push repository")).toBeTruthy()
    })
})

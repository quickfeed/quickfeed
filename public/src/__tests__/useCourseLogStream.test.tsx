import { create } from "@bufbuild/protobuf"
import { timestampDate, timestampFromDate } from "@bufbuild/protobuf/wkt"
import { Code, ConnectError } from "@connectrpc/connect"
import { act, renderHook, waitFor } from "@testing-library/react"
import { Provider } from "overmind-react"
import type { ReactNode } from "react"
import type { CourseLog, LogCursor } from "../../proto/qf/requests_pb"
import { CourseLogEntry_Level, CourseLogSchema, LogCursorSchema, LogPositionSchema } from "../../proto/qf/requests_pb"
import { ConnStatus } from "../Helpers"
import type { CourseLogFilters } from "../hooks/useCourseLogStream"
import { useCourseLogStream } from "../hooks/useCourseLogStream"
import { ApiClient } from "../overmind/namespaces/global/effects"
import { initializeOvermind, mock } from "./TestHelpers"
import { makeStream, type TestStream } from "./testStream"

const HOUR = 60 * 60 * 1000
const preset: CourseLogFilters = { range: { kind: "preset", duration: 2 * HOUR }, repository: "", level: CourseLogEntry_Level.DEBUG }
const tuesday = new Date(2026, 8, 22, 14, 30)
const dates = (to: Date | null = null): CourseLogFilters => ({ ...preset, range: { kind: "dates", from: tuesday, to } })

/** cursorOf stands in for the cursor the server gives each entry: the same
 *  one for the same text, and a different one for each text. */
const cursors = new Map<string, LogCursor>()
const cursorOf = (text: string): LogCursor => {
    let cursor = cursors.get(text)
    if (!cursor) {
        cursor = create(LogCursorSchema, {
            date: timestampFromDate(new Date(Date.UTC(2023, 10, 14))),
            offset: BigInt(100 * (cursors.size + 1)),
        })
        cursors.set(text, cursor)
    }
    return cursor
}

const message = (messages: string[], extra: { repositories?: string[], truncated?: boolean } = {}) => create(CourseLogSchema, {
    entries: messages.map((text, i) => ({
        message: text,
        time: timestampFromDate(new Date(1_700_000_000_000 + i * 1000)),
        cursor: cursorOf(text),
    })),
    ...extra,
})

/** entryAt builds a one-entry message at an exact instant. */
const entryAt = (text: string, seconds: bigint, nanos: number) => create(CourseLogSchema, {
    entries: [{ message: text, time: { seconds, nanos }, cursor: cursorOf(text) }],
})

type Request = Parameters<ApiClient["client"]["courseLogStream"]>[0]

// The request type carries message *initializers*, not built messages, so
// reading one back needs a round trip through create.
const position = (p: Request["from"]) => create(LogPositionSchema, p)
/** timeOf returns p's time in milliseconds, or null if p is not a time. */
const timeOf = (p: Request["from"]): number | null => {
    const pos = position(p).position
    return pos.case === "time" ? timestampDate(pos.value).getTime() : null
}
/** cursorIn returns p's cursor, or null if p is not a cursor. */
const cursorIn = (p: Request["from"]): LogCursor | null => {
    const pos = position(p).position
    return pos.case === "cursor" ? pos.value : null
}

const setup = (filters: CourseLogFilters = preset) => {
    const opened: { request: Request, stream: TestStream<CourseLog> }[] = []
    const api = new ApiClient()
    api.client = {
        ...api.client,
        courseLogStream: mock("courseLogStream", request => {
            const stream = makeStream<CourseLog>()
            opened.push({ request, stream })
            return stream.iterable
        }),
    }
    const state = initializeOvermind({}, api)
    const wrapper = ({ children }: { children: ReactNode }) => <Provider value={state}>{children}</Provider>
    return {
        ...renderHook(({ courseID }) => useCourseLogStream(courseID, filters), { wrapper, initialProps: { courseID: 1n } }),
        opened,
    }
}

// Waits for the nth stream to be opened; the hook opens them from an effect,
// and a reconnect waits out its backoff first.
const nthStream = async (opened: { stream: TestStream<CourseLog> }[], n: number, timeout = 1000) => {
    await waitFor(() => expect(opened.length).toBeGreaterThan(n), { timeout })
    return opened[n].stream
}

// Gives a stray request every chance to be sent before asserting none was.
const settle = () => act(async () => { await new Promise(resolve => setTimeout(resolve, 50)) })

describe("useCourseLogStream", () => {
    test("seeds from the backlog and appends live entries", async () => {
        const { result, opened } = setup()
        const stream = await nthStream(opened, 0)
        await act(async () => stream.push(message(["backlog one", "backlog two"], { repositories: ["student-a"] })))
        await waitFor(() => expect(result.current.loading).toBe(false))
        expect(result.current.result?.entries.map(e => e.message)).toEqual(["backlog one", "backlog two"])
        expect(result.current.status).toBe(ConnStatus.CONNECTED)

        await act(async () => stream.push(message(["live"], { repositories: ["student-b"] })))
        expect(result.current.result?.entries.map(e => e.message)).toEqual(["backlog one", "backlog two", "live"])
        // A repository first seen live joins the filter options, kept sorted.
        expect(result.current.result?.repositories).toEqual(["student-a", "student-b"])
    })

    test("a preset asks for the newest entries since its start, and leaves To unset so the stream follows", async () => {
        const { opened } = setup()
        await nthStream(opened, 0)
        const { request } = opened[0]
        expect(request.to).toBeUndefined()
        expect(request.newest).toBe(true)
        const from = timeOf(request.from) ?? 0
        expect(Date.now() - from).toBeGreaterThanOrEqual(2 * HOUR - 5000)
        expect(Date.now() - from).toBeLessThanOrEqual(2 * HOUR + 5000)
    })

    test("dates ask for the oldest entries from From, and pass To on", async () => {
        const to = new Date(2026, 8, 22, 16, 0)
        const { opened } = setup(dates(to))
        await nthStream(opened, 0)
        const { request } = opened[0]
        expect(request.newest).toBe(false)
        expect(timeOf(request.from)).toBe(tuesday.getTime())
        expect(timeOf(request.to)).toBe(to.getTime())
    })

    test("refresh reopens the stream with the new filters and drops what was loaded", async () => {
        const { result, opened } = setup()
        const first = await nthStream(opened, 0)
        await act(async () => first.push(message(["old"])))
        await waitFor(() => expect(result.current.loading).toBe(false))

        act(() => result.current.refresh({ ...preset, repository: "student-a", level: CourseLogEntry_Level.ERROR }))
        expect(result.current.result).toBeNull()
        const second = await nthStream(opened, 1)
        expect(opened[1].request).toMatchObject({ repository: "student-a", level: CourseLogEntry_Level.ERROR })
        await act(async () => second.push(message(["new"])))
        expect(result.current.result?.entries.map(e => e.message)).toEqual(["new"])
    })

    test("a course change reopens the stream and clears the repository filter", async () => {
        const { result, rerender, opened } = setup()
        const first = await nthStream(opened, 0)
        act(() => result.current.refresh({ ...preset, repository: "student-a", level: CourseLogEntry_Level.ERROR }))
        await nthStream(opened, 1)
        first.close()

        rerender({ courseID: 2n })
        await nthStream(opened, 2)
        expect(opened[2].request).toMatchObject({ courseID: 2n, repository: "", level: CourseLogEntry_Level.ERROR })
        expect(result.current.result).toBeNull()
    })

    test("reports a dropped stream and reconnects forward from the newest entry it has", async () => {
        const { result, opened } = setup()
        const first = await nthStream(opened, 0)
        await act(async () => first.push(message(["backlog"])))
        await waitFor(() => expect(result.current.loading).toBe(false))

        await act(async () => first.fail(new ConnectError("connection reset", Code.Unavailable)))
        await waitFor(() => expect(result.current.status).toBe(ConnStatus.RECONNECTING))
        expect(result.current.error).toContain("connection reset")

        const second = await nthStream(opened, 1, 3000)
        // The reconnect reads on from the newest entry already on screen, and
        // oldest first, so an outage longer than the limit is paged through
        // rather than skipped.
        expect(cursorIn(opened[1].request.from)).toEqual(cursorOf("backlog"))
        expect(opened[1].request.newest).toBe(false)
        await act(async () => second.push(message(["after reconnect"])))
        await waitFor(() => expect(result.current.status).toBe(ConnStatus.CONNECTED))
        // The entries already on screen are kept across the reconnect.
        expect(result.current.result?.entries.map(e => e.message)).toEqual(["backlog", "after reconnect"])
        expect(result.current.error).toBeNull()
    })

    test("reconnects after a cancel the page did not ask for, rather than loading for good", async () => {
        const { result, opened } = setup()
        const first = await nthStream(opened, 0)
        // A proxy or the browser can cancel a stream the page still wants.
        await act(async () => first.fail(new ConnectError("stream reset", Code.Canceled)))
        await waitFor(() => expect(result.current.status).toBe(ConnStatus.RECONNECTING))
        expect(result.current.error).toContain("stream reset")

        const second = await nthStream(opened, 1, 3000)
        await act(async () => second.push(message(["after the cancel"])))
        await waitFor(() => expect(result.current.loading).toBe(false))
        expect(result.current.status).toBe(ConnStatus.CONNECTED)
    })

    test("stops rather than retrying when the request is rejected", async () => {
        const { result, opened } = setup()
        const first = await nthStream(opened, 0)
        await act(async () => first.fail(new ConnectError("not teacher", Code.PermissionDenied)))
        await waitFor(() => expect(result.current.error).toContain("not teacher"))
        expect(result.current.status).toBe(ConnStatus.DISCONNECTED)
        await settle()
        expect(opened).toHaveLength(1)
    })

    test("loadOlder prepends the page before the oldest entry on screen, within the range", async () => {
        const { result, opened } = setup()
        const stream = await nthStream(opened, 0)
        await act(async () => stream.push(message(["oldest loaded", "newer"], { truncated: true })))
        await waitFor(() => expect(result.current.result?.moreOlder).toBe(true))

        await act(async () => void result.current.loadOlder())
        const older = await nthStream(opened, 1)
        // Bounded at the oldest entry, which the cursor leaves out, and at the
        // preset's start, which is the range the view began with.
        expect(cursorIn(opened[1].request.to)).toEqual(cursorOf("oldest loaded"))
        expect(timeOf(opened[1].request.from)).toBe(timeOf(opened[0].request.from))
        expect(opened[1].request.newest).toBe(true)

        await act(async () => { older.push(message(["earlier"])); older.close() })
        await waitFor(() => expect(result.current.result?.entries).toHaveLength(3))
        expect(result.current.result?.entries.map(e => e.message)).toEqual(["earlier", "oldest loaded", "newer"])
        // The older page fit, so the range is loaded back to its start.
        expect(result.current.result?.moreOlder).toBe(false)
    })

    test("loadOlder keeps offering more while the older page was cut short too", async () => {
        const { result, opened } = setup()
        const stream = await nthStream(opened, 0)
        await act(async () => stream.push(message(["oldest loaded"], { truncated: true })))
        await waitFor(() => expect(result.current.loading).toBe(false))

        await act(async () => void result.current.loadOlder())
        const older = await nthStream(opened, 1)
        await act(async () => { older.push(message(["earlier"], { truncated: true })); older.close() })
        await waitFor(() => expect(result.current.result?.entries).toHaveLength(2))
        expect(result.current.result?.moreOlder).toBe(true)

        await act(async () => void result.current.loadOlder())
        await nthStream(opened, 2)
        expect(cursorIn(opened[2].request.to)).toEqual(cursorOf("earlier"))
    })

    test("a forward page cut short waits for loadNewer, then follows once caught up", async () => {
        const { result, opened } = setup(dates())
        const first = await nthStream(opened, 0)
        await act(async () => { first.push(message(["page one, first", "page one, last"], { truncated: true })); first.close() })
        await waitFor(() => expect(result.current.result?.moreNewer).toBe(true))
        // Not Live: the server sent its page and ended the stream.
        expect(result.current.status).toBe(ConnStatus.DISCONNECTED)
        await settle()
        expect(opened).toHaveLength(1)

        act(() => result.current.loadNewer())
        const second = await nthStream(opened, 1)
        expect(result.current.result?.moreNewer).toBe(false)
        expect(cursorIn(opened[1].request.from)).toEqual(cursorOf("page one, last"))
        expect(opened[1].request.newest).toBe(false)
        expect(opened[1].request.to).toBeUndefined()

        await act(async () => second.push(message(["page two"])))
        await waitFor(() => expect(result.current.status).toBe(ConnStatus.CONNECTED))
        await act(async () => second.push(message(["live"])))
        expect(result.current.result?.entries.map(e => e.message)).toEqual(["page one, first", "page one, last", "page two", "live"])
    })

    test("a range with To ends after its page, without reconnecting or reporting Live", async () => {
        const { result, opened } = setup(dates(new Date(2026, 8, 22, 16, 0)))
        const first = await nthStream(opened, 0)
        await act(async () => { first.push(message(["the whole range"])); first.close() })
        await waitFor(() => expect(result.current.loading).toBe(false))
        await settle()
        expect(opened).toHaveLength(1)
        expect(result.current.status).toBe(ConnStatus.DISCONNECTED)
        expect(result.current.result?.moreNewer).toBe(false)
    })

    test("a range with To cut short pages forward up to To", async () => {
        const to = new Date(2026, 8, 22, 16, 0)
        const { result, opened } = setup(dates(to))
        const first = await nthStream(opened, 0)
        await act(async () => { first.push(message(["first page"], { truncated: true })); first.close() })
        await waitFor(() => expect(result.current.result?.moreNewer).toBe(true))

        act(() => result.current.loadNewer())
        await nthStream(opened, 1)
        expect(cursorIn(opened[1].request.from)).toEqual(cursorOf("first page"))
        expect(timeOf(opened[1].request.to)).toBe(to.getTime())
    })

    test("closes the stream when the page goes away", async () => {
        const { unmount, opened } = setup()
        await nthStream(opened, 0)
        unmount()
        // A second stream would mean the abort did not reach the iterator.
        await new Promise(resolve => setTimeout(resolve, 50))
        expect(opened).toHaveLength(1)
    })

    test("resumes after the newest entry's cursor, not its timestamp", async () => {
        const { result, opened } = setup()
        const first = await nthStream(opened, 0)
        // Two entries at one instant: a timestamp could not resume between
        // them, and one stamped earlier may yet be written after both.
        const tie = create(CourseLogSchema, {
            entries: [
                { message: "first of a tie", time: { seconds: 1_700_000_000n, nanos: 5 }, cursor: cursorOf("first of a tie") },
                { message: "second of a tie", time: { seconds: 1_700_000_000n, nanos: 5 }, cursor: cursorOf("second of a tie") },
            ],
        })
        await act(async () => first.push(tie))
        await waitFor(() => expect(result.current.loading).toBe(false))
        // The stream's report that it fell behind is no entry of the log, and
        // carries no cursor; it must not reset the resume point.
        const gap = create(CourseLogSchema, {
            entries: [{ message: "course log stream fell behind", time: { seconds: 1_700_000_001n, nanos: 0 } }],
        })
        await act(async () => first.push(gap))
        await act(async () => first.fail(new ConnectError("connection reset", Code.Unavailable)))

        await nthStream(opened, 1, 3000)
        expect(cursorIn(opened[1].request.from)).toEqual(cursorOf("second of a tie"))
    })

    test("a second loadOlder while the first is in flight asks for nothing", async () => {
        const { result, opened } = setup()
        const stream = await nthStream(opened, 0)
        await act(async () => stream.push(message(["oldest loaded"], { truncated: true })))
        await waitFor(() => expect(result.current.loading).toBe(false))

        act(() => { void result.current.loadOlder(); void result.current.loadOlder() })
        const older = await nthStream(opened, 1)
        await act(async () => { older.push(message(["earlier"])); older.close() })
        await waitFor(() => expect(result.current.result?.entries).toHaveLength(2))
        // A second request would have prepended the same page twice.
        expect(opened).toHaveLength(2)
        expect(result.current.result?.entries.map(e => e.message)).toEqual(["earlier", "oldest loaded"])
    })

    test("a message from an abandoned stream cannot report Live or move the resume point", async () => {
        const { result, opened } = setup()
        const first = await nthStream(opened, 0)
        await act(async () => first.push(message(["backlog"])))
        await waitFor(() => expect(result.current.loading).toBe(false))

        // Refresh abandons the first stream and opens a second, which has not
        // said anything yet.
        act(() => result.current.refresh({ ...preset, level: CourseLogEntry_Level.ERROR }))
        const second = await nthStream(opened, 1)
        expect(result.current.status).not.toBe(ConnStatus.CONNECTED)

        // The abandoned stream yields a message it had already buffered, from
        // far ahead of anything the new stream has seen.
        await act(async () => first.push(entryAt("from the abandoned stream", 1_900_000_000n, 0)))
        expect(result.current.status).not.toBe(ConnStatus.CONNECTED)

        // Drop the new stream and see where its reconnect resumes.
        await act(async () => second.fail(new ConnectError("connection reset", Code.Unavailable)))
        await nthStream(opened, 2, 3000)
        // Seeded from the stale entry, the reconnect would resume after it,
        // and every entry the new stream has not yet sent would be skipped.
        expect(cursorIn(opened[2].request.from)).toBeNull()
        expect(timeOf(opened[2].request.from)).toBe(timeOf(opened[1].request.from))
    })

    test("a loadOlder that lands after a refresh leaves the new view alone", async () => {
        const { result, opened } = setup()
        const first = await nthStream(opened, 0)
        await act(async () => first.push(message(["backlog"], { truncated: true })))
        await waitFor(() => expect(result.current.loading).toBe(false))

        // A bounded request is in flight when the query changes under it.
        act(() => void result.current.loadOlder())
        const older = await nthStream(opened, 1)
        act(() => result.current.refresh({ ...preset, level: CourseLogEntry_Level.ERROR }))
        const second = await nthStream(opened, 2)
        await act(async () => { second.push(message(["fresh"])) })
        await act(async () => { older.push(message(["older"])); older.close() })

        expect(result.current.result?.entries.map(e => e.message)).toEqual(["fresh"])
    })

    test("loadOlder rejects on failure, leaving the loaded entries alone", async () => {
        const { result, opened } = setup()
        const stream = await nthStream(opened, 0)
        await act(async () => stream.push(message(["oldest loaded"], { truncated: true })))
        await waitFor(() => expect(result.current.loading).toBe(false))

        let rejected: unknown = null
        await act(async () => {
            const pending = result.current.loadOlder().catch((err: unknown) => { rejected = err })
            const older = await nthStream(opened, 1)
            older.fail(new ConnectError("server said no", Code.Unavailable))
            await pending
        })
        expect(ConnectError.from(rejected).message).toContain("server said no")
        // The live stream and what it loaded are untouched.
        expect(result.current.result?.entries.map(e => e.message)).toEqual(["oldest loaded"])
        expect(result.current.result?.moreOlder).toBe(true)
        expect(result.current.error).toBeNull()
    })
})

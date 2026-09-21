import { create } from "@bufbuild/protobuf"
import { timestampDate, timestampFromDate, TimestampSchema } from "@bufbuild/protobuf/wkt"
import { Code, ConnectError } from "@connectrpc/connect"
import { act, renderHook, waitFor } from "@testing-library/react"
import { Provider } from "overmind-react"
import type { ReactNode } from "react"
import type { CourseLog } from "../../proto/qf/requests_pb"
import { CourseLogEntry_Level, CourseLogSchema } from "../../proto/qf/requests_pb"
import { ConnStatus } from "../Helpers"
import { useCourseLogStream } from "../hooks/useCourseLogStream"
import { ApiClient } from "../overmind/namespaces/global/effects"
import { initializeOvermind, mock } from "./TestHelpers"
import { makeStream, type TestStream } from "./testStream"

const HOUR = 60 * 60 * 1000
const filters = { window: 2 * HOUR, repository: "", level: CourseLogEntry_Level.DEBUG }

const message = (messages: string[], extra: { repositories?: string[], truncated?: boolean } = {}) => create(CourseLogSchema, {
    entries: messages.map((text, i) => ({ message: text, time: timestampFromDate(new Date(1_700_000_000_000 + i * 1000)) })),
    ...extra,
})

/** entryAt builds a one-entry message at an exact instant, for the boundaries
 *  a whole millisecond cannot express. */
const entryAt = (text: string, seconds: bigint, nanos: number) => create(CourseLogSchema, {
    entries: [{ message: text, time: { seconds, nanos } }],
})

type Request = Parameters<ApiClient["client"]["courseLogStream"]>[0]

// The request type carries timestamp *initializers*, not built messages, so
// reading one back needs a round trip through create.
const bound = (ts: Request["from"]) => create(TimestampSchema, ts)
// Rounds to the nearest millisecond, so only for assertions coarser than one.
const millis = (ts: Request["from"]): number => timestampDate(bound(ts)).getTime()
// The instant one nanosecond before ts, which is where a bounded request for
// everything older than ts has to end.
const justBefore = (ts: Request["from"]) => {
    const t = bound(ts)
    return t.nanos > 0
        ? { seconds: t.seconds, nanos: t.nanos - 1 }
        : { seconds: t.seconds - 1n, nanos: 999_999_999 }
}

const setup = () => {
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

    test("asks for the chosen window and leaves To unset so the stream tails", async () => {
        const { opened } = setup()
        await nthStream(opened, 0)
        const { request } = opened[0]
        expect(request.to).toBeUndefined()
        const from = millis(request.from)
        expect(Date.now() - from).toBeGreaterThanOrEqual(2 * HOUR - 5000)
        expect(Date.now() - from).toBeLessThanOrEqual(2 * HOUR + 5000)
    })

    test("refresh reopens the stream with the new filters and drops what was loaded", async () => {
        const { result, opened } = setup()
        const first = await nthStream(opened, 0)
        await act(async () => first.push(message(["old"])))
        await waitFor(() => expect(result.current.loading).toBe(false))

        act(() => result.current.refresh({ ...filters, repository: "student-a", level: CourseLogEntry_Level.ERROR }))
        expect(result.current.result).toBeNull()
        const second = await nthStream(opened, 1)
        expect(opened[1].request).toMatchObject({ repository: "student-a", level: CourseLogEntry_Level.ERROR })
        await act(async () => second.push(message(["new"])))
        expect(result.current.result?.entries.map(e => e.message)).toEqual(["new"])
    })

    test("a course change reopens the stream and clears the repository filter", async () => {
        const { result, rerender, opened } = setup()
        const first = await nthStream(opened, 0)
        act(() => result.current.refresh({ ...filters, repository: "student-a", level: CourseLogEntry_Level.ERROR }))
        await nthStream(opened, 1)
        first.close()

        rerender({ courseID: 2n })
        await nthStream(opened, 2)
        expect(opened[2].request).toMatchObject({ courseID: 2n, repository: "", level: CourseLogEntry_Level.ERROR })
        expect(result.current.result).toBeNull()
    })

    test("reports a dropped stream and reconnects after the newest entry it has", async () => {
        const { result, opened } = setup()
        const first = await nthStream(opened, 0)
        await act(async () => first.push(message(["backlog"])))
        await waitFor(() => expect(result.current.loading).toBe(false))

        await act(async () => first.fail(new ConnectError("connection reset", Code.Unavailable)))
        await waitFor(() => expect(result.current.status).toBe(ConnStatus.RECONNECTING))
        expect(result.current.error).toContain("connection reset")

        const second = await nthStream(opened, 1, 3000)
        // The reconnect resumes just after the newest entry already on screen,
        // so the fresh backlog cannot repeat it.
        expect(bound(opened[1].request.from)).toMatchObject({ seconds: 1_700_000_000n, nanos: 1 })
        await act(async () => second.push(message(["after reconnect"])))
        await waitFor(() => expect(result.current.status).toBe(ConnStatus.CONNECTED))
        // The entries already on screen are kept across the reconnect.
        expect(result.current.result?.entries.map(e => e.message)).toEqual(["backlog", "after reconnect"])
        expect(result.current.error).toBeNull()
    })

    test("stops rather than retrying when the request is rejected", async () => {
        const { result, opened } = setup()
        const first = await nthStream(opened, 0)
        await act(async () => first.fail(new ConnectError("not teacher", Code.PermissionDenied)))
        await waitFor(() => expect(result.current.error).toContain("not teacher"))
        expect(result.current.status).toBe(ConnStatus.DISCONNECTED)
        // Give a retry every chance to happen before asserting that none did.
        await act(async () => { await new Promise(resolve => setTimeout(resolve, 50)) })
        expect(opened).toHaveLength(1)
    })

    test("loadOlder prepends the window before the one already on screen", async () => {
        const { result, opened } = setup()
        const stream = await nthStream(opened, 0)
        await act(async () => stream.push(message(["oldest loaded", "newer"])))
        await waitFor(() => expect(result.current.loading).toBe(false))

        act(() => void result.current.loadOlder())
        const older = await nthStream(opened, 1)
        const start = opened[0].request.from
        // The older window ends where the view begins, not at whichever entry
        // happened to be oldest. The server's bounds are inclusive, so a
        // nanosecond before is what excludes that instant and nothing else.
        expect(bound(opened[1].request.to)).toMatchObject(justBefore(start))
        expect(millis(opened[1].request.from)).toBe(millis(start) - 1 - 2 * HOUR)

        await act(async () => { older.push(message(["earlier"])); older.close() })
        await waitFor(() => expect(result.current.result?.entries).toHaveLength(3))
        expect(result.current.result?.entries[0].message).toBe("earlier")
    })

    test("loadOlder steps past a window that held nothing", async () => {
        const { result, opened } = setup()
        const stream = await nthStream(opened, 0)
        await act(async () => { stream.push(message([])) })
        await waitFor(() => expect(result.current.loading).toBe(false))

        // Nothing is loaded, so there is no oldest entry to page back from;
        // the window the teacher asked for is the boundary all the same.
        act(() => void result.current.loadOlder())
        const first = await nthStream(opened, 1)
        await act(async () => { first.push(message([])); first.close() })

        act(() => void result.current.loadOlder())
        const second = await nthStream(opened, 2)
        await act(async () => { second.push(message([])); second.close() })

        // Each click asks for the window before the last one asked for, so a
        // quiet stretch is crossed rather than asked about again for ever.
        expect(millis(opened[1].request.from)).toBe(millis(opened[0].request.from) - 1 - 2 * HOUR)
        expect(millis(opened[2].request.from)).toBe(millis(opened[1].request.from) - 1 - 2 * HOUR)
        expect(bound(opened[2].request.to)).toMatchObject(justBefore(opened[1].request.from))
    })

    test("closes the stream when the page goes away", async () => {
        const { unmount, opened } = setup()
        await nthStream(opened, 0)
        unmount()
        // A second stream would mean the abort did not reach the iterator.
        await new Promise(resolve => setTimeout(resolve, 50))
        expect(opened).toHaveLength(1)
    })

    test("resumes at the exact instant after the newest entry, not the next millisecond", async () => {
        const { result, opened } = setup()
        const first = await nthStream(opened, 0)
        // Two entries within one millisecond, as a burst of CI records are.
        await act(async () => first.push(entryAt("mid-millisecond", 1_700_000_000n, 123_456_789)))
        await waitFor(() => expect(result.current.loading).toBe(false))
        await act(async () => first.fail(new ConnectError("connection reset", Code.Unavailable)))

        await nthStream(opened, 1, 3000)
        // Rounding up to 1_700_000_000_124 ms would step past anything logged
        // in the rest of that millisecond, losing it for good.
        expect(bound(opened[1].request.from)).toMatchObject({ seconds: 1_700_000_000n, nanos: 123_456_790 })
    })

    test("loadOlder reports that the older window was cut off by the limit", async () => {
        const { result, opened } = setup()
        const stream = await nthStream(opened, 0)
        await act(async () => stream.push(message(["oldest loaded"])))
        await waitFor(() => expect(result.current.loading).toBe(false))
        expect(result.current.result?.truncated).toBe(false)

        act(() => void result.current.loadOlder())
        const older = await nthStream(opened, 1)
        await act(async () => { older.push(message(["earlier"], { truncated: true })); older.close() })
        await waitFor(() => expect(result.current.result?.truncated).toBe(true))
    })

    test("a second loadOlder while the first is in flight asks for nothing", async () => {
        const { result, opened } = setup()
        const stream = await nthStream(opened, 0)
        await act(async () => stream.push(message(["oldest loaded"])))
        await waitFor(() => expect(result.current.loading).toBe(false))

        act(() => { void result.current.loadOlder(); void result.current.loadOlder() })
        const older = await nthStream(opened, 1)
        await act(async () => { older.push(message(["earlier"])); older.close() })
        await waitFor(() => expect(result.current.result?.entries).toHaveLength(2))
        // A second request would have prepended the same window twice.
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
        act(() => result.current.refresh({ ...filters, level: CourseLogEntry_Level.ERROR }))
        const second = await nthStream(opened, 1)
        expect(result.current.status).not.toBe(ConnStatus.CONNECTED)

        // The abandoned stream yields a message it had already buffered, from
        // far ahead of anything the new stream has seen.
        await act(async () => first.push(entryAt("from the abandoned stream", 1_900_000_000n, 0)))
        expect(result.current.status).not.toBe(ConnStatus.CONNECTED)

        // Drop the new stream and see where its reconnect resumes.
        await act(async () => second.fail(new ConnectError("connection reset", Code.Unavailable)))
        await nthStream(opened, 2, 3000)
        // Seeded from the stale entry this would be 1_900_000_000.000000001,
        // and every entry logged since would be skipped for good.
        expect(millis(opened[2].request.from)).toBe(millis(opened[1].request.from))
    })

    test("a loadOlder that lands after a refresh leaves the new window's boundary alone", async () => {
        const { result, opened } = setup()
        const first = await nthStream(opened, 0)
        await act(async () => first.push(message(["backlog"])))
        await waitFor(() => expect(result.current.loading).toBe(false))

        // A bounded request is in flight when the query changes under it.
        act(() => void result.current.loadOlder())
        const older = await nthStream(opened, 1)
        act(() => result.current.refresh({ ...filters, level: CourseLogEntry_Level.ERROR }))
        const second = await nthStream(opened, 2)
        await act(async () => { older.push(message(["older"])); older.close() })
        await act(async () => { second.push(message(["fresh"])) })

        // Load older now pages back from the window on screen. Had the stale
        // request moved the boundary, this would ask for a window two hours
        // further back again, skipping one the teacher never saw.
        act(() => void result.current.loadOlder())
        await nthStream(opened, 3)
        expect(millis(opened[3].request.from)).toBe(millis(opened[2].request.from) - 1 - 2 * HOUR)
    })

    test("loadOlder rejects on failure, leaving the loaded entries alone", async () => {
        const { result, opened } = setup()
        const stream = await nthStream(opened, 0)
        await act(async () => stream.push(message(["oldest loaded"])))
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
        expect(result.current.error).toBeNull()
    })
})

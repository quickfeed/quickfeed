import { create } from "@bufbuild/protobuf"
import type { Timestamp } from "@bufbuild/protobuf/wkt"
import { TimestampSchema, timestampFromDate } from "@bufbuild/protobuf/wkt"
import { Code, ConnectError } from "@connectrpc/connect"
import { useEffect, useRef, useState } from "react"
import type { CourseLog, CourseLogEntry, CourseLogEntry_Level } from "../../proto/qf/requests_pb"
import { ConnStatus } from "../Helpers"
import { useGrpc } from "../overmind"

export interface CourseLogFilters {
    /** How far back the log should reach, in milliseconds. */
    window: number
    repository: string
    level: CourseLogEntry_Level
}

interface CourseLogQuery {
    courseID: bigint
    filters: CourseLogFilters
}

interface CourseLogResult {
    entries: CourseLogEntry[]
    repositories: string[]
    truncated: boolean
}

/** A result and the query it answers, so a result outlives its query by
 *  exactly nothing: changing the query hides the old entries at once. */
interface CourseLogResponse {
    query: CourseLogQuery
    result: CourseLogResult | null
    error: string | null
}

const EMPTY: CourseLogResult = { entries: [], repositories: [], truncated: false }

// RECONNECT_DELAY is how long to wait before reopening a dropped stream, and
// the step by which that wait grows; the log view is a background window, so
// it backs off rather than hammering a server that is down.
const RECONNECT_DELAY = 1000
const MAX_RECONNECT_DELAY = 30 * 1000

const NANOS_PER_SECOND = 1_000_000_000

/** shift moves a bound by one nanosecond, the finest step the log records.
 *  The server's interval is inclusive at both ends, so a nanosecond is what
 *  excludes exactly one entry and nothing else. A millisecond, which is all a
 *  Date can hold, is too coarse: a burst can log several entries within one,
 *  and a bound stepped past them would skip them for good. */
const shift = (ts: Timestamp, step: 1 | -1): Timestamp => {
    const nanos = ts.nanos + step
    if (nanos < 0) {
        return create(TimestampSchema, { seconds: ts.seconds - 1n, nanos: NANOS_PER_SECOND - 1 })
    }
    if (nanos >= NANOS_PER_SECOND) {
        return create(TimestampSchema, { seconds: ts.seconds + 1n, nanos: 0 })
    }
    return create(TimestampSchema, { seconds: ts.seconds, nanos })
}

/** floorMs is ts as whole milliseconds, rounded down. The shared helper rounds
 *  to nearest, which would move a bound past an entry rather than short of it. */
const floorMs = (ts: Timestamp): number =>
    Number(ts.seconds) * 1000 + Math.floor(ts.nanos / 1_000_000)

/** newest returns the later of two timestamps, either of which may be absent. */
const newest = (a: Timestamp | null, b?: Timestamp): Timestamp | null => {
    if (!b) {
        return a
    }
    if (!a || b.seconds > a.seconds || (b.seconds === a.seconds && b.nanos > a.nanos)) {
        return b
    }
    return a
}

const union = (previous: string[], added: string[]): string[] =>
    added.length === 0 ? previous : [...new Set([...previous, ...added])].sort((a, b) => a.localeCompare(b))

/** append adds a stream message to the end of what is already on screen. */
const append = (previous: CourseLogResult, message: CourseLog): CourseLogResult => ({
    entries: previous.entries.concat(message.entries),
    repositories: union(previous.repositories, message.repositories),
    // Only the backlog can be cut off by the limit; a live message never is.
    truncated: previous.truncated || message.truncated,
})

/** prepend adds an older window in front of what is already on screen. The
 *  older window is a bounded query like the backlog, so it can hit the limit
 *  too; the warning must stand for entries missing from either end. */
const prepend = (previous: CourseLogResult, message: CourseLog): CourseLogResult => ({
    entries: message.entries.concat(previous.entries),
    repositories: union(previous.repositories, message.repositories),
    truncated: previous.truncated || message.truncated,
})

/**
 * useCourseLogStream keeps a live view of one course's log. It opens a
 * CourseLogStream for the current query, seeds the view from the backlog the
 * first message carries, and appends each later entry as it is logged. The
 * stream is closed when the query changes, the course changes, or the page
 * unmounts.
 */
export const useCourseLogStream = (courseID: bigint, initialFilters: CourseLogFilters) => {
    const { api } = useGrpc().global
    const [query, setQuery] = useState<CourseLogQuery>(() => ({ courseID, filters: initialFilters }))
    const [response, setResponse] = useState<CourseLogResponse | null>(null)
    const [status, setStatus] = useState(ConnStatus.DISCONNECTED)
    // Held in a ref so a reconnect resumes after the newest entry on screen
    // without making the effect depend on every entry that arrives.
    const latest = useRef<Timestamp | null>(null)
    // Held in a ref so a second Load older click while the first is in flight
    // is a no-op, rather than asking for the same window twice and prepending
    // it twice.
    const loadingOlder = useRef(false)
    // How far back the view has asked for, which is where Load older picks up.
    // The oldest entry on screen cannot serve: a window that held nothing
    // leaves it where it was, so every later click would ask for that same
    // empty window again, and an empty view would have no boundary at all.
    const oldestBound = useRef<Timestamp | null>(null)
    // The query the running stream belongs to, so a bounded request that
    // finishes after the query changed can tell that the boundary it would
    // move is no longer the one it measured.
    const active = useRef<CourseLogQuery | null>(null)

    if (query.courseID !== courseID) {
        setQuery({ courseID, filters: { ...query.filters, repository: "" } })
    }

    // Matching the query to its answer hides the previous course's entries as
    // soon as the query changes, before the new stream has said anything.
    const current = response?.query === query ? response : null

    useEffect(() => {
        const controller = new AbortController()
        latest.current = null
        // Where the chosen window begins, read once: a reconnect before the
        // first entry arrives then resumes from where the view began rather
        // than sliding the window forward past what it has not seen.
        const start = timestampFromDate(new Date(Date.now() - query.filters.window))
        oldestBound.current = start
        active.current = query

        // A reconnect asks for entries after the newest one on screen, so the
        // fresh backlog does not repeat what the teacher already has.
        const from = (): Timestamp => latest.current ? shift(latest.current, 1) : start
        // A message that arrives after the stream was abandoned belongs to a
        // query no longer on screen; writing it would hide the new one's
        // answer behind a stale response.
        const update = (change: (previous: CourseLogResponse | null) => CourseLogResponse) => {
            if (controller.signal.aborted) {
                return
            }
            setResponse(previous => change(previous?.query === query ? previous : null))
        }

        const run = async () => {
            for (let delay = RECONNECT_DELAY; !controller.signal.aborted;) {
                try {
                    const stream = api.client.courseLogStream({
                        courseID: query.courseID,
                        from: from(),
                        repository: query.filters.repository,
                        level: query.filters.level,
                    }, { signal: controller.signal })
                    for await (const message of stream) {
                        // An abandoned stream can still yield a message it had
                        // already buffered. It answers a query no longer on
                        // screen: reporting it Live would mislabel the new
                        // connection, and letting its entries through to latest
                        // would send the new stream's reconnect past entries it
                        // has never seen. The update below guards the response;
                        // these two have to be guarded here.
                        if (controller.signal.aborted) {
                            return
                        }
                        delay = RECONNECT_DELAY
                        setStatus(ConnStatus.CONNECTED)
                        for (const entry of message.entries) {
                            latest.current = newest(latest.current, entry.time)
                        }
                        update(previous => ({ query, result: append(previous?.result ?? EMPTY, message), error: null }))
                    }
                } catch (err) {
                    const error = ConnectError.from(err)
                    if (controller.signal.aborted || error.code === Code.Canceled) {
                        return
                    }
                    update(previous => ({ query, result: previous?.result ?? null, error: error.message }))
                    // A rejected request will be rejected again; only report it.
                    if (error.code === Code.PermissionDenied || error.code === Code.InvalidArgument) {
                        setStatus(ConnStatus.DISCONNECTED)
                        return
                    }
                }
                if (controller.signal.aborted) {
                    return
                }
                setStatus(ConnStatus.RECONNECTING)
                await new Promise(resolve => setTimeout(resolve, delay))
                delay = Math.min(delay * 2, MAX_RECONNECT_DELAY)
            }
        }
        void run()
        return () => {
            controller.abort()
            setStatus(ConnStatus.DISCONNECTED)
        }
    }, [api, query])

    /** refresh reopens the stream with new filters, discarding what is loaded. */
    const refresh = (filters: CourseLogFilters) => setQuery({ courseID, filters: { ...filters } })

    /**
     * loadOlder fetches the window before the one the view already covers, for
     * a teacher who started watching too late. It is a bounded request, so it
     * ends on its own without disturbing the live stream. It rejects if the
     * request fails, leaving the loaded entries as they were, so the caller can
     * say so without tearing down the live view.
     */
    const loadOlder = async () => {
        const bound = oldestBound.current
        if (!bound || loadingOlder.current) {
            return
        }
        loadingOlder.current = true
        try {
            const to = shift(bound, -1)
            const from = timestampFromDate(new Date(floorMs(to) - query.filters.window))
            const stream = api.client.courseLogStream({
                courseID: query.courseID,
                from,
                to,
                repository: query.filters.repository,
                level: query.filters.level,
            })
            for await (const message of stream) {
                setResponse(previous => previous?.query === query
                    ? { ...previous, result: prepend(previous.result ?? EMPTY, message) }
                    : previous)
            }
            // Moved once the window has been read and whether or not it held
            // anything, so that a quiet stretch is stepped over rather than
            // asked for again. A failed request throws before this, leaving the
            // boundary where it was so the same window can be retried. A
            // Refresh or a course change while this was in flight has already
            // set the boundary for the window now on screen, and this request
            // measured the one before it, so it has nothing left to say.
            if (active.current === query) {
                oldestBound.current = from
            }
        } finally {
            loadingOlder.current = false
        }
    }

    return {
        result: current?.result ?? null,
        error: current?.error ?? null,
        status,
        loading: current === null,
        refresh,
        loadOlder,
    }
}

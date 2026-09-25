import { create } from "@bufbuild/protobuf"
import { timestampFromDate } from "@bufbuild/protobuf/wkt"
import { Code, ConnectError } from "@connectrpc/connect"
import { useEffect, useRef, useState } from "react"
import type { CourseLogEntry, CourseLogEntry_Level, LogCursor, LogPosition } from "../../proto/qf/requests_pb"
import { LogPositionSchema } from "../../proto/qf/requests_pb"
import { ConnStatus } from "../Helpers"
import { useGrpc } from "../overmind"

/** CourseLogRange is the part of a course's log the view covers. A preset is
 *  the last `duration` milliseconds up to now: the view starts at its newest
 *  end and follows new entries. Dates start the view at `from`, and it follows
 *  new entries once it has caught up, unless `to` ends the range. */
export type CourseLogRange =
    | { kind: "preset", duration: number }
    | { kind: "dates", from: Date, to: Date | null }

export interface CourseLogFilters {
    range: CourseLogRange
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
    /** The limit left out entries before the oldest one on screen. */
    moreOlder: boolean
    /** The limit left out entries after the newest one on screen. */
    moreNewer: boolean
}

/** CourseLogResponse pairs a result with the query it answers, so changing
 *  the query hides the old entries at once. */
interface CourseLogResponse {
    query: CourseLogQuery
    result: CourseLogResult | null
    error: string | null
}

const EMPTY: CourseLogResult = { entries: [], repositories: [], moreOlder: false, moreNewer: false }

// RECONNECT_DELAY is the initial wait before reopening a dropped stream; the
// wait doubles on each retry, up to MAX_RECONNECT_DELAY.
const RECONNECT_DELAY = 1000
const MAX_RECONNECT_DELAY = 30 * 1000

const timePosition = (date: Date): LogPosition =>
    create(LogPositionSchema, { position: { case: "time", value: timestampFromDate(date) } })

const cursorPosition = (cursor: LogCursor): LogPosition =>
    create(LogPositionSchema, { position: { case: "cursor", value: cursor } })

const union = (previous: string[], added: string[]): string[] =>
    added.length === 0 ? previous : [...new Set([...previous, ...added])].sort((a, b) => a.localeCompare(b))

/**
 * useCourseLogStream keeps a view of one course's log. It opens a
 * CourseLogStream for the current query and grows the view as the stream
 * answers: forward from its newest entry as new ones are logged or loaded,
 * and back from its oldest on Load older. The stream is closed when the query
 * changes, the course changes, or the page unmounts.
 */
export const useCourseLogStream = (courseID: bigint, initialFilters: CourseLogFilters) => {
    const { api } = useGrpc().global
    const [query, setQuery] = useState<CourseLogQuery>(() => ({ courseID, filters: initialFilters }))
    const [response, setResponse] = useState<CourseLogResponse | null>(null)
    const [status, setStatus] = useState(ConnStatus.DISCONNECTED)
    // Where the range begins, fixed when the query is applied, so that Load
    // older asks for the range the view began with.
    const start = useRef<LogPosition | null>(null)
    // Makes a Load older click while one is in flight a no-op.
    const loadingOlder = useRef(false)
    // Set while the stream waits for Load newer; calling it asks for the next
    // page forward.
    const resume = useRef<(() => void) | null>(null)

    if (query.courseID !== courseID) {
        setQuery({ courseID, filters: { ...query.filters, repository: "" } })
    }

    // Matching the query to its answer hides the previous course's entries as
    // soon as the query changes, before the new stream has said anything.
    const current = response?.query === query ? response : null

    useEffect(() => {
        const controller = new AbortController()
        const { range, repository, level } = query.filters
        const from = timePosition(range.kind === "preset" ? new Date(Date.now() - range.duration) : range.from)
        const to = range.kind === "dates" && range.to ? timePosition(range.to) : undefined
        start.current = from
        // The cursor of the newest entry on screen, which the next request
        // continues from.
        let latest: LogCursor | undefined
        // Ignore messages for a stream that was abandoned.
        const update = (change: (previous: CourseLogResponse | null) => CourseLogResponse) => {
            if (controller.signal.aborted) {
                return
            }
            setResponse(previous => change(previous?.query === query ? previous : null))
        }
        const waitForLoadNewer = () => new Promise<void>(resolve => {
            resume.current = resolve
            controller.signal.addEventListener("abort", () => resolve())
        })

        const run = async () => {
            for (let delay = RECONNECT_DELAY; !controller.signal.aborted;) {
                // A preset starts at the newest end of its range; everything
                // after that, and every range of dates, reads forward from the
                // newest entry on screen, or from the start of the range.
                const newest = range.kind === "preset" && latest === undefined
                let moreNewer = false
                try {
                    const stream = api.client.courseLogStream({
                        courseID: query.courseID,
                        from: latest ? cursorPosition(latest) : from,
                        to,
                        newest,
                        repository,
                        level,
                    }, { signal: controller.signal })
                    let first = true
                    for await (const message of stream) {
                        // An abandoned stream can still yield a buffered
                        // message; it must not touch status or latest.
                        if (controller.signal.aborted) {
                            return
                        }
                        delay = RECONNECT_DELAY
                        // Only the first message is a backlog the limit can cut
                        // short. A stream with To set, or cut short going
                        // forward, ends after it, so it is not Live.
                        const moreOlder = first && newest && message.truncated
                        moreNewer = first && !newest && message.truncated
                        if (!to && !moreNewer) {
                            setStatus(ConnStatus.CONNECTED)
                        }
                        // Entries arrive in the order written; gap reports
                        // have no cursor.
                        for (const entry of message.entries) {
                            if (entry.cursor) {
                                latest = entry.cursor
                            }
                        }
                        const isFirst = first
                        update(previous => {
                            const result = previous?.result ?? EMPTY
                            return {
                                query,
                                error: null,
                                result: {
                                    entries: result.entries.concat(message.entries),
                                    repositories: union(result.repositories, message.repositories),
                                    moreOlder: result.moreOlder || moreOlder,
                                    moreNewer: isFirst ? moreNewer : result.moreNewer,
                                },
                            }
                        })
                        first = false
                    }
                    if (controller.signal.aborted) {
                        return
                    }
                    // A stream that ends without following has sent its page:
                    // either the range is done, or it waits for Load newer.
                    // Any other end is the server going away.
                    if (to || moreNewer) {
                        setStatus(ConnStatus.DISCONNECTED)
                        if (!moreNewer) {
                            return
                        }
                        await waitForLoadNewer()
                        continue
                    }
                } catch (err) {
                    // Only the page's own abort ends the stream quietly; any
                    // other cancel, such as from a proxy, is a dropped stream.
                    if (controller.signal.aborted) {
                        return
                    }
                    const error = ConnectError.from(err)
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
            resume.current = null
            setStatus(ConnStatus.DISCONNECTED)
        }
    }, [api, query])

    /** refresh reopens the stream with new filters, discarding what is loaded. */
    const refresh = (filters: CourseLogFilters) => setQuery({ courseID, filters: { ...filters } })

    /** loadNewer asks for the next page after the newest entry on screen. Once
     *  a page reaches the present, the stream follows new entries again. */
    const loadNewer = () => {
        const next = resume.current
        resume.current = null
        if (!next) {
            return
        }
        setResponse(previous => previous?.query === query && previous.result
            ? { ...previous, result: { ...previous.result, moreNewer: false } }
            : previous)
        next()
    }

    /** loadOlder prepends the page before the oldest entry on screen, without
     *  disturbing the live stream. It rejects if the request fails, leaving
     *  the loaded entries as they were. */
    const loadOlder = async () => {
        const oldest = current?.result?.entries.find(entry => entry.cursor)?.cursor
        if (!oldest || !start.current || loadingOlder.current) {
            return
        }
        loadingOlder.current = true
        try {
            const stream = api.client.courseLogStream({
                courseID: query.courseID,
                from: start.current,
                to: cursorPosition(oldest),
                newest: true,
                repository: query.filters.repository,
                level: query.filters.level,
            })
            for await (const message of stream) {
                setResponse(previous => previous?.query === query && previous.result
                    ? {
                        ...previous,
                        result: {
                            ...previous.result,
                            entries: message.entries.concat(previous.result.entries),
                            repositories: union(previous.result.repositories, message.repositories),
                            moreOlder: message.truncated,
                        },
                    }
                    : previous)
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
        loadNewer,
    }
}

import { timestampFromDate } from "@bufbuild/protobuf/wkt"
import { useEffect, useState } from "react"
import type { CourseLog, CourseLogEntry_Level } from "../../proto/qf/requests_pb"
import { useGrpc } from "../overmind"

export interface CourseLogFilters {
    from: string
    // An empty To lets the server bound each request by its current time.
    to: string
    repository: string
    level: CourseLogEntry_Level
}

interface CourseLogQuery {
    courseID: bigint
    filters: CourseLogFilters
}

interface CourseLogResponse {
    query: CourseLogQuery
    result: CourseLog | null
    error: string | null
}

export const useCourseLogs = (courseID: bigint, initialFilters: CourseLogFilters) => {
    const { api } = useGrpc().global
    const [query, setQuery] = useState<CourseLogQuery>(() => ({ courseID, filters: initialFilters }))
    const [response, setResponse] = useState<CourseLogResponse | null>(null)

    if (query.courseID !== courseID) {
        setQuery({ courseID, filters: { ...query.filters, repository: "" } })
    }

    useEffect(() => {
        let active = true
        const { from, to, repository, level } = query.filters
        void api.client.getCourseLog({
            courseID: query.courseID,
            from: from ? timestampFromDate(new Date(from)) : undefined,
            to: to ? timestampFromDate(new Date(to)) : undefined,
            repository,
            level,
        }).then(({ message, error }) => {
            if (active) {
                setResponse({ query, result: error ? null : message, error: error?.message ?? null })
            }
        })
        return () => { active = false }
    }, [api, query])

    // Query identity also hides old results immediately, before the effect runs.
    const current = response?.query === query ? response : null
    const refresh = (filters: CourseLogFilters) => setQuery({ courseID, filters: { ...filters } })
    return { result: current?.result ?? null, error: current?.error ?? null, loading: !current, refresh }
}

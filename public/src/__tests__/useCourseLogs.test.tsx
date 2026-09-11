import { create } from "@bufbuild/protobuf"
import { ConnectError } from "@connectrpc/connect"
import { act, renderHook, waitFor } from "@testing-library/react"
import { Provider } from "overmind-react"
import type { ReactNode } from "react"
import { CourseLogEntry_Level, CourseLogSchema } from "../../proto/qf/requests_pb"
import { useCourseLogs } from "../hooks/useCourseLogs"
import { ApiClient } from "../overmind/namespaces/global/effects"
import { initializeOvermind, mock } from "./TestHelpers"

type LogResponse = Awaited<ReturnType<ApiClient["client"]["getCourseLog"]>>
const filters = { from: "2026-03-09T12:00", to: "", repository: "", level: CourseLogEntry_Level.DEBUG }
const response = (message: string): LogResponse => ({
    message: create(CourseLogSchema, { entries: [{ message }] }),
    error: null,
})

const setup = () => {
    const requests: { request: Parameters<ApiClient["client"]["getCourseLog"]>[0]; resolve: (response: LogResponse) => void }[] = []
    const api = new ApiClient()
    api.client = {
        ...api.client,
        getCourseLog: mock("getCourseLog", request => new Promise(resolve => { requests.push({ request, resolve }) })),
    }
    const state = initializeOvermind({}, api)
    const wrapper = ({ children }: { children: ReactNode }) => <Provider value={state}>{children}</Provider>
    return {
        ...renderHook(({ courseID }) => useCourseLogs(courseID, filters), { wrapper, initialProps: { courseID: 1n } }),
        requests,
    }
}

describe("useCourseLogs", () => {
    test("reuses applied filters on course changes and clears the repository", async () => {
        const { result, rerender, requests } = setup()
        const applied = { ...filters, repository: "student-a", level: CourseLogEntry_Level.ERROR }
        act(() => result.current.refresh(applied))
        await act(async () => requests[1].resolve(response("course one")))
        rerender({ courseID: 2n })
        expect(requests[2].request).toMatchObject({ courseID: 2n, repository: "", level: CourseLogEntry_Level.ERROR })
        expect(requests[2].request.from).toEqual(requests[1].request.from)
        expect(requests[2].request.to).toBeUndefined()
        expect(result.current).toMatchObject({ loading: true, result: null, error: null })
        await act(async () => requests[2].resolve(response("course two")))
        expect(result.current.result?.entries[0].message).toBe("course two")
    })

    test.each([false, true])("ignores stale success and error responses (new request completed: %s)", async completed => {
        const { result, rerender, requests } = setup()
        rerender({ courseID: 2n })
        rerender({ courseID: 1n })
        if (completed) {
            await act(async () => requests[2].resolve(response("current course")))
        }
        await act(async () => {
            requests[0].resolve(response("stale course"))
            requests[1].resolve({ ...response(""), error: new ConnectError("stale error") })
        })
        expect(result.current.error).toBeNull()
        expect(result.current.loading).toBe(!completed)
        expect(result.current.result?.entries[0].message ?? null).toBe(completed ? "current course" : null)
    })

    test("ignores an earlier refresh within the same course", async () => {
        const { result, requests } = setup()
        act(() => result.current.refresh({ ...filters, level: CourseLogEntry_Level.ERROR }))
        await act(async () => requests[1].resolve(response("new filter")))
        await act(async () => requests[0].resolve(response("old filter")))
        expect(result.current.result?.entries[0].message).toBe("new filter")
    })

    test("clears a current error when retrying", async () => {
        const { result, requests } = setup()
        await act(async () => requests[0].resolve({ ...response(""), error: new ConnectError("unavailable") }))
        expect(result.current.error).toContain("unavailable")
        act(() => result.current.refresh(filters))
        expect(result.current).toMatchObject({ loading: true, result: null, error: null })
        await act(async () => requests[1].resolve(response("retried")))
        await waitFor(() => expect(result.current.loading).toBe(false))
        expect(result.current.result?.entries[0].message).toBe("retried")
    })
})

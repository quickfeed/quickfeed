import { create } from "@bufbuild/protobuf"
import { timestampFromDate } from "@bufbuild/protobuf/wkt"
import { ConnectError } from "@connectrpc/connect"
import { fireEvent, render, screen, waitFor } from "@testing-library/react"
import { Provider } from "overmind-react"
import { MemoryRouter, Route, Routes } from "react-router"
import { CourseLogEntry_Level, CourseLogEntrySchema, CourseLogSchema } from "../../proto/qf/requests_pb"
import { UserSchema } from "../../proto/qf/types_pb"
import CourseLogs from "../components/teacher/CourseLogs"
import TeacherPage from "../pages/TeacherPage"
import { ApiClient } from "../overmind/namespaces/global/effects"
import { MockData } from "./mock_data/mockData"
import { initializeOvermind, mock } from "./TestHelpers"

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

describe("CourseLogs", () => {
    test("fetches and renders entries on mount", async () => {
        const api = new ApiClient()
        api.client = {
            ...api.client,
            getCourseLog: mock("getCourseLog", async () => ({ // skipcq: JS-0116
                message: create(CourseLogSchema, {
                    entries: [entry({ message: "resolved push repository" }), entry({ message: "cloned assignments repository", repository: "student-b" })],
                    repositories: ["student-a", "student-b"],
                    truncated: false,
                }),
                error: null,
            })),
        }
        renderCourseLogs(api)

        expect(await screen.findByText("resolved push repository")).toBeTruthy()
        expect(screen.getByText("cloned assignments repository")).toBeTruthy()
        // The repository selector lists every repository from the response.
        expect(screen.getByRole("option", { name: "student-a" })).toBeTruthy()
        expect(screen.getByRole("option", { name: "student-b" })).toBeTruthy()
    })

    test("renders the structured fields and source of an entry", async () => {
        const api = new ApiClient()
        api.client = {
            ...api.client,
            getCourseLog: mock("getCourseLog", async () => ({ // skipcq: JS-0116
                message: create(CourseLogSchema, {
                    entries: [entry({
                        message: "test run failed",
                        level: CourseLogEntry_Level.ERROR,
                        source: "ci/run_tests.go:120",
                        fields: { assignment: "lab1", output: "--- FAIL: TestFoo\n    foo_test.go:12: want 1, got 2" },
                    })],
                    repositories: ["student-a"],
                }),
                error: null,
            })),
        }
        renderCourseLogs(api)

        // The CI output a teacher needs lives in fields; it must be on screen,
        // not only in the copied and downloaded text.
        expect(await screen.findByText(/--- FAIL: TestFoo/)).toBeTruthy()
        expect(screen.getByText(/assignment=lab1/)).toBeTruthy()
        expect(screen.getByText("ci/run_tests.go:120")).toBeTruthy()
    })

    test("shows an error state when the request fails", async () => {
        const api = new ApiClient()
        api.client = {
            ...api.client,
            getCourseLog: mock("getCourseLog", async () => ({ // skipcq: JS-0116
                message: create(CourseLogSchema),
                error: new ConnectError("permission denied"),
            })),
        }
        renderCourseLogs(api)

        expect(await screen.findByText(/Failed to load course logs/)).toBeTruthy()
    })

    test("shows an empty state when there are no matching entries", async () => {
        const api = new ApiClient()
        api.client = {
            ...api.client,
            getCourseLog: mock("getCourseLog", async () => ({ // skipcq: JS-0116
                message: create(CourseLogSchema, { entries: [], repositories: [], truncated: false }),
                error: null,
            })),
        }
        renderCourseLogs(api)

        expect(await screen.findByText("No log entries match the current filters")).toBeTruthy()
    })

    test("reports the server-side cutoff, not the locally filtered count", async () => {
        const api = new ApiClient()
        api.client = {
            ...api.client,
            getCourseLog: mock("getCourseLog", async () => ({ // skipcq: JS-0116
                message: create(CourseLogSchema, {
                    entries: [entry({ message: "resolved push repository" }), entry(), entry()],
                    repositories: ["student-a"],
                    truncated: true,
                }),
                error: null,
            })),
        }
        renderCourseLogs(api)

        await waitFor(() => expect(screen.getByText(/Result limited to the newest 3 entries/)).toBeTruthy())

        // Filtering locally changes neither the cutoff nor the count it reports,
        // and the warning must survive a filter that matches nothing.
        const search = screen.getByPlaceholderText("Filter loaded entries")
        fireEvent.keyUp(search, { target: { value: "no such entry" } })

        await waitFor(() => expect(screen.getByText("No log entries match the current filters")).toBeTruthy())
        expect(screen.getByText(/Result limited to the newest 3 entries/)).toBeTruthy()
    })

    test("filters loaded entries locally without refetching", async () => {
        let calls = 0
        const api = new ApiClient()
        api.client = {
            ...api.client,
            getCourseLog: mock("getCourseLog", async () => { // skipcq: JS-0116
                calls++
                return {
                    message: create(CourseLogSchema, {
                        entries: [entry({ message: "resolved push repository" }), entry({ message: "cloned assignments repository" })],
                        repositories: ["student-a"],
                        truncated: false,
                    }),
                    error: null,
                }
            }),
        }
        renderCourseLogs(api)
        await screen.findByText("resolved push repository")
        expect(calls).toBe(1)

        const search = screen.getByPlaceholderText("Filter loaded entries")
        fireEvent.keyUp(search, { target: { value: "cloned" } })

        await waitFor(() => {
            expect(screen.getByText("cloned assignments repository")).toBeTruthy()
            expect(screen.queryByText("resolved push repository")).toBeFalsy()
        })
        // The search filter is applied client-side; it must not trigger another fetch.
        expect(calls).toBe(1)
    })
    test("lets an untouched To follow the clock, and honors an edited one", async () => {
        const requests: { to?: unknown }[] = []
        const api = new ApiClient()
        api.client = {
            ...api.client,
            getCourseLog: mock("getCourseLog", async (request) => { // skipcq: JS-0116
                requests.push(request)
                return {
                    message: create(CourseLogSchema, { entries: [entry()], repositories: ["student-a"] }),
                    error: null,
                }
            }),
        }
        renderCourseLogs(api)
        await screen.findByText("resolved push repository")

        // An untouched To is left out, so the server bounds the interval by its
        // own clock and Refresh picks up whatever was logged since the page loaded.
        expect(requests[0].to).toBeUndefined()

        const toInput = screen.getByLabelText("To")
        fireEvent.change(toInput, { target: { value: "2026-03-10T12:00" } })
        screen.getByRole("button", { name: "Refresh" }).click()

        await waitFor(() => expect(requests).toHaveLength(2))
        expect(requests[1].to).toBeDefined()
    })
})

describe("CourseLogs tile", () => {
    test("navigates from the teacher tile to the logs page", async () => {
        const api = new ApiClient()
        api.client = {
            ...api.client,
            getCourseLog: mock("getCourseLog", async () => ({ // skipcq: JS-0116
                message: create(CourseLogSchema, { entries: [entry()], repositories: ["student-a"], truncated: false }),
                error: null,
            })),
        }
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

import { create } from "@bufbuild/protobuf"
import { render, screen } from "@testing-library/react"
import { Provider } from "overmind-react"
import { MemoryRouter, Route, Routes, useParams } from "react-router"
import type { Course } from "../../proto/qf/types_pb"
import { CourseSchema, UserSchema } from "../../proto/qf/types_pb"
import CourseCodeRedirect from "../components/CourseCodeRedirect"
import { initializeOvermind } from "./TestHelpers"

const self = create(UserSchema, { ID: BigInt(1), Name: "Test User" })

const course = (ID: number, code: string, year: number): Course => create(CourseSchema, {
    ID: BigInt(ID),
    code,
    name: `Course ${code}`,
    year,
})

const CourseStub = () => {
    const { id } = useParams()
    return <div>Course page for {id}</div>
}

const renderRedirect = (path: string, courses: Course[], isLoading: boolean) => {
    const mockedOvermind = initializeOvermind({ self, courses, isLoading })
    render(
        <Provider value={mockedOvermind}>
            <MemoryRouter initialEntries={[path]}>
                <Routes>
                    <Route path="/course/:id" element={<CourseStub />} />
                    <Route path="/:code" element={<CourseCodeRedirect />} />
                    <Route path="*" element={<div>Dashboard</div>} />
                </Routes>
            </MemoryRouter>
        </Provider>
    )
}

describe("CourseCodeRedirect", () => {
    test("waits while user data is loading instead of redirecting to the dashboard", () => {
        // Before getCourses() resolves, state.courses is empty. A redirect at this
        // point would lose the course code, which is the bug in #1564.
        renderRedirect("/dat515", [], true)

        expect(screen.getByText("Loading...")).toBeTruthy()
        expect(screen.queryByText("Dashboard")).toBeNull()
    })

    test("redirects to the most recent course with the given code once loaded", () => {
        renderRedirect("/dat515", [course(1, "DAT515", 2024), course(2, "DAT515", 2025)], false)

        expect(screen.getByText("Course page for 2")).toBeTruthy()
    })

    test("redirects to the dashboard when no course has the given code", () => {
        renderRedirect("/nosuchcourse", [course(1, "DAT515", 2025)], false)

        expect(screen.getByText("Dashboard")).toBeTruthy()
    })
})

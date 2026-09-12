import { create } from "@bufbuild/protobuf"
import { fireEvent, render, screen } from "@testing-library/react"
import { Provider } from "overmind-react"
import { MemoryRouter, Route, Routes } from "react-router"
import { UserSchema } from "../../proto/qf/types_pb"
import TeacherPage from "../pages/TeacherPage"
import { MockData } from "./mock_data/mockData"
import { initializeOvermind } from "./TestHelpers"

// Course 1 of the mocked courses; the teacher below is enrolled in it as a teacher.
const course = MockData.mockedCourses()[0]

const teacherState = {
    self: create(UserSchema, { ID: BigInt(1), Name: "Teacher", IsAdmin: false }),
    activeCourse: course.ID,
    courses: MockData.mockedCourses(),
    enrollments: MockData.mockedEnrollments().enrollments.filter(e => e.userID === BigInt(1)),
    assignments: MockData.mockedCourseAssignments(),
    repositories: MockData.mockedRepositories(),
}

const renderTeacherPage = (path: string) => {
    render(
        <Provider value={initializeOvermind(teacherState)}>
            <MemoryRouter initialEntries={[path]}>
                <Routes>
                    <Route path="/course/:id/*" element={<TeacherPage />} />
                </Routes>
            </MemoryRouter>
        </Provider>
    )
}

describe("EditCourse", () => {
    test("offers the teacher an Edit Course tile on the course page", () => {
        renderTeacherPage("/course/1")

        expect(screen.getByRole("button", { name: "Edit Course" })).toBeTruthy()
    })

    test("renders the course form with the current course prefilled", () => {
        renderTeacherPage("/course/1/edit")

        expect(screen.getByDisplayValue(course.name)).toBeTruthy()
        expect(screen.getByDisplayValue(course.code)).toBeTruthy()
        expect(screen.getByDisplayValue(course.tag)).toBeTruthy()
        expect(screen.getByDisplayValue(course.year.toString())).toBeTruthy()
        expect(screen.getByRole("button", { name: /Save Changes/ })).toBeTruthy()
    })

    test("navigates from the tile to the course form", () => {
        renderTeacherPage("/course/1")
        // The form is only reachable through the tile; the teacher has no
        // other route to it from the course page.
        expect(screen.queryByDisplayValue(course.name)).toBeNull()

        fireEvent.click(screen.getByRole("button", { name: "Edit Course" }))

        expect(screen.getByDisplayValue(course.name)).toBeTruthy()
    })

    test("reports a course that is not in state", () => {
        renderTeacherPage("/course/99/edit")

        expect(screen.getByText("Course not found.")).toBeTruthy()
        expect(screen.queryByDisplayValue(course.name)).toBeNull()
    })
})

import { create } from "@bufbuild/protobuf"
import { render, screen } from "@testing-library/react"
import { Provider } from "overmind-react"
import { MemoryRouter, Route, Routes, useParams } from "react-router"
import type { Course, Enrollment } from "../../proto/qf/types_pb"
import { CourseSchema, Enrollment_DisplayState, Enrollment_UserStatus, EnrollmentSchema, UserSchema } from "../../proto/qf/types_pb"
import Dashboard from "../components/Dashboard"
import { initializeOvermind } from "./TestHelpers"

const self = create(UserSchema, {
    ID: BigInt(1),
    Name: "Test User",
    Email: "test@example.com",
    StudentID: "1234567",
})

const course = (ID: number, code: string): Course => create(CourseSchema, {
    ID: BigInt(ID),
    code,
    name: `Course ${code}`,
    tag: "Fall",
    year: new Date().getFullYear(),
})

const enrollment = (ID: number, courseID: number, status: Enrollment_UserStatus, state: Enrollment_DisplayState): Enrollment =>
    create(EnrollmentSchema, {
        ID: BigInt(ID),
        courseID: BigInt(courseID),
        userID: self.ID,
        status,
        state,
    })

/** CourseStub stands in for the course page, so that a redirect to /course/:id is observable. */
const CourseStub = () => {
    const { id } = useParams()
    return <div>Course page for {id}</div>
}

const renderDashboard = (courses: Course[], enrollments: Enrollment[]) => {
    const mockedOvermind = initializeOvermind({ self, courses, enrollments, isLoading: false })
    render(
        <Provider value={mockedOvermind}>
            <MemoryRouter initialEntries={["/"]}>
                <Routes>
                    <Route path="/" element={<Dashboard />} />
                    <Route path="/course/:id" element={<CourseStub />} />
                    <Route path="/courses" element={<div>All courses page</div>} />
                </Routes>
            </MemoryRouter>
        </Provider>
    )
}

describe("Dashboard", () => {
    it("redirects to the course page when the user has a single favorite course", () => {
        const courses = [course(1, "DAT100"), course(2, "DAT200")]
        const enrollments = [
            enrollment(1, 1, Enrollment_UserStatus.TEACHER, Enrollment_DisplayState.VISIBLE),
            enrollment(2, 2, Enrollment_UserStatus.STUDENT, Enrollment_DisplayState.HIDDEN),
        ]
        renderDashboard(courses, enrollments)
        expect(screen.getByText("Course page for 1")).toBeDefined()
    })

    it("does not redirect when the user has more than one favorite course", () => {
        const courses = [course(1, "DAT100"), course(2, "DAT200")]
        const enrollments = [
            enrollment(1, 1, Enrollment_UserStatus.TEACHER, Enrollment_DisplayState.VISIBLE),
            enrollment(2, 2, Enrollment_UserStatus.STUDENT, Enrollment_DisplayState.VISIBLE),
        ]
        renderDashboard(courses, enrollments)
        expect(screen.queryByText(/Course page for/)).toBeNull()
        expect(screen.getByText("DAT100")).toBeDefined()
        expect(screen.getByText("DAT200")).toBeDefined()
    })

    it("does not redirect when the user is enrolled, but has no favorite courses", () => {
        const courses = [course(1, "DAT100"), course(2, "DAT200")]
        const enrollments = [
            enrollment(1, 1, Enrollment_UserStatus.TEACHER, Enrollment_DisplayState.HIDDEN),
            enrollment(2, 2, Enrollment_UserStatus.STUDENT, Enrollment_DisplayState.HIDDEN),
        ]
        renderDashboard(courses, enrollments)
        expect(screen.queryByText(/Course page for/)).toBeNull()
        expect(screen.queryByText("All courses page")).toBeNull()
        expect(screen.queryByText("DAT100")).toBeNull()
        expect(screen.queryByText("DAT200")).toBeNull()
    })

    it("redirects to the course list when the user has no enrollments", () => {
        renderDashboard([course(1, "DAT100")], [])
        expect(screen.getByText("All courses page")).toBeDefined()
    })
})

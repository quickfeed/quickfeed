import { create } from "@bufbuild/protobuf"
import { fireEvent, render, screen } from "@testing-library/react"
import { Provider } from "overmind-react"
import { MemoryRouter, Route, Routes, useNavigate, useParams } from "react-router"
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

/** CourseStub stands in for the course page, so that a redirect to /course/:id is observable.
 *  Its Back button lets a test observe what the redirect left behind in the history. */
const CourseStub = () => {
    const { id } = useParams()
    const navigate = useNavigate()
    return (
        <div>
            <div>Course page for {id}</div>
            <button onClick={() => navigate(-1)}>Back</button>
        </div>
    )
}

const renderDashboard = (courses: Course[], enrollments: Enrollment[], initialEntries = ["/"]) => {
    const mockedOvermind = initializeOvermind({ self, courses, enrollments, isLoading: false })
    render(
        <Provider value={mockedOvermind}>
            <MemoryRouter initialEntries={initialEntries} initialIndex={initialEntries.length - 1}>
                <Routes>
                    <Route path="/previous" element={<div>Previous page</div>} />
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
        // The dashboard shows none of the hidden enrollments, so it links to the full list.
        expect(screen.getByText("See all your courses")).toBeDefined()
    })

    it("lists courses available for enrollment below the favorites", () => {
        const courses = [course(1, "DAT100"), course(2, "DAT200"), course(3, "DAT300")]
        const enrollments = [
            enrollment(1, 1, Enrollment_UserStatus.TEACHER, Enrollment_DisplayState.VISIBLE),
            enrollment(2, 2, Enrollment_UserStatus.STUDENT, Enrollment_DisplayState.VISIBLE),
        ]
        renderDashboard(courses, enrollments)
        expect(screen.getByText("Available Courses")).toBeDefined()
        expect(screen.getByText("DAT300")).toBeDefined()
        expect(screen.getByRole("button", { name: "Enroll" })).toBeDefined()
        // Nothing is hidden from the user, so no link to the full course list is needed.
        expect(screen.queryByText("See all your courses")).toBeNull()
    })

    it("does not show the available courses section when enrolled in every course", () => {
        const courses = [course(1, "DAT100"), course(2, "DAT200")]
        const enrollments = [
            enrollment(1, 1, Enrollment_UserStatus.TEACHER, Enrollment_DisplayState.VISIBLE),
            enrollment(2, 2, Enrollment_UserStatus.STUDENT, Enrollment_DisplayState.VISIBLE),
        ]
        renderDashboard(courses, enrollments)
        expect(screen.getByText("DAT100")).toBeDefined()
        expect(screen.getByText("DAT200")).toBeDefined()
        expect(screen.queryByText("Available Courses")).toBeNull()
    })

    it("shows the available courses section when the user has no favorite courses", () => {
        const courses = [course(1, "DAT100"), course(2, "DAT200"), course(3, "DAT300")]
        const enrollments = [
            enrollment(1, 1, Enrollment_UserStatus.TEACHER, Enrollment_DisplayState.HIDDEN),
            enrollment(2, 2, Enrollment_UserStatus.STUDENT, Enrollment_DisplayState.HIDDEN),
        ]
        renderDashboard(courses, enrollments)
        expect(screen.getByText("Available Courses")).toBeDefined()
        expect(screen.getByText("DAT300")).toBeDefined()
        expect(screen.getByText("See all your courses")).toBeDefined()
    })

    // The backend creates enrollment requests as pending and visible.
    it.each([
        ["visible", Enrollment_DisplayState.VISIBLE],
        ["hidden", Enrollment_DisplayState.HIDDEN],
    ])("does not show %s pending enrollments on the dashboard", (_, state) => {
        const courses = [course(1, "DAT100"), course(2, "DAT200"), course(3, "DAT300")]
        const enrollments = [
            enrollment(1, 1, Enrollment_UserStatus.TEACHER, Enrollment_DisplayState.VISIBLE),
            enrollment(2, 2, Enrollment_UserStatus.STUDENT, Enrollment_DisplayState.VISIBLE),
            enrollment(3, 3, Enrollment_UserStatus.PENDING, state),
        ]
        renderDashboard(courses, enrollments)
        expect(screen.getByText("DAT100")).toBeDefined()
        expect(screen.getByText("DAT200")).toBeDefined()
        expect(screen.queryByText("DAT300")).toBeNull()
        expect(screen.queryByText("Available Courses")).toBeNull()
        expect(screen.getByText("See all your courses")).toBeDefined()
    })

    it("redirects to the course list when the user has no enrollments", () => {
        renderDashboard([course(1, "DAT100")], [])
        expect(screen.getByText("All courses page")).toBeDefined()
    })

    it("does not redirect when the favorite course has not been loaded", () => {
        // Only DAT200 is loaded; the single favorite refers to course 1, which is not.
        const courses = [course(2, "DAT200")]
        const enrollments = [
            enrollment(1, 1, Enrollment_UserStatus.TEACHER, Enrollment_DisplayState.VISIBLE),
        ]
        renderDashboard(courses, enrollments)
        expect(screen.queryByText(/Course page for/)).toBeNull()
        expect(screen.queryByText("All courses page")).toBeNull()
    })

    it("replaces the dashboard in history, so that going back skips the redirect", () => {
        const courses = [course(1, "DAT100")]
        const enrollments = [
            enrollment(1, 1, Enrollment_UserStatus.TEACHER, Enrollment_DisplayState.VISIBLE),
        ]
        renderDashboard(courses, enrollments, ["/previous", "/"])
        expect(screen.getByText("Course page for 1")).toBeDefined()

        fireEvent.click(screen.getByRole("button", { name: "Back" }))

        expect(screen.getByText("Previous page")).toBeDefined()
        expect(screen.queryByText(/Course page for/)).toBeNull()
    })
})

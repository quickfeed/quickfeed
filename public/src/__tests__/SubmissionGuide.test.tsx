import { create } from "@bufbuild/protobuf"
import { fireEvent, render, screen, waitFor } from "@testing-library/react"
import { Provider } from "overmind-react"
import { MemoryRouter, Route, Routes } from "react-router"
import { CourseSchema, Repository_Type, UserSchema } from "../../proto/qf/types_pb"
import { CourseLinks } from "../components/CourseLinks"
import NavBarLabs from "../components/navbar/NavBarLabs"
import SubmissionGuide from "../components/student/SubmissionGuide"
import { initializeOvermind } from "./TestHelpers"
import { vi } from "vitest"


const courseID = 12n
const repositories = {
    [courseID.toString()]: {
        [Repository_Type.USER]: "https://github.com/dat515-2026/student-labs",
        [Repository_Type.ASSIGNMENTS]: "https://github.com/dat515-2026/assignments",
    }
}

const mockedOvermind = initializeOvermind({
    self: create(UserSchema, { ID: 7n, Name: "Student", Login: "student" }),
    courses: [create(CourseSchema, { ID: courseID, code: "DAT515" })],
    repositories,
    activeCourse: courseID,
})

const renderAtCourseRoute = (element: React.ReactNode) => render(
    <Provider value={mockedOvermind}>
        <MemoryRouter initialEntries={[`/course/${courseID}/submission-guide`]}>
            <Routes>
                <Route path="/course/:id/submission-guide" element={element} />
            </Routes>
        </MemoryRouter>
    </Provider>
)

describe("SubmissionGuide", () => {
    it("explains the QuickFeed workflow with course-specific repository commands", () => {
        renderAtCourseRoute(<SubmissionGuide />)

        expect(screen.getByRole("heading", { name: "Submitting assignments" })).toBeDefined()
        expect(screen.getByText(/QuickFeed treats a push .* as a submission/)).toBeDefined()
        expect(screen.getByText("gh repo clone dat515-2026/student-labs dat515")).toBeDefined()
        expect(screen.getByText("git remote add upstream https://github.com/dat515-2026/assignments")).toBeDefined()
        expect(screen.getByText(/After your group is approved/)).toBeDefined()
    })

    it("copies single-line commands and omits copy buttons from multi-line commands", async () => {
        const writeText = vi.fn().mockResolvedValue(undefined)
        Object.defineProperty(navigator, "clipboard", {
            configurable: true,
            value: { writeText },
        })
        renderAtCourseRoute(<SubmissionGuide />)

        expect(screen.getAllByRole("button", { name: /^Copy command:/ })).toHaveLength(7)
        const multiLineCommand = screen.getByText((_, element) =>
            element?.tagName === "CODE" && element.textContent?.startsWith("git status\ngit add") === true
        )
        expect(multiLineCommand.closest("div")?.querySelector("button")).toBeNull()

        fireEvent.click(screen.getByRole("button", { name: "Copy command: gh auth login" }))

        await waitFor(() => expect(writeText).toHaveBeenCalledWith("gh auth login"))
        expect(screen.getByRole("button", { name: "Copy command: gh auth login" }).textContent).toContain("Copied")
    })

    it("is linked from the course resources", () => {
        renderAtCourseRoute(<CourseLinks />)

        expect(screen.getByRole("link", { name: "Submission Guide" }).getAttribute("href"))
            .toBe(`/course/${courseID}/submission-guide`)
    })

    it("remains available in the course navigation before assignments load", () => {
        renderAtCourseRoute(<NavBarLabs />)

        expect(screen.getByRole("button", { name: "Submission Guide" })).toBeDefined()
    })
})

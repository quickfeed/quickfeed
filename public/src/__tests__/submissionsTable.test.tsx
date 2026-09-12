import { create } from "@bufbuild/protobuf"
import { render, screen, within } from "@testing-library/react"
import { Provider } from "overmind-react"
import { vi } from "vitest"
import { Enrollment_UserStatus, EnrollmentSchema, UserSchema } from "../../proto/qf/types_pb"
import SubmissionsTable from "../components/submissions-table/SubmissionsTable"
import { MockData } from "./mock_data/mockData"
import { initializeOvermind } from "./TestHelpers"

/** Enrollment for a student without a group; all mocked enrollments are in a group. */
const soloEnrollment = create(EnrollmentSchema, {
    ID: BigInt(7),
    courseID: BigInt(1),
    userID: BigInt(6),
    status: Enrollment_UserStatus.STUDENT,
    user: create(UserSchema, {
        ID: BigInt(6),
        Name: "Ola Nordmann",
        Login: "olan",
    }),
})

const renderTable = (groupView = false) => {
    const enrollments = MockData.mockedEnrollments().enrollments.filter(e => e.courseID === BigInt(1))
    const mockedOvermind = initializeOvermind({
        activeCourse: BigInt(1),
        groupView,
        courses: MockData.mockedCourses(),
        courseEnrollments: { "1": [...enrollments, soloEnrollment] },
        groups: { "1": MockData.mockedGroups().groups },
        assignments: { "1": MockData.mockedAssignments().filter(a => a.CourseID === BigInt(1)) },
        submissionsForCourse: MockData.mockedCourseSubmissions(BigInt(1)),
    })
    render(
        <Provider value={mockedOvermind}>
            <SubmissionsTable onSubmissionClick={vi.fn()} />
        </Provider>
    )
}

/** Returns the table row containing the given member name. */
const rowFor = (name: string) => {
    const row = screen.getByText(name).closest("tr")
    expect(row).not.toBeNull()
    return within(row as HTMLTableRowElement)
}

describe("SubmissionsTable shows the GitHub login and group name", () => {
    test("student in a group shows both the login and the group name", () => {
        renderTable()
        const row = rowFor("Bob Bobsen")
        expect(row.getByText("@Bob")).toBeDefined()
        expect(row.getByText("Group 2")).toBeDefined()
    })

    test("student without a group shows only the login", () => {
        renderTable()
        const row = rowFor("Ola Nordmann")
        expect(row.getByText("@olan")).toBeDefined()
        expect(row.queryByText(/^Group /)).toBeNull()
    })

    test("group rows show the group name without a login", () => {
        renderTable(true)
        const row = rowFor("Group 1")
        expect(row.queryByText(/^@/)).toBeNull()
    })
})

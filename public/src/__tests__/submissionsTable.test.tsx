import { clone, create } from "@bufbuild/protobuf"
import { act, fireEvent, render, screen, within } from "@testing-library/react"
import { Provider } from "overmind-react"
import { MemoryRouter, Route, Routes } from "react-router"
import { vi } from "vitest"
import { Enrollment_UserStatus, EnrollmentSchema, GroupSchema, UserSchema } from "../../proto/qf/types_pb"
import SubmissionsTable from "../components/submissions-table/SubmissionsTable"
import { ApiClient } from "../overmind/namespaces/global/effects"
import type { State } from "../overmind/state"
import { MockData } from "./mock_data/mockData"
import { initializeOvermind, mock } from "./TestHelpers"

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

const renderTable = (groupView = false, overrides: Partial<State> = {}, api?: ApiClient) => {
    const enrollments = MockData.mockedEnrollments().enrollments.filter(e => e.courseID === BigInt(1))
    const courses = MockData.mockedCourses()
    courses[0].ScmOrganizationName = "test-course"
    const mockedOvermind = initializeOvermind({
        activeCourse: BigInt(1),
        groupView,
        courses,
        courseEnrollments: { "1": [...enrollments, soloEnrollment] },
        groups: { "1": MockData.mockedGroups().groups },
        assignments: { "1": MockData.mockedAssignments().filter(a => a.CourseID === BigInt(1)) },
        submissionsForCourse: MockData.mockedCourseSubmissions(BigInt(1)),
        ...overrides,
    }, api)
    render(
        <Provider value={mockedOvermind}>
            <MemoryRouter initialEntries={["/course/1/results"]}>
                <Routes>
                    <Route path="/course/1/results" element={<SubmissionsTable onSubmissionClick={vi.fn()} />} />
                    <Route path="/course/1/members/7" element={<div>Student details</div>} />
                </Routes>
            </MemoryRouter>
        </Provider>
    )
    return mockedOvermind
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

const enrollmentWithGroup = () => create(EnrollmentSchema, {
    ...soloEnrollment,
    groupID: 2n,
    group: create(GroupSchema, { ID: 2n, courseID: 1n, name: "Old group" }),
})

describe("SubmissionsTable member navigation and search", () => {
    test("name opens student details using the enrollment ID", () => {
        renderTable()
        const link = rowFor("Ola Nordmann").getByRole("link", { name: "Ola Nordmann" })
        expect(link.getAttribute("href")).toBe("/course/1/members/7")
        fireEvent.click(link)
        expect(screen.getByText("Student details")).toBeDefined()
    })

    test("login and group open their course repositories", () => {
        renderTable()
        const row = rowFor("Bob Bobsen")
        for (const [name, repo] of [["@Bob", "Bob-labs"], ["Group 2", "Group 2"]]) {
            const link = row.getByRole("link", { name })
            expect(link.getAttribute("href")).toBe(`https://github.com/test-course/${repo}`)
            expect(link.getAttribute("target")).toBe("_blank")
            expect(link.getAttribute("rel")).toBe("noopener noreferrer")
        }
    })

    test("group rows still link to the group repository", () => {
        renderTable(true)
        expect(rowFor("Group 1").getByRole("link").getAttribute("href"))
            .toBe("https://github.com/test-course/Group 1")
    })

    test.each(["ola", "olan", "@olan", "group 2"])("searches by %s", async query => {
        const overmind = renderTable()
        await act(async () => { await overmind.actions.global.setQuery(query) })
        expect(screen.getByText(query === "group 2" ? "Bob Bobsen" : "Ola Nordmann")).toBeDefined()
        expect(screen.queryByText("Test User")).toBeNull()
        expect(screen.queryByText(query === "group 2" ? "Ola Nordmann" : "Bob Bobsen")).toBeNull()
    })
})

describe("SubmissionsTable group details", () => {
    test.each([false, true])("uses nested group only while list is unloaded: %s", loaded => {
        renderTable(false, {
            courseEnrollments: { "1": [enrollmentWithGroup()] },
            groups: loaded ? { "1": [] } : {},
        })
        const row = rowFor("Ola Nordmann")
        expect(row.queryByText("Old group") !== null).toBe(!loaded)
    })

    test("updates displayed group, repository link and search after a rename", async () => {
        const api = new ApiClient()
        api.client = {
            ...api.client,
            updateGroup: mock("updateGroup", async group => ({ error: null, message: create(GroupSchema, group) })),
        }
        const overmind = renderTable(false, {
            courseEnrollments: { "1": [enrollmentWithGroup()] },
        }, api)
        expect(rowFor("Ola Nordmann").queryByText("Old group")).toBeNull()
        expect(rowFor("Ola Nordmann").getByText("Group 2")).toBeDefined()
        const group = clone(GroupSchema, overmind.state.groups["1"][1])
        group.name = "Renamed group"
        await act(async () => { await overmind.actions.global.updateGroup(group) })
        const row = rowFor("Ola Nordmann")
        expect(row.queryByText("Group 2")).toBeNull()
        expect(row.getByRole("link", { name: "Renamed group" }).getAttribute("href"))
            .toBe("https://github.com/test-course/Renamed group")
        await act(async () => { await overmind.actions.global.setQuery("renamed") })
        expect(screen.getByText("Ola Nordmann")).toBeDefined()
        await act(async () => { await overmind.actions.global.setQuery("old group") })
        expect(screen.queryByText("Ola Nordmann")).toBeNull()
    })
})

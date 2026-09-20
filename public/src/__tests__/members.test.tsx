import { clone, create } from "@bufbuild/protobuf"
import { act, fireEvent, render, screen } from "@testing-library/react"
import { Provider } from "overmind-react"
import { MemoryRouter, Route, Routes } from "react-router"
import { Enrollment_UserStatus, EnrollmentSchema, GroupSchema, NotesSchema, UserSchema } from "../../proto/qf/types_pb"
import Members from "../components/Members"
import { ApiClient } from "../overmind/namespaces/global/effects"
import type { State } from "../overmind/state"
import { MockData } from "./mock_data/mockData"
import { initializeOvermind, mock } from "./TestHelpers"

/** Enrollment whose nested group message still carries the name the group had when the course was loaded. */
const staleEnrollment = () => create(EnrollmentSchema, {
    ID: BigInt(7),
    courseID: BigInt(1),
    userID: BigInt(6),
    status: Enrollment_UserStatus.STUDENT,
    groupID: BigInt(2),
    group: create(GroupSchema, { ID: BigInt(2), courseID: BigInt(1), name: "Old group" }),
    user: create(UserSchema, { ID: BigInt(6), Name: "Ola Nordmann", Login: "olan" }),
})

const renderMembers = (overrides: Partial<State> = {}) => {
    const api = new ApiClient()
    api.client = {
        ...api.client,
        getCourseNotes: mock("getCourseNotes", async () => ({ error: null, message: create(NotesSchema) })), // skipcq: JS-0116
        updateGroup: mock("updateGroup", async group => ({ error: null, message: create(GroupSchema, group) })),
    }
    const mockedOvermind = initializeOvermind({
        activeCourse: BigInt(1),
        courses: MockData.mockedCourses(),
        courseEnrollments: { "1": [staleEnrollment()] },
        groups: { "1": MockData.mockedGroups().groups },
        ...overrides,
    }, api)
    render(
        <Provider value={mockedOvermind}>
            <MemoryRouter initialEntries={["/course/1/members"]}>
                <Routes>
                    <Route path="/course/:id/members" element={<Members />} />
                </Routes>
            </MemoryRouter>
        </Provider>
    )
    return mockedOvermind
}

describe("Members group column", () => {
    test("prefers the loaded group list over the nested enrollment group", () => {
        renderMembers()
        expect(screen.getByText("Group 2")).toBeDefined()
        expect(screen.queryByText("Old group")).toBeNull()
    })

    test("falls back to the nested group while the group list is unloaded", () => {
        renderMembers({ groups: {} })
        expect(screen.getByText("Old group")).toBeDefined()
    })

    test("updates the displayed group after a rename", async () => {
        const overmind = renderMembers()
        expect(screen.getByText("Group 2")).toBeDefined()
        const group = clone(GroupSchema, overmind.state.groups["1"][1])
        group.name = "Renamed group"
        await act(async () => { await overmind.actions.global.updateGroup(group) })
        expect(screen.getByText("Renamed group")).toBeDefined()
        expect(screen.queryByText("Group 2")).toBeNull()
    })
})

/** Two students whose nested group names sort in the opposite order to the live ones:
    nested "Zebra"/"Alpha" would put Bob first, the live "Group 1"/"Group 2" put Alice first. */
const staleSortEnrollments = () => [
    create(EnrollmentSchema, {
        ID: BigInt(8),
        courseID: BigInt(1),
        userID: BigInt(8),
        status: Enrollment_UserStatus.STUDENT,
        groupID: BigInt(1),
        group: create(GroupSchema, { ID: BigInt(1), courseID: BigInt(1), name: "Zebra" }),
        user: create(UserSchema, { ID: BigInt(8), Name: "Alice", Login: "alice" }),
    }),
    create(EnrollmentSchema, {
        ID: BigInt(9),
        courseID: BigInt(1),
        userID: BigInt(9),
        status: Enrollment_UserStatus.STUDENT,
        groupID: BigInt(2),
        group: create(GroupSchema, { ID: BigInt(2), courseID: BigInt(1), name: "Alpha" }),
        user: create(UserSchema, { ID: BigInt(9), Name: "Bob", Login: "bob" }),
    }),
]

/** Names of the table's members, in the order the rows are rendered. The name is the
    row's first cell. The table is aria-hidden, so it has to be read from the DOM. */
const memberOrder = () => Array.from(document.querySelectorAll("tbody tr"))
    .map(row => row.querySelector("th")?.textContent?.trim())

describe("Members group column sorting", () => {
    test("sorts on the loaded group names, not the nested ones", () => {
        renderMembers({ courseEnrollments: { "1": staleSortEnrollments() } })
        fireEvent.click(screen.getByText("Group"))
        expect(memberOrder()).toEqual(["Alice", "Bob"])
    })

    test("re-sorts after a rename", async () => {
        const overmind = renderMembers({ courseEnrollments: { "1": staleSortEnrollments() } })
        fireEvent.click(screen.getByText("Group"))
        expect(memberOrder()).toEqual(["Alice", "Bob"])
        const group = clone(GroupSchema, overmind.state.groups["1"][0])
        group.name = "Zulu group"
        await act(async () => { await overmind.actions.global.updateGroup(group) })
        expect(memberOrder()).toEqual(["Bob", "Alice"])
    })
})

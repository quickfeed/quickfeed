import { clone, create } from "@bufbuild/protobuf"
import { act, render, screen } from "@testing-library/react"
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

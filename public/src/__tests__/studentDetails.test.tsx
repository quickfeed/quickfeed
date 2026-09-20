import { clone, create } from "@bufbuild/protobuf"
import { act, render, screen } from "@testing-library/react"
import { Provider } from "overmind-react"
import { MemoryRouter, Route, Routes } from "react-router"
import { Enrollment_UserStatus, EnrollmentSchema, GroupSchema, NotesSchema, UserSchema } from "../../proto/qf/types_pb"
import StudentDetails from "../components/StudentDetails"
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

const renderDetails = (overrides: Partial<State> = {}) => {
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
        // Pretend the course submissions are loaded so the view does not fetch them.
        loadedCourse: { "1": true },
        ...overrides,
    }, api)
    render(
        <Provider value={mockedOvermind}>
            <MemoryRouter initialEntries={["/course/1/members/7"]}>
                <Routes>
                    <Route path="/course/:id/members/:enrollmentID" element={<StudentDetails />} />
                </Routes>
            </MemoryRouter>
        </Provider>
    )
    return mockedOvermind
}

/** The identity card renders the group line in a span; the note target selector
    renders a same-looking "Group: ..." option, so the query is narrowed to the span. */
const groupLine = (name: string) => screen.queryByText(`Group: ${name}`, { selector: "span" })

describe("StudentDetails group line", () => {
    test("prefers the loaded group list over the nested enrollment group", () => {
        renderDetails()
        expect(groupLine("Group 2")).not.toBeNull()
        expect(groupLine("Old group")).toBeNull()
    })

    test("falls back to the nested group while the group list is unloaded", () => {
        renderDetails({ groups: {} })
        expect(groupLine("Old group")).not.toBeNull()
    })

    test("updates the displayed group after a rename", async () => {
        const overmind = renderDetails()
        expect(groupLine("Group 2")).not.toBeNull()
        const group = clone(GroupSchema, overmind.state.groups["1"][1])
        group.name = "Renamed group"
        await act(async () => { await overmind.actions.global.updateGroup(group) })
        expect(groupLine("Renamed group")).not.toBeNull()
        expect(groupLine("Group 2")).toBeNull()
    })
})

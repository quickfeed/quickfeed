import { create } from "@bufbuild/protobuf"
import { act, fireEvent, render, screen } from "@testing-library/react"
import { Provider } from "overmind-react"
import { MemoryRouter, Route, Routes } from "react-router"
import type { Group } from "../../proto/qf/types_pb"
import { Enrollment_UserStatus, EnrollmentSchema, Group_GroupStatus, GroupSchema, UserSchema } from "../../proto/qf/types_pb"
import Groups from "../components/Groups"
import { ApiClient } from "../overmind/namespaces/global/effects"
import { MockData } from "./mock_data/mockData"
import { initializeOvermind, mock } from "./TestHelpers"

const courseID = BigInt(1)
const teacher = create(UserSchema, { ID: BigInt(1), Name: "Tina Teacher", Login: "tina" })
const student = create(UserSchema, { ID: BigInt(2), Name: "Sam Student", Login: "sam" })

/** By default the teacher is a member of group 1, which used to hide every way for them to create a group. */
const teacherEnrollment = (groupID: bigint) => create(EnrollmentSchema, {
    ID: BigInt(1), courseID, userID: teacher.ID, user: teacher,
    status: Enrollment_UserStatus.TEACHER, groupID,
    group: groupID ? create(GroupSchema, { ID: groupID, courseID, name: "Group 1" }) : undefined,
})

const studentEnrollment = () => create(EnrollmentSchema, {
    ID: BigInt(2), courseID, userID: student.ID, user: student,
    status: Enrollment_UserStatus.STUDENT,
})

const setup = (teacherGroupID = BigInt(1)) => {
    const created: Group[] = []
    const api = new ApiClient()
    api.client = {
        ...api.client,
        createGroup: mock("createGroup", async req => {
            const group = create(GroupSchema, { ...req, ID: BigInt(3), status: Group_GroupStatus.PENDING })
            created.push(group)
            return { error: null, message: group }
        }),
    }
    const overmind = initializeOvermind({
        self: teacher,
        activeCourse: courseID,
        courses: MockData.mockedCourses(),
        enrollments: [teacherEnrollment(teacherGroupID)],
        status: { "1": Enrollment_UserStatus.TEACHER },
        courseEnrollments: { "1": [teacherEnrollment(teacherGroupID), studentEnrollment()] },
        groups: { "1": MockData.mockedGroups().groups },
    }, api)
    return { overmind, created }
}

const renderGroups = async (overmind: ReturnType<typeof setup>["overmind"]) => {
    await act(async () => {
        render(
            <Provider value={overmind}>
                <MemoryRouter initialEntries={["/course/1/groups"]}>
                    <Routes>
                        <Route path="/course/:id/groups" element={<Groups />} />
                    </Routes>
                </MemoryRouter>
            </Provider>
        )
    })
}

describe("Teacher group creation", () => {
    test("teacher in a group can create a group for students from the group list", async () => {
        const { overmind, created } = setup()
        await renderGroups(overmind)

        fireEvent.click(screen.getByText("New Group"))
        fireEvent.keyUp(screen.getByPlaceholderText("Enter group name..."), { target: { value: "new_group" } })
        // The add button sits in the row with the student's name.
        const row = screen.getByText(student.Name).parentElement as HTMLElement
        fireEvent.click(row.querySelector("button") as HTMLButtonElement)
        await act(async () => { fireEvent.click(screen.getByText("Create Group")) })

        expect(created).toHaveLength(1)
        expect(created[0].name).toBe("new_group")
        expect(created[0].users.map(u => u.ID)).toEqual([student.ID])
        // Back on the list, with the new group pending approval.
        expect(screen.queryByText("Create Group")).toBeNull()
        expect(screen.getByText("new_group")).toBeDefined()
        expect(overmind.state.groups["1"].map(g => g.ID)).toContain(BigInt(3))
        // The teacher is not a member, so their own group is unchanged.
        expect(overmind.state.userGroup["1"].ID).toBe(BigInt(1))
        // The student can no longer be added to another group.
        expect(overmind.state.courseEnrollments["1"][1].groupID).toBe(BigInt(3))
    })

    test("cancel returns to the group list without creating a group", async () => {
        const { overmind, created } = setup()
        await renderGroups(overmind)

        fireEvent.click(screen.getByText("New Group"))
        fireEvent.click(screen.getByText("Cancel"))

        expect(created).toHaveLength(0)
        expect(overmind.state.activeGroup).toBeNull()
        expect(screen.getByText("New Group")).toBeDefined()
    })

    test("teacher not in a group can add themselves", async () => {
        const { overmind } = setup(BigInt(0))
        await renderGroups(overmind)

        fireEvent.click(screen.getByText("New Group"))
        fireEvent.click(screen.getByText("Students"))

        expect(screen.getByText(teacher.Name)).toBeDefined()
    })
})

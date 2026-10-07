import { clone, create } from "@bufbuild/protobuf"
import { act, fireEvent, render, screen } from "@testing-library/react"
import { Provider } from "overmind-react"
import { MemoryRouter, Route, Routes } from "react-router"
import type { Group } from "../../proto/qf/types_pb"
import { VoidSchema } from "../../proto/qf/requests_pb"
import { Enrollment_UserStatus, EnrollmentSchema, Group_GroupStatus, GroupSchema, UserSchema } from "../../proto/qf/types_pb"
import GroupForm from "../components/group/GroupForm"
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
    group: groupID ? create(GroupSchema, { ID: groupID, courseID, name: "group_1" }) : undefined,
})

const studentEnrollment = () => create(EnrollmentSchema, {
    ID: BigInt(2), courseID, userID: student.ID, user: student,
    status: Enrollment_UserStatus.STUDENT,
})

/** courseGroups returns an approved group_1, with the teacher as its member if teacherGroupID is 1, and a pending group_2. */
const courseGroups = (teacherGroupID: bigint) => [
    create(GroupSchema, {
        ID: BigInt(1), courseID, name: "group_1", status: Group_GroupStatus.APPROVED,
        users: teacherGroupID === BigInt(1) ? [clone(UserSchema, teacher)] : [],
    }),
    create(GroupSchema, { ID: BigInt(2), courseID, name: "group_2", status: Group_GroupStatus.PENDING }),
]

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
        // Copy the request, as it still belongs to the state tree.
        updateGroup: mock("updateGroup", async req => ({ error: null, message: clone(GroupSchema, create(GroupSchema, req)) })),
        deleteGroup: mock("deleteGroup", async () => ({ error: null, message: create(VoidSchema) })),
        isEmptyRepo: mock("isEmptyRepo", async () => ({ error: null, message: create(VoidSchema) })),
    }
    const overmind = initializeOvermind({
        self: teacher,
        activeCourse: courseID,
        courses: MockData.mockedCourses(),
        enrollments: [teacherEnrollment(teacherGroupID)],
        status: { "1": Enrollment_UserStatus.TEACHER },
        courseEnrollments: { "1": [teacherEnrollment(teacherGroupID), studentEnrollment()] },
        groups: { "1": courseGroups(teacherGroupID) },
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

const nameInput = () => screen.getByPlaceholderText("e.g. team_rocket")
const checkbox = (user: typeof teacher) => screen.getByRole("checkbox", { name: new RegExp(user.Name) }) as HTMLInputElement
const createButton = () => screen.getByRole("button", { name: "Create group" }) as HTMLButtonElement

describe("Teacher group creation", () => {
    test("teacher in a group can create a group for students from the group list", async () => {
        const { overmind, created } = setup()
        await renderGroups(overmind)

        fireEvent.click(screen.getByText("New Group"))
        fireEvent.change(nameInput(), { target: { value: "new_group" } })
        fireEvent.click(checkbox(student))
        await act(async () => { fireEvent.click(createButton()) })

        expect(created).toHaveLength(1)
        expect(created[0].name).toBe("new_group")
        expect(created[0].users.map(u => u.ID)).toEqual([student.ID])
        // Back on the list, with the new group pending approval.
        expect(screen.queryByRole("button", { name: "Create group" })).toBeNull()
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
        fireEvent.click(screen.getByRole("tab", { name: /Teachers/ }))
        fireEvent.click(checkbox(teacher))

        expect(overmind.state.activeGroup?.users.map(u => u.ID)).toEqual([teacher.ID])
    })

    test("users in another group are shown but cannot be picked", async () => {
        const { overmind } = setup()
        await renderGroups(overmind)

        fireEvent.click(screen.getByText("New Group"))
        fireEvent.click(screen.getByRole("tab", { name: /Teachers/ }))

        expect(checkbox(teacher).disabled).toBe(true)
        expect(screen.getByText("In group_1")).toBeDefined()
    })
})

describe("Group membership in the member list", () => {
    const openNewGroup = (tab: "Students" | "Teachers") => {
        fireEvent.click(screen.getByText("New Group"))
        fireEvent.click(screen.getByRole("tab", { name: new RegExp(tab) }))
    }

    test("a member removed from a group can join another", async () => {
        const { overmind } = setup()
        await renderGroups(overmind)

        // group_2 is pending, so it is listed first.
        fireEvent.click(screen.getAllByText("Edit")[1])
        fireEvent.click(screen.getByRole("button", { name: `Remove ${teacher.Name}` }))
        // A group needs at least one member.
        fireEvent.click(checkbox(student))
        await act(async () => { fireEvent.click(screen.getByRole("button", { name: "Save changes" })) })
        openNewGroup("Teachers")

        expect(checkbox(teacher).disabled).toBe(false)
        expect(screen.queryByText("In group_1")).toBeNull()
    })

    test("members of a deleted group can join another", async () => {
        window.confirm = vi.fn(() => true)
        const { overmind } = setup()
        await renderGroups(overmind)

        await act(async () => { fireEvent.click(screen.getAllByText("Delete")[1]) })
        openNewGroup("Teachers")

        expect(checkbox(teacher).disabled).toBe(false)
    })

    test("a member added to a group cannot join another", async () => {
        const { overmind } = setup()
        await renderGroups(overmind)

        fireEvent.click(screen.getAllByText("Edit")[0])
        fireEvent.click(checkbox(student))
        await act(async () => { fireEvent.click(screen.getByRole("button", { name: "Save changes" })) })
        openNewGroup("Students")

        expect(checkbox(student).disabled).toBe(true)
        expect(screen.getByText("In group_2")).toBeDefined()
    })
})

describe("Group form validation", () => {
    test("create is disabled until the group has a valid name and a member", async () => {
        const { overmind } = setup()
        await renderGroups(overmind)
        fireEvent.click(screen.getByText("New Group"))

        expect(createButton().disabled).toBe(true)
        expect(screen.getByText("Enter a group name to continue.")).toBeDefined()

        fireEvent.change(nameInput(), { target: { value: "bad!" } })
        expect(screen.getByText(/can only contain letters/)).toBeDefined()
        expect(createButton().disabled).toBe(true)

        fireEvent.change(nameInput(), { target: { value: "good" } })
        expect(screen.getByText("Add at least one member to continue.")).toBeDefined()
        expect(createButton().disabled).toBe(true)

        fireEvent.click(checkbox(student))
        expect(createButton().disabled).toBe(false)
    })

    test("spaces in the name become underscores", async () => {
        const { overmind } = setup()
        await renderGroups(overmind)
        fireEvent.click(screen.getByText("New Group"))

        fireEvent.change(nameInput(), { target: { value: "team rocket" } })

        expect(overmind.state.activeGroup?.name).toBe("team_rocket")
    })

    test("enter in the search adds the only matching person", async () => {
        const { overmind } = setup()
        await renderGroups(overmind)
        fireEvent.click(screen.getByText("New Group"))

        const search = screen.getByRole("searchbox", { name: "Search members" })
        fireEvent.change(search, { target: { value: "sam" } })
        fireEvent.keyDown(search, { key: "Enter" })

        expect(overmind.state.activeGroup?.users.map(u => u.ID)).toEqual([student.ID])
        expect((search as HTMLInputElement).value).toBe("")
    })
})

describe("Student group creation", () => {
    test("the student is a member and cannot remove themselves", async () => {
        const classmate = create(UserSchema, { ID: BigInt(4), Name: "Cal Classmate", Login: "cal" })
        const overmind = initializeOvermind({
            self: student,
            activeCourse: courseID,
            enrollments: [studentEnrollment()],
            status: { "1": Enrollment_UserStatus.STUDENT },
            courseEnrollments: {
                "1": [studentEnrollment(), create(EnrollmentSchema, {
                    ID: BigInt(4), courseID, userID: classmate.ID, user: classmate, status: Enrollment_UserStatus.STUDENT,
                })],
            },
        })
        await act(async () => {
            render(
                <Provider value={overmind}>
                    <MemoryRouter initialEntries={["/course/1/group"]}>
                        <Routes>
                            <Route path="/course/:id/group" element={<GroupForm />} />
                        </Routes>
                    </MemoryRouter>
                </Provider>
            )
        })

        expect(screen.getByText("Create your group")).toBeDefined()
        expect(checkbox(student).checked).toBe(true)
        expect(checkbox(student).disabled).toBe(true)
        expect(screen.queryByRole("button", { name: `Remove ${student.Name}` })).toBeNull()
        expect(screen.queryByText("Cancel")).toBeNull()
        expect(checkbox(classmate).disabled).toBe(false)
    })
})

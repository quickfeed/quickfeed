import { create } from "@bufbuild/protobuf"
import { Code, ConnectError } from "@connectrpc/connect"
import { render, screen } from "@testing-library/react"
import { createOvermindMock } from "overmind"
import { Provider } from "overmind-react"
import { MemoryRouter } from "react-router"
import { VoidSchema } from "../../proto/qf/requests_pb"
import { UserSchema } from "../../proto/qf/types_pb"
import Profile from "../components/profile/Profile"
import { Color } from "../Helpers"
import { config } from "../overmind"
import { ApiClient } from "../overmind/namespaces/global/effects"
import { initializeOvermind, mock } from "./TestHelpers"


describe("Profile", () => {
    it("Renders with logged in user", () => {
        const mockedOvermind = createOvermindMock(config, (state) => {
            state.self = create(UserSchema, {
                ID: BigInt(1),
                Name: "Test User",
                AvatarURL: "https://example.com/avatar.png",
            })
        })
        render(
            <Provider value={mockedOvermind}>
                <MemoryRouter>
                    <Profile />
                </MemoryRouter>
            </Provider>
        )
        const loggedIn = mockedOvermind.state.isLoggedIn
        expect(loggedIn).toBe(true)
        expect(screen.getByRole("heading").textContent).toBe("Hi, Test User")
    })

    it("Logged in is false if the user is invalid", () => {
        const mockedOvermind = createOvermindMock(config, (state) => {
            state.self = create(UserSchema, {
                ID: BigInt(0),
            })
        })
        const loggedIn = mockedOvermind.state.isLoggedIn
        expect(loggedIn).toBe(false)
    })
})

describe("Duplicate profile errors from the server", () => {
    const duplicateStudentID = "a QuickFeed account with this student ID already exists"

    it("updateUser returns false and leaves the user unchanged", async () => {
        const api = new ApiClient()
        api.client = {
            ...api.client,
            // eslint-disable-next-line @typescript-eslint/no-unused-vars
            updateUser: mock("updateUser", async (_request) => { // skipcq: JS-0116
                return { message: create(VoidSchema), error: new ConnectError(duplicateStudentID, Code.AlreadyExists) }
            }),
        }
        const self = create(UserSchema, { ID: BigInt(1), Name: "Test User", StudentID: "1001" })
        const { state, actions } = initializeOvermind({ self }, api)

        const updated = await actions.global.updateUser(create(UserSchema, { ID: BigInt(1), StudentID: "1002" }))
        expect(updated).toBe(false)
        expect(state.self.StudentID).toEqual("1001")
    })

    it("errorHandler alerts the user with the server's message", () => {
        const { state, actions } = initializeOvermind({ self: create(UserSchema, { ID: BigInt(1) }) })
        actions.global.errorHandler({
            method: "UpdateUser",
            error: new ConnectError(duplicateStudentID, Code.AlreadyExists),
        })
        expect(state.alerts).toHaveLength(1)
        expect(state.alerts[0].text).toEqual(duplicateStudentID)
        expect(state.alerts[0].color).toEqual(Color.RED)
    })
})

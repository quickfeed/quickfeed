import { create } from "@bufbuild/protobuf"
import { render, screen } from "@testing-library/react"
import { createOvermindMock } from "overmind"
import { Provider } from "overmind-react"
import React from "react"
import { MemoryRouter } from "react-router-dom"
import { UserSchema } from "../../proto/qf/types_pb"
import Profile from "../components/profile/Profile"
import { config } from "../overmind"


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

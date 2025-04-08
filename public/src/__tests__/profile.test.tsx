import { UserSchema } from "../../proto/qf/types_pb"
import { Provider } from "overmind-react"
import { createOvermindMock } from "overmind"
import { config } from "../overmind"
import Profile from "../components/profile/Profile"
import { Router } from "react-router-dom"
import { createMemoryHistory } from "history"
import React from "react"
import { render, screen } from "@testing-library/react"
import { create } from "@bufbuild/protobuf"


describe("Profile", () => {
    it("Renders with logged in user", () => {
        const mockedOvermind = createOvermindMock(config, (state) => {
            state.self = create(UserSchema, {
                ID: BigInt(1),
                Name: "Test User",
                AvatarURL: "https://example.com/avatar.png",
            })
        })
        const history = createMemoryHistory()
        render(
            <Provider value={mockedOvermind}>
                <Router history={history}>
                    <Profile />
                </Router>
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

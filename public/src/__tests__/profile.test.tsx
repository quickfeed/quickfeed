import { User } from "../../proto/qf/types_pb"
import { Provider } from "overmind-react"
import { createOvermindMock } from "overmind"
import { config } from "../overmind"
import Profile from "../components/profile/Profile"
import { Router } from "react-router-dom"
import { createMemoryHistory } from "history"
import React from "react"
import { render, screen } from "@testing-library/react"


describe("Profile", () => {
    it("Renders with logged in user", () => {
        const mockedOvermind = createOvermindMock(config, (state) => {
            state.self = new User({
                ID: BigInt(1),
                Name: "Test User",
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
            state.self = new User({
                ID: BigInt(0),
            })
        })
        const loggedIn = mockedOvermind.state.isLoggedIn
        expect(loggedIn).toBe(false)
    })
})

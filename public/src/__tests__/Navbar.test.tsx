import { create } from "@bufbuild/protobuf"
import { render, screen } from "@testing-library/react"
import { createOvermindMock } from "overmind"
import { Provider } from "overmind-react"
import React from "react"
import { MemoryRouter } from "react-router-dom"
import { UserSchema } from "../../proto/qf/types_pb"
import NavBar from "../components/NavBar"
import { config } from "../overmind"


describe("Visibility when logged in", () => {
    const mockedOvermind = createOvermindMock(config, (state) => {
        state.self = create(UserSchema, {
            ID: BigInt(1),
            Name: "Test User",
            IsAdmin: true,
            // Set AvatarURL to a valid URL to avoid console errors
            // In production, we always have a non-empty AvatarURL
            AvatarURL: "https://example.com/avatar.jpg",
        })
    })

    beforeEach(() => {
        render(
            <Provider value={mockedOvermind}>
                <MemoryRouter>
                    <NavBar />
                </MemoryRouter>
            </Provider>
        )
    })

    it("Sign in is not visible when logged in", () => {
        const signIn = "Sign in with"
        const link = screen.getAllByRole("link")
        link.forEach(element => {
            expect(element.textContent).not.toEqual(signIn)
        })
    })

    it("When user is logged in, hamburger menu should appear", () => {
        const hamburger = "☰"
        const element = screen.getByText(hamburger)
        expect(element.tagName).toEqual("SPAN")
    })
})

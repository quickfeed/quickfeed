import NavBar from "../components/NavBar";
import React from "react";
import { UserSchema } from "../../proto/qf/types_pb";
import { createOvermindMock } from "overmind";
import { config } from "../overmind";
import { MemoryRouter } from "react-router-dom";
import { Provider } from "overmind-react";
import { render, screen } from "@testing-library/react";
import { create } from "@bufbuild/protobuf";
describe("Visibility when logged in", () => {
    const mockedOvermind = createOvermindMock(config, (state) => {
        state.self = create(UserSchema, {
            ID: BigInt(1),
            Name: "Test User",
            IsAdmin: true,
            AvatarURL: "https://example.com/avatar.jpg",
        });
    });
    beforeEach(() => {
        render(React.createElement(Provider, { value: mockedOvermind },
            React.createElement(MemoryRouter, null,
                React.createElement(NavBar, null))));
    });
    it("Sign in is not visible when logged in", () => {
        const signIn = "Sign in with";
        const link = screen.getAllByRole("link");
        link.forEach(element => {
            expect(element.textContent).not.toEqual(signIn);
        });
    });
    it("When user is logged in, hamburger menu should appear", () => {
        const hamburger = "☰";
        const element = screen.getByText(hamburger);
        expect(element.tagName).toEqual("SPAN");
    });
});

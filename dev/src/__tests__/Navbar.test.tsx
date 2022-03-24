import NavBar from "../components/NavBar"
import { configure, mount } from "enzyme"
import Adapter from "@wojtekmaj/enzyme-adapter-react-17"
import React from "react"
import { User } from "../../proto/ag/ag_pb"
import { createOvermindMock } from "overmind"
import { config } from "../overmind"
import { createMemoryHistory } from "history"
import { Router } from "react-router-dom"
import Enzyme from "enzyme"
import EnzymeAdapter from "@wojtekmaj/enzyme-adapter-react-17"
import { Provider } from "overmind-react"

configure({ adapter: new Adapter() });
Enzyme.configure( { adapter: new EnzymeAdapter() });
const history = createMemoryHistory()
const mockedOvermind = createOvermindMock(config, (state) => {
        state.self = new User().setId(1).setName("Test User")
})
const wrapped = mount(<Provider value={mockedOvermind}>
            <Router history={history}>
               <NavBar />
            </Router>
        </Provider>
    )

describe("Visibility when logged in", () => {
    it("When user is logged in, hamburger menu should appear", () => {
        const hamburger = "☰"
        expect(wrapped.find(".clickable").text()).toEqual(hamburger)
    })

    it("Sign in is not visible when logged in", () => {
        const signIn = "Sign in with"
        const link = wrapped.find("a").at(1)
        var exists = true
        if (link.text() !== signIn){
            exists = false
        }
        expect(exists).toBe(false)
    })
});

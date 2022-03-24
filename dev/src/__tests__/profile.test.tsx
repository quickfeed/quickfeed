import {User} from "../../proto/ag/ag_pb"
import { Provider } from "overmind-react"
import { createOvermindMock} from "overmind"
import { config } from "../overmind"
import Profile from "../components/profile/Profile"
import { Router } from "react-router-dom"
import {createMemoryHistory} from "history"
import { configure, render } from "enzyme"
import Adapter from "@wojtekmaj/enzyme-adapter-react-17"
import React from "react"
import { state } from "../overmind/state"

configure({ adapter: new Adapter() });
React.useLayoutEffect = React.useEffect 

describe("Profile", () =>{
    it("Renders with logged in user", () =>{
        const mockedOvermind = createOvermindMock(config, (state) => {
            state.self = new User().setId(1).setName("Test User")
        })
        const history = createMemoryHistory()
        const cheerio = render(
            <Provider value={mockedOvermind}>
                <Router history={history}>
                    <Profile />
                </Router>
            </Provider>
        )
        const loggedIn = mockedOvermind.state.isLoggedIn
        expect(loggedIn).toBe(true)
        expect(cheerio.find("h1").text()).toBe("Hi, Test User")
    });
    
    it("Logged in is false if the user is invalid", () =>{
        const mockedOvermind = createOvermindMock(config, (state) => {
            state.self = new User().setId(0)
        })
        const loggedIn = mockedOvermind.state.isLoggedIn
        expect(loggedIn).toBe(false)
    });
})

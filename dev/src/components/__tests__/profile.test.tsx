
import {User} from "../../../proto/ag/ag_pb";
import { Provider } from 'overmind-react'
import { createOvermindMock} from "overmind";
import { config } from "../../overmind";
import Profile from "../profile/Profile";
import { Router } from "react-router-dom";
import {createMemoryHistory} from "history";
import { configure, render } from "enzyme";
import Adapter from "@wojtekmaj/enzyme-adapter-react-17"
import React from "react";

configure({ adapter: new Adapter() });
React.useLayoutEffect = React.useEffect 



describe("Profile", () =>{
    it("renders with logged in user", () =>{
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

        var loggedIn = mockedOvermind.state.isLoggedIn
        expect(loggedIn).toBe(true)
        expect(cheerio.find("h1").text()).toBe("Hi, Test User")
    });
    
    it("Logged in is false if there is no user", () =>{
        const mockedOvermind = createOvermindMock(config, (state) => {
        })
        var loggedIn = mockedOvermind.state.isLoggedIn
        expect(loggedIn).toBe(false)
    });

})


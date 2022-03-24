import StudentPage from "../pages/StudentPage"
import {configure, mount} from "enzyme"
import {isStudent} from "../Helpers"
import React from "react"
import { createMemoryHistory } from "history"
import { Router } from "react-router-dom"
import Enzyme from "enzyme"
import EnzymeAdapter from "@wojtekmaj/enzyme-adapter-react-17"
import { Provider } from "overmind-react"
import { createOvermindMock } from "overmind";
import { User } from "../../proto/ag/ag_pb"



Enzyme.configure( { adapter: new EnzymeAdapter() });
const history = createMemoryHistory()

const mockedOvermind = createOvermindMock(config, (state) => {
        state.self = new User().setId(1).setName("Test User")
})
const wrapped = mount(<Provider value={mockedOvermind}>
            <Router history={history}>
                <StudentPage></StudentPage>
            </Router>
        </Provider>
    )

// const title = "Create Course"
// let wrapped = shallow(<CourseUti></CourseCreationInfo>);
// describe("Title should be equal to", () => {
//     it('Should render correctly', ()=> {
//         expect(wrapped).toMatchSnapshot();
//     });

//     it('Renders titles children', () => {
//         expect(wrapped.find('h1').text()).toEqual(title)
//     });
// });

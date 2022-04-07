import { updateEnrollment } from "../overmind/actions"
import { Enrollment, User } from "../../proto/ag/ag_pb"
import { createOvermindMock } from "overmind"
import { config } from "../overmind"
import { configure, render } from "enzyme"
import Adapter from "@wojtekmaj/enzyme-adapter-react-17"
import { createMemoryHistory } from "history"
import React from "react"
import Members from "../components/Members"
import { Route, Router } from "react-router"
import { Provider } from "overmind-react"

React.useLayoutEffect = React.useEffect

describe("UpdateEnrollment", () => {
    it("Pending student gets accecpted", () => {
        const user = new User().setId(1).setName("Test User").setStudentid("12345687")
        const mockedOvermind = createOvermindMock(config, (state) => {
            state.self = user
        })
        const enrollment = new Enrollment().setId(1).setCourseid(1).setStatus(1).setUser(user)
        window.confirm = jest.fn(() => true)
        updateEnrollment(mockedOvermind, { enrollment: enrollment, status: Enrollment.UserStatus.STUDENT })
        expect(enrollment.getStatus()).toEqual(2)
    })
    it("Demote teacher to student", () => {
        const user2 = new User().setId(1).setName("Test User").setStudentid("12345687")
        const mockedOvermind = createOvermindMock(config, (state) => {
            state.self = user2
        })
        const enrollment = new Enrollment().setId(1).setCourseid(1).setStatus(3).setUser(user2)
        window.confirm = jest.fn(() => true)
        updateEnrollment(mockedOvermind, { enrollment: enrollment, status: Enrollment.UserStatus.STUDENT })
        expect(enrollment.getStatus()).toEqual(2)
    })
    it("Promote student to teacher", () => {
        const user3 = new User().setId(1).setName("Test User").setStudentid("12345687")
        const mockedOvermind = createOvermindMock(config, (state) => {
            state.self = user3
        })
        const enrollment = new Enrollment().setId(1).setCourseid(1).setStatus(2).setUser(user3)
        window.confirm = jest.fn(() => true)
        updateEnrollment(mockedOvermind, { enrollment: enrollment, status: Enrollment.UserStatus.TEACHER })
        expect(enrollment.getStatus()).toEqual(3)
    })
})

configure({ adapter: new Adapter() })

describe("UpdateEnrollment in webpage", () => {
    it("If status is teacher, button should display demote", () => {
        const user = new User().setId(1).setName("Test User").setStudentid("6583969706").setEmail("test@gmail.com")
        const enrollment = new Enrollment().setId(2).setCourseid(1).setStatus(3).setUser(user).setSlipdaysremaining(3).setLastactivitydate("10 Mar").setTotalapproved(0)

        const mockedOvermind = createOvermindMock(config, (state) => {
            state.self = user
            state.activeCourse = 1
            state.courseEnrollments = { [1]: [enrollment] }
        })
        const history = createMemoryHistory()
        history.push("/course/1/members")

        React.useState = jest.fn().mockReturnValue("True")
        const wrapped = render(
            <Provider value={mockedOvermind}>
                <Router history={history} >
                    <Route path="/course/:id/members" component={Members} />
                </Router>
            </Provider>
        )
        expect(wrapped.find("i").first().text()).toEqual("Demote")
    })
    it("If status is student, button should display promote", () => {
        const user = new User().setId(1).setName("Test User").setStudentid("6583969706").setEmail("test@gmail.com")
        const enrollment = new Enrollment().setId(2).setCourseid(1).setStatus(2).setUser(user).setSlipdaysremaining(3).setLastactivitydate("10 Mar").setTotalapproved(0)

        const mockedOvermind = createOvermindMock(config, (state) => {
            state.self = user
            state.activeCourse = 1
            state.courseEnrollments = { [1]: [enrollment] }
        })
        const history = createMemoryHistory()
        history.push("/course/1/members")

        React.useState = jest.fn().mockReturnValue("True")
        const wrapped = render(
            <Provider value={mockedOvermind}>
                <Router history={history} >
                    <Route path="/course/:id/members" component={Members} />
                </Router>
            </Provider>
        )
        expect(wrapped.find("i").first().text()).toEqual("Promote")
    })
})

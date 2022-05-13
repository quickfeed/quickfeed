import { Enrollment, User } from "../../proto/ag/ag_pb"
import { createOvermindMock } from "overmind"
import { config } from "../overmind"
import { configure, mount, render } from "enzyme"
import Adapter from "@wojtekmaj/enzyme-adapter-react-17"
import { createMemoryHistory } from "history"
import React from "react"
import Members from "../components/Members"
import { Route, Router } from "react-router"
import { Provider } from "overmind-react"
import { MockGrpcManager } from "../MockGRPCManager"
import enzyme from "enzyme"

React.useLayoutEffect = React.useEffect

describe("UpdateEnrollment", () => {
    const mockedOvermind = createOvermindMock(config, {
        grpcMan: new MockGrpcManager()
    })
    it("Pending student gets accecpted", async () => {
        await mockedOvermind.actions.getEnrollmentsByCourse({ courseID: 2, statuses: [] })
        // This is a user with course status pending
        window.confirm = jest.fn(() => true)
        var enrollment = mockedOvermind.state.courseEnrollments[2][1]
        const status = Enrollment.UserStatus.STUDENT
        mockedOvermind.actions.updateEnrollment({ enrollment: enrollment, status: status })
        await expect(enrollment.getStatus()).toEqual(status)
    })

    it("Demote teacher to student", async () => {
        await mockedOvermind.actions.getEnrollmentsByCourse({ courseID: 2, statuses: [] })
        // This is a user with course status teacher
        window.confirm = jest.fn(() => true)
        var enrollment = mockedOvermind.state.courseEnrollments[2][0]
        const status = Enrollment.UserStatus.STUDENT
        mockedOvermind.actions.updateEnrollment({ enrollment: enrollment, status: status })
        expect(enrollment.getStatus()).toEqual(status)
    })

    it("Promote student to teacher", async () => {
        await mockedOvermind.actions.getEnrollmentsByCourse({ courseID: 1, statuses: [] })
        // This is a user with course status student
        window.confirm = jest.fn(() => true)
        var enrollment = mockedOvermind.state.courseEnrollments[1][0]
        var status = Enrollment.UserStatus.TEACHER
        mockedOvermind.actions.updateEnrollment({ enrollment: enrollment, status: status })
        expect(enrollment.getStatus()).toEqual(status)
    })
})

configure({ adapter: new Adapter() })

describe("UpdateEnrollment in webpage", () => {
    it("If status is teacher, button should display demote", () => {
        const user = new User().setId(1).setName("Test User").setStudentid("6583969706").setEmail("test@gmail.com")
        const enrollment = new Enrollment().setId(2).setCourseid(1).setStatus(3).setUser(user)
            .setSlipdaysremaining(3).setLastactivitydate("10 Mar").setTotalapproved(0)

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

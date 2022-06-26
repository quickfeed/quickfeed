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
import { MockGrpcManager } from "../MockGRPCManager"


React.useLayoutEffect = React.useEffect
configure({ adapter: new Adapter() })

describe("UpdateEnrollment", () => {
    const mockedOvermind = createOvermindMock(config, {
        grpcMan: new MockGrpcManager()
    })

    const updateEnrollmentTests: { desc: string, courseID: number, userID: number, want: Enrollment.UserStatus }[] = [
        // Refer to addLocalCourseStudent() in MockGRPCManager.ts for a list of available enrollments
        { desc: "Pending student gets accepted", courseID: 2, userID: 2, want: Enrollment.UserStatus.STUDENT },
        { desc: "Demote teacher to student", courseID: 2, userID: 1, want: Enrollment.UserStatus.STUDENT },
        { desc: "Promote student to teacher", courseID: 1, userID: 2, want: Enrollment.UserStatus.TEACHER },
    ]

    beforeAll(async () => {
        // Load enrollments into state before running tests
        await mockedOvermind.actions.getEnrollmentsByCourse({ courseID: 2, statuses: [] })
        await mockedOvermind.actions.getEnrollmentsByCourse({ courseID: 1, statuses: [] })
    })

    updateEnrollmentTests.forEach(({ desc, courseID, userID, want }) => {
        it(desc, async () => {
            window.confirm = jest.fn(() => true)
            const enrollment = mockedOvermind.state.courseEnrollments[courseID].find(e => e.getUserid() === userID)
            expect(enrollment).toBeDefined()
            mockedOvermind.actions.updateEnrollment({ enrollment: enrollment as Enrollment, status: want })
            expect((enrollment as Enrollment).getStatus()).toEqual(want)
        })
    })
})

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
        const enrollment = new Enrollment().setId(2).setCourseid(1).setStatus(2).setUser(user)
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
        expect(wrapped.find("i").first().text()).toEqual("Promote")
    })
})

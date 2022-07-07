import { Enrollment, User } from "../../proto/qf/qf_pb"
import { createOvermindMock } from "overmind"
import { config } from "../overmind"
import { createMemoryHistory } from "history"
import React from "react"
import Members from "../components/Members"
import { Route, Router } from "react-router"
import { Provider } from "overmind-react"
import { initializeOvermind } from "./TestHelpers"
import { render, screen } from "@testing-library/react"


describe("UpdateEnrollment", () => {
    const mockedOvermind = initializeOvermind({})

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

    test.each(updateEnrollmentTests)(`$desc`, async (test) => {
        const enrollment = mockedOvermind.state.courseEnrollments[test.courseID].find(e => e.userid === test.userID)
        if (enrollment === undefined) {
            throw new Error(`No enrollment found for user ${test.userID} in course ${test.courseID}`)
        }
        mockedOvermind.actions.setActiveCourse(test.courseID)
        window.confirm = jest.fn(() => true)
        await mockedOvermind.actions.updateEnrollment({ enrollment: enrollment, status: test.want })
        expect(enrollment.status).toEqual(test.want)
    })
})

describe("UpdateEnrollment in webpage", () => {
    it("If status is teacher, button should display demote", () => {
        const user = new User().setId(1).setName("Test User").setStudentid("6583969706").setEmail("test@gmail.com")
        const enrollment = new Enrollment().setId(2).setCourseid(1).setStatus(3).setUser(user)
            .setSlipdaysremaining(3).setLastactivitydate("10 Mar").setTotalapproved(0).toObject()

        const mockedOvermind = createOvermindMock(config, (state) => {
            state.self = user.toObject()
            state.activeCourse = 1
            state.courseEnrollments = { [1]: [enrollment] }
        })
        const history = createMemoryHistory()
        history.push("/course/1/members")

        React.useState = jest.fn().mockReturnValue("True")
        render(
            <Provider value={mockedOvermind}>
                <Router history={history} >
                    <Route path="/course/:id/members" component={Members} />
                </Router>
            </Provider>
        )
        expect(screen.getByText("Demote")).toBeTruthy()
        expect(screen.queryByText("Promote")).toBeFalsy()
    })

    it("If status is student, button should display promote", () => {
        const user = new User().setId(1).setName("Test User").setStudentid("6583969706").setEmail("test@gmail.com")
        const enrollment = new Enrollment().setId(2).setCourseid(1).setStatus(2).setUser(user)
            .setSlipdaysremaining(3).setLastactivitydate("10 Mar").setTotalapproved(0).toObject()

        const mockedOvermind = createOvermindMock(config, (state) => {
            state.self = user.toObject()
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
        expect(screen.getByText("Promote")).toBeTruthy()
        expect(screen.queryByText("Demote")).toBeFalsy()
    })
})

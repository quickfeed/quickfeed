import { Enrollment, Enrollment_UserStatus, EnrollmentSchema, EnrollmentsSchema, UserSchema } from "../../proto/qf/types_pb"
import { createOvermindMock } from "overmind"
import { config } from "../overmind"
import { createMemoryHistory } from "history"
import React from "react"
import Members from "../components/Members"
import { Route, Router } from "react-router"
import { Provider } from "overmind-react"
import { render, screen } from "@testing-library/react"
import { MockData } from "./mock_data/mockData"
import { VoidSchema } from "../../proto/qf/requests_pb"
import { initializeOvermind, mock } from "./TestHelpers"
import { ApiClient } from "../overmind/effects"
import { create } from "@bufbuild/protobuf"
import { timestampFromDate } from "@bufbuild/protobuf/wkt"


describe("UpdateEnrollment", () => {
    const api = new ApiClient()
    api.client = {
        ...api.client,
        getEnrollments: mock("getEnrollments", async (request) => {
            const enrollments: Enrollment[] = []
            MockData.mockedEnrollments().enrollments.forEach(e => {
                if (request.FetchMode?.case === "courseID") {
                    if (e.courseID === request.FetchMode.value) {
                        enrollments.push(e)
                    }
                } else if (request.FetchMode?.case === "userID") {
                    if (e.userID === request.FetchMode.value) {
                        enrollments.push(e)
                    }
                } else {
                    enrollments.push(e)
                }
            })
            return { message: create(EnrollmentsSchema, { enrollments }), error: null }
        }),
        updateEnrollments: mock("updateEnrollments", async (request) => {
            const enrollments = request.enrollments ?? []
            if (enrollments.length === 0) {
                return { message: create(VoidSchema), error: null }
            }
            enrollments.forEach(e => {
                const enrollment = MockData.mockedEnrollments().enrollments.find(en => en.ID === e.ID)
                if (!enrollment || e.status === undefined) {
                    return
                }
                enrollment.status = e.status
            })
            return { message: create(VoidSchema), error: null }
        }),
    }
    const mockedOvermind = initializeOvermind({}, api)

    const updateEnrollmentTests: { desc: string, courseID: bigint, userID: bigint, want: Enrollment_UserStatus }[] = [
        // Refer to addLocalCourseStudent() in MockGRPCManager.ts for a list of available enrollments
        { desc: "Pending student gets accepted", courseID: BigInt(2), userID: BigInt(2), want: Enrollment_UserStatus.STUDENT },
        { desc: "Demote teacher to student", courseID: BigInt(2), userID: BigInt(1), want: Enrollment_UserStatus.STUDENT },
        { desc: "Promote student to teacher", courseID: BigInt(1), userID: BigInt(2), want: Enrollment_UserStatus.TEACHER },
    ]



    beforeAll(async () => {
        // mock getEnrollmentsByCourse() to load enrollments into state
        // Load enrollments into state before running tests
        await mockedOvermind.actions.getEnrollmentsByCourse({ courseID: BigInt(2), statuses: [] })
        await mockedOvermind.actions.getEnrollmentsByCourse({ courseID: BigInt(1), statuses: [] })
    })

    test.each(updateEnrollmentTests)(`$desc`, async (test) => {
        const enrollment = mockedOvermind.state.courseEnrollments[test.courseID.toString()].find(e => e.userID === test.userID)
        if (!enrollment) {
            throw new Error(`No enrollment found for user ${test.userID} in course ${test.courseID}`)
        }
        mockedOvermind.actions.setActiveCourse(test.courseID)
        window.confirm = jest.fn(() => true)
        await mockedOvermind.actions.updateEnrollment({ enrollment, status: test.want })
        expect(enrollment.status).toEqual(test.want)
    })
})

describe("UpdateEnrollment in webpage", () => {
    it("If status is teacher, button should display demote", () => {
        const user = create(UserSchema, { ID: BigInt(1), Name: "Test User", StudentID: "6583969706", Email: "test@gmail.com" })
        const enrollment = create(EnrollmentSchema, {
            ID: BigInt(2),
            courseID: BigInt(1),
            status: 3,
            user,
            slipDaysRemaining: 3,
            lastActivityDate: timestampFromDate(new Date(2022, 3, 10)),
            totalApproved: BigInt(0),
        })

        const mockedOvermind = createOvermindMock(config, (state) => {
            state.self = user
            state.activeCourse = BigInt(1)
            state.courseEnrollments = { "1": [enrollment] }
        })
        const history = createMemoryHistory()
        history.push("/course/1/members")

        render(
            <Provider value={mockedOvermind}>
                <Router history={history} >
                    <Route path="/course/:id/members" component={Members} />
                </Router>
            </Provider>
        )

        const editButton = screen.getByText("Edit")
        editButton.click()

        expect(screen.getByText("Demote")).toBeTruthy()
        expect(screen.queryByText("Promote")).toBeFalsy()
    })

    it("If status is student, button should display promote", () => {
        const user = create(UserSchema, {
            ID: BigInt(1),
            Name: "Test User",
            StudentID: "6583969706",
            Email: "test@gmail.com"
        })
        const enrollment = create(EnrollmentSchema, {
            ID: BigInt(2),
            courseID: BigInt(1),
            status: 2,
            user,
            slipDaysRemaining: 3,
            lastActivityDate: timestampFromDate(new Date(2022, 3, 10)),
            totalApproved: BigInt(0),
        })
        const mockedOvermind = createOvermindMock(config, (state) => {
            state.self = user
            state.activeCourse = BigInt(1)
            state.courseEnrollments = { "1": [enrollment] }
        })
        const history = createMemoryHistory()
        history.push("/course/1/members")

        render(
            <Provider value={mockedOvermind}>
                <Router history={history} >
                    <Route path="/course/:id/members" component={Members} />
                </Router>
            </Provider>
        )

        const editButton = screen.getByText("Edit")
        editButton.click()

        expect(screen.getByText("Promote")).toBeTruthy()
        expect(screen.queryByText("Demote")).toBeFalsy()
    })
})

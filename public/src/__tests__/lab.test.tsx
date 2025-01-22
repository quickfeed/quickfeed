import React from "react"
import { Assignment, Assignments, Submissions, User } from "../../proto/qf/types_pb"
import { createMemoryHistory } from "history"
import { Route, Router } from "react-router-dom"
import { Provider } from "overmind-react"
import { act, render, screen } from "@testing-library/react"
import Lab from "../components/Lab"
import { MockData } from "./mock_data/mockData"
import { ApiClient } from "../overmind/effects"
import { initializeOvermind, mock } from "./TestHelpers"
import { ConnectError } from "@bufbuild/connect"


describe("Labs", () => {
const api = new ApiClient()
    api.client = {
        ...api.client,
        getAssignments: mock("getAssignments", async (request) => {
            const course = MockData.mockedCourses().find(c => c.ID === request.courseID)
            if (!course) {
                return { message: new Assignments(), error: new ConnectError("course not found") }
            }
            const assignments = MockData.mockedAssignments().filter(a => a.CourseID === request.courseID)
            return { message: new Assignments({assignments}), error: null }
        }),
        getSubmissions: mock("getSubmissions", async (request) => {
            const course = MockData.mockedCourses().find(c => c.ID === request.CourseID)
            if (!course) {
                return { message: new Submissions(), error: new ConnectError("course not found") }
            }
            const submissions = MockData.mockedSubmissions().submissions.filter(s => s.userID === request.FetchMode?.value)
            return { message: new Submissions({submissions}), error: null }
        })
        
    }
    const history = createMemoryHistory()
    let mockedOvermind = initializeOvermind({}, api)

    beforeEach(() => {
        mockedOvermind = initializeOvermind({
        self: new User({
            ID: BigInt(1),
            Name: "Test User",
            IsAdmin: true,
        }),
        enrollments: MockData.mockedEnrollments().enrollments,
        courses: MockData.mockedCourses(),
        repositories: MockData.mockedRepositories()
    }, api)
        history.push("/course/1/lab/1")
        render(
            <Provider value={mockedOvermind}>
                <Router history={history}>
                    <Route path="/course/:id/lab/:lab">
                        <Lab />
                    </Route>
                </Router>
            </Provider>
        )
    })

    const assertContent = (content: string) => {
        const element = screen.getByText(content)
        expect(element).toBeDefined()
    }

    const fetchAssignments = async () => {
        await act(async () => {
            await mockedOvermind.actions.getAssignments()
        })
    }

    test("No assignment", () => {
        // Lab should show "Assignment not found" if the assignment is not found
        assertContent("Assignment not found")
    })

    test("No submission", async () => {
        // Lab should show "Assignment not found" if the assignment is not found
        assertContent("Assignment not found")
        fetchAssignments()
        expect(mockedOvermind.state.assignments["1"]).toBeDefined()
        // after the assignment is fetched it should show "No submission found"
        assertContent("No submission found")
    })

    test("Submission found", async () => {
        // TODO:  The previous tests are covered here, we could remove them
        // Lab should show "Assignment not found" if the assignment is not found
        assertContent("Assignment not found")
        fetchAssignments()
        expect(mockedOvermind.state.assignments["1"]).toBeDefined()
        // after the assignment is fetched it should show "No submission found"
        assertContent("No submission found")
        // fetch submissions
        await act(async () => {
            await mockedOvermind.actions.getUserSubmissions(1n)
        })
        const submissions = mockedOvermind.state.submissions.ForAssignment(new Assignment({ID: 1n, CourseID: 1n}))
        expect(submissions).toBeDefined()
        expect(submissions.length).toBe(1)

        // after the submission is fetched it should show the submission
        // we specifically check for the build log
        assertContent("Build log for submission 1")
        // trigger a receive event (this is what happens when a submission is received via streaming)
        await act(async () => {
            const modifiedSubmission = submissions[0].clone()
            modifiedSubmission.BuildInfo!.BuildLog = "This is a build log"
            mockedOvermind.actions.receiveSubmission(modifiedSubmission)
        })
        // verify that the updated submission is shown
        assertContent("This is a build log")
    })
})

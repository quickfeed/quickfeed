import React from "react"
import { AssignmentSchema, AssignmentsSchema, SubmissionSchema, SubmissionsSchema, UserSchema } from "../../proto/qf/types_pb"
import { createMemoryHistory } from "history"
import { Route, Router } from "react-router-dom"
import { Provider } from "overmind-react"
import { act, render, screen } from "@testing-library/react"
import Lab from "../components/Lab"
import { MockData } from "./mock_data/mockData"
import { ApiClient } from "../overmind/effects"
import { initializeOvermind, mock } from "./TestHelpers"
import { create, clone } from "@bufbuild/protobuf"
import { ConnectError } from "@connectrpc/connect"
import { KnownMessage } from "../components/CenteredMessage"

describe("Lab view correctly re-renders on state change", () => {
    const api = new ApiClient()
    api.client = {
        ...api.client,
        getAssignments: mock("getAssignments", async (request) => { // skipcq: JS-0116
            const course = MockData.mockedCourses().find(c => c.ID === request.courseID)
            if (!course) {
                return { message: create(AssignmentsSchema), error: new ConnectError("course not found") }
            }
            const assignments = MockData.mockedAssignments().filter(a => a.CourseID === request.courseID)
            return { message: create(AssignmentsSchema, { assignments }), error: null }
        }),
        getSubmissions: mock("getSubmissions", async (request) => { // skipcq: JS-0116
            const course = MockData.mockedCourses().find(c => c.ID === request.CourseID)
            if (!course) {
                return { message: create(SubmissionsSchema), error: new ConnectError("course not found") }
            }
            const submissions = MockData.mockedSubmissions().submissions.filter(s => s.userID === request.FetchMode?.value)
            return { message: create(SubmissionsSchema, { submissions }), error: null }
        })

    }
    const history = createMemoryHistory()
    let mockedOvermind = initializeOvermind({}, api)

    beforeEach(() => {
        mockedOvermind = initializeOvermind({
            self: create(UserSchema, {
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
        assertContent(KnownMessage.NoAssignment)
    })

    test("No submission", async () => {
        // Lab should show "Assignment not found" if the assignment is not found
        assertContent(KnownMessage.NoAssignment)
        await fetchAssignments()
        expect(mockedOvermind.state.assignments["1"]).toBeDefined()
        // after the assignment is fetched it should show "Select a submission from the results table"
        assertContent(KnownMessage.NoSubmission)
    })

    test("Submission found", async () => {
        // TODO:  The previous tests are covered here, we could remove them
        // Lab should show "Assignment not found" if the assignment is not found
        assertContent(KnownMessage.NoAssignment)
        await fetchAssignments()
        expect(mockedOvermind.state.assignments["1"]).toBeDefined()
        // after the assignment is fetched it should show "Select a submission from the results table"
        assertContent(KnownMessage.NoSubmission)

        // fetch submissions for the user
        await act(async () => {
            await mockedOvermind.actions.getUserSubmissions(1n)
        })
        const submissions = mockedOvermind.state.submissions.ForAssignment(create(AssignmentSchema, { ID: 1n, CourseID: 1n }))
        expect(submissions).toBeDefined()
        expect(submissions.length).toBe(1)

        // after the submission is fetched it should show the submission
        // we specifically check for the build log
        assertContent("Build log for submission 1")
        // trigger a receive event (this is what happens when a submission is received via streaming)
        await act(async () => {
            const modifiedSubmission = clone(SubmissionSchema, submissions[0])
            if (modifiedSubmission.BuildInfo) {
                modifiedSubmission.BuildInfo.BuildLog = "This is a build log"
            }
            mockedOvermind.actions.receiveSubmission(modifiedSubmission)
        })
        // verify that the updated submission is shown
        assertContent("This is a build log")
    })
})

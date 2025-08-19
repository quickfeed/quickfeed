import React from "react";
import { AssignmentSchema, AssignmentsSchema, SubmissionSchema, SubmissionsSchema, UserSchema } from "../../proto/qf/types_pb";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { Provider } from "overmind-react";
import { act, render, screen } from "@testing-library/react";
import Lab from "../components/Lab";
import { MockData } from "./mock_data/mockData";
import { ApiClient } from "../overmind/namespaces/global/effects";
import { initializeOvermind, mock } from "./TestHelpers";
import { create, clone } from "@bufbuild/protobuf";
import { ConnectError } from "@connectrpc/connect";
import { KnownMessage } from "../components/CenteredMessage";
describe("Lab view correctly re-renders on state change", () => {
    const api = new ApiClient();
    api.client = {
        ...api.client,
        getAssignments: mock("getAssignments", async (request) => {
            const course = MockData.mockedCourses().find(c => c.ID === request.courseID);
            if (!course) {
                return { message: create(AssignmentsSchema), error: new ConnectError("course not found") };
            }
            const assignments = MockData.mockedAssignments().filter(a => a.CourseID === request.courseID);
            return { message: create(AssignmentsSchema, { assignments }), error: null };
        }),
        getSubmissions: mock("getSubmissions", async (request) => {
            const course = MockData.mockedCourses().find(c => c.ID === request.CourseID);
            if (!course) {
                return { message: create(SubmissionsSchema), error: new ConnectError("course not found") };
            }
            const submissions = MockData.mockedSubmissions().submissions.filter(s => s.userID === request.FetchMode?.value);
            return { message: create(SubmissionsSchema, { submissions }), error: null };
        })
    };
    let mockedOvermind = initializeOvermind({}, api);
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
        }, api);
        render(React.createElement(Provider, { value: mockedOvermind },
            React.createElement(MemoryRouter, { initialEntries: ["/course/1/lab/1"] },
                React.createElement(Routes, null,
                    React.createElement(Route, { path: "/course/:id/lab/:lab", element: React.createElement(Lab, null) })))));
    });
    const assertContent = (content) => {
        const element = screen.getByText(content);
        expect(element).toBeDefined();
    };
    const fetchAssignments = async () => {
        await act(async () => {
            await mockedOvermind.actions.global.getAssignments();
        });
    };
    test("No assignment", () => {
        assertContent(KnownMessage.StudentNoAssignment);
    });
    test("No submission", async () => {
        assertContent(KnownMessage.StudentNoAssignment);
        await fetchAssignments();
        expect(mockedOvermind.state.assignments["1"]).toBeDefined();
        assertContent(KnownMessage.StudentNoSubmission);
    });
    test("Submission found", async () => {
        assertContent(KnownMessage.StudentNoAssignment);
        await fetchAssignments();
        expect(mockedOvermind.state.assignments["1"]).toBeDefined();
        assertContent(KnownMessage.StudentNoSubmission);
        await act(async () => {
            await mockedOvermind.actions.global.getUserSubmissions(1n);
        });
        const submissions = mockedOvermind.state.submissions.ForAssignment(create(AssignmentSchema, { ID: 1n, CourseID: 1n }));
        expect(submissions).toBeDefined();
        expect(submissions.length).toBe(1);
        assertContent("Build log for submission 1");
        act(() => {
            const modifiedSubmission = clone(SubmissionSchema, submissions[0]);
            if (modifiedSubmission.BuildInfo) {
                modifiedSubmission.BuildInfo.BuildLog = "This is a build log";
            }
            mockedOvermind.actions.global.receiveSubmission(modifiedSubmission);
        });
        assertContent("This is a build log");
    });
});

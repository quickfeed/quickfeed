import { CourseSchema, Enrollment_UserStatus, EnrollmentSchema, UserSchema } from "../../proto/qf/types_pb";
import { createOvermindMock } from "overmind";
import { config } from "../overmind";
import React, { act } from "react";
import Members from "../components/Members";
import { Route, MemoryRouter, Routes } from "react-router-dom";
import { Provider } from "overmind-react";
import { render, screen } from "@testing-library/react";
import { MockData } from "./mock_data/mockData";
import { VoidSchema } from "../../proto/qf/requests_pb";
import { initializeOvermind, mock } from "./TestHelpers";
import { ApiClient } from "../overmind/namespaces/global/effects";
import { create } from "@bufbuild/protobuf";
import { timestampFromDate } from "@bufbuild/protobuf/wkt";
import { ConnectError } from "@connectrpc/connect";
describe("UpdateEnrollment", () => {
    const api = new ApiClient();
    api.client = {
        ...api.client,
        getCourse: mock("getCourse", async (request) => {
            const course = MockData.mockedCourses().find(c => c.ID === request.courseID);
            if (!course) {
                return { message: create(CourseSchema), error: new ConnectError("course not found") };
            }
            course.enrollments = MockData.mockedEnrollments().enrollments.filter(e => e.courseID === request.courseID);
            return { message: course, error: null };
        }),
        updateEnrollments: mock("updateEnrollments", async (request) => {
            const enrollments = request.enrollments ?? [];
            if (enrollments.length === 0) {
                return { message: create(VoidSchema), error: null };
            }
            enrollments.forEach(e => {
                const enrollment = MockData.mockedEnrollments().enrollments.find(en => en.ID === e.ID);
                if (!enrollment || e.status === undefined) {
                    return;
                }
                enrollment.status = e.status;
            });
            return { message: create(VoidSchema), error: null };
        }),
    };
    const mockedOvermind = initializeOvermind({}, api);
    const updateEnrollmentTests = [
        { desc: "Pending student gets accepted", courseID: BigInt(2), userID: BigInt(2), want: Enrollment_UserStatus.STUDENT },
        { desc: "Demote teacher to student", courseID: BigInt(2), userID: BigInt(1), want: Enrollment_UserStatus.STUDENT },
        { desc: "Promote student to teacher", courseID: BigInt(1), userID: BigInt(2), want: Enrollment_UserStatus.TEACHER },
    ];
    beforeAll(async () => {
        await mockedOvermind.actions.global.getCourseData({ courseID: BigInt(2) });
        await mockedOvermind.actions.global.getCourseData({ courseID: BigInt(1) });
    });
    test.each(updateEnrollmentTests)(`$desc`, async (test) => {
        const enrollment = mockedOvermind.state.courseEnrollments[test.courseID.toString()].find(e => e.userID === test.userID);
        if (!enrollment) {
            throw new Error(`No enrollment found for user ${test.userID} in course ${test.courseID}`);
        }
        mockedOvermind.actions.global.setActiveCourse(test.courseID);
        window.confirm = jest.fn(() => true);
        await mockedOvermind.actions.global.updateEnrollment({ enrollment, status: test.want });
        expect(enrollment.status).toEqual(test.want);
    });
});
describe("UpdateEnrollment in webpage", () => {
    it("If status is teacher, button should display demote", () => {
        const user = create(UserSchema, { ID: BigInt(1), Name: "Test User", StudentID: "6583969706", Email: "test@gmail.com" });
        const enrollment = create(EnrollmentSchema, {
            ID: BigInt(2),
            courseID: BigInt(1),
            status: Enrollment_UserStatus.TEACHER,
            user,
            slipDaysRemaining: 3,
            lastActivityDate: timestampFromDate(new Date(2022, 3, 10)),
            totalApproved: BigInt(0),
        });
        const mockedOvermind = createOvermindMock(config, (state) => {
            state.self = user;
            state.activeCourse = BigInt(1);
            state.courseEnrollments = { "1": [enrollment] };
        });
        render(React.createElement(Provider, { value: mockedOvermind },
            React.createElement(MemoryRouter, { initialEntries: ["/course/1/members"] },
                React.createElement(Routes, null,
                    React.createElement(Route, { path: "/course/:id/members", element: React.createElement(Members, null) })))));
        const editButton = screen.getByText("Edit");
        expect(editButton).toBeTruthy();
        act(() => {
            editButton.click();
        });
        expect(screen.getByText("Demote")).toBeTruthy();
        expect(screen.queryByText("Promote")).toBeFalsy();
    });
    it("If status is student, button should display promote", () => {
        const user = create(UserSchema, {
            ID: BigInt(1),
            Name: "Test User",
            StudentID: "6583969706",
            Email: "test@gmail.com"
        });
        const enrollment = create(EnrollmentSchema, {
            ID: BigInt(2),
            courseID: BigInt(1),
            status: Enrollment_UserStatus.STUDENT,
            user,
            slipDaysRemaining: 3,
            lastActivityDate: timestampFromDate(new Date(2022, 3, 10)),
            totalApproved: BigInt(0),
        });
        const mockedOvermind = createOvermindMock(config, (state) => {
            state.self = user;
            state.activeCourse = BigInt(1);
            state.courseEnrollments = { "1": [enrollment] };
        });
        render(React.createElement(Provider, { value: mockedOvermind },
            React.createElement(MemoryRouter, { initialEntries: ["/course/1/members"] },
                React.createElement(Routes, null,
                    React.createElement(Route, { path: "/course/:id/members", element: React.createElement(Members, null) })))));
        const editButton = screen.getByText("Edit");
        act(() => {
            editButton.click();
        });
        expect(screen.getByText("Promote")).toBeTruthy();
        expect(screen.queryByText("Demote")).toBeFalsy();
    });
});

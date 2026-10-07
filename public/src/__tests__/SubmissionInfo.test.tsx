import { create } from "@bufbuild/protobuf"
import { render, screen } from "@testing-library/react"
import { Provider } from "overmind-react"
import { BuildInfoSchema, RunStatus } from "../../proto/kit/score/score_pb"
import { AssignmentSchema, Enrollment_UserStatus, EnrollmentSchema, GradeSchema, Submission_Status, SubmissionSchema, TestInfoSchema } from "../../proto/qf/types_pb"
import SubmissionInfo from "../components/submissions/SubmissionInfo"
import { initializeOvermind } from "./TestHelpers"

describe("SubmissionInfo", () => {
    it.each([
        {
            name: "environment failure",
            status: RunStatus.NO_SCORES,
            scoreText: "This run did not update the score.",
        },
        {
            name: "compilation failure",
            status: RunStatus.BUILD_FAILURE,
            scoreText: "The score was recorded as zero.",
        },
    ])("explains the score after a $name", ({ status, scoreText }) => {
        const courseID = 1n
        const userID = 2n
        const overmind = initializeOvermind({
            enrollments: [create(EnrollmentSchema, { courseID, userID })],
        })
        const assignment = create(AssignmentSchema, { ID: 3n, CourseID: courseID, name: "lab1" })
        const submission = create(SubmissionSchema, {
            AssignmentID: assignment.ID,
            userID,
            BuildInfo: create(BuildInfoSchema, { Status: status }),
        })

        render(
            <Provider value={overmind}>
                <SubmissionInfo submission={submission} assignment={assignment} />
            </Provider>
        )

        expect(screen.getByText(new RegExp(scoreText))).toBeDefined()
        expect(screen.queryByText(/last successful run/)).toBeNull()
    })

    describe("for an assignment with tests and reviews", () => {
        const courseID = 1n
        const userID = 2n
        const assignment = create(AssignmentSchema, {
            ID: 3n,
            CourseID: courseID,
            name: "lab1",
            reviewers: 1,
            reviewWeight: 40,
            ExpectedTests: [create(TestInfoSchema, { TestName: "TestA", MaxScore: 1, Weight: 1 })],
        })
        const renderInfo = (status: Enrollment_UserStatus, grade: Submission_Status) => {
            const overmind = initializeOvermind({
                activeCourse: courseID,
                enrollments: [create(EnrollmentSchema, { courseID, userID, status })],
            })
            const submission = create(SubmissionSchema, {
                AssignmentID: assignment.ID,
                userID,
                testScore: 90,
                reviewScore: 70,
                score: 82,
                Grades: [create(GradeSchema, { UserID: userID, Status: grade })],
            })
            render(
                <Provider value={overmind}>
                    <SubmissionInfo submission={submission} assignment={assignment} />
                </Provider>
            )
        }

        it("shows the test, review, and total scores with their weights", () => {
            renderInfo(Enrollment_UserStatus.TEACHER, Submission_Status.NONE)
            expect(screen.getByText("Tests Passed")).toBeDefined()
            expect(screen.getByText("90% (weight 60%)")).toBeDefined()
            expect(screen.getByText("70% (weight 40%)")).toBeDefined()
            expect(screen.getByText("82%")).toBeDefined()
        })

        it("hides the review score from a student until the submission is graded", () => {
            renderInfo(Enrollment_UserStatus.STUDENT, Submission_Status.NONE)
            expect(screen.getByText("90% (weight 60%)")).toBeDefined()
            expect(screen.getByText("Not graded yet")).toBeDefined()
            expect(screen.queryByText("Total score")).toBeNull()
        })

        it("shows a student the review score once the submission is graded", () => {
            renderInfo(Enrollment_UserStatus.STUDENT, Submission_Status.APPROVED)
            expect(screen.getByText("70% (weight 40%)")).toBeDefined()
            expect(screen.getByText("82%")).toBeDefined()
        })
    })
})

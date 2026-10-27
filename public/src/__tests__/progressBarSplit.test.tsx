import { create } from "@bufbuild/protobuf"
import { render } from "@testing-library/react"
import { Provider } from "overmind-react"
import { AssignmentSchema, GradeSchema, Submission_Status, SubmissionSchema, TestInfoSchema, UserSchema } from "../../proto/qf/types_pb"
import ProgressBar from "../components/ProgressBar"
import { SubmissionsForUser } from "../Helpers"
import { initializeOvermind } from "./TestHelpers"

describe("ProgressBar for an assignment with tests and reviews", () => {
    const userID = 1n
    const assignment = create(AssignmentSchema, {
        ID: 3n,
        scoreLimit: 90,
        reviewers: 1,
        reviewWeight: 20,
        ExpectedTests: [create(TestInfoSchema, { TestName: "TestA", MaxScore: 1, Weight: 1 })],
    })
    const renderBar = (status: Submission_Status, split: boolean, score = 52) => {
        const submission = create(SubmissionSchema, {
            AssignmentID: assignment.ID,
            userID,
            testScore: 40,
            reviewScore: score > 0 ? 100 : 0,
            score,
            Grades: [create(GradeSchema, { UserID: userID, Status: status })],
        })
        const submissions = new SubmissionsForUser()
        submissions.setSubmissions(1n, "USER", [submission])
        const overmind = initializeOvermind({
            assignments: { "1": [assignment] },
            submissions,
            self: create(UserSchema, { ID: userID }),
        })
        const { container } = render(
            <Provider value={overmind}>
                <ProgressBar courseID="1" submission={submission} split={split} />
            </Provider>
        )
        return container
    }
    // segments returns the bar's segments as [width, text] pairs.
    const segments = (container: HTMLElement) =>
        Array.from(container.querySelectorAll<HTMLElement>(".relative > div")).map(div => [div.style.width, div.textContent])

    it("splits a graded submission's score into its test and review parts", () => {
        const container = renderBar(Submission_Status.APPROVED, true)
        expect(segments(container)).toEqual([["32%", "Tests 32 %"], ["20%", "Review 20 %"]])
    })

    it("shows the remaining score to pass only until the submission is approved", () => {
        const container = renderBar(Submission_Status.REVISION, true)
        expect(segments(container)).toEqual([["32%", "Tests 32 %"], ["20%", "Review 20 %"], ["38%", "38 %"]])
    })

    it("shows a student the tests' part of an ungraded submission, and the review's part as pending", () => {
        // The server hides the score of such a submission as 0.
        const container = renderBar(Submission_Status.NONE, true, 0)
        expect(segments(container)).toEqual([["32%", "Tests 32 %"], ["20%", "Review pending"]])
    })

    it("shows a student the review's part as pending without splitting the bar", () => {
        const container = renderBar(Submission_Status.NONE, false, 0)
        expect(segments(container)).toEqual([["32%", "Tests 32 %"], ["20%", "Review pending"]])
    })

    it("does not split unless asked to", () => {
        const container = renderBar(Submission_Status.APPROVED, false)
        expect(segments(container)).toEqual([["52%", "52 %"]])
    })
})

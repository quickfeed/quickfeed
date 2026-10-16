import { create } from "@bufbuild/protobuf"
import { AssignmentSchema, Submission_Status, SubmissionSchema, TestInfoSchema } from "../../proto/qf/types_pb"
import { assignmentStatusText, runsTests } from "../Helpers"

describe("runsTests", () => {
    const tests = [create(TestInfoSchema, { TestName: "TestA", MaxScore: 1, Weight: 1 })]
    it.each([
        { name: "tests only", reviewers: 0, ExpectedTests: tests, want: true },
        { name: "no reviewers and no tests", reviewers: 0, ExpectedTests: [], want: true },
        { name: "reviews only", reviewers: 1, ExpectedTests: [], want: false },
        { name: "tests and reviews", reviewers: 1, ExpectedTests: tests, want: true },
    ])("returns $want for an assignment with $name", ({ reviewers, ExpectedTests, want }) => {
        expect(runsTests(create(AssignmentSchema, { reviewers, ExpectedTests }))).toBe(want)
    })
})

describe("assignmentStatusText", () => {
    it.each([
        { name: "awaits review of a reviewed submission", reviewers: 1, score: 0, want: "Awaiting review" },
        { name: "asks for the score limit of a tested submission", reviewers: 0, score: 50, want: "Need 80% score for approval" },
        { name: "awaits approval of a tested submission", reviewers: 0, score: 90, want: "Awaiting approval" },
    ])("$name", ({ reviewers, score, want }) => {
        const assignment = create(AssignmentSchema, { reviewers, scoreLimit: 80 })
        const submission = create(SubmissionSchema, { score })
        expect(assignmentStatusText(assignment, submission, Submission_Status.NONE)).toBe(want)
    })
})

import { create } from "@bufbuild/protobuf"
import { AssignmentSchema, TestInfoSchema } from "../../proto/qf/types_pb"
import { runsTests } from "../Helpers"

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

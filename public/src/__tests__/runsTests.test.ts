import { create } from "@bufbuild/protobuf"
import { describe, expect, it } from "vitest"
import { AssignmentSchema, TestInfoSchema } from "../../proto/qf/types_pb"
import { runsTests } from "../Helpers"

describe("runsTests", () => {
    const expectedTests = [create(TestInfoSchema, { TestName: "TestA" })]
    const cases = [
        { name: "tests only", reviewers: 0, tests: expectedTests, want: true },
        { name: "tests and review", reviewers: 1, tests: expectedTests, want: true },
        { name: "no tests.json, no review", reviewers: 0, tests: [], want: true },
        { name: "review only", reviewers: 2, tests: [], want: false },
    ]
    for (const c of cases) {
        it(c.name, () => {
            const assignment = create(AssignmentSchema, { reviewers: c.reviewers, ExpectedTests: c.tests })
            expect(runsTests(assignment)).toBe(c.want)
        })
    }
})

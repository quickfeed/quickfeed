import { create } from "@bufbuild/protobuf"
import { cleanup, fireEvent, render, screen } from "@testing-library/react"
import { afterEach, describe, expect, it } from "vitest"
import { ScoreSchema, TestStatus } from "../../proto/kit/score/score_pb"
import { SubmissionSchema } from "../../proto/qf/types_pb"
import SubmissionScores from "../components/submissions/SubmissionScores"

afterEach(() => {
    cleanup()
    window.history.replaceState(null, "", window.location.pathname)
})

const scores = [
    create(ScoreSchema, {
        ID: 1n, TestName: "TestStack", Score: 5, MaxScore: 5, Weight: 1,
        Status: TestStatus.PASSED, Elapsed: 0.02,
        TestOutput: "stack: pushing 3 elements",
    }),
    create(ScoreSchema, {
        ID: 2n, TestName: "TestQueue", Score: 2, MaxScore: 5, Weight: 2,
        Status: TestStatus.FAILED, Elapsed: 0.01,
        TestDetails: "queue_test.go:44: Pop() = <nil>, want: x",
        TestOutput: "queue: dequeue returned nothing",
    }),
    create(ScoreSchema, {
        ID: 3n, TestName: "TestHeap", Score: 0, MaxScore: 4, Weight: 1,
        Status: TestStatus.SKIPPED,
    }),
]

const renderScores = (Scores = scores) =>
    render(<SubmissionScores submission={create(SubmissionSchema, { ID: 1n, score: 64, Scores })} />)

/** toggle returns the disclosure button of the named test. */
const toggle = (testName: string) =>
    screen.getByRole("button", { name: new RegExp(testName) })

describe("SubmissionScores", () => {
    it("says how many tests failed", () => {
        renderScores()
        expect(screen.getByText(/1 of 3 tests failed/)).toBeDefined()
    })

    it("says so when every test passed", () => {
        renderScores([scores[0]])
        expect(screen.getByText(/All 1 test passed/)).toBeDefined()
    })

    it("opens a failing test and leaves the others closed", () => {
        renderScores()
        // A student opening a submission wants the failure, not a hunt for it.
        expect(toggle("TestQueue").getAttribute("aria-expanded")).toBe("true")
        expect(screen.getByText("queue_test.go:44: Pop() = <nil>, want: x")).toBeDefined()

        expect(toggle("TestStack").getAttribute("aria-expanded")).toBe("false")
        expect(screen.queryByText("stack: pushing 3 elements")).toBeNull()
    })

    it("opens and closes a test when its row is clicked", () => {
        renderScores()
        fireEvent.click(toggle("TestStack"))
        expect(toggle("TestStack").getAttribute("aria-expanded")).toBe("true")
        expect(screen.getByText("stack: pushing 3 elements")).toBeDefined()

        fireEvent.click(toggle("TestStack"))
        expect(toggle("TestStack").getAttribute("aria-expanded")).toBe("false")
        expect(screen.queryByText("stack: pushing 3 elements")).toBeNull()
    })

    it("opens and closes every test at once", () => {
        renderScores()
        fireEvent.click(screen.getByRole("button", { name: /Expand all/ }))
        expect(toggle("TestStack").getAttribute("aria-expanded")).toBe("true")
        expect(toggle("TestHeap").getAttribute("aria-expanded")).toBe("true")

        fireEvent.click(screen.getByRole("button", { name: /Collapse all/ }))
        expect(toggle("TestQueue").getAttribute("aria-expanded")).toBe("false")
    })

    it("shows only the failing tests when asked", () => {
        renderScores()
        fireEvent.click(screen.getByRole("button", { name: /Failures only/ }))
        expect(screen.queryByRole("button", { name: /TestStack/ })).toBeNull()
        expect(toggle("TestQueue")).toBeDefined()
    })

    it("filters the tests by name", () => {
        renderScores()
        fireEvent.change(screen.getByRole("searchbox"), { target: { value: "heap" } })
        expect(toggle("TestHeap")).toBeDefined()
        expect(screen.queryByRole("button", { name: /TestStack/ })).toBeNull()
    })

    it("reports the outcome the run recorded, not the score", () => {
        renderScores()
        // A test can fail while holding partial credit, so the badge and the
        // score answer different questions.
        expect(toggle("TestQueue").textContent).toMatch(/Failed/)
        expect(toggle("TestHeap").textContent).toMatch(/Skipped/)
        expect(toggle("TestStack").textContent).toMatch(/Passed/)
    })

    it("says when a test recorded no output", () => {
        renderScores()
        fireEvent.click(toggle("TestHeap"))
        expect(screen.getByText(/recorded no output/)).toBeDefined()
    })

    it("opens the test a link points at, instead of the failures", () => {
        // A teacher can send a student straight to one test's output.
        window.history.replaceState(null, "", "#test=TestStack")
        renderScores()
        expect(toggle("TestStack").getAttribute("aria-expanded")).toBe("true")
        expect(toggle("TestQueue").getAttribute("aria-expanded")).toBe("false")
    })

    it("names the open test in the address, so the view can be linked", () => {
        renderScores()
        fireEvent.click(toggle("TestStack"))
        expect(window.location.hash).toBe("#test=TestStack")

        fireEvent.click(toggle("TestStack"))
        expect(window.location.hash).toBe("")
    })

    it("shortens the container paths a run records", () => {
        // runtime.Caller records the path inside the container, which is noise
        // to a student and pushes the message itself off the line.
        renderScores([create(ScoreSchema, {
            ID: 4n, TestName: "TestPath", Score: 0, MaxScore: 1, Weight: 1,
            Status: TestStatus.FAILED,
            TestDetails: "/quickfeed/tests/lab1/queue_test.go:44: Pop() = <nil>, want: x",
        })])
        expect(screen.getByText("queue_test.go:44: Pop() = <nil>, want: x")).toBeDefined()
    })

    it("falls back to the build log for a submission with no per-test output", () => {
        // Submissions recorded before the run output was attributed to tests
        // have no status and no output; there is nothing to expand.
        renderScores([create(ScoreSchema, { ID: 9n, TestName: "TestOld", Score: 1, MaxScore: 2, Weight: 1 })])
        expect(screen.queryByRole("button", { name: /TestOld/ })).toBeNull()
        expect(screen.getByText("TestOld")).toBeDefined()
    })
})

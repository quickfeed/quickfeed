import React from "react"
import { Assignment, AssignmentSchema, Submission, SubmissionSchema } from "../../proto/qf/types_pb"
import ProgressBar from "../components/ProgressBar"
import ProgressIndicator from "../components/ProgressIndicator"
import { initializeOvermind } from "./TestHelpers"
import { Provider } from "overmind-react"
import { render } from "@testing-library/react"
import { create } from "@bufbuild/protobuf"
import { SubmissionsForUser } from "../Helpers"

type ProgressBarTest = {
    desc: string,
    submission: Submission,
    assignment: Assignment,
    want: string
}


describe("ProgressBar", () => {
    React.useLayoutEffect = React.useEffect


    const progressBarTests: ProgressBarTest[] = [
        {
            desc: "100% Progress Bar",
            submission: create(SubmissionSchema, {
                score: 100,
            }),
            assignment: create(AssignmentSchema, { scoreLimit: 100 }),
            want: "100 %"
        },
        {
            desc: "0% Progress Bar",
            submission: create(SubmissionSchema, { score: 0 }),
            assignment: create(AssignmentSchema, { scoreLimit: 100 }),
            want: "0 %"
        },
        {
            desc: "50% Progress Bar",
            submission: create(SubmissionSchema, { score: 50 }),
            assignment: create(AssignmentSchema, { scoreLimit: 100 }),
            want: "50 %"
        },
        {
            desc: "50% Progress Bar, with 75% scorelimit",
            submission: create(SubmissionSchema, { score: 50 }),
            assignment: create(AssignmentSchema, { scoreLimit: 75 }),
            want: "50 %"
        },
        {
            desc: "75% Progress Bar, with 50% scorelimit",
            submission: create(SubmissionSchema, { score: 75 }),
            assignment: create(AssignmentSchema, { scoreLimit: 50 }),
            want: "75 %"
        },
        {
            desc: "75% Progress Bar, with 75% scorelimit",
            submission: create(SubmissionSchema, { score: 75 }),
            assignment: create(AssignmentSchema, { scoreLimit: 75 }),
            want: "75 %"
        },
        {
            desc: "Progress Bar without score",
            submission: create(SubmissionSchema),
            assignment: create(AssignmentSchema, { scoreLimit: 100 }),
            want: "0 %"
        },
        {
            desc: "Progress Bar without scorelimit",
            submission: create(SubmissionSchema, { score: 50 }),
            assignment: create(AssignmentSchema),
            want: "50 %"
        },
        {
            desc: "Progress Bar without score and scorelimit",
            submission: create(SubmissionSchema),
            assignment: create(AssignmentSchema),
            want: "0 %"
        },
    ]

    test.each(progressBarTests)("[ProgressBar] $desc", (test) => {
        labTest(test)
    })


    test.each(progressBarTests)("[ProgressIndicator] $desc", (test) => {
        const submissions = new SubmissionsForUser()
        submissions.setSubmissions(1n, "USER", [test.submission])
        const overmind = initializeOvermind({ assignments: { "1": [test.assignment] }, submissions })
        const { container } = render(
            <Provider value={overmind}>
                <ProgressIndicator courseID={"1"} submission={test.submission} />
            </Provider>
        )

        const bar = container.getElementsByTagName("div").item(0)
        expect(bar?.style).toHaveProperty("right", `${100 - test.submission.score}%`)

        // Check for the appropriate Tailwind class based on status and score
        const expectedClassSuffix = test.submission.score >= test.assignment.scoreLimit
            ? "border-b-success"
            : "border-b-primary"
        expect(bar?.className).toContain(expectedClassSuffix)
    })
})

const labTest = (test: ProgressBarTest) => {
    const submissions = new SubmissionsForUser()
    submissions.setSubmissions(1n, "USER", [test.submission])
    const overmind = initializeOvermind({ assignments: { "1": [test.assignment] }, submissions })

    const { container } = render(
        <Provider value={overmind}>
            <ProgressBar courseID={"1"} submission={test.submission} />
        </Provider>
    )

    // Check if there should be a secondary progress bar
    const hasSecondary = test.submission.score < test.assignment.scoreLimit
    const score = test.submission.score

    const bars = container.querySelectorAll('[role="progressbar"]')
    expect(bars).toHaveLength(hasSecondary ? 2 : 1)

    // Check primary progress bar
    const primary = bars[0]
    expect(primary.getAttribute("style")).toContain(`width: ${score}%`)

    // Only expect text to be shown when score > 10 (UI constraint for small bars)
    if (score > 10) {
        expect(primary.textContent).toContain(test.want)
    }

    if (hasSecondary) {
        const secondary = bars[1]
        const expectedWidth = test.assignment.scoreLimit - test.submission.score
        expect(secondary.getAttribute("style")).toContain(`width: ${expectedWidth}%`)
        expect(secondary.getAttribute("style")).toContain(`left: ${test.submission.score}%`)
    }
}

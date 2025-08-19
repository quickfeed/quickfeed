import React from "react";
import { AssignmentSchema, SubmissionSchema } from "../../proto/qf/types_pb";
import ProgressBar, { Progress } from "../components/ProgressBar";
import { initializeOvermind } from "./TestHelpers";
import { Provider } from "overmind-react";
import { render } from "@testing-library/react";
import { create } from "@bufbuild/protobuf";
import { SubmissionsForUser } from "../Helpers";
describe("ProgressBar", () => {
    React.useLayoutEffect = React.useEffect;
    const progressBarTests = [
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
    ];
    test.each(progressBarTests)(`[Progress.LAB] $desc`, (test) => {
        labTest(test);
    });
    test.each(progressBarTests)(`[Progress.NAV] $desc`, (test) => {
        const submissions = new SubmissionsForUser();
        submissions.setSubmissions(1n, "USER", [test.submission]);
        const overmind = initializeOvermind({ assignments: { "1": [test.assignment] }, submissions });
        const { container } = render(React.createElement(Provider, { value: overmind },
            React.createElement(ProgressBar, { courseID: "1", submission: test.submission, type: Progress.NAV })));
        const bar = container.getElementsByTagName("div").item(0);
        expect(bar?.style).toHaveProperty("right", `${100 - test.submission.score}%`);
        const color = test.submission.score >= test.assignment.scoreLimit
            ? "2px solid green"
            : "2px solid yellow";
        expect(bar?.style).toHaveProperty("border-bottom", color);
    });
});
const labTest = (test) => {
    const submissions = new SubmissionsForUser();
    submissions.setSubmissions(1n, "USER", [test.submission]);
    const overmind = initializeOvermind({ assignments: { "1": [test.assignment] }, submissions });
    const { container } = render(React.createElement(Provider, { value: overmind },
        React.createElement(ProgressBar, { courseID: "1", submission: test.submission, type: Progress.LAB })));
    const hasSecondary = test.submission.score < test.assignment.scoreLimit;
    const score = test.submission.score;
    const bars = container.getElementsByClassName("progress-bar");
    expect(bars).toHaveLength(hasSecondary ? 2 : 1);
    if (hasSecondary) {
        const secondary = container.getElementsByClassName("progressbar-secondary").item(0);
        if (!secondary) {
            fail();
        }
        expect(secondary.getAttribute("style")).toContain(`width: ${test.assignment.scoreLimit - test.submission.score}%`);
        expect(secondary.textContent).toEqual(`${test.assignment.scoreLimit - test.submission.score} %`);
    }
    expect(container.getElementsByClassName("progress-bar bg-primary").item(0)?.getAttribute("style")).toContain(`width: ${score}%`);
    expect(bars[0].textContent).toContain(test.want);
};

import { create } from "@bufbuild/protobuf"
import { render, screen } from "@testing-library/react"
import { Provider } from "overmind-react"
import { BuildInfoSchema, RunStatus } from "../../proto/kit/score/score_pb"
import { AssignmentSchema, EnrollmentSchema, SubmissionSchema } from "../../proto/qf/types_pb"
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
})

import { create } from "@bufbuild/protobuf"
import { render, screen } from "@testing-library/react"
import { Provider } from "overmind-react"
import { SubmissionSchema } from "../../proto/qf/types_pb"
import ManageSubmissionStatus from "../components/ManageSubmissionStatus"
import { initializeOvermind } from "./TestHelpers"

describe("ManageSubmissionStatus", () => {
    it.each([
        { canRebuild: true, want: true },
        { canRebuild: false, want: false },
    ])("shows the Rebuild button: $want, when canRebuild is $canRebuild", ({ canRebuild, want }) => {
        const overmind = initializeOvermind({ selectedSubmission: create(SubmissionSchema, { ID: 1n }) })
        render(
            <Provider value={overmind}>
                <ManageSubmissionStatus courseID="1" canRebuild={canRebuild} />
            </Provider>
        )
        expect(screen.queryByText("Rebuild") !== null).toBe(want)
    })
})

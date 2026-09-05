import { create } from "@bufbuild/protobuf"
import { ConnectError } from "@connectrpc/connect"
import { RepositoryIssuesSchema } from "../../proto/qf/requests_pb"
import { Color } from "../Helpers"
import { ApiClient } from "../overmind/namespaces/global/effects"
import { initializeOvermind, mock } from "./TestHelpers"

describe("updateAssignments alerts on the outcome reported by the server", () => {
    const updateAssignmentsTests: { desc: string, count: number, wantAlert: { text: string, color: Color } | null }[] = [
        {
            desc: "No issues found, shows green success alert",
            count: 0,
            wantAlert: { text: "Assignments updated", color: Color.GREEN },
        },
        {
            desc: "One issue found, shows singular yellow warning alert",
            count: 1,
            wantAlert: { text: "Assignments updated with 1 issue. See Course Logs for details.", color: Color.YELLOW },
        },
        {
            desc: "Multiple issues found, shows plural yellow warning alert",
            count: 3,
            wantAlert: { text: "Assignments updated with 3 issues. See Course Logs for details.", color: Color.YELLOW },
        },
    ]

    test.each(updateAssignmentsTests)(`$desc`, async (test) => {
        const api = new ApiClient()
        api.client = {
            ...api.client,
            updateAssignments: mock("updateAssignments", async () => { // skipcq: JS-0116
                return { message: create(RepositoryIssuesSchema, { count: test.count }), error: null }
            }),
        }
        const { state, actions } = initializeOvermind({}, api)
        await actions.global.updateAssignments(BigInt(1))
        expect(state.alerts).toHaveLength(1)
        expect(state.alerts[0]).toMatchObject(test.wantAlert)
    })

    it("Does not alert if the request fails", async () => {
        const api = new ApiClient()
        api.client = {
            ...api.client,
            updateAssignments: mock("updateAssignments", async () => { // skipcq: JS-0116
                return { message: create(RepositoryIssuesSchema), error: new ConnectError("failed to update assignments from tests repository") }
            }),
        }
        const { state, actions } = initializeOvermind({}, api)
        await actions.global.updateAssignments(BigInt(1))
        expect(state.alerts).toHaveLength(0)
    })
})

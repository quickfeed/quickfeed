import { User } from "../../proto/qf/types_pb"
import { Void } from "../../proto/qf/requests_pb"
import { initializeOvermind, mock } from "./TestHelpers"
import { ApiClient } from "../overmind/effects"



describe("Correct permission status should be set", () => {

    const updateAdminTests: { desc: string, user: User, confirm: boolean, want: boolean }[] = [
        {
            desc: "If user is not admin, promote to admin",
            user: new User({
                ID: BigInt(1),
                IsAdmin: false,
            }),
            confirm: true,
            want: true
        },
        {
            desc: "If user is admin, demote to non-admin",
            user: new User({
                ID: BigInt(1),
                IsAdmin: true,
            }),
            confirm: true,
            want: false
        },
        {
            desc: "If user does not confirm, do not change status",
            user: new User({
                ID: BigInt(1),
                IsAdmin: true,
            }),
            confirm: false,
            want: true
        }
    ]
    test.each(updateAdminTests)(`$desc`, async (test) => {
        const api = new ApiClient()
        api.client = {
            ...api.client,
            // eslint-disable-next-line @typescript-eslint/no-unused-vars
            updateUser: mock("updateUser", async (_request) => { // skipcq: JS-0116
                return { message: new Void(), error: null }
            }),
        }
        const { state, actions } = initializeOvermind({ allUsers: [test.user], review: { reviewer: new User() } }, api)
        window.confirm = jest.fn(() => test.confirm)
        await actions.updateAdmin(test.user)
        expect(state.allUsers[0].IsAdmin).toEqual(test.want)
    })
})

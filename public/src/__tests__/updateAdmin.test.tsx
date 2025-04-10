import { User, UserSchema } from "../../proto/qf/types_pb"
import { initializeOvermind, mock } from "./TestHelpers"
import { ApiClient } from "../overmind/namespaces/global/effects"
import { create } from "@bufbuild/protobuf"
import { VoidSchema } from "../../proto/qf/requests_pb"



describe("Correct permission status should be set", () => {

    const updateAdminTests: { desc: string, user: User, confirm: boolean, want: boolean }[] = [
        {
            desc: "If user is not admin, promote to admin",
            user: create(UserSchema, {
                ID: BigInt(1),
                IsAdmin: false,
            }),
            confirm: true,
            want: true
        },
        {
            desc: "If user is admin, demote to non-admin",
            user: create(UserSchema, {
                ID: BigInt(1),
                IsAdmin: true,
            }),
            confirm: true,
            want: false
        },
        {
            desc: "If user does not confirm, do not change status",
            user: create(UserSchema, {
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
            updateUser: mock("updateUser", async (_request) => {
                return { message: create(VoidSchema), error: null }
            }),
        }
        const { state, actions } = initializeOvermind({ allUsers: [test.user], review: { reviewer: create(UserSchema) } }, api)
        window.confirm = jest.fn(() => test.confirm)
        await actions.global.updateAdmin(test.user)
        expect(state.allUsers[0].IsAdmin).toEqual(test.want)
    })
})

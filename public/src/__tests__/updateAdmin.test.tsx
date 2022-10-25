import { User } from "../../gen/qf/types_pb"
import { initializeOvermind } from "./TestHelpers"


describe("Correct permission status should be set", () => {

    const updateAdminTests: { desc: string, user: User, confirm: boolean, want: boolean }[] = [
        {
            desc: "If user is not admin, promote to admin",
            user: new User({
                ID: BigInt(1),
                isAdmin: false,
            }),
            confirm: true,
            want: true
        },
        {
            desc: "If user is admin, demote to non-admin",
            user: new User({
                ID: BigInt(1),
                isAdmin: true,
            }),
            confirm: true,
            want: false
        },
        {
            desc: "If user does not confirm, do not change status",
            user: new User({
                ID: BigInt(1),
                isAdmin: true,
            }),
            confirm: false,
            want: true
        }
    ]
    test.each(updateAdminTests)(`$desc`, async (test) => {
        const { state, actions } = initializeOvermind({ allUsers: [test.user] })
        window.confirm = jest.fn(() => test.confirm)
        await actions.updateAdmin(test.user)
        expect(state.allUsers[0].isAdmin).toEqual(test.want)
    })
})

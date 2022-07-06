import { User } from "../../proto/qf/qf_pb"
import { initializeOvermind } from "./TestHelpers"


describe("Correct permission status should be set", () => {

    const updateAdminTests: { desc: string, user: User.AsObject, confirm: boolean, want: boolean }[] = [
        {
            desc: "If user is not admin, promote to admin",
            user: new User().setId(1).setIsadmin(false).toObject(),
            confirm: true,
            want: true
        },
        {
            desc: "If user is admin, demote to non-admin",
            user: new User().setId(1).setIsadmin(true).toObject(),
            confirm: true,
            want: false
        },
        {
            desc: "If user does not confirm, do not change status",
            user: new User().setId(1).setIsadmin(true).toObject(),
            confirm: false,
            want: true
        }
    ]
    test.each(updateAdminTests)(`$desc`, async (test) => {
        const { state, actions } = initializeOvermind({ allUsers: [test.user] })
        window.confirm = jest.fn(() => test.confirm)
        await actions.updateAdmin(test.user)
        expect(state.allUsers[0].isadmin).toEqual(test.want)
    })
})

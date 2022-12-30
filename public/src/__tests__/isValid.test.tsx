import { User } from "../../proto/qf/types_pb"
import { initializeOvermind } from "./TestHelpers"

describe("User and enrollment validation", () => {

    const tests = [
        {
            desc: "User should be valid",
            user: new User({
                ID: BigInt(1),
                Name: "Test User",
                Email: "mail@mail.com",
                StudentID: "1234567"
            }),
            expect: true
        },
        {
            desc: "User should not be valid if name is empty",
            user: new User({
                ID: BigInt(2),
                Email: "mail@mail.com",
                StudentID: "1234567"
            }),
            expect: false
        },
        {
            desc: "User should not be valid if email is empty",
            user: new User({
                ID: BigInt(1),
                Name: "Test User 3",
                StudentID: "1234567"
            }),
            expect: false
        },
        {
            desc: "User should not be valid if studentId is empty",
            user: new User({
                ID: BigInt(4),
                Name: "Test User 4",
                Email: "mail@mail.com"
            }),
            expect: false
        },
        {
            desc: "User should not be valid if name, email and studentId is empty",
            user: new User({
                ID: BigInt(5)
            }),
            expect: false
        },
    ]

    test.each(tests)(`$desc`, (test) => {
        const { state } = initializeOvermind({ self: test.user })
        expect(state.isValid).toBe(test.expect)
    })

    const emailTests = [
        {
            desc: "Email should be valid",
            email: "hei@mail.com",
            expect: true
        },
        {
            desc: "Email should not be valid",
            email: "hei@mail",
            expect: false
        }
    ]
    test.each(emailTests)(`$desc`, (test) => {
        const regex = /\S+@\S+\.\S+/
        const match = test.email.match(regex)
        // If no match is found, match is null, otherwise it is an array
        // Converting to boolean, null is false, array is true
        expect(Boolean(match)).toBe(test.expect)
    })
})

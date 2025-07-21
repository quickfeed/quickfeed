import { isExpired } from "../../Helpers"
import { timeStamp } from "../TestHelpers"

describe("isExpired", () => {
    const { T, F } = { T: true, F: false }
    const tests = [
        { expired: F, timestamp: timeStamp() }, // today
        { expired: F, timestamp: timeStamp({ months: 1 }) }, // next month
        { expired: T, timestamp: timeStamp({ months: 2 }) }, // two months in future
        { expired: T, timestamp: timeStamp({ years: -1 }) }, // previous year
        { expired: T, timestamp: timeStamp({ years: 1 }) }, // next year
    ]

    test.each(tests)("isExpired: expect $timestamp to be $expired", ({ timestamp, expired }) => {
        const result = isExpired(timestamp)
        expect(result).toBe(expired)
    })
})

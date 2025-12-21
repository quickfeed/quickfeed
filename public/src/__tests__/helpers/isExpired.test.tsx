import { isExpired } from "../../Helpers"
import { timeStamp } from "../TestHelpers"

describe("isExpired", () => {
    const { T, F } = { T: true, F: false }
    const tests = [
        { expired: F, timestamp: timeStamp() }, // today - not expired
        { expired: F, timestamp: timeStamp({ days: 20 }) }, // 20 days in future - not expired
        { expired: F, timestamp: timeStamp({ days: -20 }) }, // 20 days ago - not expired (within 1 month)
        { expired: T, timestamp: timeStamp({ days: -40 }) }, // 40 days ago - expired (> 1 month)
        { expired: T, timestamp: timeStamp({ months: -2 }) }, // 2 months ago - expired
        { expired: T, timestamp: timeStamp({ years: -1 }) }, // 1 year ago - expired
    ]

    test.each(tests)("isExpired: expect $timestamp to be $expired", ({ timestamp, expired }) => {
        const result = isExpired(timestamp)
        expect(result).toBe(expired)
    })
})

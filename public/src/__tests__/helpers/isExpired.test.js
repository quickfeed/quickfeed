import { isExpired } from "../../Helpers";
import { timeStamp } from "../TestHelpers";
describe("isExpired", () => {
    const { T, F } = { T: true, F: false };
    const tests = [
        { expired: F, timestamp: timeStamp() },
        { expired: F, timestamp: timeStamp({ months: 1 }) },
        { expired: T, timestamp: timeStamp({ months: 2 }) },
        { expired: T, timestamp: timeStamp({ years: -1 }) },
        { expired: T, timestamp: timeStamp({ years: 1 }) },
    ];
    test.each(tests)("isExpired: expect $timestamp to be $expired", ({ timestamp, expired }) => {
        const result = isExpired(timestamp);
        expect(result).toBe(expired);
    });
});

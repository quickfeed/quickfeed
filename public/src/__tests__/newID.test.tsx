import { newID } from "../Helpers";

describe("newID", () => {
    const numTests = 100000;
    it("should not generate identical IDs", () => {
        const observed = new Map<Number, boolean>();
        for (let i = 0; i < numTests; i++) {
            const id = newID();
            expect(observed.has(id)).toEqual(false);
            observed.set(id, true);
        }
        expect(observed.size).toEqual(numTests);
    });
})
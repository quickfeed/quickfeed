import { validateGroup } from "../Helpers";
describe('validateGroup', () => {
    const { T, F } = { T: true, F: false };
    const tests = [
        { valid: F, users: [], name: '' },
        { valid: F, users: [], name: "test" },
        { valid: T, users: [1n], name: "test" },
        { valid: T, users: [1n, 2n], name: "test" },
        { valid: T, users: [1n, 2n], name: "test" },
        { valid: F, users: [1n, 2n], name: "test test" },
        { valid: F, users: [1n, 2n], name: '' },
        { valid: F, users: [1n, 2n], name: ' ' },
        { valid: T, users: [1n, 2n], name: 'Group' },
        { valid: T, users: [1n, 2n], name: '1' },
        { valid: T, users: [1n, 2n], name: '123456789' },
        { valid: F, users: [1n, 2n], name: 'abcdefghijklmnopqrstuvwxyz' },
        { valid: F, users: [1n, 2n], name: 'ABCDEFGHIJKLMNOPQRSTUVWXYZ' },
        { valid: T, users: [1n, 2n], name: '0123456789abcdefghij' },
        { valid: F, users: [1n, 2n], name: 'Group Name' },
        { valid: F, users: [1n, 2n], name: 'Group Name 1' },
        { valid: T, users: [1n, 2n], name: 'GroupName' },
        { valid: T, users: [1n, 2n], name: 'Group-name1' },
        { valid: T, users: [1n, 2n], name: 'Group_name1' },
        { valid: T, users: [1n, 2n], name: 'Group_name-1' },
        { valid: F, users: [1n, 2n], name: 'Group_Å' },
        { valid: F, users: [1n, 2n], name: 'Group_Å1' },
        { valid: F, users: [1n, 2n], name: 'Group_Æ1' },
        { valid: T, users: [1n, 2n], name: 'Group_A1' },
    ];
    test.each(tests)(`Group name: expect "$name" to be $valid`, ({ name, users, valid }) => {
        const group = { name, users: users, courseID: 1n };
        const result = validateGroup(group);
        expect(result.valid).toBe(valid);
    });
});

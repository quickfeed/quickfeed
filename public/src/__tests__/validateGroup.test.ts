import { validateGroup } from "../Helpers";
import { CourseGroup } from "../overmind/state";


describe('validateGroup', () => {
    const { T, F } = { T: true, F: false }
    const tests = [
        { valid: F, users: [], name: '' }, // empty group
        { valid: F, users: [], name: "test" }, // empty group
        { valid: T, users: [1n], name: "test" },
        { valid: T, users: [1n, 2n], name: "test" },
        { valid: T, users: [1n, 2n], name: "test" },
        { valid: F, users: [1n, 2n], name: "test test" }, // space in name
        { valid: F, users: [1n, 2n], name: '' }, // empty name
        { valid: F, users: [1n, 2n], name: ' ' }, // space name
        { valid: T, users: [1n, 2n], name: 'Group' },
        { valid: T, users: [1n, 2n], name: '1' },
        { valid: T, users: [1n, 2n], name: '123456789' },
        { valid: T, users: [1n, 2n], name: 'abcdefghijklmnopqrstuvwxyz' },
        { valid: T, users: [1n, 2n], name: 'ABCDEFGHIJKLMNOPQRSTUVWXYZ' },
        { valid: F, users: [1n, 2n], name: 'Group Name' }, // space in name
        { valid: F, users: [1n, 2n], name: 'Group Name 1' }, // space in name
        { valid: T, users: [1n, 2n], name: 'GroupName' },
        { valid: T, users: [1n, 2n], name: 'Group-name1' },
        { valid: T, users: [1n, 2n], name: 'Group_name1' },
        { valid: T, users: [1n, 2n], name: 'Group_name-1' },
        { valid: F, users: [1n, 2n], name: 'Group_Å' }, // Å
        { valid: F, users: [1n, 2n], name: 'Group_Å1' }, // Å
        { valid: F, users: [1n, 2n], name: 'Group_Æ1' }, // Æ
        { valid: T, users: [1n, 2n], name: 'Group_A1' },
    ]

    test.each(tests)(`Group name: expect "$name" to be $valid`, ({ name, users, valid }) => {
        const group: CourseGroup = { name, users: users, courseID: 1n }
        const result = validateGroup(group);
        expect(result.valid).toBe(valid);
    })
})
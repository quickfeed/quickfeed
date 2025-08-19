import { create } from "@bufbuild/protobuf";
import { EnrollmentSchema, GroupSchema } from "../../proto/qf/types_pb";
import { generateRow } from "../components/ComponentsHelpers";
import { MockData } from "./mock_data/mockData";
import { initializeOvermind } from "./TestHelpers";
import { Icons } from "../components/Icons";
describe("generateRow", () => {
    const tests = [
        {
            desc: `Enrollment{ID: 1, groupID: 1} should have rows {1, ${Icons.NotAvailable}, 3, 4}`,
            enrollment: create(EnrollmentSchema, { ID: 1n, groupID: 1n }),
            generator: (s) => ({ value: `${s.ID}` }),
            want: [{ value: "1" }, Icons.NotAvailable, { value: "3" }, { value: "4" }]
        },
        {
            desc: `Enrollment{ID: 2, groupID: 0} should have rows {${Icons.NotAvailable}, 2, ${Icons.NotAvailable}, 7}`,
            enrollment: create(EnrollmentSchema, { ID: 3n }),
            generator: (s) => ({ value: `${s.ID}` }),
            want: [Icons.NotAvailable, { value: "2" }, Icons.NotAvailable, { value: "7" }]
        },
        {
            desc: `Enrollment{ID: 3, groupID: 0} should have rows {6, ${Icons.NotAvailable}, ${Icons.NotAvailable}, ${Icons.NotAvailable}}`,
            enrollment: create(EnrollmentSchema, { ID: 5n }),
            generator: (s) => ({ value: `${s.ID}` }),
            want: [{ value: "6" }, Icons.NotAvailable, Icons.NotAvailable, Icons.NotAvailable]
        },
        {
            desc: "Group{ID: 1} should have rows {4}",
            enrollment: create(GroupSchema, { ID: 1n }),
            generator: (s) => ({ value: `${s.ID}` }),
            want: [{ value: "", link: "https://github.com//" }, { value: "4" }]
        },
        {
            desc: `Group{ID: 2} should have rows {${Icons.NotAvailable}}`,
            enrollment: create(GroupSchema, { ID: 2n }),
            generator: (s) => ({ value: `${s.ID}` }),
            want: [{ value: "", link: "https://github.com//" }, Icons.NotAvailable]
        }
    ];
    const { state } = initializeOvermind({ courses: MockData.mockedCourses(), assignments: MockData.mockedCourseAssignments(), activeCourse: 1n, submissionsForCourse: MockData.mockedCourseSubmissions(1n) });
    test.each(tests)(`$desc`, (test) => {
        const rows = generateRow(test.enrollment, state.getAssignmentsMap(state.activeCourse), state.submissionsForCourse, test.generator, state.individualSubmissionView, state.courses.find(c => c.ID === state.activeCourse), false);
        expect(rows).toEqual(test.want);
    });
});


import { create } from "@bufbuild/protobuf"
import { EnrollmentSchema, GroupSchema, Submission } from "../../proto/qf/types_pb"
import { generateRow } from "../components/ComponentsHelpers"
import { MockData } from "./mock_data/mockData"
import { initializeOvermind } from "./TestHelpers"
import { Icons } from "../components/Icons"

describe("generateRow", () => {
    const tests = [
        {
            // Enrolled user with enrollment ID: 1 and groupID: 1
            // Individual submissions:
            // - Submission ID: 1, Assignment ID: 1
            // - Submission ID: 3, Assignment ID: 3
            // Group submission:
            // - Submission ID: 4, Assignment ID: 4
            desc: `Enrollment{ID: 1, groupID: 1} should have rows {1, ${Icons.NotAvailable}, 3, 4}`,
            enrollment: create(EnrollmentSchema, { ID: 1n, groupID: 1n }),
            generator: (s: Submission) => ({ value: `${s.ID}` }),
            want: [{ value: "1" }, Icons.NotAvailable, { value: "3" }, { value: "4" }]
        },
        {
            // Enrolled user with enrollment ID: 2 and groupID: 0 (not in a group)
            // Individual submissions:
            // - Submission ID: 2, Assignment ID: 2
            // - Submission ID: 7, Assignment ID: 4
            // Individual submission for group assignment
            // should be included as the user is not in a group
            desc: `Enrollment{ID: 2, groupID: 0} should have rows {${Icons.NotAvailable}, 2, ${Icons.NotAvailable}, 7}`,
            enrollment: create(EnrollmentSchema, { ID: 3n }),
            generator: (s: Submission) => ({ value: `${s.ID}` }),
            want: [Icons.NotAvailable, { value: "2" }, Icons.NotAvailable, { value: "7" }]
        },
        {
            // Enrolled user with enrollment ID: 3 and groupID: 0 (not in a group)
            // Individual submissions:
            // - Submission ID: 6, Assignment ID: 1
            desc: `Enrollment{ID: 3, groupID: 0} should have rows {6, ${Icons.NotAvailable}, ${Icons.NotAvailable}, ${Icons.NotAvailable}}`,
            enrollment: create(EnrollmentSchema, { ID: 5n }),
            generator: (s: Submission) => ({ value: `${s.ID}` }),
            want: [{ value: "6" }, Icons.NotAvailable, Icons.NotAvailable, Icons.NotAvailable]
        },
        {
            // Group with ID: 1
            // Group submissions:
            // - Submission ID: 4, Assignment ID: 4
            desc: "Group{ID: 1} should have rows {4}",
            enrollment: create(GroupSchema, { ID: 1n }),
            generator: (s: Submission) => ({ value: `${s.ID}` }),
            want: [{ value: "", link: "https://github.com//" }, { value: "4" }]
        },
        {
            // Group with ID: 2
            // Has no submissions
            desc: `Group{ID: 2} should have rows {${Icons.NotAvailable}}`,
            enrollment: create(GroupSchema, { ID: 2n }),
            generator: (s: Submission) => ({ value: `${s.ID}` }),
            want: [{ value: "", link: "https://github.com//" }, Icons.NotAvailable]
        }
    ]

    const { state } = initializeOvermind({ courses: MockData.mockedCourses(), assignments: MockData.mockedCourseAssignments(), activeCourse: 1n, submissionsForCourse: MockData.mockedCourseSubmissions(1n) })
    test.each(tests)(`$desc`, (test) => {
        const rows = generateRow(test.enrollment, state.getAssignmentsMap(state.activeCourse), state.submissionsForCourse, test.generator, state.individualSubmissionView, state.courses.find(c => c.ID === state.activeCourse), false)
        expect(rows).toEqual(test.want)
    })

})


import { Enrollment, Submission } from "../../proto/qf/types_pb"
import {  generateRow } from "../components/ComponentsHelpers"
import { MockData } from "./mock_data/mockData"
import { initializeOvermind } from "./TestHelpers"

describe("ComponentsHelpers", () => {
    const tests = [
        {
            // Enrolled user with enrollment ID: 1 and groupID: 1
            // Individual submissions:
            // - Submission ID: 1, Assignment ID: 1
            // - Submission ID: 3, Assignment ID: 3
            // Group submission:
            // - Submission ID: 4, Assignment ID: 1
            desc: "Enrollment{ID: 1, groupID: 1} should generate correct rows",
            enrollment: new Enrollment({ ID: 1n, groupID: 1n }),
            generator: (s: Submission) => ({ value: `${s.ID}` }),
            want: [{ value: "1" }, "N/A", { value: "3" }, { value: "4" }]
        },
        {
            // Enrolled user with enrollment ID: 2 and groupID: 0 (not in a group)
            // Individual submissions:
            // - Submission ID: 2, Assignment ID: 2
            // - Submission ID: 7, Assignment ID: 4
            // Individual submission for group assignment
            // should be included as the user is not in a group
            desc: "Enrollment{ID: 2, groupID: 0} should generate correct rows",
            enrollment: new Enrollment({ ID: 3n }),
            generator: (s: Submission) => ({ value: `${s.ID}` }),
            want: ["N/A", {value: "2" }, "N/A", { value: "7"}]
        },
        {
            // Enrolled user with enrollment ID: 3 and groupID: 0 (not in a group)
            // Individual submissions:
            // - Submission ID: 5, Assignment ID: 3
            desc: "Enrollment{ID: 3, groupID: 0} should generate correct rows",
            enrollment: new Enrollment({ ID: 5n }),
            generator: (s: Submission) => ({ value: `${s.ID}` }),
            want: [{ value: "6" }, "N/A", "N/A", "N/A"]
        }
    ]
    
    const { state } = initializeOvermind({ courses: MockData.mockedCourses(), assignments: MockData.mockedCourseAssignments(), activeCourse: 1n, submissionsForCourse: MockData.mockedCourseSubmissions(1n) })
    test.each(tests)(`$desc`, async (test) => {
        const rows = generateRow(test.enrollment, state.getAssignmentsMap(state.activeCourse), state.submissionsForCourse, test.generator, state.courses.find(c => c.ID === state.activeCourse), false)
        expect(rows).toEqual(test.want)
    })

})
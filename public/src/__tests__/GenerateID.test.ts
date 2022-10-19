import { Course, GradingBenchmark } from "../../gen/qf/types_pb"
import { MockGrpcManager } from "../MockGRPCManager"

(BigInt.prototype as any).toJSON = function () {
    return this.toString();
}

// The functionality tested here is only used in the MockGRPCManager class.
// - generateID should generate an ID that is not already in use.
// - The ID should be automatically incremented
describe('GenerateID', () => {
    let grpcMan: MockGrpcManager
    let types: typeof grpcMan.generate

    beforeEach(() => {
        grpcMan = new MockGrpcManager()
        types = grpcMan.generate
    })

    it('should generate an ID that is not already in use', async () => {
        const id = grpcMan.idMap.get(types.Course)

        // There should be generated 5 course IDs
        // as the grpcMan is initialized with 5 courses
        expect(id).toBe(5)

        // The next course ID should be 6
        const nextId = grpcMan.generateID(types.Course)
        expect(nextId).toBe(6n)

        // Enrollments ID should be 6
        const enrollmentsId = grpcMan.idMap.get(types.Enrollment)
        expect(enrollmentsId).toBe(6)

        // New course should have ID 7
        const course = new Course({courseCreatorID: BigInt(5)})
        const gotCourse = (await grpcMan.createCourse(course)).data
        if (gotCourse) {
            expect(gotCourse.ID).toBe(7n)
        } else {
            fail('Course was not created')
        }

        // Creating a course also enrolls the course creator in the course
        // so the enrollments ID should be incremented
        const enrollmentsId2 = grpcMan.idMap.get(types.Enrollment)
        expect(enrollmentsId2).toBe(7)

        // The next course ID should be 8
        const nextId2 = grpcMan.generateID(types.Course)
        expect(nextId2).toBe(8n)
    })

    it('should auto-increment the ID for the Group type', async () => {
        const id = grpcMan.idMap.get(types.Group)

        // There should be generated 2 IDs
        // as the grpcMan is initialized with 2 groups
        expect(id).toBe(2)

        // The next ID should be 3
        const nextId = grpcMan.generateID(types.Group)
        expect(nextId).toBe(3n)

        // New group should have ID 4
        const gotGroup = (await grpcMan.createGroup(BigInt(1), "Test", [BigInt(1), BigInt(2), BigInt(3)])).data
        if (gotGroup) {
            expect(gotGroup.ID).toBe(4n)
        } else {
            fail('Group was not created')
        }

        // Delete group
        await grpcMan.deleteGroup(BigInt(1), BigInt(4))

        // The next ID should be 5
        const gotGroup2 = (await grpcMan.createGroup(BigInt(1), "Test", [BigInt(1), BigInt(2), BigInt(3)])).data
        if (gotGroup2) {
            expect(gotGroup2.ID).toBe(5n)
        } else {
            fail('Group was not created')
        }

        // The next ID should be 6
        const nextId2 = grpcMan.generateID(types.Group)
        expect(nextId2).toBe(6n)
    })

    it('should auto-increment the ID for the Enrollment type', async () => {
        // There should be generated 6 IDs
        // as the grpcMan is initialized with 6 enrollments
        const id = grpcMan.idMap.get(types.Enrollment)
        expect(id).toBe(6)

        // The next ID should be 7
        const nextId = grpcMan.generateID(types.Enrollment)
        expect(nextId).toBe(7n)

        // New enrollment should have ID 8
        await grpcMan.createEnrollment(BigInt(1), BigInt(1))
        expect(grpcMan.idMap.get(types.Enrollment)).toBe(8)
    })

    it('should auto-increment the ID for the TemplateBenchmark type', async () => {
        const id = grpcMan.idMap.get(types.TemplateBenchmark)
        expect(id).toBe(2)

        const benchmark = new GradingBenchmark()
        const gotBenchmark = (await grpcMan.createBenchmark(benchmark)).data

        if (gotBenchmark) {
            expect(gotBenchmark.ID).toBe(3n)
        }
    })
})

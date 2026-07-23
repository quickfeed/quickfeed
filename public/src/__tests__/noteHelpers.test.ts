import { create } from "@bufbuild/protobuf"
import { EnrollmentSchema, GroupSchema, NoteSchema, SubmissionSchema, UserSchema } from "../../proto/qf/types_pb"
import { noteCountsByEnrollment, notesForEnrollment, noteTargetKey, studentNoteTargets, submissionNoteTargets } from "../components/notes/noteHelpers"

describe("note helpers", () => {
    test("student note counts include direct enrollment notes and group notes", () => {
        const enrollment = create(EnrollmentSchema, { ID: 1n, groupID: 10n })
        const otherEnrollment = create(EnrollmentSchema, { ID: 2n })
        const notes = [
            create(NoteSchema, { ID: 1n, EnrollmentID: 1n, body: "student note" }),
            create(NoteSchema, { ID: 2n, GroupID: 10n, body: "group note" }),
            create(NoteSchema, { ID: 3n, GroupID: 20n, body: "other group note" }),
        ]

        expect(notesForEnrollment(notes, enrollment).map(note => note.ID)).toEqual([1n, 2n])
        expect(noteCountsByEnrollment(notes, [enrollment, otherEnrollment]).get(enrollment.ID)).toBe(2)
        expect(noteCountsByEnrollment(notes, [enrollment, otherEnrollment]).get(otherEnrollment.ID)).toBeUndefined()
    })

    test("student targets use stable keys for direct and group notes", () => {
        const enrollment = create(EnrollmentSchema, {
            ID: 7n,
            groupID: 12n,
            group: create(GroupSchema, { ID: 12n, name: "team-a" }),
        })

        expect(studentNoteTargets(enrollment).map(target => target.key)).toEqual([
            "enrollment:7",
            "group:12",
        ])
    })

    test("submission targets resolve group members to enrollment targets", () => {
        const user = create(UserSchema, { ID: 3n, Name: "Ada" })
        const enrollment = create(EnrollmentSchema, { ID: 9n, userID: user.ID, user })
        const group = create(GroupSchema, { ID: 4n, name: "team-b", users: [user] })
        const submission = create(SubmissionSchema, { ID: 5n, groupID: group.ID })

        const targets = submissionNoteTargets(submission, [enrollment], [group])

        expect(targets.map(target => target.key)).toEqual([
            "submission:5",
            "group:4",
            "enrollment:9",
        ])
        expect(targets.map(target => noteTargetKey(target.value))).toEqual(targets.map(target => target.key))
    })
})

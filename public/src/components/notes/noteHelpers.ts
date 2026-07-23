import type { Enrollment, Group, Note, Submission } from "../../../proto/qf/types_pb"
import type { NoteTarget } from "../../overmind/namespaces/notes/actions"

/** A labelled target the staff member may attach a new note to. */
export type LabelledTarget = { key: string, label: string, value: NoteTarget }

/** A short description of which entity a note is attached to. */
export type TargetInfo = { icon: string, text: string }

export const noteTargetKey = (target: NoteTarget): string => {
    if (target.SubmissionID && target.SubmissionID > 0n) {
        return `submission:${target.SubmissionID}`
    }
    if (target.GroupID && target.GroupID > 0n) {
        return `group:${target.GroupID}`
    }
    if (target.EnrollmentID && target.EnrollmentID > 0n) {
        return `enrollment:${target.EnrollmentID}`
    }
    return "none"
}

const labelledTarget = (label: string, value: NoteTarget): LabelledTarget => ({
    key: noteTargetKey(value),
    label,
    value,
})

export const submissionNoteTargets = (submission: Submission, enrollments: Enrollment[], groups: Group[]): LabelledTarget[] => {
    const enrollmentByUserID = new Map(enrollments.map(enrollment => [enrollment.userID, enrollment]))
    const groupByID = new Map(groups.map(group => [group.ID, group]))
    const targets: LabelledTarget[] = [labelledTarget("This submission", { SubmissionID: submission.ID })]

    if (submission.groupID > 0n) {
        const group = groupByID.get(submission.groupID)
        targets.push(labelledTarget(group ? `Group: ${group.name}` : "Group", { GroupID: submission.groupID }))
        group?.users.forEach(user => {
            const enrollment = enrollmentByUserID.get(user.ID)
            if (enrollment) {
                targets.push(labelledTarget(`Student: ${user.Name}`, { EnrollmentID: enrollment.ID }))
            }
        })
        return targets
    }

    if (submission.userID > 0n) {
        const enrollment = enrollmentByUserID.get(submission.userID)
        if (enrollment) {
            targets.push(labelledTarget(enrollment.user ? `Student: ${enrollment.user.Name}` : "Student", { EnrollmentID: enrollment.ID }))
            const group = enrollment.groupID > 0n ? groupByID.get(enrollment.groupID) : undefined
            if (group) {
                targets.push(labelledTarget(`Group: ${group.name}`, { GroupID: group.ID }))
            }
        }
    }
    return targets
}

export const studentNoteTargets = (enrollment: Enrollment): LabelledTarget[] => {
    const targets = [labelledTarget("Student", { EnrollmentID: enrollment.ID })]
    if (enrollment.groupID > 0n) {
        targets.push(labelledTarget(enrollment.group?.name ? `Group: ${enrollment.group.name}` : "Group", { GroupID: enrollment.groupID }))
    }
    return targets
}

export const notesForEnrollment = (notes: Note[], enrollment: Enrollment): Note[] => notes.filter(note =>
    note.EnrollmentID === enrollment.ID || (enrollment.groupID > 0n && note.GroupID === enrollment.groupID)
)

export const noteCountsByEnrollment = (notes: Note[], enrollments: Enrollment[]): Map<bigint, number> => {
    const counts = new Map<bigint, number>()
    for (const enrollment of enrollments) {
        const count = notesForEnrollment(notes, enrollment).length
        if (count > 0) {
            counts.set(enrollment.ID, count)
        }
    }
    return counts
}

export const submissionNoteTargetInfo = (note: Note, enrollments: Enrollment[], groups: Group[]): TargetInfo => {
    const enrollmentByID = new Map(enrollments.map(enrollment => [enrollment.ID, enrollment]))
    const groupByID = new Map(groups.map(group => [group.ID, group]))
    if (note.EnrollmentID > 0n) {
        return { icon: "fa-user", text: enrollmentByID.get(note.EnrollmentID)?.user?.Name ?? "Student" }
    }
    if (note.GroupID > 0n) {
        return { icon: "fa-users", text: groupByID.get(note.GroupID)?.name ?? "Group" }
    }
    return { icon: "fa-file-lines", text: "Submission" }
}

export const studentNoteTargetInfo = (note: Note, enrollment: Enrollment, groups: Group[]): TargetInfo => {
    if (note.GroupID > 0n) {
        const group = groups.find(group => group.ID === note.GroupID)
        return { icon: "fa-users", text: group?.name ?? "Group" }
    }
    return { icon: "fa-user", text: enrollment.user?.Name ?? "Student" }
}

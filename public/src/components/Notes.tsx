import React, { useEffect, useState } from "react"
import type { Note } from "../../proto/qf/types_pb"
import { getFormattedTime } from "../Helpers"
import { useActions, useAppState } from "../overmind"
import type { NoteTarget } from "../overmind/namespaces/notes/actions"

/** A labelled target the staff member may attach a new note to. */
export type LabelledTarget = { label: string, value: NoteTarget }

/** A short description of which entity a note is attached to. */
export type TargetInfo = { icon: string, text: string }

/**
 * Notes renders the internal staff notes for the currently selected submission,
 * including the associated group and enrollment notes. Notes are never shown to
 * students; this component is only rendered in teacher views. Authors (and
 * admins) may edit or delete their own notes.
 */
const Notes = () => {
    const state = useAppState()
    const actions = useActions().notes
    const submission = state.selectedSubmission
    const [open, setOpen] = useState(true)

    useEffect(() => {
        actions.getNotes()
    }, [actions, submission?.ID])

    if (!submission) {
        return null
    }

    const notes = state.notes.notes.get(submission.ID) ?? []
    const enrollments = state.courseEnrollments[state.activeCourse.toString()] ?? []
    const groups = state.groups[state.activeCourse.toString()] ?? []

    // TODO(jostein): We should probably have this mapping in the overmind state
    // TODO(jostein): however, individual elements in maps are not reactive, so we need to recreate the map whenever the underlying array changes.
    const enrollmentByID = new Map(enrollments.map(e => [e.ID, e]))
    const enrollmentByUserID = new Map(enrollments.map(e => [e.userID, e]))
    const groupByID = new Map(groups.map(g => [g.ID, g]))

    const targetInfo = (note: Note): TargetInfo => {
        if (note.EnrollmentID > 0n) {
            return { icon: "fa-user", text: enrollmentByID.get(note.EnrollmentID)?.user?.Name ?? "Student" }
        }
        if (note.GroupID > 0n) {
            return { icon: "fa-users", text: groupByID.get(note.GroupID)?.name ?? "Group" }
        }
        return { icon: "fa-file-lines", text: "Submission" }
    }

    // Targets are derived from the submission so the list always matches its owner.
    const targets: LabelledTarget[] = [{ label: "This submission", value: { SubmissionID: submission.ID } }]
    if (submission.groupID > 0n) {
        const group = groupByID.get(submission.groupID)
        targets.push({ label: group ? `Group: ${group.name}` : "Group", value: { GroupID: submission.groupID } })
        // Groups carry `users`, not enrollments, so resolve each member's enrollment ID.
        // otherwise we could've just iterated over `group.enrollments` and used the enrollment ID directly.
        group?.users.forEach(user => {
            const enrollment = enrollmentByUserID.get(user.ID)
            if (enrollment) {
                targets.push({ label: `Student: ${user.Name}`, value: { EnrollmentID: enrollment.ID } })
            }
        })
    } else if (submission.userID > 0n) {
        const enrollment = enrollmentByUserID.get(submission.userID)
        if (enrollment) {
            targets.push({ label: enrollment.user ? `Student: ${enrollment.user.Name}` : "Student", value: { EnrollmentID: enrollment.ID } })
            const group = enrollment.groupID > 0n ? groupByID.get(enrollment.groupID) : undefined
            if (group) {
                targets.push({ label: `Group: ${group.name}`, value: { GroupID: group.ID } })
            }
        }
    }

    return (
        <div className="card bg-base-100 shadow-xl mb-4">
            <button type="button"
                className="flex items-center gap-2 w-full bg-warning text-warning-content px-6 py-3 rounded-t-2xl"
                aria-expanded={open}
                onClick={() => setOpen(prev => !prev)}
            >
                <i className="fas fa-lock" />
                <h3 className="text-md font-bold">Internal Notes</h3>
                {notes.length > 0 && <span className="badge badge-sm">{notes.length}</span>}
                <i className={`fas fa-chevron-${open ? "up" : "down"} ml-auto`} />
            </button>

            {open && (
                <div className="card-body p-0">
                    <NotePanelBody notes={notes} targets={targets} targetInfo={targetInfo} />
                </div>
            )}
        </div>
    )
}

/**
 * NotePanelBody renders the list of notes and the add/edit form. It is shared by
 * the collapsible submission panel and the per-student notes modal.
 */
export const NotePanelBody = ({ notes, targets, targetInfo }: { notes: Note[], targets: LabelledTarget[], targetInfo?: (note: Note) => TargetInfo }) => {
    const state = useAppState()
    const canModify = (note: Note) => note.AuthorID === state.self.ID || state.self.IsAdmin
    const authorName = (authorID: bigint) => state.courseTeachers[authorID.toString()]?.Name ?? "Staff"

    return (
        <>
            <ul className="divide-y divide-base-300">
                {notes.length === 0 && (
                    <li className="px-6 py-4 text-sm text-base-content/60">No notes yet.</li>
                )}
                {notes.map(note => (
                    <NoteItem key={note.ID.toString()}
                        note={note}
                        authorName={authorName(note.AuthorID)}
                        target={targetInfo?.(note)}
                        canModify={canModify(note)}
                    />
                ))}
            </ul>

            {state.notes.editing === 0n && <NoteForm targets={targets} />}
        </>
    )
}

/** NoteItem renders a single note, with edit/delete controls for its author or an admin. */
const NoteItem = ({ note, authorName, target, canModify }: { note: Note, authorName: string, target?: TargetInfo, canModify: boolean }) => {
    const state = useAppState()
    const actions = useActions().notes
    const isEditing = state.notes.editing === note.ID

    if (isEditing) {
        return (
            <li className="px-6 py-4">
                <textarea className="textarea textarea-bordered w-full" rows={3}
                    value={state.notes.draft}
                    onChange={e => actions.setDraft(e.target.value)}
                />
                <div className="flex gap-2 mt-2">
                    <button className="btn btn-sm btn-primary" onClick={() => actions.updateNote(note)}>Save</button>
                    <button className="btn btn-sm btn-ghost" onClick={() => actions.cancelEditing()}>Cancel</button>
                </div>
            </li>
        )
    }

    return (
        <li className="px-6 py-4">
            {target && (
                <div className="flex items-center gap-1 text-xs text-base-content/70 mb-1">
                    <i className={`fas ${target.icon}`} />
                    <span className="font-semibold">{target.text}</span>
                </div>
            )}
            <p className="whitespace-pre-wrap">{note.body}</p>
            <div className="flex items-center justify-between mt-2 text-xs text-base-content/60">
                <span>{authorName} · {getFormattedTime(note.editedAt ?? note.createdAt)}</span>
                {canModify && (
                    <div className="flex gap-2">
                        <button className="link link-hover" onClick={() => actions.startEditing(note)}>Edit</button>
                        <button className="link link-hover text-error" onClick={() => actions.deleteNote(note)}>Delete</button>
                    </div>
                )}
            </div>
        </li>
    )
}

/** NoteForm lets staff draft a new note and choose which target to attach it to. */
const NoteForm = ({ targets }: { targets: LabelledTarget[] }) => {
    const state = useAppState()
    const actions = useActions().notes
    const [targetIndex, setTargetIndex] = useState(0)

    return (
        <div className="px-6 py-4 border-t border-base-300">
            <textarea className="textarea textarea-bordered w-full" rows={3} placeholder="Add an internal note…"
                value={state.notes.draft}
                onChange={e => actions.setDraft(e.target.value)}
            />
            <div className="flex items-center gap-2 mt-2">
                {targets.length > 1 && (
                    <select className="select select-bordered select-sm"
                        value={targetIndex}
                        onChange={e => setTargetIndex(Number(e.target.value))}
                    >
                        {targets.map((t, i) => <option key={t.label} value={i}>{t.label}</option>)}
                    </select>
                )}
                <button className="btn btn-sm btn-primary"
                    disabled={state.notes.draft.trim().length === 0}
                    onClick={() => actions.createNote(targets[targetIndex].value)}
                >
                    Add note
                </button>
            </div>
        </div>
    )
}

export default Notes

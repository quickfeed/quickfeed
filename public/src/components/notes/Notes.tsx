import { useEffect, useState } from "react"
import type { Note } from "../../../proto/qf/types_pb"
import { getFormattedTime } from "../../Helpers"
import { useActions, useAppState } from "../../overmind"
import type { NoteScope } from "../../overmind/namespaces/notes/state"
import type { LabelledTarget, TargetInfo } from "./noteHelpers"
import { submissionNoteTargetInfo, submissionNoteTargets } from "./noteHelpers"

/**
 * Notes renders the internal staff notes for the currently selected submission,
 * including the associated group and enrollment notes. Notes are never shown to
 * students; this component is only rendered in teacher views. Only a note's
 * author may edit or delete it.
 */
const Notes = () => {
    const state = useAppState()
    const actions = useActions().notes
    const submission = state.selectedSubmission
    const [open, setOpen] = useState(false)

    useEffect(() => {
        actions.getNotes()
    }, [actions, submission?.ID])

    if (!submission) {
        return null
    }

    const notes = state.notes.notes.get(submission.ID) ?? []
    const enrollments = state.courseEnrollments[state.activeCourse.toString()] ?? []
    const groups = state.groups[state.activeCourse.toString()] ?? []

    const targets = submissionNoteTargets(submission, enrollments, groups)
    const targetInfo = (note: Note): TargetInfo => submissionNoteTargetInfo(note, enrollments, groups)

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
                    <NotePanelBody notes={notes} targets={targets} targetInfo={targetInfo} scope="submission" />
                </div>
            )}
        </div>
    )
}

/**
 * NotePanelBody renders the list of notes and the add/edit form. It is shared by
 * the collapsible submission panel and the per-student details view. The scope
 * says which note list this panel shows, so a mutation reloads that list.
 */
export const NotePanelBody = ({ notes, targets, targetInfo, scope }: { notes: Note[], targets: LabelledTarget[], targetInfo?: (note: Note) => TargetInfo, scope: NoteScope }) => {
    const state = useAppState()
    const canModify = (note: Note) => note.AuthorID === state.self.ID
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
                        scope={scope}
                    />
                ))}
            </ul>

            <NoteForm targets={targets} scope={scope} />
        </>
    )
}

/** NoteItem renders a single note, with edit/delete controls for its author. */
const NoteItem = ({ note, authorName, target, canModify, scope }: { note: Note, authorName: string, target?: TargetInfo, canModify: boolean, scope: NoteScope }) => {
    const state = useAppState()
    const actions = useActions().notes
    const isEditing = state.notes.editing === note.ID

    if (isEditing) {
        return (
            <li className="px-6 py-4">
                <textarea className="textarea textarea-bordered w-full" rows={3}
                    value={state.notes.editDraft}
                    onChange={e => actions.setEditDraft(e.target.value)}
                />
                <div className="flex gap-2 mt-2">
                    <button className="btn btn-sm btn-primary"
                        disabled={state.notes.editDraft.trim().length === 0}
                        onClick={() => actions.updateNote({ note, scope })}
                    >Save</button>
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
                        <button className="link link-hover text-error" onClick={() => actions.deleteNote({ note, scope })}>Delete</button>
                    </div>
                )}
            </div>
        </li>
    )
}

/** NoteForm lets staff draft a new note and choose which target to attach it to. */
const NoteForm = ({ targets, scope }: { targets: LabelledTarget[], scope: NoteScope }) => {
    const state = useAppState()
    const actions = useActions().notes
    const [targetKey, setTargetKey] = useState(targets[0]?.key ?? "")
    // Falls back to the first target when targetKey no longer matches one, e.g.
    // after the group or enrollment it pointed to disappears from the list.
    const selectedTarget = targets.find(target => target.key === targetKey) ?? targets[0]

    return (
        <div className="px-6 py-4 border-t border-base-300">
            <textarea className="textarea textarea-bordered w-full" rows={3} placeholder="Add an internal note…"
                value={state.notes.draft}
                onChange={e => actions.setDraft(e.target.value)}
            />
            <div className="flex items-center gap-2 mt-2">
                {targets.length > 1 && (
                    <select className="select select-bordered select-sm"
                        value={selectedTarget?.key ?? ""}
                        onChange={e => setTargetKey(e.target.value)}
                    >
                        {targets.map(target => <option key={target.key} value={target.key}>{target.label}</option>)}
                    </select>
                )}
                <button className="btn btn-sm btn-primary"
                    disabled={!selectedTarget || state.notes.draft.trim().length === 0}
                    onClick={() => selectedTarget && actions.createNote({ target: selectedTarget.value, scope })}
                >
                    Add note
                </button>
            </div>
        </div>
    )
}

export default Notes

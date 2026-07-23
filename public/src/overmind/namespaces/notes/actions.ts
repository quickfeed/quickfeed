import { create } from "@bufbuild/protobuf"
import type { Context } from "../.."
import type { Note } from "../../../../proto/qf/types_pb"
import { NoteSchema } from "../../../../proto/qf/types_pb"

/* NoteTarget identifies which entity a new note is attached to.
   Exactly one of the fields should be set. */
export type NoteTarget = { SubmissionID?: bigint, GroupID?: bigint, EnrollmentID?: bigint }

/* getNotes fetches all notes relevant to the currently selected submission,
   including the associated group and enrollment notes. */
export const getNotes = async ({ state, effects }: Context): Promise<void> => {
    state.notes.scope = "submission"
    const submission = state.selectedSubmission
    if (!submission || !state.activeCourse) {
        return
    }
    const response = await effects.global.api.client.getNotes({
        courseID: state.activeCourse,
        submissionID: submission.ID,
    })
    if (response.error) {
        return
    }
    const notes = new Map(state.notes.notes)
    notes.set(submission.ID, response.message.notes)
    state.notes.notes = notes
}

/* getCourseNotes fetches every note in the active course, used by staff
   overviews such as the members page to show per-student notes. */
export const getCourseNotes = async ({ state, effects }: Context): Promise<void> => {
    state.notes.scope = "course"
    if (!state.activeCourse) {
        return
    }
    const response = await effects.global.api.client.getCourseNotes({ courseID: state.activeCourse })
    if (response.error) {
        return
    }
    state.notes.courseNotes = response.message.notes
}

/* refresh reloads whichever note scope was last loaded, so mutations update the
   correct view (a submission's notes or the course-wide list). */
export const refresh = async ({ state, actions }: Context): Promise<void> => {
    if (state.notes.scope === "course") {
        await actions.notes.getCourseNotes()
    } else {
        await actions.notes.getNotes()
    }
}

/* createNote creates a note attached to the given target (submission, group, or
   enrollment) and refreshes the notes for the current submission. */
export const createNote = async ({ state, actions, effects }: Context, target: NoteTarget): Promise<void> => {
    const body = state.notes.draft.trim()
    if (!body || !state.activeCourse) {
        return
    }
    const note = create(NoteSchema, { body, ...target })
    const response = await effects.global.api.client.createNote({
        courseID: state.activeCourse,
        note,
    })
    if (response.error) {
        return
    }
    state.notes.draft = ""
    await actions.notes.refresh()
}

/* updateNote saves an edited note body. Only the author or an admin may succeed. */
export const updateNote = async ({ state, actions, effects }: Context, note: Note): Promise<void> => {
    const body = state.notes.draft.trim()
    if (!body || !state.activeCourse) {
        return
    }
    const response = await effects.global.api.client.updateNote({
        courseID: state.activeCourse,
        note: { ...note, body },
    })
    if (response.error) {
        return
    }
    state.notes.editing = 0n
    state.notes.draft = ""
    await actions.notes.refresh()
}

/* deleteNote removes a note. Only the author or an admin may succeed. */
export const deleteNote = async ({ state, actions, effects }: Context, note: Note): Promise<void> => {
    if (!state.activeCourse || !confirm("Are you sure you want to delete this note?")) {
        return
    }
    const response = await effects.global.api.client.deleteNote({
        courseID: state.activeCourse,
        note,
    })
    if (response.error) {
        return
    }
    await actions.notes.refresh()
}

/* startEditing prepares the form to edit an existing note. */
export const startEditing = ({ state }: Context, note: Note): void => {
    state.notes.editing = note.ID
    state.notes.draft = note.body
}

/* cancelEditing clears any in-progress edit or draft. */
export const cancelEditing = ({ state }: Context): void => {
    state.notes.editing = 0n
    state.notes.draft = ""
}

/* setDraft updates the body of the note being drafted or edited. */
export const setDraft = ({ state }: Context, body: string): void => {
    state.notes.draft = body
}

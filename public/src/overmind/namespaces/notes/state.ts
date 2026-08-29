import type { Note } from "../../../../proto/qf/types_pb"

/* NoteScope identifies which note list a view is showing, and therefore which
   list must be reloaded after the view creates, edits, or deletes a note. */
export type NoteScope = "submission" | "course"

export type NotesState = {
    /* Internal staff notes for the currently selected submission, keyed by submission ID.
       The list includes the submission's own notes plus the associated group and enrollment notes. */
    notes: Map<bigint, Note[]>

    /* All internal notes for the active course, used by staff overviews such as the members page. */
    courseNotes: Note[]

    /* The body of the new note being drafted. Kept separate from editDraft so that
       starting an edit does not discard a new note the user has begun writing. */
    draft: string

    /* The ID of the note currently being edited, or 0 if none */
    editing: bigint

    /* The body of the note currently being edited */
    editDraft: string
}

export const state: NotesState = {
    notes: new Map(),
    courseNotes: [],
    draft: "",
    editing: 0n,
    editDraft: "",
}

import type { Note } from "../../../../proto/qf/types_pb"

export type NotesState = {
    /* Internal staff notes for the currently selected submission, keyed by submission ID.
       The list includes the submission's own notes plus the associated group and enrollment notes. */
    notes: Map<bigint, Note[]>

    /* All internal notes for the active course, used by staff overviews such as the members page. */
    courseNotes: Note[]

    /* The scope last loaded; determines what is refreshed after a note is created, edited, or deleted. */
    scope: "submission" | "course"

    /* The body of the note currently being drafted */
    draft: string

    /* The ID of the note currently being edited, or 0 if none */
    editing: bigint
}

export const state: NotesState = {
    notes: new Map(),
    courseNotes: [],
    scope: "submission",
    draft: "",
    editing: 0n,
}

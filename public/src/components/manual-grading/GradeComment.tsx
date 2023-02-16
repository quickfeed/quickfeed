import React, { Dispatch, SetStateAction } from "react"
import { GradingBenchmark, GradingCriterion } from "../../../proto/qf/types_pb"
import { useActions, useAppState } from "../../overmind"

type GradeCommentProps = {
    grade: GradingBenchmark | GradingCriterion,
    editing: boolean,
    setEditing: Dispatch<SetStateAction<boolean>>
}

const GradeComment = ({ grade, editing, setEditing }: GradeCommentProps): JSX.Element | null => {
    const actions = useActions()
    const state = useAppState()

    /* Don't allow grading if user is not a teacher or editing is false */
    if (!state.isTeacher || !editing) {
        return null
    }

    // handleChange saves the comment when clicking outside the text area, or when pressing enter.
    // Clicking outside, pressing enter, or pressing escape will set editing to false.
    // Changes are discarded if the user presses escape.
    const handleChange = (event: React.FocusEvent<HTMLTextAreaElement> | React.KeyboardEvent<HTMLTextAreaElement>) => {
        // Handle if event is keyboard event
        if ("key" in event) {
            if (event.key !== "Escape") {
                // Exit early if the key is not an escape or enter key
                return
            }
            if (event.key === "Escape") {
                setEditing(false)
                return
            }
        }

        const { value } = event.currentTarget
        setEditing(false)
        // Exit early if the value is unchanged
        if (value === grade.comment) {
            return
        }
        actions.review.updateComment({ grade: grade, comment: value })
    }

    return (
        <tr>
            <th colSpan={3}>
                <textarea rows={20} autoFocus onBlur={handleChange} onKeyUp={handleChange} defaultValue={grade.comment} className="form-control"></textarea>
            </th>
        </tr>
    )

}

export default GradeComment

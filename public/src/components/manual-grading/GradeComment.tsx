import React, { Dispatch, SetStateAction } from "react"
import { GradingBenchmark, GradingCriterion } from "../../../proto/qf/types_pb"
import { useActions, useAppState } from "../../overmind"

type GradeCommentProps = {
    grade: GradingBenchmark | GradingCriterion,
    editing: boolean,
    setEditing: Dispatch<SetStateAction<boolean>>
}

const GradeComment = ({ grade, editing, setEditing }: GradeCommentProps) => {
    const actions = useActions()
    const state = useAppState()

    /* Don't allow grading if user is not a teacher or editing is false */
    if (!state.isTeacher || !editing) {
        return null
    }

    const updateComment = (value: string) => {
        setEditing(false)
        // Exit early if the value is unchanged
        if (value === grade.comment) {
            return
        }
        actions.review.updateComment({ grade: grade, comment: value })
    }

    // handleBlur saves the comment when clicking outside the text area.
    const handleBlur = (event: React.FocusEvent<HTMLTextAreaElement>) => {
        const { value } = event.currentTarget
        updateComment(value)
    }

    // handleKeyUp saves the comment when pressing Ctrl/Cmd+Enter or Ctrl/Cmd+Q/q.
    // It also cancels the edit when pressing Escape.
    const handleKeyUp = (event: React.KeyboardEvent<HTMLTextAreaElement>) => {
        if (event.key === "Escape") {
            setEditing(false)
            return
        }
        if ((event.key === "Enter" || event.key === "q" || event.key === "Q") && (event.ctrlKey || event.metaKey)) {
            const { value } = event.currentTarget
            updateComment(value)
        }
    }

    return (
        <tr>
            <th colSpan={3}>
                <textarea
                    rows={20}
                    autoFocus
                    onBlur={handleBlur}
                    onKeyUp={handleKeyUp}
                    defaultValue={grade.comment}
                    className="w-full p-4 bg-base-100 border border-base-300 rounded-lg text-base-content placeholder-base-content/50 focus:border-primary focus:ring-2 focus:ring-primary/20 focus:outline-none transition-all duration-200 resize-y"
                    placeholder="Enter your comment... (Ctrl/Cmd+Enter to save, Escape to cancel)"
                /> {/* skipcq: JS-0757 */}
            </th>
        </tr>
    )

}

export default GradeComment

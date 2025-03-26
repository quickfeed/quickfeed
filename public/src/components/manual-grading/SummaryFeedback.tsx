
import React, { useCallback, useState } from "react"
import { Review } from "../../../proto/qf/types_pb"
import { useActions, useAppState } from "../../overmind"
import CriterionComment from "./Comment"


const SummaryFeedback = ({ review }: { review: Review }) => {
    const state = useAppState()
    const actions = useActions()
    const [editing, setEditing] = useState<boolean>(false)

    const summaryFeedback = <td colSpan={3}><CriterionComment comment={review.feedback.length > 0 ? review.feedback : "No summary feedback"} /></td>

    const handleChange = useCallback((event: React.FormEvent<HTMLTextAreaElement>) => {
        const { value } = event.currentTarget
        setEditing(false)
        // Exit early if the value is unchanged
        if (value === review.feedback) {
            return
        }
        actions.review.updateFeedback({ feedback: value })
    }, [actions.review, review.feedback])

    if (!state.isTeacher) {
        return <tr>{summaryFeedback}</tr>
    }

    return (
        <>
            <tr onClick={() => setEditing(true)}>
                {summaryFeedback}
            </tr>
            {editing &&
                <tr>
                    <td colSpan={3}>
                        <textarea rows={20} autoFocus onBlur={handleChange} defaultValue={review.feedback} className="form-control" />
                    </td>
                </tr>
            }
        </>
    )
}

export default SummaryFeedback

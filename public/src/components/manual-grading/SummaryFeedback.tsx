
import React, { useState } from "react"
import { Review } from "../../../proto/qf/types_pb"
import { useActions, useAppState } from "../../overmind"


const SummaryFeedback = ({ review }: { review: Review.AsObject }) => {
    const state = useAppState()
    const actions = useActions()
    const [editing, setEditing] = useState<boolean>(false)

    const summaryFeedback = <td colSpan={3}>{review.feedback.length > 0 ? review.feedback : "No summary feedback"}</td>

    if (!state.isTeacher) {
        return summaryFeedback
    }

    const handleChange = (event: React.FormEvent<HTMLInputElement>) => {
        const { value } = event.currentTarget
        setEditing(false)
        // Exit early if the value is unchanged
        if (value === review.feedback) {
            return
        }
        actions.review.updateFeedback({ feedback: value })
    }

    return (
        <>
            <tr onClick={() => setEditing(true)}>
                {summaryFeedback}
            </tr>
            {editing &&
                <tr>
                    <th colSpan={3}>
                        <input autoFocus onBlur={handleChange} defaultValue={review.feedback} className="form-control" type="text" />
                    </th>
                </tr>
            }
        </>
    )
}

export default SummaryFeedback

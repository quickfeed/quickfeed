
import React, { useState } from "react"
import { Review } from "../../../proto/ag/ag_pb"
import { useActions, useAppState } from "../../overmind"


const SummaryFeedback = ({ review }: { review: Review }) => {
    const state = useAppState()
    const actions = useActions()
    const [editing, setEditing] = useState<boolean>(false)

    const summaryFeedback = <td colSpan={3}>{review.getFeedback().length > 0 ? review.getFeedback() : "No summary feedback"}</td>

    if (!state.isTeacher) {
        return summaryFeedback
    }

    const handleChange = (event: React.FormEvent<HTMLInputElement>) => {
        const { value } = event.currentTarget
        setEditing(false)
        // Exit early if the value is unchanged
        if (value === review.getFeedback()) {
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
                        <input autoFocus onBlur={handleChange} defaultValue={review.getFeedback()} className="form-control" type="text" />
                    </th>
                </tr>
            }
        </>
    )
}

export default SummaryFeedback

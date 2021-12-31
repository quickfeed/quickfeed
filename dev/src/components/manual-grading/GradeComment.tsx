import React from "react"
import { Dispatch, SetStateAction } from "react"
import { GradingBenchmark, GradingCriterion } from "../../../proto/ag/ag_pb"
import { useActions, useAppState } from "../../overmind"


const GradeComment = ({grade, editing, setEditing}: {grade: GradingBenchmark | GradingCriterion, editing: boolean, setEditing: Dispatch<SetStateAction<boolean>>}): JSX.Element => {
    const actions = useActions()
    const state = useAppState()
    
    if (!state.isTeacher) {
        return <></>
    }

    const handleChange = (event: React.FormEvent<HTMLInputElement>) => {
        const { value } = event.currentTarget
        setEditing(false)
        // Exit early if the value is unchanged
        if (value === grade.getComment()) {
            return
        }
        grade.setComment(value)
        actions.review.updateReview()
    }

    if (editing) {
        return (
            <tr>
                <th colSpan={3}>
                    <input autoFocus onBlur={(e) => handleChange(e)} defaultValue={grade.getComment()} className="form-control" type="text"></input>
                </th>
            </tr>
        )
    }
    return <></>
}

export default GradeComment
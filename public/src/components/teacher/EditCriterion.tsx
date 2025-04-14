import React, { useState } from "react"
import { Assignment, GradingCriterion, GradingCriterionSchema } from "../../../proto/qf/types_pb"
import { useActions } from "../../overmind"
import { clone, create } from "@bufbuild/protobuf"


const EditCriterion = ({ originalCriterion, benchmarkID, assignment }: { originalCriterion?: GradingCriterion, benchmarkID: bigint, assignment: Assignment }) => {
    const actions = useActions().global

    const [editing, setEditing] = useState<boolean>(false)
    const [add, setAdd] = useState<boolean>(originalCriterion ? false : true)

    // Clone the criterion, or create a new one if none was passed in
    const criterion = originalCriterion
        ? clone(GradingCriterionSchema, originalCriterion)
        : create(GradingCriterionSchema)


    const resetCriterion = () => {
        // Reset the criterion and enable add button
        criterion.description = ""
        setAdd(true)
    }

    const handleCriteria = async (event: React.KeyboardEvent<HTMLInputElement>) => {
        const { value } = event.currentTarget
        if (event.key === "Enter") {
            // Set the criterion's benchmark ID
            // This could already be set if a criterion was passed in
            criterion.BenchmarkID = benchmarkID
            const success = await actions.createOrUpdateCriterion({ criterion, assignment })
            if (!success) {
                resetCriterion()
            }
            setEditing(false)
        } else {
            criterion.description = value
        }
    }

    const handleBlur = () => {
        if (originalCriterion) {
            // Restore the original criterion
            criterion.description = originalCriterion.description
        } else {
            resetCriterion()
        }
        setEditing(false)
    }

    if (add) {
        return (
            <div className="list-group-item">
                <button className="btn btn-success" name="submit" onClick={() => { setAdd(false); setEditing(true) }}>Add Criteria</button>
            </div>
        )
    }

    const input = <input className="form-control" type="text" autoFocus onBlur={handleBlur} defaultValue={criterion.description} name="criterion" onKeyUp={e => handleCriteria(e)} /> // skipcq: JS-0757
    const textAndButton = (
        <span onClick={() => setEditing(!editing)} role="button" aria-hidden="true">
            {criterion.description}
            <button className="p-2 badge badge-danger float-right clickable" onClick={() => actions.deleteCriterion({ criterion: originalCriterion, assignment })}>
                Delete Criteria
            </button>
        </span>
    )
    return (
        <div className="list-group-item">
            {editing ? input : textAndButton}
        </div>
    )
}

export default EditCriterion

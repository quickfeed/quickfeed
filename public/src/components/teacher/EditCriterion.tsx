import React, { useState } from "react"
import { Assignment, GradingCriterion, GradingCriterionSchema } from "../../../proto/qf/types_pb"
import { useActions } from "../../overmind"
import { clone, create } from "@bufbuild/protobuf"


const EditCriterion = ({ originalCriterion, benchmarkID, assignment }: { originalCriterion?: GradingCriterion, benchmarkID: bigint, assignment: Assignment }) => {
    const actions = useActions()

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
            const success = await actions.createOrUpdateCriterion({ criterion: criterion, assignment: assignment })
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

    return (
        <div className="list-group-item" onClick={() => setEditing(!editing)} role="button" aria-hidden="true">
            {editing
                ? <input className="form-control" type="text" onBlur={handleBlur} onClick={handleBlur} autoFocus defaultValue={criterion.description} name="criterion" onKeyUp={e => handleCriteria(e)} />
                : <span>{criterion.description}<span className="p-2 badge badge-danger float-right clickable" onClick={() => actions.deleteCriterion({ criterion: originalCriterion, assignment: assignment })}>Delete Criteria</span></span>
            }
        </div>
    )
}

export default EditCriterion

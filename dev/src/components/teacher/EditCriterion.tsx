import React, { useState } from "react"
import { Assignment, GradingCriterion } from "../../../proto/ag/ag_pb"
import { Converter } from "../../convert"
import { useActions } from "../../overmind"


const EditCriterion = ({ originalCriterion, benchmarkID, assignment }: { originalCriterion?: GradingCriterion.AsObject, benchmarkID: number, assignment: Assignment.AsObject }): JSX.Element => {
    const actions = useActions()

    const [editing, setEditing] = useState<boolean>(false)
    const [add, setAdd] = useState<boolean>(originalCriterion ? false : true)

    // Clone the criterion, or create a new one if none was passed in
    const criterion = originalCriterion
        ? Converter.clone(originalCriterion)
        : Converter.create<GradingCriterion.AsObject>(GradingCriterion)

    const handleCriteria = (event: React.KeyboardEvent<HTMLInputElement>) => {
        const { value } = event.currentTarget
        if (event.key === "Enter") {
            // Set the criterion's benchmark ID
            // This could already be set if a criterion was passed in
            criterion.benchmarkid = benchmarkID
            actions.createOrUpdateCriterion({ criterion: criterion, assignment: assignment })
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
            // Reset the criterion and enable add button
            criterion.description = ""
            setAdd(true)
        }
        setEditing(false)
    }

    if (add) {
        return (
            <div className="list-group-item">
                <button className="btn btn-primary" name="submit" onClick={() => { setAdd(false); setEditing(true) }}>Add</button>
            </div>
        )
    }

    return (
        <div className="list-group-item" onClick={() => setEditing(!editing)}>
            {editing
                ? <input className="form-control" type="text" onBlur={() => { handleBlur() }} autoFocus defaultValue={criterion.description} name="criterion" onKeyUp={e => { handleCriteria(e) }} />
                : <><span>{criterion.description}</span><span className="badge badge-danger float-right clickable" onClick={() => actions.deleteCriterion({ criterion: originalCriterion, assignment: assignment })}>Delete</span></>
            }
        </div>
    )
}

export default EditCriterion

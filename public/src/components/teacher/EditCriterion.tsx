import React, { useState, memo } from "react"
import { GradingCriterion } from "../../../proto/qf/types_pb"

interface EditCriterionProps {
    originalCriterion?: GradingCriterion
    updateCriterion: (event: React.KeyboardEvent<HTMLInputElement>, criterion: GradingCriterion) => void
    deleteCriterion?: () => void
}

const EditCriterion = memo(({ originalCriterion, updateCriterion, deleteCriterion }: EditCriterionProps) => {
    const [editing, setEditing] = useState<boolean>(false)
    const [add, setAdd] = useState<boolean>(originalCriterion ? false : true)

    // Clone the criterion, or create a new one if none was passed in
    const criterion = originalCriterion
        ? originalCriterion.clone()
        : new GradingCriterion()

    const handleCriterion = (event: React.KeyboardEvent<HTMLInputElement>) => {
        updateCriterion(event, criterion)
        setEditing(false)
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
                <button className="btn btn-primary" name="submit" onClick={() => { setAdd(false); setEditing(true) }}>Add</button> {/* skipcq: JS-0417 */}
            </div>
        )
    }

    return (
        <div className="list-group-item" onClick={() => setEditing(!editing)}> {/* skipcq: JS-0417 */}
            {editing
                ? <input className="form-control" type="text" onBlur={handleBlur} autoFocus defaultValue={criterion.description} name="criterion" onKeyUp={handleCriterion} />
                : <><span>{criterion.description}</span><span className="badge badge-danger float-right clickable" onClick={deleteCriterion}>Delete</span></>}
        </div>
    )
})

EditCriterion.displayName = "EditCriterion"

export default EditCriterion

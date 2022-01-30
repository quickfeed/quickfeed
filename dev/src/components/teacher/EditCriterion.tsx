import { json } from "overmind"
import React, { useState } from "react"
import { Assignment, GradingCriterion } from "../../../proto/ag/ag_pb"
import { useActions } from "../../overmind"


const EditCriterion = ({ criterion, benchmarkID, assignment }: { criterion?: GradingCriterion, benchmarkID: number, assignment: Assignment }): JSX.Element => {
    const actions = useActions()

    const [editing, setEditing] = useState<boolean>(false)
    const [add, setAdd] = useState<boolean>(criterion ? false : true)
    const [newCriterion, setNewCriterion] = useState<GradingCriterion>(new GradingCriterion().setBenchmarkid(benchmarkID))
    // Clone the criterion to use as backup in case of cancel
    const copy = json(criterion)?.cloneMessage()
    const c = json(criterion) ?? newCriterion

    const handleCriteria = (event: React.KeyboardEvent<HTMLInputElement>) => {
        const { value } = event.currentTarget
        if (event.key === "Enter") {
            actions.createOrUpdateCriterion({ criterion: c, assignment: assignment })
            setEditing(false)
        } else {
            criterion ? c.setDescription(value) : setNewCriterion(c.setDescription(value))
        }
    }

    const handleBlur = () => {
        if (criterion && copy) {
            // Restore the original criterion
            c.setDescription(copy.getDescription())
        } else {
            // Reset the criterion and enable add button
            setNewCriterion(c.setDescription(""))
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
                ? <input className="form-control" type="text" onBlur={() => { handleBlur() }} autoFocus defaultValue={c?.getDescription()} name="criterion" onKeyUp={e => { handleCriteria(e) }}></input>
                : <><span>{c.getDescription()}</span><span className="badge badge-danger float-right clickable" onClick={() => actions.deleteCriterion({ criterion: criterion, assignment: assignment })}>Delete</span></>
            }
        </div>
    )
}

export default EditCriterion

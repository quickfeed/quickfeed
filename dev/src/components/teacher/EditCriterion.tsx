import React, { useState } from "react"
import { Assignment, GradingCriterion } from "../../../proto/ag/ag_pb"
import { useActions } from "../../overmind"


const EditCriterion = ({ criterion, benchmarkID, assignment }: { criterion?: GradingCriterion, benchmarkID: number, assignment: Assignment }): JSX.Element => {
    const actions = useActions()

    const [editing, setEditing] = useState<boolean>(false)
    const [add, setAdd] = useState<boolean>(criterion ? false : true)

    const c = criterion ?? new GradingCriterion().setBenchmarkid(benchmarkID)

    const handleCriteria = (event: React.KeyboardEvent<HTMLInputElement>) => {
        const { value } = event.currentTarget
        if (event.key === "Enter") {
            actions.createOrUpdateCriterion({ criterion: c, assignment: assignment })
            event.currentTarget.blur()
        } else {
            c.setDescription(value)
        }
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
                ? <input className="form-control" type="text" onBlur={() => { setEditing(false) }} autoFocus defaultValue={c?.getDescription()} name="criterion" onKeyUp={e => { handleCriteria(e) }}></input>
                : <><span>{c.getDescription()}</span><span className="float-right" onClick={() => actions.deleteCriterion({ criterion: criterion, assignment: assignment })}>Delete</span></>
            }
        </div>
    )
}

export default EditCriterion

import React, { useState } from "react";
import { Assignment, GradingCriterion } from "../../../proto/ag/ag_pb";
import { useActions, useGrpc } from "../../overmind";



const EditCriterion = ({ criterion, benchmarkID, assignment }: { criterion?: GradingCriterion, benchmarkID: number, assignment: Assignment }): JSX.Element => {
    const grpc = useGrpc().grpcMan
    const actions = useActions()

    const [editing, setEditing] = useState<boolean>(false)
    const [add, setAdd] = useState<boolean>(criterion ? false : true)

    const c = criterion ? criterion : new GradingCriterion()

    const handleCriteria = (event: React.FormEvent<HTMLInputElement>) => {
        console.log(c.getDescription())
        const { name, value } = event.currentTarget
        switch (name) {
            case "criterion":
                c.setDescription(value)
                break
            case "submit":
                c.getId() ? grpc.updateCriterion(c) : actions.createCriterion({ criterion: c, assignment: assignment })
                break
            default:
                break
        }
        //if (!c.getId() && )
    }

    if (add) {
        return (
            <li className="list-group-item">
                <button className="btn btn-primary" name="submit" onClick={() => { setAdd(false); setEditing(true) }}>Add</button>
            </li>
        )
    }

    return (
        <li className="list-group-item" onClick={() => setEditing(!editing)}>
            {editing ?
                <input className="form-control" type="text" onBlur={() => setEditing(false)} autoFocus defaultValue={c?.getDescription()} name="criterion" onKeyUp={e => { handleCriteria(e) }}></input>
                : <>
                    <span>{c.getDescription()}</span>
                    <span className="float-right">Delete</span>
                </>
            }
        </li>
    )
}

export default EditCriterion
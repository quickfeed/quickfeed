import { json } from "overmind"
import React, { useState } from "react"
import { Assignment, GradingBenchmark } from "../../../proto/ag/ag_pb"
import { useActions } from "../../overmind"


const EditBenchmark = ({ children, benchmark, assignment }: { children?: React.ReactNode, benchmark?: GradingBenchmark, assignment: Assignment }): JSX.Element => {
    const actions = useActions()

    const [editing, setEditing] = useState<boolean>(false)
    const [add, setAdd] = useState<boolean>(benchmark ? false : true)
    const [newBenchmark, setNewBenchmark] = useState<GradingBenchmark>(new GradingBenchmark().setAssignmentid(assignment.getId()))
    const bm = json(benchmark) ?? newBenchmark
    const copy = json(benchmark)?.cloneMessage()

    const handleBenchmark = (event: React.KeyboardEvent<HTMLInputElement>) => {
        const { value } = event.currentTarget
        if (event.key === "Enter") {
            actions.createOrUpdateBenchmark({ benchmark: bm, assignment: assignment })
            setEditing(false)
        } else {
            benchmark ? bm.setHeading(value) : setNewBenchmark(bm.setHeading(value))
        }
    }

    const handleBlur = () => {
        if (benchmark && copy) {
            // Restore the original criterion
            bm.setHeading(copy.getHeading())
        } else {
            // Reset the criterion and enable add button
            setNewBenchmark(bm.setHeading(""))
            setAdd(true)
        }
        setEditing(false)
    }

    if (add) {
        return (
            <div className="list-group-item list-group-item-primary">
                <button className="btn btn-primary" name="submit" onClick={() => { setAdd(false); setEditing(true) }}>Add Benchmark</button>
            </div>
        )
    }

    return (
        <>
            <div className="list-group-item list-group-item-primary">
                {editing
                    ? <input className="form-control" type="text" autoFocus defaultValue={bm?.getHeading()} onBlur={() => handleBlur()} onClick={() => setEditing(true)} onKeyUp={e => { handleBenchmark(e) }}></input>
                    : <span onClick={() => setEditing(true)}>{bm?.getHeading()}<span className="badge badge-danger float-right clickable" onClick={() => actions.deleteBenchmark({ benchmark: benchmark, assignment: assignment })}>Delete</span></span>
                }
            </div>
            {children}
        </>
    )
}

export default EditBenchmark

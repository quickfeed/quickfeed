import React, { useState } from "react"
import { Assignment, GradingBenchmark } from "../../../proto/ag/ag_pb"
import { Converter } from "../../convert"
import { useActions } from "../../overmind"


const EditBenchmark = ({ children, benchmark, assignment }: { children?: React.ReactNode, benchmark?: GradingBenchmark.AsObject, assignment: Assignment.AsObject }): JSX.Element => {
    const actions = useActions()

    const [editing, setEditing] = useState<boolean>(false)
    const [add, setAdd] = useState<boolean>(benchmark ? false : true)

    // Clone the criterion, or create a new one if none was passed in
    const bm = benchmark
        ? Converter.clone(benchmark)
        : Converter.create<GradingBenchmark.AsObject>(GradingBenchmark)

    const handleBenchmark = (event: React.KeyboardEvent<HTMLInputElement>) => {
        const { value } = event.currentTarget
        if (event.key === "Enter") {
            // Set the criterion's benchmark ID
            // This could already be set if a benchmark was passed in
            bm.assignmentid = assignment.id
            actions.createOrUpdateBenchmark({ benchmark: bm, assignment: assignment })
            setEditing(false)
        } else {
            bm.heading = value
        }
    }

    const handleBlur = () => {
        if (benchmark) {
            // Restore the original criterion
            bm.heading = benchmark.heading
        } else {
            // Reset the criterion and enable add button
            bm.heading = ""
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
                    ? <input className="form-control" type="text" autoFocus defaultValue={bm?.heading} onBlur={() => handleBlur()} onClick={() => setEditing(true)} onKeyUp={e => { handleBenchmark(e) }}></input>
                    : <span onClick={() => setEditing(true)}>{bm?.heading}<span className="badge badge-danger float-right clickable" onClick={() => actions.deleteBenchmark({ benchmark: benchmark, assignment: assignment })}>Delete</span></span>
                }
            </div>
            {children}
        </>
    )
}

export default EditBenchmark

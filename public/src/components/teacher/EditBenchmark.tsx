import React, { useState } from "react"
import { Assignment, GradingBenchmark, GradingBenchmarkSchema } from "../../../proto/qf/types_pb"
import { useActions } from "../../overmind"
import { clone, create } from "@bufbuild/protobuf"


const EditBenchmark = ({ children, benchmark, assignment }: { children?: React.ReactNode, benchmark?: GradingBenchmark, assignment: Assignment }) => {
    const actions = useActions().global

    const [editing, setEditing] = useState<boolean>(false)
    const [add, setAdd] = useState<boolean>(benchmark ? false : true)

    // Clone the benchmark, or create a new one if none was passed in
    const bm = benchmark
        ? clone(GradingBenchmarkSchema, benchmark)
        : create(GradingBenchmarkSchema)

    const resetBenchmark = () => {
        // Reset the benchmark and enable add button
        bm.heading = ""
        setAdd(true)
    }

    const handleBenchmark = async (event: React.KeyboardEvent<HTMLInputElement>) => {
        const { value } = event.currentTarget
        if (event.key === "Enter") {
            // Set the benchmark's assignment ID
            // This could already be set if a benchmark was passed in
            bm.AssignmentID = assignment.ID
            const success = await actions.createOrUpdateBenchmark({ benchmark: bm, assignment })
            if (!success) {
                resetBenchmark()
            }
            setEditing(false)
        } else {
            bm.heading = value
        }
    }

    const handleBlur = () => {
        if (benchmark) {
            // Restore the original benchmark
            bm.heading = benchmark.heading
        } else {
            resetBenchmark()
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
    const input = <input className="form-control" type="text" autoFocus defaultValue={bm?.heading} onBlur={handleBlur} onKeyUp={e => handleBenchmark(e)} /> // skipcq: JS-0757
    const textAndButton = (
        <span onClick={() => setEditing(!editing)} role="button" aria-hidden="true">
            {bm?.heading}
            <button className="p-2 badge badge-danger float-right clickable" onClick={() => actions.deleteBenchmark({ benchmark, assignment })}>
                Delete Benchmark
            </button>
        </span>
    )
    return (
        <>
            <div className="list-group-item list-group-item-primary">
                {editing ? input : textAndButton}
            </div>
            {children}
        </>
    )
}

export default EditBenchmark

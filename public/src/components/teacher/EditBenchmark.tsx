import React, { useState, memo } from "react"
import { GradingBenchmark } from "../../../proto/qf/types_pb"

interface EditBenchmarkProps {
    children?: React.ReactNode
    benchmark?: GradingBenchmark
    updateBenchmark: (event: React.KeyboardEvent<HTMLInputElement>, bm: GradingBenchmark) => void
    deleteBenchmark?: () => void
}

const EditBenchmark = memo(({ children, benchmark, updateBenchmark, deleteBenchmark }: EditBenchmarkProps) => {
    const [editing, setEditing] = useState<boolean>(false)
    const [add, setAdd] = useState<boolean>(benchmark ? false : true)

    // Clone the criterion, or create a new one if none was passed in
    const bm = benchmark
        ? benchmark.clone()
        : new GradingBenchmark()

    const handleBenchmark = (event: React.KeyboardEvent<HTMLInputElement>) => {
        updateBenchmark(event, bm)
        if (event.key === "Enter") {
            setEditing(false)
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
    const handleSetEditing = (editing: boolean) => () => setEditing(editing)

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
                    ? <input className="form-control" type="text" autoFocus defaultValue={bm?.heading} onBlur={handleBlur} onClick={handleSetEditing(!editing)} onKeyUp={handleBenchmark} />
                    : <span onClick={handleSetEditing(true)}>{bm?.heading}<span className="badge badge-danger float-right clickable" onClick={deleteBenchmark}>Delete</span></span>}
            </div>
            {children}
        </>
    )
})

EditBenchmark.displayName = "EditBenchmark"

export default EditBenchmark

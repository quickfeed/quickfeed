import React, { useState } from "react"
import { GradingBenchmark } from "../../../proto/qf/types_pb"
import GradeComment from "./GradeComment"


const Benchmark = ({ children, bm }: { children: React.ReactNode, bm: GradingBenchmark }) => {
    const [editing, setEditing] = useState<boolean>(false)
    return (
        <>
            <tr className="table-info">
                <th colSpan={2}>{bm.heading}</th>
                <th onClick={() => setEditing(true)}>{bm.comment}</th>
            </tr>
            <GradeComment grade={bm} editing={editing} setEditing={setEditing} />
            {children}
        </>
    )
}

export default Benchmark

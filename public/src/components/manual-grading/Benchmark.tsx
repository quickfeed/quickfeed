import React, { useState } from "react"
import { GradingBenchmark } from "../../../proto/qf/types_pb"
import GradeComment from "./GradeComment"


const Benchmark = ({ children, bm }: { children: React.ReactNode, bm: GradingBenchmark }) => {
    const [editing, setEditing] = useState<boolean>(false)
    return (
        <>
            <tr className="bg-base-200 border-b-2 border-base-300 border-t border-base-300">
                <th colSpan={2} className="text-base font-bold py-3 px-4">{bm.heading}</th>
                <th onClick={() => setEditing(true)} className="py-3 px-4">{bm.comment}</th>
            </tr>
            <GradeComment grade={bm} editing={editing} setEditing={setEditing} />
            {children}
        </>
    )
}

export default Benchmark

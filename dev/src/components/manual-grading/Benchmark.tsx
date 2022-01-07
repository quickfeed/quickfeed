import React, { useState } from "react"
import { GradingBenchmark } from "../../../proto/ag/ag_pb"
import GradeComment from "./GradeComment"


const Benchmark = ({ children, bm }: { children: React.ReactNode, bm: GradingBenchmark }): JSX.Element => {
    const [editing, setEditing] = useState<boolean>(false)
    return (
        <>
            <tr className="table-info">
                <th colSpan={2}>{bm.getHeading()}</th>
                <th onClick={() => setEditing(true)}>{bm.getComment()}</th>
            </tr>
            <GradeComment grade={bm} editing={editing} setEditing={setEditing} />
            {children}
        </>
    )
}

export default Benchmark
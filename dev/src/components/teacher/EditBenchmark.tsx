import { json } from "overmind"
import React, { useState } from "react"
import { Assignment, GradingBenchmark } from "../../../proto/ag/ag_pb"
import { useActions, useGrpc } from "../../overmind"



const EditBenchmark = ({ children, benchmark, assignment }: { children?: React.ReactNode, benchmark?: GradingBenchmark, assignment: Assignment }): JSX.Element => {
    const grpc = useGrpc().grpcMan
    const actions = useActions()
    const bm = json(benchmark) ?? new GradingBenchmark()

    const [editing, setEditing] = useState<boolean>(false)

    const handleBenchmark = (event: React.KeyboardEvent<HTMLInputElement>) => {
        const { value } = event.currentTarget
        if (event.key === "Enter") {
            if (bm.getId() > 0) {
                grpc.updateBenchmark(bm)
            } else {
                actions.createBenchmark({ benchmark: bm, assignment: assignment })
            }
        } else {
            bm.setHeading(value)
        }
    }

    return (
        <>
            <div className="list-group-item list-group-item-primary">
                {editing
                    ? <input className="form-control" type="text" defaultValue={bm?.getHeading()} onClick={() => setEditing(true)} onKeyUp={e => { handleBenchmark(e) }}></input>
                    : <span onClick={() => setEditing(true)}>{bm?.getHeading()}</span>
                }
            </div>
            {children}
        </>
    )
}

export default EditBenchmark

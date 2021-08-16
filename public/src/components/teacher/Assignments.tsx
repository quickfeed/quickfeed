import { json } from "overmind"
import React, { useState } from "react"
import { Assignment } from "../../../proto/ag/ag_pb"
import { getCourseID } from "../../Helpers"
import { useAppState, useGrpc } from "../../overmind"
import CriterionForm from "../forms/CriterionForm"


// TODO: Implement benchmark adding

const Assignments = () => {
    const courseID = getCourseID()
    const assignments = useAppState().assignments[courseID]

    const grpc = useGrpc().grpcMan

    const [editing, setEditing] = useState<Assignment>()
    const [editBenchmarkID, setBenchmarkID] = useState<number>(0)
    const [addBenchmark, setAddBenchmark] = useState<boolean>(false)
    const [editCriteriaID, setCriteriaID] = useState<number>()

    const assignmentslist = assignments.map(assignment => {
        return (
            <tr onClick={() => {setEditing(json(assignment))}}>
                <th colSpan={2}>{assignment.getName()}</th>
                <td>{assignment.getIsgrouplab() ? "V" : "X"}</td>
                <td>{assignment.getReviewers()}</td>
            </tr>
            )
    })

    const EditAssignment = () => {
        return (
        <>
        <table className="table table-curved table-striped">
            <thead className={"thead-dark"}>
                <th colSpan={2}>{editing?.getName()}</th>
                <th>Points</th>
            </thead>
            <tbody>
            {editing?.getGradingbenchmarksList().map(bm => {
                return (
                    <>
                    <tr className="table-info"> {bm}
                        <th colSpan={3}>{bm.getHeading()}</th>
                    </tr>
                    {bm.getCriteriaList().map(c => {
                        if (editCriteriaID == c.getId()) {
                            return (
                            <CriterionForm criterion={c} setEditing={setCriteriaID} />
                            )
                        }
                        return (
                        <tr onClick={() => setCriteriaID(c.getId())}>
                            <th colSpan={2}>{c.getDescription()}</th>
                            <td>{c.getPoints()}</td>
                        </tr>
                        )
                    })}
                    {editBenchmarkID !== bm.getId() ? 
                        <button onClick={() => setBenchmarkID(bm.getId())}>Add Criterion</button> 
                        :
                        <CriterionForm benchmarkID={bm.getId()} assignment={editing} setEditing={setBenchmarkID} />
                    }
                    </>
                )
            })}
            {!addBenchmark ?
            <button onClick={() => setAddBenchmark(true)}>Add Benchmark</button> 
            : <input type="text"></input>
            }
            </tbody>
        </table>
        </>)
    }

    return (
        <div className="box row">
            <div className="col">
                <table className="table table-curved table-striped">
                    <thead className={"thead-dark"}>
                        <th colSpan={2}>Assignment</th>
                        <th>Group</th>
                        <th>Reviewers</th>
                    </thead>
                    <tbody>
                    {assignmentslist.length > 0 ? assignmentslist : "This course has no assignments."}
                    </tbody>

                </table>
            </div>
            <div className="col">
                {editing && editing.getReviewers() === 0 ? `This assignment (${editing.getName()}) is not for manual grading.` : null}
                {editing && editing.getReviewers() > 0 ? EditAssignment() : null}
            </div>
        </div>

    )

}

export default Assignments
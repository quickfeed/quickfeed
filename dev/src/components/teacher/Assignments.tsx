import { json } from "overmind"
import React, { useState } from "react"
import { Assignment } from "../../../proto/ag/ag_pb"
import { getCourseID, isManuallyGraded } from "../../Helpers"
import { useAppState } from "../../overmind"
import DynamicTable, { CellElement } from "../DynamicTable"
import CriterionForm from "../forms/CriterionForm"


// TODO: Implement benchmark adding

const Assignments = (): JSX.Element => {
    const courseID = getCourseID()
    const assignments = useAppState().assignments[courseID]

    const [editing, setEditing] = useState<Assignment>()
    const [editBenchmarkID, setBenchmarkID] = useState<number>(0)
    const [addBenchmark, setAddBenchmark] = useState<boolean>(false)
    const [editCriteriaID, setCriteriaID] = useState<number>(-1)

    const assignmentsData = assignments.map(assignment => {
        const data: (string | JSX.Element | CellElement)[] = []
        data.push({className: "clickable", value: assignment.getName(), onClick: () => setEditing(assignment)})
        data.push(assignment.getIsgrouplab() ? "V" : "X")
        data.push(assignment.getReviewers().toString())
        return data
    })

    const EditAssignment = () => {
        if (editing) { 
            return editing.getGradingbenchmarksList().map(bm => {  
                const data = bm.getCriteriaList().map(c => {
                    const data: (string | JSX.Element | CellElement)[] = []
                    if (editCriteriaID == c.getId()) {
                        data.push(<CriterionForm criterion={c} setEditing={setCriteriaID} />)
                    }
                    else {
                        data.push({value: c.getDescription(), onClick: () => setCriteriaID(c.getId())})
                        data.push(c.getPoints().toString())
                    }
                    return data
                })
                if (editBenchmarkID !== bm.getId()) {
                        data.push([<div key={8}><button onClick={() => setBenchmarkID(bm.getId())}>Add Criterion</button> </div>])
                } 
                 else {
                        data.push([<CriterionForm key={89} benchmarkID={bm.getId()} assignment={editing} setEditing={setBenchmarkID} />])
                }
                return <DynamicTable key={bm.getId()} data={data} header={[bm.getHeading(), "Points"]} />
            })
            
        }
        return []
    }

    return (
        <div className="box row">
            <div className="col">
                {assignmentsData.length > 0 ? 
                    <DynamicTable header={["Assignment", "Group", "Reviewers"]} data={assignmentsData} /> 
                    : 
                    "This course has no assignments."
                }

            </div>
            <div className="col mb-5">
                {editing && !isManuallyGraded(editing) ? `This assignment (${editing.getName()}) is not for manual grading.` : null}
                {editing && isManuallyGraded(editing) ? EditAssignment() : null}
                {(editing && isManuallyGraded(editing) && addBenchmark) ? null : <button onClick={() => setAddBenchmark(true)}>Add Benchmark</button>}
            </div>
        </div>

    )

}

export default Assignments
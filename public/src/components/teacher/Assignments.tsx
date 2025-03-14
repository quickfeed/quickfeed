import React, { useState } from "react"
import { Assignment } from "../../../proto/qf/types_pb"
import { getCourseID, isManuallyGraded, Color, hasBenchmarks, hasCriteria } from "../../Helpers"
import { useActions, useAppState } from "../../overmind"
import Button, { ButtonType } from "../admin/Button"
import EditBenchmark from "./EditBenchmark"
import EditCriterion from "./EditCriterion"


/** This component displays all assignments for the active course and:
 *  for assignments that are not manually graded, allows teachers to rebuild all submissions.
 *  for manually graded assignments, allows teachers to add or remove criteria and benchmarks for the assignment */
const Assignments = () => {
    const courseID = getCourseID()
    const actions = useActions()
    const state = useAppState()

    const assignmentElement = (assignment: Assignment) => {
        const [hidden, setHidden] = useState<boolean>(false)
        const [buttonText, setButtonText] = useState<string>("Rebuild all tests")

        /* rebuild all tests for this assignment */
        const rebuild = async () => {
            if (confirm(`Warning! This will rebuild all submissions for ${assignment.name}. This may take several minutes. Are you sure you want to continue?`)) {
                setButtonText("Rebuilding...")
                const success = await actions.rebuildAllSubmissions({ assignmentID: assignment.ID, courseID: courseID })
                if (success) {
                    setButtonText("Finished rebuilding")
                } else {
                    setButtonText("Failed to rebuild")
                }
            }
        }

        const assignmentForm = hasBenchmarks(assignment) ? assignment.gradingBenchmarks.map((bm) => (
            <EditBenchmark key={bm.ID.toString()}
                benchmark={bm}
                assignment={assignment}
            >
                {/* Show all criteria for this benchmark */}
                {hasCriteria(bm) && bm.criteria?.map((crit) => (
                    <EditCriterion key={crit.ID.toString()}
                        originalCriterion={crit}
                        assignment={assignment}
                        benchmarkID={bm.ID}
                    />
                ))}
                {/* Always show one criterion form in case of benchmarks without any */}
                <EditCriterion key={bm.criteria.length}
                    assignment={assignment}
                    benchmarkID={bm.ID}
                />
            </EditBenchmark>
        )) : null

        return (
            <ul key={assignment.ID.toString()} className="list-group">
                <div onClick={() => setHidden(!hidden)} role="button" aria-hidden="true">
                    <li key="assignment" className="list-group-item">
                        {assignment.name}
                    </li>
                </div>
                {hidden && (
                    <li key="form" className="list-group-item">
                        {/* Only show the rebuild button if the assignment is not manually graded */}
                        {isManuallyGraded(assignment)
                            ? <> {assignmentForm} <EditBenchmark key={assignment.gradingBenchmarks.length} assignment={assignment} /></>
                            : <Button text={buttonText} color={Color.BLUE} type={ButtonType.BUTTON} onClick={rebuild} />
                        }
                    </li>
                )}
            </ul>
        )
    }

    const list = state.assignments[courseID.toString()]?.map(assignment => assignmentElement(assignment))
    return (
        <div className="column">
            {list}
        </div>
    )
}

export default Assignments

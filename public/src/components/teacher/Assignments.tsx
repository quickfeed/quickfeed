import React, { useState, useCallback } from "react"
import { Assignment, GradingBenchmark, GradingCriterion } from "../../../proto/qf/types_pb"
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

    /* rebuild all tests for this assignment */
    const rebuild = useCallback((name: string, id: bigint, setButtonText: React.Dispatch<React.SetStateAction<string>>) => async () => {
        if (confirm(`Warning! This will rebuild all submissions for ${name}. This may take several minutes. Are you sure you want to continue?`)) {
            setButtonText("Rebuilding...")
            const success = await actions.rebuildAllSubmissions({ assignmentID: id, courseID: courseID })
            if (success) {
                setButtonText("Finished rebuilding")
            } else {
                setButtonText("Failed to rebuild")
            }
        }
    }, [actions, courseID])

    const updateBenchmark = useCallback((assignment: Assignment) => (event: React.KeyboardEvent<HTMLInputElement>, bm: GradingBenchmark) => {
        const { value } = event.currentTarget
        if (event.key === "Enter") {
            // Set the criterion's benchmark ID
            // This could already be set if a benchmark was passed in
            bm.AssignmentID = assignment.ID
            actions.createOrUpdateBenchmark({ benchmark: bm, assignment: assignment })
        } else {
            bm.heading = value
        }
    }, [actions])

    const deleteBenchmark = useCallback((benchmark: GradingBenchmark, assignment: Assignment) => () => actions.deleteBenchmark({ benchmark: benchmark, assignment: assignment }), [actions])

    const updateCriterion = useCallback((benchmarkID: bigint, assignment: Assignment) => (event: React.KeyboardEvent<HTMLInputElement>, criterion: GradingCriterion) => {
        const { value } = event.currentTarget
        if (event.key === "Enter") {
            // Set the criterion's benchmark ID
            // This could already be set if a criterion was passed in
            criterion.BenchmarkID = benchmarkID
            actions.createOrUpdateCriterion({ criterion: criterion, assignment: assignment })
        } else {
            criterion.description = value
        }
    }, [actions])

    const deleteCriterion = useCallback((criterion: GradingCriterion, assignment: Assignment) => () => actions.deleteCriterion({ criterion: criterion, assignment: assignment }), [actions])

    const AssignmentElement = (assignment: Assignment) => {
        const [hidden, setHidden] = useState<boolean>(false)
        const [buttonText, setButtonText] = useState<string>("Rebuild all tests")

        const assignmentForm = hasBenchmarks(assignment) ? assignment.gradingBenchmarks.map((bm) => (
            <EditBenchmark key={bm.ID.toString()}
                benchmark={bm}
                updateBenchmark={updateBenchmark(assignment)}
                deleteBenchmark={deleteBenchmark(bm, assignment)}
            >
                {/* Show all criteria for this benchmark */}
                {hasCriteria(bm) && bm.criteria?.map((crit) => (
                    <EditCriterion key={crit.ID.toString()}
                        originalCriterion={crit}
                        updateCriterion={updateCriterion(bm.ID, assignment)}
                        deleteCriterion={deleteCriterion(crit, assignment)}
                    />
                ))}
                {/* Always show one criterion form in case of benchmarks without any */}
                <EditCriterion key={bm.criteria.length} updateCriterion={updateCriterion(bm.ID, assignment)} />
            </EditBenchmark>
        )) : null

        const editOrRebuild = isManuallyGraded(assignment)
            ? <> {assignmentForm} <EditBenchmark key={assignment.gradingBenchmarks.length} updateBenchmark={updateBenchmark(assignment)} /></>
            : <Button text={buttonText} color={Color.BLUE} type={ButtonType.BUTTON} onClick={rebuild(assignment.name, assignment.ID, setButtonText)} />

        return (
            <ul key={assignment.ID.toString()} className="list-group">
                <div onClick={() => setHidden(!hidden)} role="button" aria-hidden="true"> {/* skipcq: JS-0417 */}
                    <li key="assignment" className="list-group-item">
                        {assignment.name}
                    </li>
                </div>
                {hidden && (
                    <li key="form" className="list-group-item">
                        {/* Only show the rebuild button if the assignment is not manually graded */}
                        {editOrRebuild}
                    </li>
                )}
            </ul>
        )
    }

    const list = state.assignments[courseID.toString()]?.map(assignment => AssignmentElement(assignment))
    return (
        <div className="column">
            {list}
        </div>
    )
}

export default Assignments

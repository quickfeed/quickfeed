import React, { useState } from "react"
import { Assignment } from "../../../proto/ag/ag_pb"
import { getCourseID, isManuallyGraded, Color, hasBenchmarks, hasCriteria } from "../../Helpers"
import { useActions, useAppState } from "../../overmind"
import Button, { ButtonType } from "../admin/Button"
import EditBenchmark from "./EditBenchmark"
import EditCriterion from "./EditCriterion"


// TODO: Needs some love. Currently a mess.

const Assignments = (): JSX.Element => {
    const courseID = getCourseID()
    const actions = useActions()
    const state = useAppState()

    const assignmentElement = (assignment: Assignment): JSX.Element => {
        const [hidden, setHidden] = useState<boolean>(false)
        const [buttonText, setButtonText] = useState<string>("Rebuild all tests")

        /* rebuild all tests for this assignment */
        const rebuild = async () => {
            if (confirm(`Warning! This will rebuild all submissions for ${assignment.getName()}. This may take several minutes. Are you sure you want to continue?`)) {
                setButtonText("Rebuilding...")
                const success = await actions.rebuildAllSubmissions({ assignmentID: assignment.getId(), courseID: courseID })
                if (success) {
                    setButtonText("Finished rebuilding")
                } else {
                    setButtonText("Failed to rebuild")
                }
            }
        }

        const assignmentForm = hasBenchmarks(assignment) ? assignment.getGradingbenchmarksList().map((bm) => (
            <EditBenchmark key={bm.getId()}
                benchmark={bm}
                assignment={assignment}
            >
                {/* Show all criteria for this benchmark */}
                {hasCriteria(bm) && bm.getCriteriaList()?.map((crit) => (
                    <EditCriterion key={crit.getId()}
                        criterion={crit}
                        assignment={assignment}
                        benchmarkID={bm.getId()}
                    />
                ))}
                {/* Always show one criterion form in case of benchmarks without any */}
                <EditCriterion key={"edit"}
                    assignment={assignment}
                    benchmarkID={bm.getId()}
                />
            </EditBenchmark>
        )) : null

        return (
            <ul key={assignment.getId()} className="list-group">
                <li key={"assignment"} className="list-group-item" onClick={() => setHidden(!hidden)}>
                    {assignment.getName()}
                </li>
                {hidden && (
                    <li key={"form"} className="list-group-item">
                        {/* Only show the rebuild button if the assignment is not manually graded */}
                        {isManuallyGraded(assignment)
                            ? <> {assignmentForm} <EditBenchmark key={"eb"} assignment={assignment} /></>
                            : <Button text={buttonText} type={ButtonType.BUTTON} color={Color.BLUE} onclick={rebuild} />
                        }
                    </li>
                )}
            </ul >
        )
    }

    const list = state.assignments[courseID]?.map(assignment => assignmentElement(assignment))
    return (
        <div className="column">
            {list}
        </div>
    )
}

export default Assignments

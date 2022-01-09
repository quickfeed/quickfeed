import { json } from "overmind"
import React, { useState } from "react"
import { Assignment } from "../../../proto/ag/ag_pb"
import { getCourseID, isManuallyGraded, Color } from "../../Helpers"
import { useActions, useAppState } from "../../overmind"
import Button, { ButtonType } from "../admin/Button"
import EditBenchmark from "./EditBenchmark"
import EditCriterion from "./EditCriterion"


// TODO: Needs some love. Currently a mess.

const Assignments = (): JSX.Element => {
    const courseID = getCourseID()
    const actions = useActions()
    const assignments = useAppState().assignments[courseID]

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

        const generateForm = json(assignment).getGradingbenchmarksList()?.map((bm, index) => (
            <EditBenchmark key={bm.getId()}
                benchmark={bm}
                assignment={assignment}
            >
                {/* Show all criteria for this benchmark */}
                {bm.getCriteriaList().map((crit, index) => (
                    <EditCriterion key={index}
                        criterion={crit}
                        assignment={assignment}
                        benchmarkID={bm.getId()} />
                ))}
                {/* Always show one criterion form in case of benchmarks without any */}
                <EditCriterion key={index}
                    assignment={assignment}
                    benchmarkID={bm.getId()}
                />
            </EditBenchmark>
        ))


        /* Only show the rebuild button if the assignment is not manually graded */
        const rebuildButton = !isManuallyGraded(assignment) ? <Button text={buttonText} type={ButtonType.BUTTON} color={Color.BLUE} onclick={rebuild} /> : null
        return (
            <ul className="list-group">
                <li className="list-group-item" onClick={() => setHidden(!hidden)}>
                    {assignment.getName()}
                </li>
                {hidden && (
                    <li className="list-group-item">
                        {rebuildButton}
                        {isManuallyGraded(assignment) && (
                            <>
                                {generateForm}
                                <EditBenchmark assignment={assignment} />
                            </>
                        )}

                    </li>
                )}
            </ul >
        )
    }

    const list = assignments.map(assignment => assignmentElement(assignment))
    return (
        <div className="column">
            {list}
        </div>

    )

}

export default Assignments
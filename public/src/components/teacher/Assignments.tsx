// ...existing code...
import React, { useState } from "react"
import { Assignment } from "../../../proto/qf/types_pb"
import { isManuallyGraded, Color, getFormattedTime, hasBenchmarks } from "../../Helpers"
import { useActions, useAppState } from "../../overmind"
import Button, { ButtonType } from "../admin/Button"
import { useCourseID } from "../../hooks/useCourseID"
import RubricDisplay from "./RubricDisplay"

const Assignments = () => {
    const courseID = useCourseID()
    const actions = useActions().global
    const state = useAppState()

    const AssignmentElement = ({ assignment }: { assignment: Assignment }) => {
        const [open, setOpen] = useState<boolean>(false)
        const [buttonText, setButtonText] = useState<string>("Rebuild all tests")
        const [isRebuilding, setIsRebuilding] = useState<boolean>(false)

        const manually = isManuallyGraded(assignment.reviewers)

        const rebuild = async () => {
            if (
                confirm(
                    `Warning! This will rebuild all submissions for ${assignment.name}. This may take several minutes. Are you sure you want to continue?`,
                )
            ) {
                setButtonText("Rebuilding...")
                setIsRebuilding(true)
                const success = await actions.rebuildAllSubmissions({
                    assignmentID: assignment.ID,
                    courseID,
                })
                setIsRebuilding(false)
                if (success) {
                    setButtonText("Rebuild Successful ✓")
                    setTimeout(() => setButtonText("Rebuild All Tests"), 3000)
                } else {
                    setButtonText("Rebuild Failed ✗")
                    setTimeout(() => setButtonText("Rebuild All Tests"), 3000)
                }
            }
            setButtonText("Rebuilding...")
            const success = await actions.rebuildAllSubmissions({ assignmentID: assignment.ID, courseID })
            setButtonText(success ? "Finished rebuilding" : "Failed to rebuild")
        }

        return (
            <div key={assignment.ID.toString()} className="card bg-base-200 shadow-md rounded-lg overflow-hidden">
                <div
                    className="flex items-center justify-between px-4 py-3 bg-base-300 hover:bg-base-200 transition-colors cursor-pointer"
                    onClick={() => setOpen(!open)}
                    role="button"
                    aria-expanded={open}
                >
                    <div className="flex items-start gap-3">
                        <div className="flex flex-col">
                            <div className="text-lg font-semibold text-base-content">{assignment.name}</div>
                            <div className="flex items-center gap-2 mt-1">
                                <span className="text-sm text-base-content/70 flex items-center gap-1">
                                    <i className="fa fa-calendar" />
                                    {getFormattedTime(assignment.deadline, true)}
                                </span>
                                {manually && (
                                    <span className="badge badge-warning badge-sm">Manual</span>
                                )}
                                {/* show score limit if available */}
                                <span className="badge badge-outline text-sm">{`Pass: ${assignment.scoreLimit}%`}</span>
                            </div>
                        </div>
                    </div>

                    <div className="flex items-center gap-3">
                        <div className="text-sm text-base-content/70 hidden md:block">
                            {/* short description or hint */}
                            Rebuild tests or manage grading
                        </div>
                        <i className={`fa fa-chevron-down transition-transform ${open ? "rotate-180" : ""}`} />
                    </div>
                </div>

                {open && (
                    <div className="p-4 border-t border-base-content/10">
                        {!manually ? (
                            <div className="flex flex-col md:flex-row md:items-center md:justify-between gap-3">
                                <div className="flex items-center gap-3">
                                    <Button
                                        text={buttonText}
                                        color={Color.BLUE}
                                        type={ButtonType.SOLID}
                                        onClick={rebuild}
                                        disabled={isRebuilding}
                                    />
                                    <div className="text-sm text-base-content/70">
                                        Rebuilds all submissions for this assignment.
                                    </div>
                                </div>

                                <div className="flex items-center gap-2">
                                    <Button
                                        text="View submissions"
                                        color={Color.GREEN}
                                        type={ButtonType.OUTLINE}
                                        onClick={() => {/* navigate to submissions list */ }}
                                    />
                                </div>
                            </div>
                        ) : (
                            <div className="text-sm text-base-content/70">
                                This assignment is manually graded. Manage criteria and benchmarks in the assignment <code className="px-1 rounded bg-base-100 text-error">criteria.json</code> file.
                                {hasBenchmarks(assignment) &&
                                    <div className="mt-3">
                                        {assignment.gradingBenchmarks.map((bm) => (
                                            <RubricDisplay key={bm.ID.toString()} benchmark={bm} />
                                        ))}
                                    </div>
                                }
                            </div>
                        )}
                    </div>
                )}
            </div>
        )
    }

    return (
        <div className="space-y-4">
            {state.assignments[courseID.toString()]?.map(assignment =>
                <AssignmentElement key={assignment.ID} assignment={assignment} />
            )}
        </div>
    )
}

export default Assignments
